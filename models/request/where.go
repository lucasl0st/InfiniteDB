/*
 * Copyright (c) 2023 Lucas Pape
 */

package request

import "encoding/json"

type Where struct {
	Field    string            `json:"field"`
	Operator Operator          `json:"operator"`
	Value    json.RawMessage   `json:"value"`
	All      []json.RawMessage `json:"all"`
	Any      []json.RawMessage `json:"any"`
}
