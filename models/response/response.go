/*
 * Copyright (c) 2023 Lucas Pape
 */

package response

import (
	"encoding/json"
	request2 "github.com/lucasl0st/InfiniteDB/models/request"
)

type GetDatabasesResponse struct {
	Databases []string `json:"databases"`
}

type CreateDatabaseResponse struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

type DeleteDatabaseResponse struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

type GetDatabaseResponse struct {
	Name   string   `json:"name"`
	Tables []string `json:"tables"`
}

type GetDatabaseTableResponse struct {
	Name      string                    `json:"name"`
	TableName string                    `json:"tableName"`
	Fields    map[string]request2.Field `json:"fields"`
	Options   request2.TableOptions     `json:"options"`
}

type CreateTableInDatabaseResponse struct {
	Name      string                    `json:"name"`
	TableName string                    `json:"tableName"`
	Fields    map[string]request2.Field `json:"fields"`
}

type DeleteTableInDatabaseResponse struct {
	Name      string `json:"name"`
	TableName string `json:"tableName"`
	Message   string `json:"message"`
}

type GetFromDatabaseTableResponse struct {
	Name      string                       `json:"name"`
	TableName string                       `json:"tableName"`
	Results   []map[string]json.RawMessage `json:"results"`
}

type InsertToDatabaseTableResponse struct {
	Name      string                     `json:"name"`
	TableName string                     `json:"tableName"`
	Object    map[string]json.RawMessage `json:"object"`
}

type RemoveFromDatabaseTableResponse struct {
	Name      string `json:"name"`
	TableName string `json:"tableName"`
	Removed   int64  `json:"removed"`
}

type UpdateInDatabaseTableResponse struct {
	Name      string                     `json:"name"`
	TableName string                     `json:"tableName"`
	Object    map[string]json.RawMessage `json:"object"`
}

type SubscribeToMetricUpdatesResponse struct {
}

type UnsubscribedFromMetricUpdatesResponse struct {
}

type MetricUpdateResponse struct {
	Metric string `json:"metric"`
	Value  any    `json:"value"`
}
