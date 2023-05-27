/*
 * Copyright (c) 2023 Lucas Pape
 */

package field

import (
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"github.com/lucasl0st/InfiniteDB/models/request"
)

type Field struct {
	Name    string              `json:"name"`
	Indexed bool                `json:"indexed"`
	Unique  bool                `json:"unique"`
	Null    bool                `json:"null"`
	Type    dbtype.DatabaseType `json:"type"`
}

type TableConfig struct {
	Fields  map[string]Field     `json:"fields"`
	Options request.TableOptions `json:"options"`
}
