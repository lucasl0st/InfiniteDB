/*
 * Copyright (c) 2023 Lucas Pape
 */

package field

import (
	"github.com/lucasl0st/InfiniteDB/idblib/object"
	"github.com/lucasl0st/InfiniteDB/util"
	"regexp"
	"sync"
)

type Index struct {
	fields      map[string]Field
	maps        map[string]map[string][]int64
	reverseMaps map[string]map[int64]string
	sync.RWMutex
}

func NewIndex(fields map[string]Field) *Index {
	index := &Index{
		fields:      fields,
		maps:        make(map[string]map[string][]int64),
		reverseMaps: make(map[string]map[int64]string),
	}

	for _, field := range fields {
		index.maps[field.Name] = make(map[string][]int64)
		index.reverseMaps[field.Name] = make(map[int64]string)
	}

	return index
}

func (i *Index) Index(o object.Object) {
	for fieldName, field := range i.fields {
		if field.Indexed {
			v := util.InterfaceToString(o.M[fieldName])
			i.add(fieldName, v, o.Id)
		}
	}
}

func (i *Index) UnIndex(o object.Object) {
	for fieldName, field := range i.fields {
		if field.Indexed {
			i.remove(fieldName, util.InterfaceToString(o.M[fieldName]), o.Id)
		}
	}
}

func (i *Index) UpdateIndex(o object.Object) {
	i.UnIndex(o)
	i.Index(o)
}

func (i *Index) add(field string, key string, id int64) {
	i.Lock()
	defer i.Unlock()
	i.maps[field][key] = append(i.maps[field][key], id)
	i.reverseMaps[field][id] = key
}

func (i *Index) remove(field string, key string, id int64) {
	i.Lock()
	defer i.Unlock()

	var elements []int64

	for _, ie := range i.maps[field][key] {
		if ie != id {
			elements = append(elements, id)
		}
	}

	i.maps[field][key] = elements
	delete(i.reverseMaps[field], id)
}

func (i *Index) GetValue(field string, id int64) string {
	i.RLock()
	defer i.RUnlock()

	return i.reverseMaps[field][id]
}

func (i *Index) Equal(field string, key string) []int64 {
	i.RLock()
	defer i.RUnlock()

	return i.maps[field][key]
}

func (i *Index) Not(field string, key string) []int64 {
	i.RLock()
	defer i.RUnlock()

	var results []int64

	for mapKey := range i.maps[field] {
		if key != mapKey {
			results = append(results, i.maps[field][mapKey]...)
		}
	}

	return results
}

func (i *Index) Match(field string, r regexp.Regexp) []int64 {
	i.RLock()
	defer i.RUnlock()

	var results []int64

	for mapKey := range i.maps[field] {
		if r.MatchString(mapKey) {
			results = append(results, i.maps[field][mapKey]...)
		}
	}

	return results
}

func (i *Index) Larger(field string, key string, parseNumber bool) ([]int64, error) {
	i.RLock()
	defer i.RUnlock()

	var results []int64

	for mapKey := range i.maps[field] {
		if parseNumber {
			keyInt, err := util.StringToNumber(key)

			if err != nil {
				return nil, err
			}

			mapKeyInt, err := util.StringToNumber(mapKey)

			if err != nil {
				return nil, err
			}

			if mapKeyInt > keyInt {
				results = append(results, i.maps[field][mapKey]...)
			}
		} else {
			if mapKey > key {
				results = append(results, i.maps[field][mapKey]...)
			}
		}
	}

	return results, nil
}

func (i *Index) Smaller(field string, key string, parseNumber bool) ([]int64, error) {
	i.RLock()
	defer i.RUnlock()

	var results []int64

	for mapKey := range i.maps[field] {
		if parseNumber {
			keyInt, err := util.StringToNumber(key)

			if err != nil {
				return nil, err
			}

			mapKeyInt, err := util.StringToNumber(mapKey)

			if err != nil {
				return nil, err
			}

			if mapKeyInt < keyInt {
				results = append(results, i.maps[field][mapKey]...)
			}
		} else {
			if mapKey < key {
				results = append(results, i.maps[field][mapKey]...)
			}
		}
	}

	return results, nil
}

func (i *Index) Between(field string, smaller string, larger string, parseNumber bool) ([]int64, error) {
	i.RLock()
	defer i.RUnlock()

	var results []int64

	for mapKey := range i.maps[field] {
		if parseNumber {
			smallerInt, err := util.StringToNumber(smaller)

			if err != nil {
				return nil, err
			}

			largerInt, err := util.StringToNumber(larger)

			if err != nil {
				return nil, err
			}

			mapKeyInt, err := util.StringToNumber(mapKey)

			if err != nil {
				return nil, err
			}

			if mapKeyInt > smallerInt && mapKeyInt < largerInt {
				results = append(results, i.maps[field][mapKey]...)
			}
		} else {
			if mapKey > smaller && mapKey < larger {
				results = append(results, i.maps[field][mapKey]...)
			}
		}
	}

	return results, nil
}
