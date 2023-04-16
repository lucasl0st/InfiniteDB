/*
 * Copyright (c) 2023 Lucas Pape
 */

package request

import "encoding/json"

type Function struct {
	Function   string                      `json:"function"`
	Parameters *map[string]json.RawMessage `json:"parameters"`
}
