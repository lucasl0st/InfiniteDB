/*
 * Copyright (c) 2023 Lucas Pape
 */

package index

import (
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"sort"
	"sync"
)

type SortedIndex struct {
	values []int64

	sorted bool

	getValue func(id int64) dbtype.DBType

	sync.RWMutex
}

func NewSortedIndex(getValue func(id int64) dbtype.DBType) *SortedIndex {
	return &SortedIndex{
		sorted:   true,
		getValue: getValue,
	}
}

func (i *SortedIndex) Add(id int64) {
	i.Lock()
	defer i.Unlock()

	i.sorted = false
	i.values = append(i.values, id)
}

func (i *SortedIndex) Remove(id int64) {
	i.Lock()
	defer i.Unlock()

	var removed []int64

	for _, compareId := range i.values {
		if compareId != id {
			removed = append(removed, compareId)
		}
	}

	i.values = removed
}

func (i *SortedIndex) sort() {
	i.Lock()
	defer i.Unlock()

	sort.Slice(i.values, func(j, k int) bool {
		return i.getValue(i.values[j]).Smaller(i.getValue(i.values[k]))
	})

	i.sorted = true
}

func (i *SortedIndex) Larger(value dbtype.DBType) []int64 {
	if !i.sorted {
		i.sort()
	}

	i.RLock()
	defer i.RUnlock()

	for k, compareId := range i.values {
		if i.getValue(compareId).Larger(value) {
			var results []int64

			for j := k; j < len(i.values); j++ {
				results = append(results, i.values[j])
			}

			return results
		}
	}

	return nil
}

func (i *SortedIndex) Smaller(value dbtype.DBType) []int64 {
	if !i.sorted {
		i.sort()
	}

	i.RLock()
	defer i.RUnlock()

	for k := len(i.values) - 1; k >= 0; k-- {
		if i.getValue(i.values[k]).Smaller(value) {
			var results []int64

			for j := k; j >= 0; j-- {
				results = append(results, i.values[j])
			}

			return results
		}
	}

	return nil
}
