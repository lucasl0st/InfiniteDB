/*
 * Copyright (c) 2023 Lucas Pape
 */

package object

import (
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
)

type Object struct {
	Id int64
	M  map[string]dbtype.DBType
}
