/*
 * Copyright (c) 2023 Lucas Pape
 */

package field

import (
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"regexp"
	"runtime"
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

func (i *Index) rangeMap(compare func(id int64, compareValue dbtype.DBType)) {
	i.RLock()
	defer i.RUnlock()

	keys := make(chan int64, len(i.values))

	for k := range i.values {
		keys <- k
	}

	close(keys)

	var wg sync.WaitGroup

	for k := 0; k < runtime.NumCPU(); k++ {
		wg.Add(1)

		go func() {
			for key := range keys {
				compare(key, i.values[key])
			}

			wg.Done()
		}()
	}

	wg.Wait()
}

func (i *Index) Equal(value dbtype.DBType) []int64 {
	var results []int64

	i.rangeMap(func(id int64, compareValue dbtype.DBType) {
		if compareValue.Equal(value) {
			results = append(results, id)
		}
	})

	return results
}

func (i *Index) Not(value dbtype.DBType) []int64 {
	var results []int64

	i.rangeMap(func(id int64, compareValue dbtype.DBType) {
		if compareValue.Not(value) {
			results = append(results, id)
		}
	})

	return results
}

func (i *Index) Match(r regexp.Regexp) []int64 {
	var results []int64

	i.rangeMap(func(id int64, compareValue dbtype.DBType) {
		if compareValue.Matches(r) {
			results = append(results, id)
		}
	})

	return results
}

func (i *Index) Larger(value dbtype.DBType) []int64 {
	var results []int64

	i.rangeMap(func(id int64, compareValue dbtype.DBType) {
		if compareValue.Larger(value) {
			results = append(results, id)
		}
	})

	return results
}

func (i *Index) Smaller(value dbtype.DBType) []int64 {
	var results []int64

	i.rangeMap(func(id int64, compareValue dbtype.DBType) {
		if compareValue.Smaller(value) {
			results = append(results, id)
		}
	})

	return results
}

func (i *Index) Between(smaller dbtype.DBType, larger dbtype.DBType) []int64 {
	var results []int64

	i.rangeMap(func(id int64, compareValue dbtype.DBType) {
		if compareValue.Between(smaller, larger) {
			results = append(results, id)
		}
	})

	return results
}
