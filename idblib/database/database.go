/*
 * Copyright (c) 2023 Lucas Pape
 */

package database

import (
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	e "github.com/lucasl0st/InfiniteDB/errors"
	"github.com/lucasl0st/InfiniteDB/idblib/field"
	"github.com/lucasl0st/InfiniteDB/idblib/metrics"
	"github.com/lucasl0st/InfiniteDB/idblib/object"
	"github.com/lucasl0st/InfiniteDB/idblib/table"
	"github.com/lucasl0st/InfiniteDB/request"
	"github.com/lucasl0st/InfiniteDB/util"
	"os"
	"strings"
	"time"
)

type Database struct {
	Name              string
	path              string
	tablesPath        string
	tables            map[string]*table.Table
	l                 util.Logger
	m                 *metrics.Metrics
	cacheSize         uint
	watchForNewTables bool
	watcher           *fsnotify.Watcher
}

func NewDatabase(name string, path string, logger util.Logger, metrics *metrics.Metrics, cacheSize uint) (*Database, int, error) {
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

	for _, tableFolder := range files {
		if !tableFolder.IsDir() {
			continue
		}

		err := d.loadTable(tableFolder.Name())

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

	for tableName := range d.tables {
		tableNames = append(tableNames, tableName)
	}

	return tableNames
}

func (d *Database) Get(tableName string, request table.Request) ([]object.Object, error) {
	if request.Query != nil {
		t := d.tables[tableName]

		if t == nil {
			return nil, e.TableDoesNotExist()
		}

		objects, additionalFields, err := t.Query(*request.Query, nil, make(table.AdditionalFields))

		if request.Sort != nil {
			objects, err = t.Sort(objects, request.Sort.Field, additionalFields, request.Sort.Direction)
		}

		objects = t.SkipAndLimit(objects, request.Skip, request.Limit)

		results := t.Storage.GetObjects(objects)

		if request.Implement != nil {
			for _, implement := range request.Implement {
				results, err = d.implement(implement, results)
			}
		}

		return results, err
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
		err = t.Remove(&o)

		if err != nil {
			return count, err
		}

		count++
	}

	return count, nil
}

func (d *Database) Insert(tableName string, o map[string]interface{}) error {
	t := d.tables[tableName]

	if t == nil {
		return e.TableDoesNotExist()
	}

	return t.Insert(o)
}

func (d *Database) Update(tableName string, o map[string]interface{}) error {
	t := d.tables[tableName]

	if t == nil {
		return e.TableDoesNotExist()
	}

	foundObjectId, err := t.FindExisting(o)

	if err != nil {
		return err
	}

	foundObject := t.Storage.GetObject(foundObjectId)

	for key, value := range o {
		foundObject.M[key] = value
	}

	return t.Update(foundObject)
}

func (d *Database) implement(implement request.Implement, objects []object.Object) ([]object.Object, error) {
	fromTable := d.tables[implement.From.Table]

	if fromTable == nil {
		return nil, e.TableDoesNotExist()
	}

	for _, o := range objects {
		f := util.InterfaceToString(o.M[implement.Field])

		implementObjects, _, err := fromTable.Query(table.Query{
			Where: &request.Where{
				Field:    implement.From.Field,
				Operator: request.EQUALS,
				Value:    &f,
			},
		}, nil, nil)

		if err != nil {
			return nil, err
		}

		if len(implementObjects) > 0 {
			as := implement.From.Table

			if implement.As != nil {
				as = *implement.As
			}

			if len(implementObjects) > 1 || (implement.ForceArray != nil && *implement.ForceArray) {
				maps := make([]map[string]interface{}, 0)

				for _, o := range implementObjects {
					maps = append(maps, fromTable.Storage.GetObject(o).M)
				}

				o.M[as] = maps
			} else {
				io := fromTable.Storage.GetObject(implementObjects[0])
				o.M[as] = io.M
			}
		}
	}

	return objects, nil
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
