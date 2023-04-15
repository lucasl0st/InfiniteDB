/*
 * Copyright (c) 2023 Lucas Pape
 */

package dump

import "github.com/lucasl0st/InfiniteDB/request"

type Table struct {
	Name    string               `json:"name"`
	Fields  []request.Field      ` json:"fields"`
	Options request.TableOptions `json:"options"`
}
