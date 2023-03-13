/*
 * Copyright (c) 2023 Lucas Pape
 */

package request

type Function struct {
	Function   string                  `json:"function"`
	Parameters *map[string]interface{} `json:"parameters"`
}
