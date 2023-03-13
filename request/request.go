/*
 * Copyright (c) 2023 Lucas Pape
 */

package request

type Request struct {
	Query     *Query      `json:"query"`
	Sort      *Sort       `json:"sort"`
	Implement []Implement `json:"implement"`
	Skip      *int64      `json:"skip"`
	Limit     *int64      `json:"limit"`
}
