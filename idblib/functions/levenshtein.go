/*
 * Copyright (c) 2023 Lucas Pape
 */

package functions

import (
	"encoding/json"
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"github.com/lucasl0st/InfiniteDB/idblib/object"
	"github.com/lucasl0st/InfiniteDB/idblib/table"
	"github.com/lucasl0st/InfiniteDB/util"
)

const fieldNameLevenshtein = "levenshtein"

type LevenshteinFunction struct {
	value     string
	fieldName string
	as        string
}

func (l *LevenshteinFunction) Run(
	table *table.Table,
	objects object.Objects,
	additionalFields table.AdditionalFields,
	parameters map[string]json.RawMessage,
) (object.Objects, table.AdditionalFields, error) {
	err := l.parseParameters(parameters)

	if err != nil {
		return nil, nil, err
	}

	str1 := []rune(l.value)

	for _, o := range objects {
		var str2 dbtype.Text

		if additionalFields[o][l.fieldName] != nil {
			str2 = additionalFields[o][l.fieldName].(dbtype.Text)
		} else {
			index, err := table.GetIndex(l.fieldName)

			if err != nil {
				return nil, nil, err
			}

			str2 = index.GetValue(o).(dbtype.Text)
		}

		if additionalFields[o] == nil {
			additionalFields[o] = make(map[string]dbtype.DBType)
		}

		lev, err := l.levenshtein(str1, []rune(str2.ToString()))

		if err != nil {
			return nil, nil, err
		}

		additionalFields[o][l.as] = lev
	}

	return objects, additionalFields, nil
}

func (l *LevenshteinFunction) parseParameters(parameters map[string]json.RawMessage) error {
	v, err := util.JsonRawToString(parameters["value"])

	if err != nil {
		return err
	}

	fieldName, err := util.JsonRawToString(parameters["fieldName"])

	if err != nil {
		return err
	}

	as := fieldNameLevenshtein

	asP, err := util.JsonRawToString(parameters["as"])

	if asP != nil && len(*asP) > 0 && err == nil {
		as = *asP
	}

	l.value = *v
	l.fieldName = *fieldName
	l.as = as

	return nil
}

func (l *LevenshteinFunction) levenshtein(str1, str2 []rune) (dbtype.Number, error) {
	s1len := len(str1)
	s2len := len(str2)
	column := make([]int, len(str1)+1)

	for y := 1; y <= s1len; y++ {
		column[y] = y
	}
	for x := 1; x <= s2len; x++ {
		column[0] = x
		lastkey := x - 1
		for y := 1; y <= s1len; y++ {
			oldkey := column[y]
			var incr int
			if str1[y-1] != str2[x-1] {
				incr = 1
			}

			column[y] = l.minimum(column[y]+1, column[y-1]+1, lastkey+incr)
			lastkey = oldkey
		}
	}

	return dbtype.NumberFromInt64(int64(column[s1len]))
}

func (l *LevenshteinFunction) minimum(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
	} else {
		if b < c {
			return b
		}
	}
	return c
}
