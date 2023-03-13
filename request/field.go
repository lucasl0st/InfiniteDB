/*
 * Copyright (c) 2023 Lucas Pape
 */

package request

type Field struct {
	Type    string `json:"type"`
	Indexed *bool  `json:"indexed"`
	Unique  *bool  `json:"unique"`
	Null    *bool  `json:"null"`
}
