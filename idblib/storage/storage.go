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
)

const objectsFileName = "objects.idb"

type Storage struct {
	file *SharedFile
	c    *cache.Cache

	fields map[string]field.Field

	addedObject   func(object idblib.Object)
	deletedObject func(object idblib.Object)

	NumberOfObjects int64

	logger idbutil.Logger

	metricAddTotalObject func()
	metricWroteObject    func()
}

func NewStorage(
	path string,
	fields map[string]field.Field,
	addedObject func(object idblib.Object),
	deletedObject func(object idblib.Object),
	cacheSize uint,
	logger idbutil.Logger,
	metricAddTotalObject func(),
	metricWroteObject func(),
) (*Storage, error) {
	s := &Storage{
		c:                    cache.New(cacheSize),
		fields:               fields,
		addedObject:          addedObject,
		deletedObject:        deletedObject,
		logger:               logger,
		metricAddTotalObject: metricAddTotalObject,
		metricWroteObject:    metricWroteObject,
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

func (s *Storage) addedLineInFile(lineNumber int64, line string) {
	var event Event

	err := json.Unmarshal([]byte(line), &event)

	if err != nil {
		s.logger.Fatal(err.Error())
	}

	if event.Type == EventTypeRemove {
		o := s.GetObject(*event.RefersTo)
		s.deletedObject(*o)

		s.NumberOfObjects--
		return
	}

	if event.Type == EventTypeUpdate {
		o := s.GetObject(*event.RefersTo)
		s.deletedObject(*o)
	}

	if event.Type == EventTypeAdd {
		s.NumberOfObjects++
		s.metricAddTotalObject()
	}

	s.addedObject(s.eventToObject(lineNumber, event))
}

func (s *Storage) GetObject(id int64) *idblib.Object {
	measurementId := metrics.StartTimingMeasurement()
	defer metrics.StopTimingMeasurement(measurementId)

	objects := s.GetObjects([]int64{id})

	if len(objects) > 1 {
		s.logger.Fatal(errors.New("too many results"))
	} else if len(objects) == 1 {
		return &objects[0]
	}

	return nil
}

func (s *Storage) GetObjects(ids []int64) []idblib.Object {
	measurementId := metrics.StartTimingMeasurement()
	defer metrics.StopTimingMeasurement(measurementId)

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

	for lineNumber, line := range lines {
		var event Event

		err = json.Unmarshal([]byte(line), &event)

		if err != nil {
			s.logger.Fatal(err.Error())
		}

		o := s.eventToObject(lineNumber, event)

		objects = append(objects, o)

		s.c.Set(o)
	}

	return objects
}

func (s *Storage) AddObject(m map[string]dbtype.DBType) {
	measurementId := metrics.StartTimingMeasurement()
	defer metrics.StopTimingMeasurement(measurementId)

	s.writeEvents([]Event{
		s.mapStringDbTypeToEvent(m, EventTypeAdd, nil),
	})
}

func (s *Storage) UpdateObject(o idblib.Object) {
	measurementId := metrics.StartTimingMeasurement()
	defer metrics.StopTimingMeasurement(measurementId)

	s.writeEvents([]Event{
		s.mapStringDbTypeToEvent(o.M, EventTypeUpdate, &o.Id),
	})
}

func (s *Storage) RemoveObject(o idblib.Object) {
	measurementId := metrics.StartTimingMeasurement()
	defer metrics.StopTimingMeasurement(measurementId)

	s.writeEvents([]Event{
		s.mapStringDbTypeToEvent(nil, EventTypeRemove, &o.Id),
	})
}

func (s *Storage) writeEvents(events []Event) {
	err := s.file.Write(events, func(event Event, lineNumber int64) string {
		bytes, err := json.Marshal(event)

		if err != nil {
			s.logger.Fatal(err.Error())
		}

		s.metricWroteObject()

		//invalidate cache
		s.c.Remove(lineNumber)

		if event.RefersTo != nil {
			s.c.Remove(*event.RefersTo)
		}

		return string(bytes)
	})

	if err != nil {
		s.logger.Fatal(err.Error())
	}
}

func (s *Storage) eventToObject(eventId int64, event Event) idblib.Object {
	o := idblib.Object{
		Id: eventId,
		M:  map[string]dbtype.DBType{},
	}

	for _, f := range s.fields {
		str, ok := event.Data[f.Name]

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

func (s *Storage) mapStringDbTypeToEvent(m map[string]dbtype.DBType, eventType EventType, refersTo *int64) Event {
	event := Event{
		Type:     eventType,
		Data:     map[string]string{},
		RefersTo: refersTo,
	}

	for key, value := range m {
		event.Data[key] = value.ToString()
	}

	return event
}
