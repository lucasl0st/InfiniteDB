/*
 * Copyright (c) 2023 Lucas Pape
 */

package index

import (
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"sync"
)

type ExactIndex struct {
	//exact value as string -> array of objectIds
	values map[string][]int64

	sync.RWMutex
}

func NewExactIndex() *ExactIndex {
	return &ExactIndex{
		values: map[string][]int64{},
	}
}

func (i *ExactIndex) Add(value dbtype.DBType, id int64) {
	i.Lock()
	defer i.Unlock()

	str := value.ToString()

	i.values[str] = append(i.values[str], id)
}

func (i *ExactIndex) Remove(value dbtype.DBType, id int64) {
	i.Lock()
	defer i.Unlock()

	str := value.ToString()

	var removed []int64

	for _, compareId := range i.values[str] {
		if compareId != id {
			removed = append(removed, compareId)
		}
	}

	i.values[str] = removed
}

func (i *ExactIndex) Get(value dbtype.DBType) []int64 {
	i.RLock()
	defer i.RUnlock()

	return i.values[value.ToString()]
}
