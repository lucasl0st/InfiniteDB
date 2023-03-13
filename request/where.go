/*
 * Copyright (c) 2023 Lucas Pape
 */

package request

type Where struct {
	Field    string        `json:"field"`
	Operator Operator      `json:"operator"`
	Value    interface{}   `json:"value"`
	All      []interface{} `json:"all"`
	Any      []interface{} `json:"any"`
}
