/*
 * Copyright (c) 2023 Lucas Pape
 */

package field

import (
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"github.com/lucasl0st/InfiniteDB/idblib/object"
	"testing"
)

var testFields = map[string]Field{
	"textField": {
		Name:    "textField",
		Indexed: true,
	},
	"numberField": {
		Name:    "numberField",
		Indexed: true,
	},
	"booleanField": {
		Name:    "booleanField",
		Indexed: true,
	},
}

func TestIndex_Equal(t *testing.T) {
	tests := []object.Object{
		{M: map[string]dbtype.DBType{
			"textField":    dbtype.TextFromString("hello"),
			"numberField":  dbtype.NumberFromNull(),
			"booleanField": dbtype.BoolFromBool(true),
		}, Id: 0},

		{M: map[string]dbtype.DBType{
			"textField":    dbtype.TextFromString("hello"),
			"numberField":  dbtype.NumberFromNull(),
			"booleanField": dbtype.BoolFromBool(true),
		}, Id: 1},

		{M: map[string]dbtype.DBType{
			"textField":    dbtype.TextFromString("hello"),
			"numberField":  dbtype.NumberFromNull(),
			"booleanField": dbtype.BoolFromBool(true),
		}, Id: 2},

		{M: map[string]dbtype.DBType{
			"textField":    dbtype.TextFromString("hello"),
			"numberField":  dbtype.NumberFromNull(),
			"booleanField": dbtype.BoolFromBool(true),
		}, Id: 3},

		{M: map[string]dbtype.DBType{
			"textField":    dbtype.TextFromString("hello"),
			"numberField":  dbtype.NumberFromNull(),
			"booleanField": dbtype.BoolFromBool(true),
		}, Id: 4},
	}

	index := NewIndex(testFields)

	for _, test := range tests {
		index.Index(test)
	}

	for _, test := range tests {
		for field := range test.M {
			ids := index.Equal(field, test.M[field])
			extra := len(ids)

			for _, test2 := range tests {
				if test.M[field].Equal(test2.M[field]) {
					found := false

					for _, id := range ids {
						if test2.Id == id {
							found = true
							extra--
							break
						}
					}

					if !found {
						t.Errorf("exected to find %v", test2.Id)
					}
				}
			}

			if extra > 0 {
				t.Errorf("found %v extra, expected 0", extra)
			}
		}
	}
}
