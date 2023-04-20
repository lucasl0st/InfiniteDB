/*
 * Copyright (c) 2023 Lucas Pape
 */

package util

import (
	"github.com/lucasl0st/InfiniteDB/models/request"
)

type Table struct {
	Name    string                   `json:"name"`
	Fields  map[string]request.Field ` json:"fields"`
	Options request.TableOptions     `json:"options"`
}
