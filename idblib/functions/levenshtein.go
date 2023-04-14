/*
 * Copyright (c) 2023 Lucas Pape
 */

package functions

import (
	e "github.com/lucasl0st/InfiniteDB/errors"
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"github.com/lucasl0st/InfiniteDB/idblib/object"
	"github.com/lucasl0st/InfiniteDB/idblib/table"
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
	parameters map[string]interface{},
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
			str2 = table.Index.GetValue(l.fieldName, o).(dbtype.Text)
		}

		if additionalFields[o] == nil {
			additionalFields[o] = make(map[string]dbtype.DBType)
		}

		additionalFields[o][l.as] = dbtype.NumberFromFloat64(float64(l.levenshtein(str1, []rune(str2.ToString()))))
	}

	return objects, additionalFields, nil
}

func (l *LevenshteinFunction) parseParameters(parameters map[string]interface{}) error {
	v, ok := parameters["value"].(string)

	if !ok {
		return e.IsNotAString("value in levenshtein function")
	}

	fieldName, ok := parameters["fieldName"].(string)

	if !ok {
		return e.IsNotAString("fieldName in levenshtein function")
	}

	as, ok := parameters["as"].(string)

	if !ok {
		as = fieldNameLevenshtein
	}

	l.value = v
	l.fieldName = fieldName
	l.as = as

	return nil
}

func (l *LevenshteinFunction) levenshtein(str1, str2 []rune) int {
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
	return column[s1len]
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
