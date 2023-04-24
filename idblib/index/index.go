/*
 * Copyright (c) 2023 Lucas Pape
 */

package index

import (
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"regexp"
)

type Index struct {
	valueIndex  *ValueIndex
	exactIndex  *ExactIndex
	sortedIndex *SortedIndex
}

func NewIndex() *Index {
	index := &Index{
		valueIndex: NewValueIndex(),
		exactIndex: NewExactIndex(),
	}

	index.sortedIndex = NewSortedIndex(index.valueIndex.GetValue)

	return index
}

func (i *Index) Add(value dbtype.DBType, id int64) {
	i.valueIndex.Add(value, id)
	i.exactIndex.Add(value, id)
	i.sortedIndex.Add(id)
}

func (i *Index) Remove(value dbtype.DBType, id int64) {
	i.sortedIndex.Remove(id)
	i.exactIndex.Remove(value, id)
	i.valueIndex.Remove(id)
}

func (i *Index) GetValue(id int64) dbtype.DBType {
	return i.valueIndex.GetValue(id)
}

func (i *Index) Equal(value dbtype.DBType) []int64 {
	return i.exactIndex.Get(value)
}

func (i *Index) Not(value dbtype.DBType) []int64 {
	return i.valueIndex.Range(func(compareValue dbtype.DBType) bool {
		return compareValue.Not(value)
	})
}

func (i *Index) Match(r regexp.Regexp) []int64 {
	return i.valueIndex.Range(func(compareValue dbtype.DBType) bool {
		return compareValue.Matches(r)
	})
}

func (i *Index) Larger(value dbtype.DBType) []int64 {
	return i.sortedIndex.Larger(value)
}

func (i *Index) Smaller(value dbtype.DBType) []int64 {
	return i.sortedIndex.Smaller(value)
}

func (i *Index) Between(smaller dbtype.DBType, larger dbtype.DBType) []int64 {
	return i.valueIndex.Range(func(compareValue dbtype.DBType) bool {
		return compareValue.Between(smaller, larger)
	})
}
