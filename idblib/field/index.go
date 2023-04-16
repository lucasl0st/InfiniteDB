/*
 * Copyright (c) 2023 Lucas Pape
 */

package field

import (
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"github.com/lucasl0st/InfiniteDB/idblib/object"
	"hash/fnv"
	"regexp"
	"sync"
)

const InternalObjectIdField = "INTERNAL_OBJECT_ID"

type Index struct {
	fields         map[string]Field
	maps           map[string]map[uint32][]int64
	reverseMaps    map[string]map[int64]dbtype.DBType
	mapLock        sync.RWMutex
	reverseMapLock sync.RWMutex
}

func NewIndex(fields map[string]Field) *Index {
	index := &Index{
		fields:      fields,
		maps:        make(map[string]map[uint32][]int64),
		reverseMaps: make(map[string]map[int64]dbtype.DBType),
	}

	for _, field := range fields {
		index.maps[field.Name] = make(map[uint32][]int64)
		index.reverseMaps[field.Name] = make(map[int64]dbtype.DBType)
	}

	index.maps[InternalObjectIdField] = make(map[uint32][]int64)
	index.reverseMaps[InternalObjectIdField] = make(map[int64]dbtype.DBType)

	return index
}

func (i *Index) Index(o object.Object) {
	for fieldName, field := range i.fields {
		if field.Indexed && field.Name != InternalObjectIdField {
			i.add(fieldName, o.M[fieldName], o.Id)
		}
	}

	n, err := dbtype.NumberFromInt64(o.Id)

	if err != nil {
		panic(err.Error())
	}

	i.add(InternalObjectIdField, n, o.Id)
}

func (i *Index) UnIndex(o object.Object) {
	for fieldName, field := range i.fields {
		if field.Indexed && field.Name != InternalObjectIdField {
			i.remove(fieldName, o.M[fieldName], o.Id)
		}
	}

	n, err := dbtype.NumberFromInt64(o.Id)

	if err != nil {
		panic(err.Error())
	}

	i.remove(InternalObjectIdField, n, o.Id)
}

func (i *Index) UpdateIndex(o object.Object) {
	i.UnIndex(o)
	i.Index(o)
}

func (i *Index) hash(value dbtype.DBType) uint32 {
	h := fnv.New32a()
	h.Write([]byte(value.ToString()))
	return h.Sum32()
}

func (i *Index) add(field string, value dbtype.DBType, id int64) {
	i.mapLock.Lock()
	i.reverseMapLock.Lock()
	defer i.mapLock.Unlock()
	defer i.reverseMapLock.Unlock()

	hash := i.hash(value)

	i.maps[field][hash] = append(i.maps[field][hash], id)
	i.reverseMaps[field][id] = value
}

func (i *Index) remove(field string, value dbtype.DBType, id int64) {
	i.mapLock.Lock()
	i.reverseMapLock.Lock()
	defer i.mapLock.Unlock()
	defer i.reverseMapLock.Unlock()

	var elements []int64

	hash := i.hash(value)

	for _, ie := range i.maps[field][hash] {
		if ie != id {
			elements = append(elements, id)
		}
	}

	i.maps[field][hash] = elements
	delete(i.reverseMaps[field], id)
}

func (i *Index) GetValue(field string, id int64) dbtype.DBType {
	i.reverseMapLock.RLock()
	defer i.reverseMapLock.RUnlock()

	return i.reverseMaps[field][id]
}

func (i *Index) getValues(field string, ids []int64) []dbtype.DBType {
	var results []dbtype.DBType

	for _, id := range ids {
		results = append(results, i.GetValue(field, id))
	}

	return results
}

func (i *Index) Equal(field string, value dbtype.DBType) []int64 {
	i.mapLock.RLock()
	defer i.mapLock.RUnlock()

	hash := i.hash(value)

	var results []int64

	for _, id := range i.maps[field][hash] {
		if i.GetValue(field, id).Equal(value) {
			results = append(results, id)
		}
	}

	return results
}

func (i *Index) Not(field string, value dbtype.DBType) []int64 {
	i.mapLock.RLock()
	defer i.mapLock.RUnlock()

	hash := i.hash(value)

	var results []int64

	for compareHash, ids := range i.maps[field] {
		if hash != compareHash {
			results = append(results, ids...)
		}
	}

	return results
}

func (i *Index) Match(field string, r regexp.Regexp) []int64 {
	i.reverseMapLock.RLock()
	defer i.reverseMapLock.RUnlock()

	var results []int64

	for id, compareValue := range i.reverseMaps[field] {
		if compareValue.Matches(r) {
			results = append(results, id)
		}
	}

	return results
}

func (i *Index) Larger(field string, value dbtype.DBType) ([]int64, error) {
	i.reverseMapLock.RLock()
	defer i.reverseMapLock.RUnlock()

	var results []int64

	for id, compareValue := range i.reverseMaps[field] {
		if compareValue.Larger(value) {
			results = append(results, id)
		}
	}

	return results, nil
}

func (i *Index) Smaller(field string, value dbtype.DBType) ([]int64, error) {
	i.reverseMapLock.RLock()
	defer i.reverseMapLock.RUnlock()

	var results []int64

	for id, compareValue := range i.reverseMaps[field] {
		if compareValue.Smaller(value) {
			results = append(results, id)
		}
	}

	return results, nil
}

func (i *Index) Between(field string, smaller dbtype.DBType, larger dbtype.DBType) ([]int64, error) {
	i.reverseMapLock.RLock()
	defer i.reverseMapLock.RUnlock()

	var results []int64

	for id, compareValue := range i.reverseMaps[field] {
		if compareValue.Between(smaller, larger) {
			results = append(results, id)
		}
	}

	return results, nil
}
