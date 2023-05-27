/*
 * Copyright (c) 2023 Lucas Pape
 */

package file

import (
	e "github.com/lucasl0st/InfiniteDB/models/errors"
	"os"
	"time"
)

type Lock struct {
	path     string
	HaveLock bool
}

func NewLock(path string) *Lock {
	return &Lock{
		path:     path,
		HaveLock: false,
	}
}

func (l *Lock) Lock() error {
	if !l.HaveLock {
		locked, err := l.locked()

		if err != nil {
			return err
		}

		if locked {
			time.Sleep(time.Millisecond * 100)
			return l.Lock()
		}

		file, err := os.Create(l.path)

		if err != nil {
			return err
		}

		err = file.Close()

		if err != nil {
			return err
		}

		l.HaveLock = true
	}

	return nil
}

func (l *Lock) Unlock() error {
	if !l.HaveLock {
		return e.DontHaveLock()
	}

	err := os.Remove(l.path)

	if err != nil {
		return err
	}

	l.HaveLock = false

	return nil
}

func (l *Lock) locked() (bool, error) {
	return fileExists(l.path)
}
