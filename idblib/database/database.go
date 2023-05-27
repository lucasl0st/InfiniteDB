/*
 * Copyright (c) 2023 Lucas Pape
 */

package database

import (
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/lucasl0st/InfiniteDB/idblib/field"
	"github.com/lucasl0st/InfiniteDB/idblib/metrics"
	"github.com/lucasl0st/InfiniteDB/idblib/object"
	"github.com/lucasl0st/InfiniteDB/idblib/table"
	idbutil "github.com/lucasl0st/InfiniteDB/idblib/util"
	e "github.com/lucasl0st/InfiniteDB/models/errors"
	"github.com/lucasl0st/InfiniteDB/models/request"
	"github.com/lucasl0st/InfiniteDB/util"
	"os"
	"strings"
	"sync"
	"time"
)

type Database struct {
	Name              string
	path              string
	tablesPath        string
	tables            map[string]*table.Table
	l                 idbutil.Logger
	m                 *metrics.Metrics
	cacheSize         uint
	watchForNewTables bool
	watcher           *fsnotify.Watcher
}

func NewDatabase(name string, path string, logger idbutil.Logger, metrics *metrics.Metrics, cacheSize uint) (*Database, int, error) {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		return nil, 0, err
	}

	tablesPath := path + name + "/tables/"

	err = watcher.Add(tablesPath)

	if err != nil {
		return nil, 0, err
	}

	database := &Database{
		Name:              name,
		path:              path,
		tablesPath:        tablesPath,
		tables:            make(map[string]*table.Table),
		l:                 logger,
		m:                 metrics,
		cacheSize:         cacheSize,
		watchForNewTables: true,
		watcher:           watcher,
	}

	err = database.loadTables()

	go func() {
		for {
			if !database.watchForNewTables {
				return
			}

			database.tablesWatcher()
		}
	}()

	return database, len(database.tables), err
}

func CreateDatabase(path string, name string) error {
	runMiddleware, createDatabase := table.CreateDatabaseMiddleware(name)

	if runMiddleware {
		return createDatabase()
	}

	err := os.MkdirAll(path+"/"+name+"/tables", os.ModePerm)

	if err != nil {
		return err
	}

	return nil
}

func (d *Database) tablesWatcher() {
	select {
	case event, ok := <-d.watcher.Events:
		if !ok {
			return
		}

		tableName := strings.ReplaceAll(event.Name, d.tablesPath, "")
		isTable := !strings.Contains(tableName, "/")

		if !isTable {
			return
		}

		if event.Has(fsnotify.Create) {
			time.Sleep(time.Millisecond * 100)

			err := d.loadTables()

			if err != nil {
				d.l.Fatal(err.Error())
			}
		} else if event.Has(fsnotify.Remove) {
			t := d.tables[tableName]

			if t == nil {
				return
			}

			t.Storage.Kill()
			delete(d.tables, tableName)
		}
	}
}

func (d *Database) loadTables() error {
	files, err := os.ReadDir(d.tablesPath)

	if err != nil {
		return err
	}

	errs := make(chan error, len(files))
	var wg sync.WaitGroup

	for _, tableFolder := range files {
		if !tableFolder.IsDir() {
			continue
		}

		wg.Add(1)

		go func(tableName string) {
			defer wg.Done()

			errs <- d.loadTable(tableName)
		}(tableFolder.Name())
	}

	wg.Wait()
	close(errs)

	for err = range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Database) loadTable(name string) error {
	if d.tables[name] != nil {
		return nil
	}

	start := time.Now()

	bytes, err := os.ReadFile(d.tablesPath + name + "/table.json")

	if err != nil {
		return err
	}

	var fields field.TableConfig
	err = json.Unmarshal(bytes, &fields)

	if err != nil {
		return err
	}

	t, err := table.NewTable(d.Name, name, d.tablesPath, fields, d.l, d.m, d.cacheSize)

	if err != nil {
		return err
	}

	d.tables[name] = t

	elapsed := time.Since(start)

	d.l.Println("loaded table "+name+" with "+fmt.Sprint(t.Storage.NumberOfObjects())+" objects, took ", elapsed)

	return nil
}

func (d *Database) CreateTable(name string, fields map[string]field.Field, options request.TableOptions) error {
	if d.tables[name] != nil {
		return e.TableAlreadyExists()
	}

	err := os.MkdirAll(d.tablesPath+name, os.ModePerm)

	if err != nil {
		return err
	}

	config := field.TableConfig{
		Fields:  fields,
		Options: options,
	}

	bytes, err := json.Marshal(config)

	if err != nil {
		return err
	}

	err = os.WriteFile(d.tablesPath+name+"/table.json", bytes, 0644)

	if err != nil {
		return err
	}

	return d.loadTable(name)
}

func (d *Database) GetTableNames() []string {
	var tableNames []string

	for _, t := range d.tables {
		tableNames = append(tableNames, t.Name)
	}

	return tableNames
}

func (d *Database) GetTable(tableName string) (map[string]request.Field, *request.TableOptions, error) {
	t := d.tables[tableName]

	if t == nil {
		return nil, nil, e.TableDoesNotExist()
	}

	fields := map[string]request.Field{}

	for name, f := range t.Config.Fields {
		fields[name] = request.Field{
			Type:    fmt.Sprint(f.Type),
			Indexed: util.Ptr(f.Indexed),
			Unique:  util.Ptr(f.Unique),
			Null:    util.Ptr(f.Null),
		}
	}

	return fields, &t.Config.Options, nil
}

func (d *Database) Get(tableName string, request table.Request) ([]map[string]json.RawMessage, error) {
	if request.Query != nil {
		t := d.tables[tableName]

		if t == nil {
			return nil, e.TableDoesNotExist()
		}

		objects, additionalFields, err := t.Query(*request.Query, nil, make(table.AdditionalFields))

		if err != nil {
			return nil, err
		}

		if request.Sort != nil {
			objects, err = t.Sort(objects, request.Sort.Field, additionalFields, request.Sort.Direction)
		}

		if err != nil {
			return nil, err
		}

		objects = t.SkipAndLimit(objects, request.Skip, request.Limit)

		results := t.Storage.GetObjects(objects)

		implementObjectsMap := map[int64]map[string]json.RawMessage{}

		if request.Implement != nil {
			for _, id := range objects {
				implementObjectsMap[id] = map[string]json.RawMessage{}
			}

			for _, implement := range request.Implement {
				i, as, err := d.implement(implement, results)

				if err != nil {
					return nil, err
				}

				for id, implementedObject := range i {
					implementObjectsMap[id][*as] = implementedObject
				}
			}
		}

		interfaceObjects, err := d.objectsToMapStringJsonRawArray(results, t, implementObjectsMap, additionalFields)

		if err != nil {
			return nil, err
		}

		return interfaceObjects, err
	} else {
		return nil, nil
	}
}

func (d *Database) Remove(tableName string, request table.Request) (int64, error) {
	objects, err := d.Get(tableName, request)

	if err != nil {
		return 0, err
	}

	t := d.tables[tableName]

	var count int64 = 0

	for _, o := range objects {
		om, err := t.JsonRawMapToObject(o)

		if err != nil {
			return 0, err
		}

		err = t.Remove(om)

		if err != nil {
			return count, err
		}

		count++
	}

	return count, nil
}

func (d *Database) Insert(tableName string, o map[string]json.RawMessage) error {
	t := d.tables[tableName]

	if t == nil {
		return e.TableDoesNotExist()
	}

	return t.Insert(o)
}

func (d *Database) Update(tableName string, o map[string]json.RawMessage) error {
	t := d.tables[tableName]

	if t == nil {
		return e.TableDoesNotExist()
	}

	return t.Update(o)
}

func (d *Database) objectsToMapStringJsonRawArray(
	objects []object.Object,
	t *table.Table,
	implementObjectsMap map[int64]map[string]json.RawMessage,
	additionalFields table.AdditionalFields,
) ([]map[string]json.RawMessage, error) {
	var results []map[string]json.RawMessage

	for _, o := range objects {
		interfaceMap, err := t.ObjectToJsonRawMap(o)

		if err != nil {
			return nil, err
		}

		implementObjects, ok := implementObjectsMap[o.Id]

		if ok {
			for key, value := range implementObjects {
				interfaceMap[key] = value
			}
		}

		additional, ok := additionalFields[o.Id]

		if ok {
			for key, value := range additional {
				interfaceMap[key] = value.ToJsonRaw()
			}
		}

		results = append(results, interfaceMap)
	}

	return results, nil
}

func (d *Database) Kill() {
	d.watchForNewTables = false

	for _, t := range d.tables {
		t.Storage.Kill()
	}
}

func (d *Database) Delete() error {
	for tableName, t := range d.tables {
		delete(d.tables, tableName)

		err := t.Delete()

		if err != nil {
			return err
		}
	}

	return os.RemoveAll(d.path + d.Name)
}

func (d *Database) DeleteTable(tableName string) error {
	t := d.tables[tableName]

	if t == nil {
		return e.TableDoesNotExist()
	}

	delete(d.tables, tableName)
	return t.Delete()
}
