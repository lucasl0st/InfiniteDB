/*
 * Copyright (c) 2023 Lucas Pape
 */

package request

type Sort struct {
	Field     string        `json:"field"`
	Direction SortDirection `json:"direction"`
}

type SortDirection string

const (
	ASC  SortDirection = "asc"
	DESC SortDirection = "desc"
)
