/*
 * Copyright (c) 2023 Lucas Pape
 */

package main

import (
	"encoding/json"
	"github.com/lucasl0st/InfiniteDB/idblib/field"
	"github.com/lucasl0st/InfiniteDB/idblib/functions"
	"github.com/lucasl0st/InfiniteDB/idblib/table"
	e "github.com/lucasl0st/InfiniteDB/models/errors"
	"github.com/lucasl0st/InfiniteDB/models/request"
)

func parseRequest(r request.Request) (*table.Request, error) {
	var err error

	var q *table.Query

	if r.Query != nil {
		q, err = parseQuery(*r.Query)

		if err != nil {
			return nil, err
		}
	}

	return &table.Request{
		Query:     q,
		Sort:      r.Sort,
		Implement: r.Implement,
		Skip:      r.Skip,
		Limit:     r.Limit,
	}, nil
}

func parseQuery(q request.Query) (*table.Query, error) {
	f, err := parseFunctions(q.Functions)

	if err != nil {
		return nil, err
	}

	var and *table.Query = nil

	if q.And != nil {
		and, err = parseQuery(*q.And)

		if err != nil {
			return nil, err
		}
	}

	var or *table.Query = nil

	if q.Or != nil {
		or, err = parseQuery(*q.Or)

		if err != nil {
			return nil, err
		}
	}

	if and != nil && or != nil {
		return nil, e.CannotHaveAndANDOrInOneQuery()
	}

	query := table.Query{
		Where:     q.Where,
		Functions: f,
		And:       and,
		Or:        or,
	}

	if query.Where.Value != nil && (query.Where.All != nil || query.Where.Any != nil) ||
		query.Where.All != nil && (query.Where.Value != nil || query.Where.Any != nil) ||
		query.Where.Any != nil && (query.Where.Value != nil || query.Where.All != nil) {
		return nil, e.OnlyValueAllOrAny()
	}

	return &query, nil
}

func parseFunctions(f []request.Function) ([]table.FunctionWithParameters, error) {
	var results []table.FunctionWithParameters

	for _, function := range f {
		ff := table.FunctionWithParameters{}

		if function.Parameters != nil {
			ff.Parameters = *function.Parameters
		} else {
			ff.Parameters = make(map[string]json.RawMessage)
		}

		switch function.Function {
		case "levenshtein":
			ff.Function = &functions.LevenshteinFunction{}
		case "max":
			ff.Function = &functions.MinMaxFunction{Max: true}
		case "min":
			ff.Function = &functions.MinMaxFunction{Max: false}
		case "distance":
			ff.Function = &functions.DistanceFunction{}
		default:
			return nil, e.NotAValidFunction()
		}

		results = append(results, ff)
	}

	return results, nil
}

func parseFields(fields map[string]request.Field) (map[string]field.Field, error) {
	resultMap := make(map[string]field.Field)

	for fieldName, f := range fields {
		var t *field.DatabaseType
		indexed := false
		unique := false
		null := false

		t = field.ParseDatabaseType(f.Type)

		if t == nil {
			return nil, e.TypeNotSupported(f.Type)
		}

		if f.Indexed != nil {
			indexed = *f.Indexed
		}

		if f.Unique != nil {
			if *f.Unique && !indexed {
				return nil, e.FieldCannotBeUniqueWithoutBeingIndexed()
			}

			unique = *f.Unique
		}

		if f.Null != nil {
			null = *f.Null
		}

		resultMap[fieldName] = field.Field{
			Name:    fieldName,
			Indexed: indexed,
			Unique:  unique,
			Type:    *t,
			Null:    null,
		}
	}

	return resultMap, nil
}
