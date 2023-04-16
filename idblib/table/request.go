/*
 * Copyright (c) 2023 Lucas Pape
 */

package table

import (
	"encoding/json"
	"github.com/lucasl0st/InfiniteDB/request"
)

type Request struct {
	Query     *Query
	Sort      *request.Sort
	Implement []request.Implement
	Skip      *int64
	Limit     *int64
}

type Query struct {
	Where     *request.Where
	Functions []FunctionWithParameters
	And       *Query
	Or        *Query
}

type FunctionWithParameters struct {
	Function   Function
	Parameters map[string]json.RawMessage
}
