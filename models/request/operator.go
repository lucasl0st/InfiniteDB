/*
 * Copyright (c) 2023 Lucas Pape
 */

package request

type Operator string

const (
	EQUALS  Operator = "="
	NOT     Operator = "!="
	MATCH   Operator = "match"
	LARGER  Operator = ">"
	SMALLER Operator = "<"
	BETWEEN Operator = "><"
)
