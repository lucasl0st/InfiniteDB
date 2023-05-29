/*
 * Copyright (c) 2023 Lucas Pape
 */

package file

import (
	"github.com/lucasl0st/InfiniteDB/idblib/metrics"
	"os"
	"sort"
	"sync"
)

type File struct {
	path string

	cachedScanner *CachedScanner
	sync.Mutex
}

func New(path string) (*File, error) {
	exists, err := fileExists(path)

	if err != nil {
		return nil, err
	}

	if !exists {
		_, err := os.Create(path)

		if err != nil {
			return nil, err
		}
	}

	return &File{
		path:          path,
		cachedScanner: NewCachedScanner(path),
	}, nil
}

func (f *File) Append(lines []string) error {
	measurementId := metrics.StartTimingMeasurement()
	defer metrics.StopTimingMeasurement(measurementId)

	f.Lock()
	defer f.Unlock()

	err := f.cachedScanner.Close()

	if err != nil {
		return err
	}

	file, err := openWritableFile(f.path)

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
	}

	return err
}

func (f *File) Read(lineNumbers []int64) ([]string, error) {
	measurementId := metrics.StartTimingMeasurement()
	defer metrics.StopTimingMeasurement(measurementId)

	if len(lineNumbers) == 0 {
		return nil, nil
	}

	sort.Slice(lineNumbers, func(i, j int) bool {
		return lineNumbers[i] < lineNumbers[j]
	})

	f.Lock()
	defer f.Unlock()

	err := f.cachedScanner.Open()

	if err != nil {
		return nil, err
	}

	var lines []string

	for _, lineNumber := range lineNumbers {
		err = f.cachedScanner.GetLineFrom(lineNumber, func(lineNumber int64, line string) bool {
			lines = append(lines, line)

			return false
		})

		if err != nil {
			return nil, err
		}
	}

	return lines, nil
}

func (f *File) ReadAtStartLine(start int64, readLine func(line string)) error {
	measurementId := metrics.StartTimingMeasurement()
	defer metrics.StopTimingMeasurement(measurementId)

	f.Lock()
	defer f.Unlock()

	err := f.cachedScanner.Open()

	if err != nil {
		return err
	}

	err = f.cachedScanner.GetLineFrom(start, func(lineNumber int64, line string) bool {
		readLine(line)
		return true
	})

	if err != nil {
		return err
	}

	return nil
}

func (f *File) NumberOfLines() (int, error) {
	f.Lock()
	defer f.Unlock()

	err := f.cachedScanner.Open()

	if err != nil {
		return 0, err
	}

	lines, err := f.cachedScanner.NumberOfLines()

	if err != nil {
		return 0, err
	}

	return lines, nil
}
