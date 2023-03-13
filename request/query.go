/*
 * Copyright (c) 2023 Lucas Pape
 */

package request

type Query struct {
	Where     *Where     `json:"where"`
	Functions []Function `json:"functions"`
	And       *Query     `json:"and"`
	Or        *Query     `json:"or"`
}
