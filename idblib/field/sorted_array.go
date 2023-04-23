/*
 * Copyright (c) 2023 Lucas Pape
 */

package field

import (
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"sort"
	"sync"
)

type SortedArray struct {
	a         []int64
	fieldName string
	sync.RWMutex
	getValueForObjectId func(fieldName string, id int64) dbtype.DBType
}

func NewSortedArray(fieldName string, getValueForObjectId func(fieldName string, id int64) dbtype.DBType) *SortedArray {
	return &SortedArray{
		fieldName:           fieldName,
		getValueForObjectId: getValueForObjectId,
	}
}

func (a *SortedArray) Insert(value dbtype.DBType, objectId int64) {
	if len(a.a) == 0 {
		a.Lock()
		defer a.Unlock()

		a.a = append(a.a, objectId)
		return
	}

	i := a.FindStartIndex(value)

	a.Lock()
	defer a.Unlock()

	a.a = append(a.a, 0)
	copy(a.a[i+1:], a.a[i:])
	a.a[i] = objectId
}

func (a *SortedArray) Get(i int) *int64 {
	a.RLock()
	defer a.RUnlock()

	if i >= a.Length() {
		return nil
	}

	return &a.a[i]
}

func (a *SortedArray) Length() int {
	return len(a.a)
}

func (a *SortedArray) Remove(value dbtype.DBType, objectId int64) {
	start := a.FindStartIndex(value)

	a.Lock()
	defer a.Unlock()

	index := -1

	for i := start; i < len(a.a); i++ {
		if a.a[i] == objectId {
			index = i
			break
		}
	}

	if index == -1 {
		return
	}

	a.a = append(a.a[:index], a.a[index+1:]...)
}

func (a *SortedArray) FindStartIndex(value dbtype.DBType) int {
	a.RLock()
	defer a.RUnlock()

	return sort.Search(len(a.a), func(i int) bool {
		return a.getValueForObjectId(a.fieldName, a.a[i]).Larger(value)
	})
}

func (a *SortedArray) FindEndIndex(value dbtype.DBType, start int) int {
	end := start

	for end < len(a.a) {
		if a.getValueForObjectId(a.fieldName, a.a[end]).Equal(value) {
			end++
		} else {
			break
		}
	}

	return end
}
