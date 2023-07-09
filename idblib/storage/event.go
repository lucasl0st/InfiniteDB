/*
 * Copyright (c) 2023 Lucas Pape
 */

package storage

type Event struct {
	Type     EventType         `json:"type"`
	Data     map[string]string `json:"data,omitempty"`
	RefersTo *int64            `json:"refersTo,omitempty"`
}

type EventType string

const (
	EventTypeAdd    EventType = "ADD"
	EventTypeUpdate EventType = "UPDATE"
	EventTypeRemove EventType = "REMOVE"
)
