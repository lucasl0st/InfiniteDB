/*
 * Copyright (c) 2023 Lucas Pape
 */

package idblib

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/gammazero/workerpool"
	"github.com/lucasl0st/InfiniteDB/idblib/database"
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"github.com/lucasl0st/InfiniteDB/idblib/field"
	"github.com/lucasl0st/InfiniteDB/idblib/metrics"
	"github.com/lucasl0st/InfiniteDB/idblib/table"
	"github.com/lucasl0st/InfiniteDB/idblib/util"
	e "github.com/lucasl0st/InfiniteDB/models/errors"
	"github.com/lucasl0st/InfiniteDB/models/metric"
	"github.com/lucasl0st/InfiniteDB/models/request"
	"github.com/lucasl0st/InfiniteDB/models/response"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type IDB struct {
	databasePath   string
	databases      map[string]*database.Database
	l              util.Logger
	m              *metrics.Metrics
	cacheSize      uint
	watcher        *fsnotify.Watcher
	watchDatabases bool
	workerPool     *workerpool.WorkerPool
	ready          bool
}

func New(databasePath string, logger util.Logger, metricsReceiver *metric.Receiver, cacheSize uint, ready func()) (*IDB, error) {
	if _, err := os.Stat(databasePath); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(databasePath, os.ModePerm)

		if err != nil {
			return nil, err
		}
	}

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		return nil, err
	}

	err = watcher.Add(databasePath)

	if err != nil {
		return nil, err
	}

	workers := runtime.NumCPU()

	idb := &IDB{
		databasePath:   databasePath,
		databases:      make(map[string]*database.Database),
		l:              logger,
		m:              metrics.New(metricsReceiver),
		cacheSize:      cacheSize,
		watcher:        watcher,
		watchDatabases: true,
		workerPool:     workerpool.New(workers),
		ready:          false,
	}

	go func() {
		err = idb.loadDatabases()

		if err != nil {
			log.Fatal(err.Error())
		}

		idb.ready = true
		ready()
	}()

	go func() {
		for idb.watchDatabases {
			idb.databaseWatcher()
		}
	}()

	return idb, nil
}

func (i *IDB) databaseWatcher() {
	select {
	case event, ok := <-i.watcher.Events:
		if !ok {
			return
		}

		databaseName := strings.ReplaceAll(event.Name, i.databasePath, "")
		isDatabase := !strings.Contains(databaseName, "/")

		if !isDatabase {
			return
		}

		if event.Has(fsnotify.Create) {
			time.Sleep(time.Millisecond * 100)
			err := i.loadDatabases()

			if err != nil {
				i.l.Fatal(err.Error())
			}
		} else if event.Has(fsnotify.Remove) {
			d := i.databases[databaseName]

			if d == nil {
				return
			}

			d.Kill()
			delete(i.databases, databaseName)
		}
	}
}

func (i *IDB) Kill() {
	i.watchDatabases = false

	for _, d := range i.databases {
		d.Kill()
	}
}

func (i *IDB) GetDatabases() (response.GetDatabasesResponse, error) {
	if !i.ready {
		return response.GetDatabasesResponse{}, e.IdbNotReady()
	}

	var databaseNames []string

	for key := range i.databases {
		databaseNames = append(databaseNames, key)
	}

	return response.GetDatabasesResponse{Databases: databaseNames}, nil
}

func (i *IDB) CreateDatabase(name string) (response.CreateDatabaseResponse, error) {
	if !i.ready {
		return response.CreateDatabaseResponse{}, e.IdbNotReady()
	}

	if i.databases[name] != nil {
		return response.CreateDatabaseResponse{}, e.DatabaseAlreadyExists()
	}

	err := database.CreateDatabase(i.databasePath, name)

	if err != nil {
		return response.CreateDatabaseResponse{}, err
	}

	err = i.loadDatabase(name)

	if err != nil {
		return response.CreateDatabaseResponse{}, err
	}

	return response.CreateDatabaseResponse{
		Message: "Created database",
		Name:    name,
	}, nil
}

func (i *IDB) DeleteDatabase(name string) (response.DeleteDatabaseResponse, error) {
	if !i.ready {
		return response.DeleteDatabaseResponse{}, e.IdbNotReady()
	}

	d := i.databases[name]

	if d == nil {
		return response.DeleteDatabaseResponse{}, e.DatabaseDoesNotExist()
	}

	delete(i.databases, name)

	err := d.Delete()

	if err != nil {
		return response.DeleteDatabaseResponse{}, err
	}

	return response.DeleteDatabaseResponse{
		Message: "Deleted database",
		Name:    name,
	}, nil
}

func (i *IDB) loadDatabase(name string) error {
	if i.databases[name] != nil {
		return nil
	}

	start := time.Now()

	d, tables, err := database.NewDatabase(name, i.databasePath, i.l, i.m, i.cacheSize)

	if err != nil {
		return err
	}

	i.databases[name] = d

	elapsed := time.Since(start)

	i.l.Println("loaded database "+name+" with "+fmt.Sprint(tables)+" tables, took ", elapsed)

	return nil
}

func (i *IDB) loadDatabases() error {
	files, err := os.ReadDir(i.databasePath)

	if err != nil {
		return err
	}

	errs := make(chan error, len(files))
	var wg sync.WaitGroup

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		wg.Add(1)

		go func(name string) {
			defer wg.Done()
			errs <- i.loadDatabase(name)
		}(file.Name())
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

func (i *IDB) GetDatabase(name string) (response.GetDatabaseResponse, error) {
	if !i.ready {
		return response.GetDatabaseResponse{}, e.IdbNotReady()
	}

	d := i.databases[name]

	if d == nil {
		return response.GetDatabaseResponse{}, e.DatabaseDoesNotExist()
	}

	tableNames := d.GetTableNames()

	return response.GetDatabaseResponse{Name: d.Name, Tables: tableNames}, nil
}

func (i *IDB) GetDatabaseTable(name string, tableName string) (response.GetDatabaseTableResponse, error) {
	if !i.ready {
		return response.GetDatabaseTableResponse{}, e.IdbNotReady()
	}

	d := i.databases[name]

	if d == nil {
		return response.GetDatabaseTableResponse{}, e.DatabaseDoesNotExist()
	}

	fields, options, err := d.GetTable(tableName)

	if err != nil {
		return response.GetDatabaseTableResponse{}, nil
	}

	return response.GetDatabaseTableResponse{
		Name:      name,
		TableName: tableName,
		Fields:    fields,
		Options:   *options,
	}, nil
}

func (i *IDB) CreateTableInDatabase(name string, tableName string, fields map[string]field.Field, options request.TableOptions) (response.CreateTableInDatabaseResponse, error) {
	if !i.ready {
		return response.CreateTableInDatabaseResponse{}, e.IdbNotReady()
	}

	d := i.databases[name]

	if d == nil {
		return response.CreateTableInDatabaseResponse{}, e.DatabaseDoesNotExist()
	}

	err := d.CreateTable(tableName, fields, options)

	if err != nil {
		return response.CreateTableInDatabaseResponse{}, err
	}

	f := map[string]request.Field{}

	for fieldName, ff := range fields {
		f[fieldName] = request.Field{
			Type:    dbtype.DatabaseTypeToString(ff.Type),
			Indexed: &ff.Indexed,
			Unique:  &ff.Unique,
			Null:    &ff.Null,
		}
	}

	return response.CreateTableInDatabaseResponse{
		Name:      name,
		TableName: tableName,
		Fields:    f,
	}, nil
}

func (i *IDB) DeleteTableInDatabase(name string, tableName string) (response.DeleteTableInDatabaseResponse, error) {
	if !i.ready {
		return response.DeleteTableInDatabaseResponse{}, e.IdbNotReady()
	}

	d := i.databases[name]

	if d == nil {
		return response.DeleteTableInDatabaseResponse{}, e.DatabaseDoesNotExist()
	}

	err := d.DeleteTable(tableName)

	if err != nil {
		return response.DeleteTableInDatabaseResponse{}, err
	}

	return response.DeleteTableInDatabaseResponse{
		Message:   "Deleted table in database",
		TableName: tableName,
		Name:      name,
	}, nil
}

func (i *IDB) GetFromDatabaseTable(name string, tableName string, request table.Request) (response.GetFromDatabaseTableResponse, error) {
	if !i.ready {
		return response.GetFromDatabaseTableResponse{}, e.IdbNotReady()
	}

	measurementId := metrics.StartTimingMeasurement()
	defer metrics.StopTimingMeasurement(measurementId)

	d := i.databases[name]

	if d == nil {
		return response.GetFromDatabaseTableResponse{}, e.DatabaseDoesNotExist()
	}

	var wg sync.WaitGroup
	wg.Add(1)

	objectsChannel := make(chan []map[string]json.RawMessage, 1)
	errChannel := make(chan error, 1)

	i.workerPool.Submit(func() {
		defer wg.Done()

		objects, err := d.Get(tableName, request)

		objectsChannel <- objects
		errChannel <- err
	})

	wg.Wait()

	close(objectsChannel)
	close(errChannel)

	objects, err := <-objectsChannel, <-errChannel

	if err != nil {
		return response.GetFromDatabaseTableResponse{}, err
	}

	return response.GetFromDatabaseTableResponse{
		Name:      name,
		TableName: tableName,
		Results:   objects,
	}, nil
}

func (i *IDB) InsertToDatabaseTable(name string, tableName string, object map[string]json.RawMessage) (response.InsertToDatabaseTableResponse, error) {
	if !i.ready {
		return response.InsertToDatabaseTableResponse{}, e.IdbNotReady()
	}

	measurementId := metrics.StartTimingMeasurement()
	defer metrics.StopTimingMeasurement(measurementId)

	d := i.databases[name]

	if d == nil {
		return response.InsertToDatabaseTableResponse{}, e.DatabaseDoesNotExist()
	}

	var wg sync.WaitGroup
	wg.Add(1)

	errChannel := make(chan error, 1)

	i.workerPool.Submit(func() {
		defer wg.Done()

		err := d.Insert(tableName, object)

		errChannel <- err
	})

	wg.Wait()

	err := <-errChannel

	if err != nil {
		return response.InsertToDatabaseTableResponse{}, err
	}

	return response.InsertToDatabaseTableResponse{
		Name:      name,
		TableName: tableName,
		Object:    object,
	}, nil
}

func (i *IDB) RemoveFromDatabaseTable(name string, tableName string, request table.Request) (response.RemoveFromDatabaseTableResponse, error) {
	if !i.ready {
		return response.RemoveFromDatabaseTableResponse{}, e.IdbNotReady()
	}

	d := i.databases[name]

	if d == nil {
		return response.RemoveFromDatabaseTableResponse{}, e.DatabaseDoesNotExist()
	}

	var wg sync.WaitGroup
	wg.Add(1)

	countChannel := make(chan int64, 1)
	errChannel := make(chan error, 1)

	i.workerPool.Submit(func() {
		defer wg.Done()

		count, err := d.Remove(tableName, request)

		countChannel <- count
		errChannel <- err
	})

	wg.Wait()

	count, err := <-countChannel, <-errChannel

	if err != nil {
		return response.RemoveFromDatabaseTableResponse{}, err
	}

	return response.RemoveFromDatabaseTableResponse{
		Name:      name,
		TableName: tableName,
		Removed:   count,
	}, nil
}

func (i *IDB) UpdateInDatabaseTable(name string, tableName string, object map[string]json.RawMessage) (response.UpdateInDatabaseTableResponse, error) {
	if !i.ready {
		return response.UpdateInDatabaseTableResponse{}, e.IdbNotReady()
	}

	d := i.databases[name]

	if d == nil {
		return response.UpdateInDatabaseTableResponse{}, e.DatabaseDoesNotExist()
	}

	var wg sync.WaitGroup
	wg.Add(1)

	errChannel := make(chan error, 1)

	i.workerPool.Submit(func() {
		err := d.Update(tableName, object)

		errChannel <- err
	})

	wg.Wait()

	err := <-errChannel

	if err != nil {
		return response.UpdateInDatabaseTableResponse{}, err
	}

	return response.UpdateInDatabaseTableResponse{
		Name:      name,
		TableName: tableName,
		Object:    object,
	}, nil
}
