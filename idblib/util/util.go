package util

import (
	"encoding/json"
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"github.com/lucasl0st/InfiniteDB/idblib/field"
	e "github.com/lucasl0st/InfiniteDB/models/errors"
	"github.com/lucasl0st/InfiniteDB/util"
)

func JsonRawToDBType(j json.RawMessage, f field.Field) (dbtype.DBType, error) {
	switch f.Type {
	case dbtype.TEXT:
		s, err := util.JsonRawToString(j)

		if err != nil {
			return nil, err
		}

		if s == nil {
			return dbtype.TextFromNull(), nil
		}

		return dbtype.TextFromString(*s), nil
	case dbtype.NUMBER:
		s, err := util.JsonRawToStringNumber(j)

		if err != nil {
			return nil, err
		}

		if s == nil {
			return dbtype.NumberFromNull(), nil
		}

		return dbtype.NumberFromString(*s)
	case dbtype.BOOL:
		s, err := util.JsonRawToStringBool(j)

		if err != nil {
			return nil, err
		}

		if s == nil {
			return dbtype.BoolFromNull(), nil
		}

		return dbtype.BoolFromString(*s), nil
	}

	return nil, e.UnknownDBTypeError()
}

func StringToDBType(s string, f field.Field) (dbtype.DBType, error) {
	switch f.Type {
	case dbtype.TEXT:
		return dbtype.TextFromString(s), nil
	case dbtype.NUMBER:
		return dbtype.NumberFromString(s)
	case dbtype.BOOL:
		return dbtype.BoolFromString(s), nil
	}

	return nil, e.UnknownDBTypeError()
}
