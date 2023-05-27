/*
 * Copyright (c) 2023 Lucas Pape
 */

package websocket

type Method string

const ShutdownMethod Method = "shutdown"
const GetDatabasesMethod Method = "getDatabases"
const CreateDatabaseMethod Method = "createDatabase"
const DeleteDatabaseMethod Method = "deleteDatabase"
const GetDatabaseMethod Method = "GetDatabaseMethod"
const GetDatabaseTableMethod Method = "getDatabaseTable"
const CreateTableInDatabaseMethod Method = "createTableInDatabase"
const DeleteTableInDatabaseMethod Method = "deleteTableInDatabase"
const GetFromDatabaseTableMethod Method = "getFromDatabaseTable"
const InsertToDatabaseTableMethod Method = "insertToDatabaseTable"
const RemoveFromDatabaseTableMethod Method = "removeFromDatabaseTable"
const UpdateInDatabaseTableMethod Method = "updateInDatabaseTable"
