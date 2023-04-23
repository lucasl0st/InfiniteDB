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

	//fieldName -> sorted values
	sortedValues *SortedArray

	sync.RWMutex
}

func NewIndex() *Index {
	index := &Index{
		values: map[int64]dbtype.DBType{},
	}

	index.sortedValues = NewSortedArray(index.getValue)

	return index
}

func (i *Index) Add(value dbtype.DBType, id int64) {
	i.Lock()
	defer i.Unlock()

	i.sortedValues.Insert(value, id)

	i.values[id] = value
}

func (i *Index) Remove(value dbtype.DBType, id int64) {
	i.Lock()
	defer i.Unlock()

	i.sortedValues.Remove(value, id)

	//TODO deleting this is bugged. It does not have any negative effect besides using memory so just leaving this for later
	//delete(i.values, id)
}

func (i *Index) GetValue(id int64) dbtype.DBType {
	i.RLock()
	defer i.RUnlock()

	return i.getValue(id)
}

func (i *Index) getValue(id int64) dbtype.DBType {
	return i.values[id]
}

func (i *Index) Equal(value dbtype.DBType) []int64 {
	i.RLock()
	defer i.RUnlock()

	start := i.sortedValues.FindStartIndex(value)
	end := i.sortedValues.FindEndIndex(value, start)

	var results []int64

	for k := start; k <= end; k++ {
		id := i.sortedValues.Get(k)

		if id == nil {
			return results
		}

		if i.getValue(*id).Equal(value) {
			results = append(results, *id)
		}
	}

	return results
}

func (i *Index) Not(value dbtype.DBType) []int64 {
	i.RLock()
	defer i.RUnlock()

	start := i.sortedValues.FindStartIndex(value)
	end := i.sortedValues.FindEndIndex(value, start)

	var results []int64

	for k := 0; k < i.sortedValues.Length(); k++ {
		if k > start && k < end {
			continue
		}

		id := i.sortedValues.Get(k)

		if i.getValue(*id).Not(value) {
			results = append(results, *id)
		}
	}

	return results
}

func (i *Index) Match(r regexp.Regexp) []int64 {
	i.RLock()
	defer i.RUnlock()

	var results []int64

	for k := 0; k < i.sortedValues.Length(); k++ {
		id := *i.sortedValues.Get(k)

		if i.values[id].Matches(r) {
			results = append(results, id)
		}
	}

	return results
}

func (i *Index) Larger(value dbtype.DBType) []int64 {
	i.RLock()
	defer i.RUnlock()

	start := i.sortedValues.FindStartIndex(value)

	var results []int64

	for k := start; k < i.sortedValues.Length(); k++ {
		results = append(results, *i.sortedValues.Get(k))
	}

	return results
}

func (i *Index) Smaller(value dbtype.DBType) []int64 {
	i.RLock()
	defer i.RUnlock()

	start := i.sortedValues.FindStartIndex(value)
	end := i.sortedValues.FindEndIndex(value, start)

	var results []int64

	for k := end; k >= 0; k-- {
		id := i.sortedValues.Get(k)

		if id == nil {
			continue
		}

		results = append(results, *id)
	}

	return results
}

func (i *Index) Between(smaller dbtype.DBType, larger dbtype.DBType) []int64 {
	i.RLock()
	defer i.RUnlock()

	start := i.sortedValues.FindStartIndex(smaller)
	startLarger := i.sortedValues.FindStartIndex(larger)
	end := i.sortedValues.FindEndIndex(larger, startLarger)

	var results []int64

	for k := start; k <= end; k++ {
		id := i.sortedValues.Get(k)

		if id == nil {
			return results
		}

		results = append(results, *id)
	}

	return results
}
