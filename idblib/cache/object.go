/*
 * Copyright (c) 2023 Lucas Pape
 */

package cache

import "github.com/lucasl0st/InfiniteDB/idblib/dbtype"

type object struct {
	m        map[string]dbtype.DBType
	priority int64
}
