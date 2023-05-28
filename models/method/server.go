/*
 * Copyright (c) 2023 Lucas Pape
 */

package method

type ServerMethod string

const ShutdownMethod ServerMethod = "shutdown"
const GetDatabasesMethod ServerMethod = "getDatabases"
const CreateDatabaseMethod ServerMethod = "createDatabase"
const DeleteDatabaseMethod ServerMethod = "deleteDatabase"
const GetDatabaseMethod ServerMethod = "GetDatabaseMethod"
const GetDatabaseTableMethod ServerMethod = "getDatabaseTable"
const CreateTableInDatabaseMethod ServerMethod = "createTableInDatabase"
const DeleteTableInDatabaseMethod ServerMethod = "deleteTableInDatabase"
const GetFromDatabaseTableMethod ServerMethod = "getFromDatabaseTable"
const InsertToDatabaseTableMethod ServerMethod = "insertToDatabaseTable"
const RemoveFromDatabaseTableMethod ServerMethod = "removeFromDatabaseTable"
const UpdateInDatabaseTableMethod ServerMethod = "updateInDatabaseTable"
const SubscribeToMetricUpdates ServerMethod = "subscribeToMetricUpdates"
const UnsubscribeFromMetricUpdates ServerMethod = "unsubscribeFromMetricUpdates"
