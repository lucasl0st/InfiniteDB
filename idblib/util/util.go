package util

import (
	e "github.com/lucasl0st/InfiniteDB/errors"
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"github.com/lucasl0st/InfiniteDB/idblib/field"
)

func DBTypeToInterface(v dbtype.DBType, f field.Field) (interface{}, error) {
	if v.IsNull() {
		return nil, nil
	}

	switch f.Type {
	case field.TEXT:
		return v.(dbtype.Text).ToString(), nil
	case field.NUMBER:
		return v.(dbtype.Number).ToFloat64(), nil
	case field.BOOL:
		return v.(dbtype.Bool).ToBool(), nil
	}

	return nil, e.UnknownDBTypeError()
}

func InterfaceToDBType(i interface{}, f field.Field) (dbtype.DBType, error) {
	switch f.Type {
	case field.TEXT:
		if i == nil {
			return dbtype.TextFromNull(), nil
		}

		s, ok := i.(string)

		if !ok {
			return nil, e.ValueIsNotText()
		}

		return dbtype.TextFromString(s), nil
	case field.NUMBER:
		if i == nil {
			return dbtype.NumberFromNull(), nil
		}

		n, ok := i.(float64)

		if !ok {
			return nil, e.ValueIsNotNumber()
		}

		return dbtype.NumberFromFloat64(n), nil
	case field.BOOL:
		if i == nil {
			return dbtype.BoolFromNull(), nil
		}

		b, ok := i.(bool)

		if !ok {
			return nil, e.ValueIsNotBool()
		}

		return dbtype.BoolFromBool(b), nil
	}

	return nil, e.UnknownDBTypeError()
}

func StringToDBType(s string, f field.Field) (dbtype.DBType, error) {
	switch f.Type {
	case field.TEXT:
		return dbtype.TextFromString(s), nil
	case field.NUMBER:
		return dbtype.NumberFromString(s)
	case field.BOOL:
		return dbtype.BoolFromString(s), nil
	}

	return nil, e.UnknownDBTypeError()
}
