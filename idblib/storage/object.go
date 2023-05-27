/*
 * Copyright (c) 2023 Lucas Pape
 */

package storage

type object struct {
	LineNumber int64             `json:"line_number"`
	Object     map[string]string `json:"object,omitempty"`
	RefersTo   *int64            `json:"refersTo,omitempty"`
	Deleted    *bool             `json:"deleted"`
}
