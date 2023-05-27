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
	parameters map[string]json.RawMessage,
) (object.Objects, table.AdditionalFields, error) {
	err := m.parseParameters(t, parameters)

	if err != nil {
		return nil, nil, err
	}

	var r dbtype.DBType

	for _, o := range objects {
		var v dbtype.DBType

		if additionalFields[o][m.fieldName] != nil {
			v = additionalFields[o][m.fieldName]
		} else {
			index, err := t.GetIndex(m.fieldName)

			if err != nil {
				return nil, nil, err
			}

			v = index.GetValue(o)
		}

		if r == nil {
			r = v
			continue
		}

		if m.Max {
			if v.Larger(r) {
				r = v
			}
		} else {
			if v.Smaller(r) {
				r = v
			}
		}
	}

	for _, o := range objects {
		if additionalFields[o] == nil {
			additionalFields[o] = make(map[string]dbtype.DBType)
		}

		additionalFields[o][m.as] = r
	}

	return objects, additionalFields, nil
}

func (m *MinMaxFunction) parseParameters(t *table.Table, parameters map[string]json.RawMessage) error {
	fieldName, err := util.JsonRawToString(parameters["fieldName"])

	if err != nil {
		return err
	}

	var as string

	if m.Max {
		as = fieldNameMax
	} else {
		as = fieldNameMin
	}

	asP, err := util.JsonRawToString(parameters["as"])

	if asP != nil && len(*asP) > 0 && err == nil {
		as = *asP
	}

	m.fieldName = *fieldName
	m.as = as

	return nil
}
