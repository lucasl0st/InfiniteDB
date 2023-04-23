/*
 * Copyright (c) 2023 Lucas Pape
 */

package field

import (
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"regexp"
	"sync"
)

const InternalObjectIdField = "INTERNAL_OBJECT_ID"

type Index struct {
	// fieldName -> objectId -> value
	values map[int64]dbtype.DBType

	sync.RWMutex
}

func NewIndex() *Index {
	index := &Index{
		values: map[int64]dbtype.DBType{},
	}

	return index
}

func (i *Index) Add(value dbtype.DBType, id int64) {
	i.Lock()
	defer i.Unlock()

	i.values[id] = value
}

func (i *Index) Remove(id int64) {
	i.Lock()
	defer i.Unlock()

	delete(i.values, id)
}

func (i *Index) GetValue(id int64) dbtype.DBType {
	i.RLock()
	defer i.RUnlock()

	return i.values[id]
}

func (i *Index) Equal(value dbtype.DBType) []int64 {
	i.RLock()
	defer i.RUnlock()

	var results []int64

	for id, compareValue := range i.values {
		if compareValue.Equal(value) {
			results = append(results, id)
		}
	}

	return results
}

func (i *Index) Not(value dbtype.DBType) []int64 {
	i.RLock()
	defer i.RUnlock()

	var results []int64

	for id, compareValue := range i.values {
		if compareValue.Not(value) {
			results = append(results, id)
		}
	}

	return results
}

func (i *Index) Match(r regexp.Regexp) []int64 {
	i.RLock()
	defer i.RUnlock()

	var results []int64

	for id, compareValue := range i.values {
		if compareValue.Matches(r) {
			results = append(results, id)
		}
	}

	return results
}

func (i *Index) Larger(value dbtype.DBType) []int64 {
	i.RLock()
	defer i.RUnlock()

	var results []int64

	for id, compareValue := range i.values {
		if compareValue.Larger(value) {
			results = append(results, id)
		}
	}

	return results
}

func (i *Index) Smaller(value dbtype.DBType) []int64 {
	i.RLock()
	defer i.RUnlock()

	var results []int64

	for id, compareValue := range i.values {
		if compareValue.Smaller(value) {
			results = append(results, id)
		}
	}

	return results
}

func (i *Index) Between(smaller dbtype.DBType, larger dbtype.DBType) []int64 {
	i.RLock()
	defer i.RUnlock()

	var results []int64

	for id, compareValue := range i.values {
		if compareValue.Between(smaller, larger) {
			results = append(results, id)
		}
	}

	return results
}
