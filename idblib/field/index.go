/*
 * Copyright (c) 2023 Lucas Pape
 */

package field

import (
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"github.com/lucasl0st/InfiniteDB/idblib/object"
	"regexp"
	"sync"
)

type Index struct {
	fields      map[string]Field
	maps        map[string]map[dbtype.DBType][]int64
	reverseMaps map[string]map[int64]dbtype.DBType
	sync.RWMutex
}

func NewIndex(fields map[string]Field) *Index {
	index := &Index{
		fields:      fields,
		maps:        make(map[string]map[dbtype.DBType][]int64),
		reverseMaps: make(map[string]map[int64]dbtype.DBType),
	}

	for _, field := range fields {
		index.maps[field.Name] = make(map[dbtype.DBType][]int64)
		index.reverseMaps[field.Name] = make(map[int64]dbtype.DBType)
	}

	return index
}

func (i *Index) Index(o object.Object) {
	for fieldName, field := range i.fields {
		if field.Indexed {
			i.add(fieldName, o.M[fieldName], o.Id)
		}
	}
}

func (i *Index) UnIndex(o object.Object) {
	for fieldName, field := range i.fields {
		if field.Indexed {
			i.remove(fieldName, o.M[fieldName], o.Id)
		}
	}
}

func (i *Index) UpdateIndex(o object.Object) {
	i.UnIndex(o)
	i.Index(o)
}

func (i *Index) add(field string, value dbtype.DBType, id int64) {
	i.Lock()
	defer i.Unlock()
	i.maps[field][value] = append(i.maps[field][value], id)
	i.reverseMaps[field][id] = value
}

func (i *Index) remove(field string, value dbtype.DBType, id int64) {
	i.Lock()
	defer i.Unlock()

	var elements []int64

	for _, ie := range i.maps[field][value] {
		if ie != id {
			elements = append(elements, id)
		}
	}

	i.maps[field][value] = elements
	delete(i.reverseMaps[field], id)
}

func (i *Index) GetValue(field string, id int64) dbtype.DBType {
	i.RLock()
	defer i.RUnlock()

	return i.reverseMaps[field][id]
}

func (i *Index) Equal(field string, value dbtype.DBType) []int64 {
	i.RLock()
	defer i.RUnlock()

	return i.maps[field][value]
}

func (i *Index) Not(field string, value dbtype.DBType) []int64 {
	i.RLock()
	defer i.RUnlock()

	var results []int64

	for compareValue, ids := range i.maps[field] {
		if compareValue.Not(value) {
			results = append(results, ids...)
		}
	}

	return results
}

func (i *Index) Match(field string, r regexp.Regexp) []int64 {
	i.RLock()
	defer i.RUnlock()

	var results []int64

	for compareValue, ids := range i.maps[field] {
		if compareValue.Matches(r) {
			results = append(results, ids...)
		}
	}

	return results
}

func (i *Index) Larger(field string, value dbtype.DBType) ([]int64, error) {
	i.RLock()
	defer i.RUnlock()

	var results []int64

	for compareValue, ids := range i.maps[field] {
		if compareValue.Larger(value) {
			results = append(results, ids...)
		}
	}

	return results, nil
}

func (i *Index) Smaller(field string, value dbtype.DBType) ([]int64, error) {
	i.RLock()
	defer i.RUnlock()

	var results []int64

	for compareValue, ids := range i.maps[field] {
		if compareValue.Smaller(value) {
			results = append(results, ids...)
		}
	}

	return results, nil
}

func (i *Index) Between(field string, smaller dbtype.DBType, larger dbtype.DBType) ([]int64, error) {
	i.RLock()
	defer i.RUnlock()

	var results []int64

	for compareValue, ids := range i.maps[field] {
		if compareValue.Between(smaller, larger) {
			results = append(results, ids...)
		}
	}

	return results, nil
}
