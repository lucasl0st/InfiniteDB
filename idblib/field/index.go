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

	valueIndex int64
	valuesLock sync.RWMutex

	// valueIndex -> value
	values map[int64]dbtype.DBType

	// valueIndex -> objectIds
	valueIndexToObjectIds     map[int64][]int64
	valueIndexToObjectIdsLock sync.RWMutex

	// fieldName -> sorted values index
	sortedValues     map[string][]int64
	sortedValuesLock sync.RWMutex

	// fieldName -> objectId -> values index
	objectIds     map[string]map[int64]int64
	objectIdsLock sync.RWMutex
}

func NewIndex(fields map[string]Field) *Index {
	index := &Index{
		fields:                fields,
		valueIndex:            0,
		values:                map[int64]dbtype.DBType{},
		valueIndexToObjectIds: map[int64][]int64{},
		sortedValues:          map[string][]int64{},
		objectIds:             map[string]map[int64]int64{},
	}

	for _, field := range fields {
		index.objectIds[field.Name] = map[int64]int64{}
	}

	index.objectIds[InternalObjectIdField] = map[int64]int64{}

	return index
}

func (i *Index) Index(o object.Object) {
	i.fieldsLock.Lock()
	defer i.fieldsLock.Unlock()

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
			i.remove(fieldName, o.Id)
		}
	}

	i.remove(InternalObjectIdField, o.Id)
}

func (i *Index) UpdateIndex(o object.Object) {
	i.UnIndex(o)
	i.Index(o)
}

func (i *Index) add(field string, value dbtype.DBType, id int64) {
	i.objectIdsLock.Lock()
	defer i.objectIdsLock.Unlock()

	i.valuesLock.Lock()
	defer i.valuesLock.Unlock()

	i.valueIndex++
	index := i.valueIndex

	i.values[index] = value

	i.insertSorted(field, value, index)

	i.valueIndexToObjectIdsLock.Lock()
	defer i.valueIndexToObjectIdsLock.Unlock()

	i.valueIndexToObjectIds[index] = append(i.valueIndexToObjectIds[index], id)
	i.objectIds[field][id] = index
}

func (i *Index) insertSorted(field string, value dbtype.DBType, valueIndex int64) {
	arr := i.sortedValues[field]

	if len(arr) == 0 {
		i.sortedValues[field] = append(arr, valueIndex)
		return
	}

	ind := i.findIndex(arr, value)

	arr = append(arr, 0)
	copy(arr[ind+1:], arr[ind:])
	arr[ind] = valueIndex

	i.sortedValues[field] = arr
}

func (i *Index) findIndex(arr []int64, value dbtype.DBType) int {
	left := 0
	right := len(arr) - 1

	for left <= right {
		mid := (left + right) / 2
		midValue := i.values[arr[mid]]

		if value.Equal(midValue) {
			return mid
		} else if value.Smaller(midValue) {
			right = mid - 1
		} else {
			left = mid + 1
		}
	}

	return left
}

func (i *Index) remove(field string, id int64) {
	i.objectIdsLock.Lock()
	defer i.objectIdsLock.Unlock()

	i.sortedValuesLock.Lock()
	defer i.sortedValuesLock.Unlock()

	i.valueIndexToObjectIdsLock.Lock()
	defer i.valueIndexToObjectIdsLock.Unlock()

	i.valuesLock.Lock()
	defer i.valuesLock.Unlock()

	valueIndex := i.objectIds[field][id]

	delete(i.objectIds[field], id)

	copy(i.sortedValues[field][valueIndex:], i.sortedValues[field][valueIndex+1:])
	i.sortedValues[field] = i.sortedValues[field][:len(i.sortedValues)-1]

	delete(i.valueIndexToObjectIds, valueIndex)

	delete(i.values, valueIndex)
}

func (i *Index) GetValue(field string, id int64) dbtype.DBType {
	i.objectIdsLock.RLock()
	defer i.objectIdsLock.RUnlock()

	i.valuesLock.RLock()
	defer i.valuesLock.RUnlock()

	return i.values[i.objectIds[field][id]]
}

func (i *Index) getObjectIds(valueIndex int64) []int64 {
	i.valueIndexToObjectIdsLock.RLock()
	defer i.valueIndexToObjectIdsLock.RUnlock()

	return i.valueIndexToObjectIds[valueIndex]
}

func (i *Index) getStartAndEndIndexes(field string, value dbtype.DBType) (int, int) {
	index := i.findIndex(i.sortedValues[field], value)

	if index >= len(i.sortedValues[field]) {
		return index, index
	}

	start := index
	end := index

	for start > 0 {
		if i.values[i.sortedValues[field][start]].Equal(value) {
			start--
		} else {
			break
		}
	}

	for end < len(i.sortedValues[field])-1 {
		if i.values[i.sortedValues[field][end]].Equal(value) {
			end++
		} else {
			break
		}
	}

	return start, end
}

func (i *Index) Equal(field string, value dbtype.DBType) []int64 {
	i.sortedValuesLock.RLock()
	defer i.sortedValuesLock.RUnlock()

	i.valuesLock.RLock()
	defer i.valuesLock.RUnlock()

	if len(i.sortedValues[field]) == 0 {
		return nil
	}

	start, end := i.getStartAndEndIndexes(field, value)

	if start >= len(i.sortedValues[field]) {
		return nil
	}

	var results []int64

	for k := start; k <= end; k++ {
		results = append(results, i.getObjectIds(i.sortedValues[field][k])...)
	}

	return results
}

func (i *Index) Not(field string, value dbtype.DBType) []int64 {
	i.sortedValuesLock.RLock()
	defer i.sortedValuesLock.RUnlock()

	i.valuesLock.RLock()
	defer i.valuesLock.RUnlock()

	if len(i.sortedValues[field]) == 0 {
		return nil
	}

	start, end := i.getStartAndEndIndexes(field, value)

	var results []int64

	for k := range i.sortedValues[field] {
		if k > start && k < end {
			continue
		}

		results = append(results, i.getObjectIds(i.sortedValues[field][k])...)
	}

	return results
}

func (i *Index) Match(field string, r regexp.Regexp) []int64 {
	i.sortedValuesLock.RLock()
	defer i.sortedValuesLock.RUnlock()

	i.valuesLock.RLock()
	defer i.valuesLock.RUnlock()

	if len(i.sortedValues[field]) == 0 {
		return nil
	}

	var results []int64

	for k := range i.sortedValues[field] {
		if i.values[i.sortedValues[field][k]].Matches(r) {
			results = append(results, i.getObjectIds(i.sortedValues[field][k])...)
		}
	}

	return results
}

func (i *Index) Larger(field string, value dbtype.DBType) []int64 {
	i.sortedValuesLock.RLock()
	defer i.sortedValuesLock.RUnlock()

	i.valuesLock.RLock()
	defer i.valuesLock.RUnlock()

	if len(i.sortedValues[field]) == 0 {
		return nil
	}

	start, _ := i.getStartAndEndIndexes(field, value)

	if start >= len(i.sortedValues[field]) {
		return nil
	}

	var results []int64

	for k := start; k < len(i.sortedValues); k++ {
		results = append(results, i.getObjectIds(i.sortedValues[field][k])...)
	}

	return results
}

func (i *Index) Smaller(field string, value dbtype.DBType) []int64 {
	i.sortedValuesLock.RLock()
	defer i.sortedValuesLock.RUnlock()

	i.valuesLock.RLock()
	defer i.valuesLock.RUnlock()

	if len(i.sortedValues[field]) == 0 {
		return nil
	}

	_, end := i.getStartAndEndIndexes(field, value)

	var results []int64

	for k := end; k >= 0; k-- {
		results = append(results, i.getObjectIds(i.sortedValues[field][k])...)
	}

	return results
}

func (i *Index) Between(field string, smaller dbtype.DBType, larger dbtype.DBType) []int64 {
	i.sortedValuesLock.RLock()
	defer i.sortedValuesLock.RUnlock()

	i.valuesLock.RLock()
	defer i.valuesLock.RUnlock()

	if len(i.sortedValues[field]) == 0 {
		return nil
	}

	start, _ := i.getStartAndEndIndexes(field, smaller)
	_, end := i.getStartAndEndIndexes(field, larger)

	var results []int64

	for k := start; k <= end; k++ {
		results = append(results, i.getObjectIds(i.sortedValues[field][k])...)
	}

	return results
}
