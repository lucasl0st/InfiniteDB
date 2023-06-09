/*
 * Copyright (c) 2023 Lucas Pape
 */

package table

import (
	"encoding/json"
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"github.com/lucasl0st/InfiniteDB/idblib/object"
)

type AdditionalFields map[int64]map[string]dbtype.DBType

type Function interface {
	Run(
		t *Table,
		objects object.Objects,
		additionalFields AdditionalFields,
		parameters map[string]json.RawMessage,
	) (object.Objects, AdditionalFields, error)
}
