/*
 * Copyright (c) 2023 Lucas Pape
 */

package storage

import (
	"encoding/json"
	"github.com/lucasl0st/InfiniteDB/idblib/cache"
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"github.com/lucasl0st/InfiniteDB/idblib/field"
	"github.com/lucasl0st/InfiniteDB/idblib/metrics"
	idblib "github.com/lucasl0st/InfiniteDB/idblib/object"
	idbutil "github.com/lucasl0st/InfiniteDB/idblib/util"
	"sort"
	"sync"
	"time"
)

const objectsFileName = "objects.idb"

type Storage struct {
	file           *File
	c              *cache.Cache
	writeQueue     []object
	writeQueueLock sync.Mutex
	AddedObject    func(object idblib.Object)
	DeletedObject  func(object idblib.Object)
	l              idbutil.Logger
	write          bool
	fields         map[string]field.Field
}

func NewStorage(
	path string,
	addedObject func(object idblib.Object),
	deletedObject func(object idblib.Object),
	logger idbutil.Logger,
	metrics *metrics.Metrics,
	cacheSize uint,
	fields map[string]field.Field,
) (*Storage, error) {
	file, err := NewFile(path+objectsFileName, logger, metrics)

	if err != nil {
		return nil, err
	}

	s := &Storage{
		file:          file,
		c:             cache.New(cacheSize),
		l:             logger,
		AddedObject:   addedObject,
		DeletedObject: deletedObject,
		write:         true,
		fields:        fields,
	}

	s.file.AddedLine = s.addedLine
	s.file.FileChanged = s.fileChanged

	err = s.initialize()

	if err != nil {
		return nil, err
	}

	go func() {
		for s.write {
			s.writer()

			time.Sleep(time.Second)
		}
	}()

	return s, nil
}

func (s *Storage) initialize() error {
	return s.file.Read()
}

func (s *Storage) GetObject(id int64) *idblib.Object {
	c := s.c.Get(id)

	if c != nil {
		return &idblib.Object{
			Id: id,
			M:  *c,
		}
	}

	storageObject, err := s.readStorageObject(id)

	if err != nil {
		s.l.Fatal(err)
	}

	if storageObject.Deleted != nil && *storageObject.Deleted {
		return nil
	}

	o := s.storageObjectToObject(*storageObject)

	s.c.Set(o)

	return &o
}

func (s *Storage) GetObjects(ids []int64) []idblib.Object {
	var objects []idblib.Object

	var toRead []int64

	for _, id := range ids {
		c := s.c.Get(id)

		if c != nil {
			objects = append(objects, idblib.Object{
				Id: id,
				M:  *c,
			})
		} else {
			toRead = append(toRead, id)
		}
	}

	sort.Slice(toRead, func(i, j int) bool {
		return toRead[i] < toRead[j]
	})

	storageObjects, err := s.readStorageObjects(toRead)

	if err != nil {
		s.l.Fatal(err)
	}

	for _, storageObject := range storageObjects {
		if storageObject.Deleted != nil && *storageObject.Deleted {
			return nil
		}

		o := s.storageObjectToObject(storageObject)

		s.c.Set(o)

		objects = append(objects, o)
	}

	return objects
}

func (s *Storage) AddObject(o idblib.Object) {
	so := s.objectToStorageObject(o)
	s.addStorageObject(so)
}

func (s *Storage) DeleteObject(o idblib.Object) {
	deleted := true

	storageObject := object{
		Deleted:  &deleted,
		RefersTo: &o.Id,
	}

	s.addStorageObject(storageObject)
}

func (s *Storage) NumberOfObjects() int64 {
	return s.file.NumberOfObjects()
}

func (s *Storage) Kill() {
	s.c.Kill()

	s.file.Kill()

	s.write = false

	s.writer()
	s.writeQueueLock.Lock()
}

func (s *Storage) addStorageObject(storageObject object) {
	s.writeQueueLock.Lock()
	s.writeQueue = append(s.writeQueue, storageObject)
	s.writeQueueLock.Unlock()
}

func (s *Storage) getStorageObjectQueue() []object {
	s.writeQueueLock.Lock()
	queue := s.writeQueue
	s.writeQueue = []object{}
	s.writeQueueLock.Unlock()

	return queue
}

func (s *Storage) addedLine(line string) {
	var storageObject object

	err := json.Unmarshal([]byte(line), &storageObject)
	if err != nil {
		s.l.Fatal(err)
	}

	if storageObject.Deleted != nil && *storageObject.Deleted {
		o := s.GetObject(*storageObject.RefersTo)
		s.DeletedObject(*o)
	} else {
		o := s.storageObjectToObject(storageObject)
		s.AddedObject(o)
	}
}

func (s *Storage) fileChanged() {
	err := s.file.Read()

	if err != nil {
		s.l.Fatal(err)
	}
}

func (s *Storage) readStorageObject(id int64) (*object, error) {
	line, err := s.file.ReadLine(id)

	var storageObject object

	err = json.Unmarshal([]byte(line), &storageObject)
	if err != nil {
		return nil, err
	}

	if storageObject.Id != id {
		panic("wrong id!")
	}

	return &storageObject, nil
}

func (s *Storage) readStorageObjects(ids []int64) ([]object, error) {
	lines, err := s.file.ReadLines(ids)

	if err != nil {
		return nil, err
	}

	var storageObjects []object

	for _, line := range lines {
		var storageObject object

		err = json.Unmarshal([]byte(line), &storageObject)
		if err != nil {
			return nil, err
		}

		storageObjects = append(storageObjects, storageObject)
	}

	if len(lines) != len(ids) {
		panic("length of received lines is not the same as received ids")
	}

	return storageObjects, nil
}

func (s *Storage) writeStorageObjects(storageObjects []object) error {
	s.file.Lock()
	defer s.file.Unlock()

	var lines []string

	for _, storageObject := range storageObjects {
		s.c.Remove(storageObject.Id)

		id, err := s.file.ReserveId()

		if err != nil {
			return err
		}

		storageObject.Id = id

		bytes, err := json.Marshal(storageObject)

		if err != nil {
			return err
		}

		lines = append(lines, string(bytes))
	}

	return s.file.AppendLines(lines)
}

func (s *Storage) writer() {
	queue := s.getStorageObjectQueue()

	if len(queue) > 0 {
		err := s.writeStorageObjects(queue)

		if err != nil {
			s.l.Fatal(err)
		}
	}
}

func (s *Storage) objectToStorageObject(o idblib.Object) object {
	storageObject := object{
		Object: map[string]string{},
	}

	for key, value := range o.M {
		storageObject.Object[key] = value.ToString()
	}

	return storageObject
}

func (s *Storage) storageObjectToObject(storageObject object) idblib.Object {
	o := idblib.Object{
		Id: storageObject.Id,
		M:  map[string]dbtype.DBType{},
	}

	for _, f := range s.fields {
		str, ok := storageObject.Object[f.Name]

		if !ok {
			continue
		}

		v, err := idbutil.StringToDBType(str, f)

		if err != nil {
			panic(err.Error())
		}

		o.M[f.Name] = v
	}

	return o
}
