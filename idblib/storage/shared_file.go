/*
 * Copyright (c) 2023 Lucas Pape
 */

package storage

import (
	"github.com/fsnotify/fsnotify"
	"github.com/lucasl0st/InfiniteDB/idblib/file"
	idbutil "github.com/lucasl0st/InfiniteDB/idblib/util"
)

type SharedFile struct {
	file *file.File
	*file.Lock

	readLines int64
	addedLine func(line string)

	watcher *fsnotify.Watcher
	watch   bool

	logger idbutil.Logger
}

func New(path string, addedLine func(line string), logger idbutil.Logger) (*SharedFile, error) {
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

func (s *SharedFile) Write(objects []object, getLine func(object object, lineNumber int64) string) error {
	err := s.Lock.Lock()

	if err != nil {
		return err
	}

	defer func() {
		err = s.Lock.Unlock()
	}()

	var lines []string

	for _, o := range objects {
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

func (s *SharedFile) Read(lineNumbers []int64) ([]string, error) {
	return s.file.Read(lineNumbers)
}

func (s *SharedFile) readChanges() error {
	return s.file.ReadAtStartLine(s.readLines, func(line string) {
		s.addedLine(line)
		s.readLines++
	})
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