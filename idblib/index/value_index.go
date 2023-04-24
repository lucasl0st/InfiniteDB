/*
 * Copyright (c) 2023 Lucas Pape
 */

package index

import (
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"runtime"
	"sync"
)

type ValueIndex struct {
	core int

	//core -> objectId -> value
	values map[int]map[int64]dbtype.DBType

	sync.RWMutex
}

func NewValueIndex() *ValueIndex {
	i := &ValueIndex{
		core:   0,
		values: map[int]map[int64]dbtype.DBType{},
	}

	for k := 0; k < runtime.NumCPU(); k++ {
		i.values[k] = map[int64]dbtype.DBType{}
	}

	return i
}

func (i *ValueIndex) Add(value dbtype.DBType, id int64) {
	i.Lock()
	defer i.Unlock()

	if i.core >= len(i.values) {
		i.core = 0
	}

	i.values[i.core][id] = value
	i.core++
}

func (i *ValueIndex) Remove(id int64) {
	i.Lock()
	defer i.Unlock()

	for k := 0; k < len(i.values); k++ {
		delete(i.values[k], id)
	}
}

func (i *ValueIndex) GetValue(id int64) dbtype.DBType {
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

func (i *ValueIndex) Range(compare func(compareValue dbtype.DBType) bool) []int64 {
	i.RLock()
	defer i.RUnlock()

	resultsChan := make(chan []int64, len(i.values)+1)
	var wg sync.WaitGroup

	for k := 0; k < len(i.values); k++ {
		wg.Add(1)

		go func(i *ValueIndex, core int) {
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
