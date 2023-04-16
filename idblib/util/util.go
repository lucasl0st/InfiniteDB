package util

import (
	"encoding/json"
	"errors"
	"fmt"
	e "github.com/lucasl0st/InfiniteDB/errors"
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"github.com/lucasl0st/InfiniteDB/idblib/field"
	"strconv"
)

func JsonRawToDBType(j json.RawMessage, f field.Field) (dbtype.DBType, error) {
	switch f.Type {
	case field.TEXT:
		s, err := JsonRawToString(j)

		if err != nil {
			return nil, err
		}

		if s == nil {
			return dbtype.TextFromNull(), nil
		}

		return dbtype.TextFromString(*s), nil
	case field.NUMBER:
		s, err := JsonRawToStringNumber(j)

		if err != nil {
			return nil, err
		}

		if s == nil {
			return dbtype.NumberFromNull(), nil
		}

		return dbtype.NumberFromString(*s)
	case field.BOOL:
		s, err := JsonRawToStringBool(j)

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

func JsonRawToString(j json.RawMessage) (*string, error) {
	s := string(j)

	if s == "null" {
		return nil, nil
	}

	if len(s) > 0 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
		return &s, nil
	}

	return nil, errors.New("is not a string")
}

func JsonRawToStringNumber(j json.RawMessage) (*string, error) {
	s := string(j)

	if s == "null" {
		return nil, nil
	}

	_, err := strconv.ParseFloat(s, 64)

	if err != nil {
		return nil, err
	}

	return &s, nil
}

func JsonRawToStringBool(j json.RawMessage) (*string, error) {
	s := string(j)

	if s == "null" {
		return nil, nil
	}

	if s == "true" || s == "false" {
		return &s, nil
	} else {
		return nil, errors.New("is not a boolean")
	}
}

func InterfaceToJsonRaw(i interface{}) json.RawMessage {
	b, err := json.Marshal(i)

	if err != nil {
		panic(err.Error())
	}

	var j json.RawMessage

	err = json.Unmarshal(b, &j)

	if err != nil {
		panic(err.Error())
	}

	return j
}

func StringToJsonRaw(s string) json.RawMessage {
	return json.RawMessage(fmt.Sprintf("\"%s\"", s))
}

func Float64ToJsonRaw(f float64) json.RawMessage {
	return json.RawMessage(fmt.Sprintf("%v", f))
}

func Int64ToJsonRaw(i int64) json.RawMessage {
	return json.RawMessage(fmt.Sprintf("%v", i))
}

func BoolToJsonRaw(b bool) json.RawMessage {
	if b {
		return json.RawMessage("true")
	} else {
		return json.RawMessage("false")
	}
}

func JsonRawMapToInterfaceMap(m map[string]json.RawMessage) map[string]interface{} {
	var r map[string]interface{}

	b, err := json.Marshal(m)

	if err != nil {
		panic(err.Error())
	}

	err = json.Unmarshal(b, &r)

	if err != nil {
		panic(err.Error())
	}

	return r
}

func InterfaceMapToJsonRawMap(m map[string]interface{}) map[string]json.RawMessage {
	r := map[string]json.RawMessage{}

	for key, value := range m {
		r[key] = InterfaceToJsonRaw(value)
	}

	return r
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
