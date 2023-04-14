/*
 * Copyright (c) 2023 Lucas Pape
 */

package functions

import (
	e "github.com/lucasl0st/InfiniteDB/errors"
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"github.com/lucasl0st/InfiniteDB/idblib/field"
	"github.com/lucasl0st/InfiniteDB/idblib/object"
	"github.com/lucasl0st/InfiniteDB/idblib/table"
)

const fieldNameMax = "max"
const fieldNameMin = "min"

type MinMaxFunction struct {
	Max       bool
	fieldName string
	as        string
}

func (m *MinMaxFunction) Run(
	t *table.Table,
	objects object.Objects,
	additionalFields table.AdditionalFields,
	parameters map[string]interface{},
) (object.Objects, table.AdditionalFields, error) {
	err := m.parseParameters(t, parameters)

	if err != nil {
		return nil, nil, err
	}

	var r float64
	r = 0

	for _, o := range objects {
		var v float64

		if additionalFields[o][m.fieldName] != nil {
			v = additionalFields[o][m.fieldName].(dbtype.Number).ToFloat64()
		} else {
			v = t.Index.GetValue(m.fieldName, o).(dbtype.Number).ToFloat64()
		}

		if m.Max {
			if v > r {
				r = v
			}
		} else {
			if v < r {
				r = v
			}
		}
	}

	for _, o := range objects {
		if additionalFields[o] == nil {
			additionalFields[o] = make(map[string]dbtype.DBType)
		}

		additionalFields[o][m.as] = dbtype.NumberFromFloat64(r)
	}

	return objects, additionalFields, nil
}

func (m *MinMaxFunction) parseParameters(t *table.Table, parameters map[string]interface{}) error {
	fieldName, ok := parameters["fieldName"].(string)

	if !ok {
		return e.IsNotAString("fieldName in min/max function")
	}

	f, ok := t.Config.Fields[fieldName]

	if !ok {
		return e.CannotFindField(fieldName)
	}

	if f.Type != field.NUMBER {
		return e.FieldHasUnsupportedTypeForThisFunction(fieldName)
	}

	as, ok := parameters["as"].(string)

	if !ok {
		if m.Max {
			as = fieldNameMax
		} else {
			as = fieldNameMin
		}
	}

	m.fieldName = fieldName
	m.as = as

	return nil
}
