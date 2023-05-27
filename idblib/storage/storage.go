/*
 * Copyright (c) 2023 Lucas Pape
 */

package storage

import (
	"encoding/json"
	"errors"
	"github.com/lucasl0st/InfiniteDB/idblib/cache"
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"github.com/lucasl0st/InfiniteDB/idblib/field"
	"github.com/lucasl0st/InfiniteDB/idblib/metrics"
	idblib "github.com/lucasl0st/InfiniteDB/idblib/object"
	idbutil "github.com/lucasl0st/InfiniteDB/idblib/util"
	"github.com/lucasl0st/InfiniteDB/util"
)

const objectsFileName = "objects.idb"

type Storage struct {
	file *SharedFile
	c    *cache.Cache

	fields map[string]field.Field

	addedObject   func(object idblib.Object)
	deletedObject func(object idblib.Object)

	NumberOfObjects int64

	logger  idbutil.Logger
	metrics *metrics.Metrics
}

func NewStorage(
	path string,
	fields map[string]field.Field,
	addedObject func(object idblib.Object),
	deletedObject func(object idblib.Object),
	cacheSize uint,
	logger idbutil.Logger,
	metrics *metrics.Metrics,
) (*Storage, error) {
	s := &Storage{
		c:             cache.New(cacheSize),
		fields:        fields,
		addedObject:   addedObject,
		deletedObject: deletedObject,
		logger:        logger,
		metrics:       metrics,
	}

	file, err := New(path+objectsFileName, s.addedLineInFile, logger)

	if err != nil {
		return nil, err
	}

	s.file = file

	return s, nil
}

func (s *Storage) Kill() {
	s.file.Kill()
}

func (s *Storage) addedLineInFile(line string) {
	var storageObject object

	err := json.Unmarshal([]byte(line), &storageObject)

	if err != nil {
		s.logger.Fatal(err.Error())
	}

	if storageObject.Deleted != nil && *storageObject.Deleted == true {
		o := s.GetObject(*storageObject.RefersTo)
		s.deletedObject(*o)

		s.NumberOfObjects--

		return
	} else if storageObject.RefersTo != nil {
		o := s.GetObject(*storageObject.RefersTo)
		s.DeleteObject(*o)
	} else {
		s.NumberOfObjects++
		s.metrics.AddTotalObject()
	}

	s.addedObject(s.storageObjectToObject(storageObject))
}

func (s *Storage) GetObject(id int64) *idblib.Object {
	objects := s.GetObjects([]int64{id})

	if len(objects) > 1 {
		s.logger.Fatal(errors.New("too many results"))
	} else if len(objects) == 1 {
		return &objects[0]
	}

	return nil
}

func (s *Storage) GetObjects(ids []int64) []idblib.Object {
	var objects []idblib.Object

	var notCached []int64

	for _, id := range ids {
		cached := s.c.Get(id)

		if cached != nil {
			objects = append(objects, idblib.Object{
				Id: id,
				M:  *cached,
			})
		} else {
			notCached = append(notCached, id)
		}
	}

	lines, err := s.file.Read(notCached)

	if err != nil {
		s.logger.Fatal(err.Error())
	}

	for _, line := range lines {
		var storageObject object

		err = json.Unmarshal([]byte(line), &storageObject)

		if err != nil {
			s.logger.Fatal(err.Error())
		}

		o := s.storageObjectToObject(storageObject)

		objects = append(objects, o)

		s.c.Set(o)
	}

	return objects
}

func (s *Storage) AddObject(m map[string]dbtype.DBType) {
	s.writeObjects([]object{
		s.mapStringDbTypeToStorageObject(m),
	})
}

func (s *Storage) UpdateObject(o idblib.Object) {
	storageObject := s.mapStringDbTypeToStorageObject(o.M)
	storageObject.RefersTo = &o.Id

	s.writeObjects([]object{
		storageObject,
	})
}

func (s *Storage) DeleteObject(o idblib.Object) {
	storageObject := s.mapStringDbTypeToStorageObject(o.M)
	storageObject.RefersTo = &o.Id
	storageObject.Deleted = util.Ptr(true)

	s.writeObjects([]object{
		storageObject,
	})
}

func (s *Storage) writeObjects(objects []object) {
	err := s.file.Write(objects, func(object object, lineNumber int64) string {
		object.LineNumber = lineNumber

		bytes, err := json.Marshal(object)

		if err != nil {
			s.logger.Fatal(err.Error())
		}

		s.metrics.WroteObject()

		s.c.Remove(object.LineNumber)

		if object.RefersTo != nil {
			s.c.Remove(*object.RefersTo)
		}

		return string(bytes)
	})

	if err != nil {
		s.logger.Fatal(err.Error())
	}
}

func (s *Storage) storageObjectToObject(storageObject object) idblib.Object {
	o := idblib.Object{
		Id: storageObject.LineNumber,
		M:  map[string]dbtype.DBType{},
	}

	for _, f := range s.fields {
		str, ok := storageObject.Object[f.Name]

		if !ok {
			continue
		}

		v, err := idbutil.StringToDBType(str, f)

		if err != nil {
			s.logger.Fatal(err.Error())
		}

		o.M[f.Name] = v
	}

	return o
}

func (s *Storage) mapStringDbTypeToStorageObject(m map[string]dbtype.DBType) object {
	storageObject := object{
		Object: map[string]string{},
	}

	for key, value := range m {
		storageObject.Object[key] = value.ToString()
	}

	return storageObject
}
