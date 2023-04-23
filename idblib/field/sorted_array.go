/*
 * Copyright (c) 2023 Lucas Pape
 */

package field

import (
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"sort"
)

type SortedArray struct {
	a                   []int64
	getValueForObjectId func(id int64) dbtype.DBType
}

func NewSortedArray(getValueForObjectId func(id int64) dbtype.DBType) *SortedArray {
	return &SortedArray{
		getValueForObjectId: getValueForObjectId,
	}
}

func (a *SortedArray) Insert(value dbtype.DBType, objectId int64) {
	if len(a.a) == 0 {
		a.a = append(a.a, objectId)
		return
	}

	i := a.FindStartIndex(value)

	a.a = append(a.a, 0)
	copy(a.a[i+1:], a.a[i:])
	a.a[i] = objectId
}

func (a *SortedArray) Get(i int) *int64 {
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
	return sort.Search(len(a.a), func(i int) bool {
		return a.getValueForObjectId(a.a[i]).Larger(value)
	})
}

func (a *SortedArray) FindEndIndex(value dbtype.DBType, start int) int {
	end := start

	for end < len(a.a) {
		if a.getValueForObjectId(a.a[end]).Equal(value) {
			end++
		} else {
			break
		}
	}

	return end
}
