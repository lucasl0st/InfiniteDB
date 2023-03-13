/*
 * Copyright (c) 2023 Lucas Pape
 */

package request

type Sort struct {
	Field     string
	Direction SortDirection
}

type SortDirection string

const (
	ASC  SortDirection = "asc"
	DESC SortDirection = "desc"
)
