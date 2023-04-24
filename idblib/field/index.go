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
	core int

	// cpu core -> fieldName -> objectId -> value
	values map[int]map[int64]dbtype.DBType

	sync.RWMutex
}

func NewIndex() *Index {
	index := &Index{
		core:   0,
		values: map[int]map[int64]dbtype.DBType{},
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		index.values[i] = map[int64]dbtype.DBType{}
	}

	return index
}

func (i *Index) Add(value dbtype.DBType, id int64) {
	i.Lock()
	defer i.Unlock()

	if i.core >= len(i.values) {
		i.core = 0
	}

	i.values[i.core][id] = value

	i.core++
}

func (i *Index) Remove(id int64) {
	i.Lock()
	defer i.Unlock()

	for k := 0; k < len(i.values); k++ {
		delete(i.values[k], id)
	}
}

func (i *Index) GetValue(id int64) dbtype.DBType {
	i.RLock()
	defer i.RUnlock()

	for k := 0; k < len(i.values); k++ {
		value, ok := i.values[k][id]

		if ok {
			return value
		}
	}

	return nil
}

func (i *Index) rangeMap(compare func(compareValue dbtype.DBType) bool) []int64 {
	i.RLock()
	defer i.RUnlock()

	resultsChan := make(chan []int64, len(i.values)+1)
	var wg sync.WaitGroup

	for k := 0; k < len(i.values); k++ {
		wg.Add(1)

		go func(i *Index, core int) {
			defer wg.Done()

			var results []int64

			for id, value := range i.values[core] {
				if compare(value) {
					results = append(results, id)
				}
			}

			resultsChan <- results
		}(i, k)
	}

	wg.Wait()

	close(resultsChan)

	var results []int64

	for result := range resultsChan {
		results = append(results, result...)
	}

	return results
}

func (i *Index) Equal(value dbtype.DBType) []int64 {
	return i.rangeMap(func(compareValue dbtype.DBType) bool {
		return compareValue.Equal(value)
	})
}

func (i *Index) Not(value dbtype.DBType) []int64 {
	return i.rangeMap(func(compareValue dbtype.DBType) bool {
		return compareValue.Not(value)
	})
}

func (i *Index) Match(r regexp.Regexp) []int64 {
	return i.rangeMap(func(compareValue dbtype.DBType) bool {
		return compareValue.Matches(r)
	})
}

func (i *Index) Larger(value dbtype.DBType) []int64 {
	return i.rangeMap(func(compareValue dbtype.DBType) bool {
		return compareValue.Larger(value)
	})
}

func (i *Index) Smaller(value dbtype.DBType) []int64 {
	return i.rangeMap(func(compareValue dbtype.DBType) bool {
		return compareValue.Smaller(value)
	})
}

func (i *Index) Between(smaller dbtype.DBType, larger dbtype.DBType) []int64 {
	return i.rangeMap(func(compareValue dbtype.DBType) bool {
		return compareValue.Between(smaller, larger)
	})
}
