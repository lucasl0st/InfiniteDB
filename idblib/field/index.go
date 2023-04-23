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

const InternalObjectIdField = "INTERNAL_OBJECT_ID"

type Index struct {
	fields     map[string]Field
	fieldsLock sync.RWMutex

	// fieldName -> objectId -> value
	values     map[string]map[int64]dbtype.DBType
	valuesLock sync.RWMutex

	//fieldName -> sorted values
	sortedValues     map[string]*SortedArray
	sortedValuesLock sync.RWMutex
}

func NewIndex(fields map[string]Field) *Index {
	index := &Index{
		fields:       fields,
		values:       map[string]map[int64]dbtype.DBType{},
		sortedValues: map[string]*SortedArray{},
	}

	for _, field := range fields {
		index.values[field.Name] = map[int64]dbtype.DBType{}

		index.sortedValues[field.Name] = NewSortedArray(field.Name, index.getValue)
	}

	index.values[InternalObjectIdField] = map[int64]dbtype.DBType{}
	index.sortedValues[InternalObjectIdField] = NewSortedArray(InternalObjectIdField, index.getValue)

	return index
}

func (i *Index) Index(o object.Object) {
	i.fieldsLock.RLock()
	defer i.fieldsLock.RUnlock()

	var wg sync.WaitGroup

	for fieldName, field := range i.fields {
		if field.Indexed && field.Name != InternalObjectIdField {
			wg.Add(1)

			go func(fieldName string) {
				i.add(fieldName, o.M[fieldName], o.Id)
				wg.Done()
			}(fieldName)
		}
	}

	n, err := dbtype.NumberFromInt64(o.Id)

	if err != nil {
		panic(err.Error())
	}

	wg.Add(1)

	go func() {
		i.add(InternalObjectIdField, n, o.Id)
		wg.Done()
	}()

	wg.Wait()
}

func (i *Index) UnIndex(o object.Object) {
	i.fieldsLock.RLock()
	defer i.fieldsLock.RUnlock()

	var wg sync.WaitGroup

	for fieldName, field := range i.fields {
		if field.Indexed && field.Name != InternalObjectIdField {
			wg.Add(1)

			go func(fieldName string) {
				i.remove(fieldName, o.M[fieldName], o.Id)
				wg.Done()
			}(fieldName)
		}
	}

	n, err := dbtype.NumberFromInt64(o.Id)

	if err != nil {
		panic(err.Error())
	}

	wg.Add(1)

	go func() {
		i.remove(InternalObjectIdField, n, o.Id)
		wg.Done()
	}()

	wg.Wait()
}

func (i *Index) UpdateIndex(o object.Object) {
	i.UnIndex(o)
	i.Index(o)
}

func (i *Index) add(field string, value dbtype.DBType, id int64) {
	i.sortedValuesLock.Lock()
	defer i.sortedValuesLock.Unlock()

	i.valuesLock.Lock()
	defer i.valuesLock.Unlock()

	i.sortedValues[field].Insert(value, id)
	i.values[field][id] = value
}

func (i *Index) remove(field string, value dbtype.DBType, id int64) {
	i.sortedValuesLock.Lock()
	defer i.sortedValuesLock.Unlock()

	i.valuesLock.Lock()
	defer i.valuesLock.Unlock()

	i.sortedValues[field].Remove(value, id)

	//TODO deleting this is bugged. It does not have any negative effect besides using memory so just leaving this for later
	//delete(i.values[field], id)
}

func (i *Index) GetValue(field string, id int64) dbtype.DBType {
	i.valuesLock.RLock()
	defer i.valuesLock.RUnlock()

	return i.getValue(field, id)
}

func (i *Index) getValue(field string, id int64) dbtype.DBType {
	return i.values[field][id]
}

func (i *Index) Equal(field string, value dbtype.DBType) []int64 {
	i.sortedValuesLock.RLock()
	defer i.sortedValuesLock.RUnlock()

	start := i.sortedValues[field].FindStartIndex(value)
	end := i.sortedValues[field].FindEndIndex(value, start)

	var results []int64

	for k := start; k <= end; k++ {
		id := i.sortedValues[field].Get(k)

		if id == nil {
			return results
		}

		if i.GetValue(field, *id).Equal(value) {
			results = append(results, *id)
		}
	}

	return results
}

func (i *Index) Not(field string, value dbtype.DBType) []int64 {
	i.sortedValuesLock.RLock()
	defer i.sortedValuesLock.RUnlock()

	start := i.sortedValues[field].FindStartIndex(value)
	end := i.sortedValues[field].FindEndIndex(value, start)

	var results []int64

	for k := 0; k < i.sortedValues[field].Length(); k++ {
		if k > start && k < end {
			continue
		}

		id := i.sortedValues[field].Get(k)

		if i.GetValue(field, *id).Not(value) {
			results = append(results, *id)
		}
	}

	return results
}

func (i *Index) Match(field string, r regexp.Regexp) []int64 {
	i.sortedValuesLock.RLock()
	defer i.sortedValuesLock.RUnlock()

	i.valuesLock.RLock()
	defer i.valuesLock.RUnlock()

	var results []int64

	for k := 0; k < i.sortedValues[field].Length(); k++ {
		id := *i.sortedValues[field].Get(k)

		if i.values[field][id].Matches(r) {
			results = append(results, id)
		}
	}

	return results
}

func (i *Index) Larger(field string, value dbtype.DBType) []int64 {
	i.sortedValuesLock.RLock()
	defer i.sortedValuesLock.RUnlock()

	start := i.sortedValues[field].FindStartIndex(value)

	var results []int64

	for k := start; k < i.sortedValues[field].Length(); k++ {
		results = append(results, *i.sortedValues[field].Get(k))
	}

	return results
}

func (i *Index) Smaller(field string, value dbtype.DBType) []int64 {
	i.sortedValuesLock.RLock()
	defer i.sortedValuesLock.RUnlock()

	start := i.sortedValues[field].FindStartIndex(value)
	end := i.sortedValues[field].FindEndIndex(value, start)

	var results []int64

	for k := end; k >= 0; k-- {
		id := i.sortedValues[field].Get(k)

		if id == nil {
			continue
		}

		results = append(results, *id)
	}

	return results
}

func (i *Index) Between(field string, smaller dbtype.DBType, larger dbtype.DBType) []int64 {
	i.sortedValuesLock.RLock()
	defer i.sortedValuesLock.RUnlock()

	start := i.sortedValues[field].FindStartIndex(smaller)
	startLarger := i.sortedValues[field].FindStartIndex(larger)
	end := i.sortedValues[field].FindEndIndex(larger, startLarger)

	var results []int64

	for k := start; k <= end; k++ {
		id := i.sortedValues[field].Get(k)

		if id == nil {
			return results
		}

		results = append(results, *id)
	}

	return results
}
