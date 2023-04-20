/*
 * Copyright (c) 2023 Lucas Pape
 */

package storage

import (
	"bufio"
	"github.com/fsnotify/fsnotify"
	"github.com/lucasl0st/InfiniteDB/idblib/metrics"
	"github.com/lucasl0st/InfiniteDB/idblib/util"
	e "github.com/lucasl0st/InfiniteDB/models/errors"
	"os"
	"time"
)

type File struct {
	path             string
	lockPath         string
	readLine         int64
	AddedLine        func(l string)
	totalLinesAtLock int64
	haveLock         bool
	watcher          *fsnotify.Watcher
	watch            bool
	FileChanged      func()
	l                util.Logger
	m                *metrics.Metrics
}

func NewFile(path string, logger util.Logger, metrics *metrics.Metrics) (*File, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_, err = os.Create(path)

		if err != nil {
			return nil, err
		}
	}

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		return nil, err
	}

	err = watcher.Add(path)

	if err != nil {
		return nil, err
	}

	f := &File{
		path:             path,
		lockPath:         path + ".lock",
		readLine:         0,
		totalLinesAtLock: 0,
		haveLock:         false,
		watcher:          watcher,
		watch:            true,
		FileChanged: func() {

		},
		l: logger,
		m: metrics,
	}

	go func() {
		for f.watch {
			f.watchChanges()
		}
	}()

	return f, nil
}

func (f *File) AppendLines(lines []string) error {
	file, err := f.writable()

	if err != nil {
		return err
	}

	defer func() {
		err = file.Close()
	}()

	for _, line := range lines {
		_, err = file.WriteString(line + "\n")

		if err != nil {
			return err
		}

		f.readLine += 1
		f.AddedLine(line)

		f.m.WroteObject()
		f.m.AddTotalObject()
	}

	return err
}

func (f *File) ReadLine(lineNumber int64) (string, error) {
	file, err := f.readOnly()

	if err != nil {
		return "", err
	}

	defer func() {
		err = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var line int64 = 0

	for scanner.Scan() {
		if line == lineNumber {
			return scanner.Text(), scanner.Err()
		}

		line++
	}

	return "", err
}

func (f *File) ReadLines(lineNumbers []int64) ([]string, error) {
	if lineNumbers == nil {
		return nil, nil
	}

	file, err := f.readOnly()

	if err != nil {
		return nil, err
	}

	defer func() {
		err = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var lines []string

	var i = 0
	var line int64 = 0

	for scanner.Scan() {
		if line == lineNumbers[i] {
			text := scanner.Text()

			for {
				lines = append(lines, text)

				i++

				if i >= len(lineNumbers) {
					break
				}

				if lineNumbers[i] != line {
					break
				}
			}

			if i >= len(lineNumbers) {
				break
			}
		}

		line++
	}

	return lines, nil
}

func (f *File) Kill() {
	f.watch = false
}

func (f *File) countLines() int64 {
	file, err := f.readOnly()

	if err != nil {
		f.l.Fatal(err)
	}

	defer func() {
		err = file.Close()

		if err != nil {
			f.l.Fatal(err)
		}
	}()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var lines int64 = 0

	for scanner.Scan() {
		lines += 1
	}

	return lines
}

func (f *File) ReserveId() (int64, error) {
	if !f.haveLock {
		return 0, e.DontHaveLock()
	}

	id := f.totalLinesAtLock
	f.totalLinesAtLock += 1

	return id, nil
}

func (f *File) Lock() {
	if !f.haveLock {
		if _, err := os.Stat(f.lockPath); err == nil {
			time.Sleep(time.Millisecond * 100)
			f.Lock()
		} else if os.IsNotExist(err) {
			file, err := os.Create(f.lockPath)

			if err != nil {
				f.l.Fatal(err)
			}

			err = file.Close()

			if err != nil {
				f.l.Fatal(err)
			}

			f.haveLock = true
		}
	}

	f.totalLinesAtLock = f.countLines()
}

func (f *File) Unlock() {
	if !f.haveLock {
		f.l.Fatal(e.DontHaveLock())
	}

	err := os.Remove(f.lockPath)

	if err != nil {
		f.l.Fatal(err)
	}

	f.haveLock = false
}

func (f *File) Read() error {
	file, err := f.readOnly()

	if err != nil {
		return err
	}

	defer func() {
		err = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var line int64 = 0

	for scanner.Scan() {
		if line >= f.readLine {
			f.AddedLine(scanner.Text())

			f.readLine += 1

			f.m.AddTotalObject()
		}

		line++
	}

	return err
}

func (f *File) writable() (*os.File, error) {
	return os.OpenFile(f.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

func (f *File) readOnly() (*os.File, error) {
	return os.OpenFile(f.path, os.O_CREATE, 0644)
}

func (f *File) watchChanges() {
	select {
	case event, ok := <-f.watcher.Events:
		if !ok {
			return
		}

		if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
			if !f.haveLock {
				f.FileChanged()
			}
		}
	}
}

func (f *File) NumberOfObjects() int64 {
	return f.readLine
}
