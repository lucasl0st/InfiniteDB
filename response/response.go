/*
 * Copyright (c) 2023 Lucas Pape
 */

package response

import (
	"github.com/lucasl0st/InfiniteDB/request"
)

type GetDatabasesResponse struct {
	Databases []string
}

type CreateDatabaseResponse struct {
	Message string
	Name    string
}

type DeleteDatabaseResponse struct {
	Message string
	Name    string
}

type GetDatabaseTablesResponse struct {
	Name   string
	Tables []GetDatabaseTablesResponseTable
}

type GetDatabaseTablesResponseTable struct {
	Name    string
	Fields  []request.Field
	Options request.TableOptions
}

type CreateTableInDatabaseResponse struct {
	Name      string
	TableName string
	Fields    map[string]request.Field
}

type DeleteTableInDatabaseResponse struct {
	Message   string
	Name      string
	TableName string
}

type GetFromDatabaseTableResponse struct {
	Name      string
	TableName string
	Results   []map[string]interface{}
}

type InsertToDatabaseTableResponse struct {
	Name      string
	TableName string
	Object    map[string]interface{}
}

type RemoveFromDatabaseTableResponse struct {
	Name      string
	TableName string
	Removed   int64
}

type UpdateInDatabaseTableResponse struct {
	Name      string
	TableName string
	Object    map[string]interface{}
}

type GetDatabaseResponse struct {
	Name string
}
