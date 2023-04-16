/*
 * Copyright (c) 2023 Lucas Pape
 */

package table

import (
	"encoding/json"
	"github.com/lucasl0st/InfiniteDB/idblib/object"
)

var QueryMiddleware func(table *Table, q Query) (bool, func(previousObjects object.Objects) (object.Objects, AdditionalFields, error))
var InsertMiddleware func(table *Table, objectM map[string]json.RawMessage) (bool, func() error)
var UpdateMiddleware func(table *Table, objectM map[string]json.RawMessage) (bool, func() error)
var RemoveMiddleware func(table *Table, object *object.Object) (bool, func() error)
var CreateDatabaseMiddleware func(name string) (bool, func() error)

func init() {
	QueryMiddleware = func(*Table, Query) (bool, func(previousObjects object.Objects) (object.Objects, AdditionalFields, error)) {
		return false, func(previousObjects object.Objects) (object.Objects, AdditionalFields, error) {
			return nil, nil, nil
		}
	}

	InsertMiddleware = func(table *Table, objectM map[string]json.RawMessage) (bool, func() error) {
		return false, func() error {
			return nil
		}
	}

	UpdateMiddleware = func(table *Table, objectM map[string]json.RawMessage) (bool, func() error) {
		return false, func() error {
			return nil
		}
	}

	RemoveMiddleware = func(table *Table, object *object.Object) (bool, func() error) {
		return false, func() error {
			return nil
		}
	}

	CreateDatabaseMiddleware = func(name string) (bool, func() error) {
		return false, func() error {
			return nil
		}
	}
}
