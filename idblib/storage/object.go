/*
 * Copyright (c) 2023 Lucas Pape
 */

package storage

type object struct {
	Id       int64             `json:"id"`
	Object   map[string]string `json:"object,omitempty"`
	Deleted  *bool             `json:"deleted"`
	RefersTo *int64            `json:"refersTo,omitempty"`
}
