/*
 * Copyright (c) 2023 Lucas Pape
 */

package file

import (
	"bufio"
	"os"
	"sync"
)

type CachedScanner struct {
	path string

	fileLock sync.Mutex
	file     *os.File

	lineCacheLock sync.RWMutex
	lineCache     map[int64]int64

	opened bool
}

func NewCachedScanner(path string) *CachedScanner {
	return &CachedScanner{
		path:      path,
		lineCache: map[int64]int64{},
		opened:    false,
	}
}

func (c *CachedScanner) Open() error {
	if c.opened {
		return nil
	}

	f, err := openReadOnlyFile(c.path)

	if err != nil {
		return err
	}

	c.file = f
	c.opened = true
	return nil
}

func (c *CachedScanner) Close() error {
	if !c.opened {
		return nil
	}

	c.fileLock.Lock()
	defer c.fileLock.Unlock()

	c.opened = false
	return c.file.Close()
}

func (c *CachedScanner) GetLineFrom(lineNumber int64, reader func(lineNumber int64, line string) bool) error {
	c.fileLock.Lock()
	defer c.fileLock.Unlock()

	c.lineCacheLock.RLock()
	_, ok := c.lineCache[lineNumber]
	c.lineCacheLock.RUnlock()

	if !ok {
		err := c.buildCache()

		if err != nil {
			return err
		}
	}

	c.lineCacheLock.RLock()
	offset, ok := c.lineCache[lineNumber]
	c.lineCacheLock.RUnlock()

	if !ok {
		return nil
	}

	_, err := c.file.Seek(offset, 0)

	if err != nil {
		return err
	}

	line := lineNumber

	scanner := bufio.NewScanner(c.file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		if !reader(line, scanner.Text()) {
			break
		}

		line++
	}

	return scanner.Err()
}

func (c *CachedScanner) buildCache() error {
	c.lineCacheLock.Lock()
	defer c.lineCacheLock.Unlock()

	var offset int64 = 0
	var line int64 = 0

	if len(c.lineCache) > 0 {
		line = int64(len(c.lineCache)) - 1
		offset = c.lineCache[line]
	}

	_, err := c.file.Seek(offset, 0)

	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(c.file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		c.lineCache[line] = offset
		offset += int64(len(scanner.Bytes())) + 1
		line++
	}

	return err
}

func (c *CachedScanner) NumberOfLines() (int, error) {
	c.fileLock.Lock()
	defer c.fileLock.Unlock()

	err := c.buildCache()

	if err != nil {
		return 0, err
	}

	return len(c.lineCache), nil
}
