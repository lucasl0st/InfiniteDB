/*
 * Copyright (c) 2023 Lucas Pape
 */

package storage

import (
	"github.com/fsnotify/fsnotify"
	"github.com/lucasl0st/InfiniteDB/idblib/file"
	"github.com/lucasl0st/InfiniteDB/idblib/metrics"
	idbutil "github.com/lucasl0st/InfiniteDB/idblib/util"
)

type SharedFile struct {
	file *file.File
	*file.Lock

	readLines int64
	addedLine func(lineNumber int64, line string)

	watcher *fsnotify.Watcher
	watch   bool

	logger idbutil.Logger
}

func New(path string, addedLine func(lineNumber int64, line string), logger idbutil.Logger) (*SharedFile, error) {
	f, err := file.New(path)

	if err != nil {
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		return nil, err
	}

	err = watcher.Add(path)

	if err != nil {
		return nil, err
	}

	s := &SharedFile{
		file:      f,
		Lock:      file.NewLock(path + ".lock"),
		readLines: 0,
		addedLine: addedLine,
		watcher:   watcher,
		watch:     true,
		logger:    logger,
	}

	err = s.readChanges()

	if err != nil {
		return nil, err
	}

	go func() {
		for s.watch {
			s.watchChanges()
		}
	}()

	return s, nil
}

func (s *SharedFile) Write(events []Event, getLine func(event Event, lineNumber int64) string) error {
	measurementId := metrics.StartTimingMeasurement()
	defer metrics.StopTimingMeasurement(measurementId)

	err := s.Lock.Lock()

	if err != nil {
		return err
	}

	defer func() {
		err = s.Lock.Unlock()
	}()

	var lines []string

	for _, o := range events {
		line := getLine(o, s.readLines)
		lines = append(lines, line)
	}

	err = s.file.Append(lines)

	if err != nil {
		return err
	}

	err = s.readChanges()

	if err != nil {
		return err
	}

	return err
}

func (s *SharedFile) Read(lineNumbers []int64) (map[int64]string, error) {
	measurementId := metrics.StartTimingMeasurement()
	defer metrics.StopTimingMeasurement(measurementId)

	return s.file.Read(lineNumbers)
}

func (s *SharedFile) readChanges() error {
	err := s.file.ReadAtStartLine(s.readLines, s.addedLine)

	if err != nil {
		return err
	}

	lines, err := s.file.NumberOfLines()

	if err != nil {
		return err
	}

	s.readLines = int64(lines)
	return nil
}

func (s *SharedFile) watchChanges() {
	select {
	case event, ok := <-s.watcher.Events:
		if !ok {
			return
		}

		if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
			if !s.HaveLock {
				err := s.readChanges()

				if err != nil {
					s.logger.Fatal(err.Error())
				}
			}
		}
	}
}

func (s *SharedFile) Kill() {
	s.watch = false
}
