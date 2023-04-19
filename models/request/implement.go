/*
 * Copyright (c) 2023 Lucas Pape
 */

package request

type Implement struct {
	From       ImplementFrom `json:"from"`
	Field      string        `json:"field"`
	As         *string       `json:"as"`
	ForceArray *bool         `json:"forceArray"`
}

type ImplementFrom struct {
	Table string `json:"table"`
	Field string `json:"field"`
}
