/*
 * Copyright (c) 2023 Lucas Pape
 */

package websocket

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	e "github.com/lucasl0st/InfiniteDB/models/errors"
	"github.com/lucasl0st/InfiniteDB/models/method"
	models "github.com/lucasl0st/InfiniteDB/models/request"
	"github.com/lucasl0st/InfiniteDB/models/response"
	"github.com/lucasl0st/InfiniteDB/server/parse"
	"github.com/lucasl0st/InfiniteDB/server/util"
)

var MethodHandlers []MethodHandler

type Handler func(a *Api, conn *websocket.Conn, request map[string]interface{}, rawRequest map[string]json.RawMessage) (any, error)

type MethodHandler struct {
	Method  method.ServerMethod
	Handler Handler
}

func init() {
	registerHandler(method.ShutdownMethod, shutdownHandler)
	registerHandler(method.GetDatabasesMethod, getDatabasesHandler)
	registerHandler(method.CreateDatabaseMethod, createDatabaseHandler)
	registerHandler(method.DeleteDatabaseMethod, deleteDatabaseHandler)
	registerHandler(method.GetDatabaseMethod, getDatabaseHandler)
	registerHandler(method.GetDatabaseTableMethod, getDatabaseTableHandler)
	registerHandler(method.CreateTableInDatabaseMethod, createTableInDatabaseHandler)
	registerHandler(method.DeleteTableInDatabaseMethod, deleteTableInDatabaseHandler)
	registerHandler(method.GetFromDatabaseTableMethod, getFromDatabaseTableHandler)
	registerHandler(method.InsertToDatabaseTableMethod, insertToDatabaseTableHandler)
	registerHandler(method.RemoveFromDatabaseTableMethod, removeFromDatabaseTableHandler)
	registerHandler(method.UpdateInDatabaseTableMethod, updateInDatabaseTableHandler)
	registerHandler(method.SubscribeToMetricUpdates, subscribeToMetricUpdates)
	registerHandler(method.UnsubscribeFromMetricUpdates, unsubscribeFromMetricUpdates)
}

func registerHandler(m method.ServerMethod, handler Handler) {
	MethodHandlers = append(MethodHandlers, struct {
		Method  method.ServerMethod
		Handler Handler
	}{
		Method:  m,
		Handler: handler,
	})
}

func getString(request map[string]interface{}, key string) (string, error) {
	name, isString := request[key].(string)

	if !isString {
		return "", e.IsNotAString(key)
	}

	err := util.ValidateName(name)

	if err != nil {
		return "", err
	}

	return name, nil
}

func getDatabaseName(request map[string]interface{}) (string, error) {
	return getString(request, "name")
}

func getTableName(request map[string]interface{}) (string, error) {
	return getString(request, "tableName")
}

func shutdownHandler(a *Api, _ *websocket.Conn, _ map[string]interface{}, _ map[string]json.RawMessage) (any, error) {
	a.shutdown()
	return nil, nil
}

func getDatabasesHandler(a *Api, _ *websocket.Conn, _ map[string]interface{}, _ map[string]json.RawMessage) (any, error) {
	return a.idb.GetDatabases()
}

func createDatabaseHandler(a *Api, _ *websocket.Conn, request map[string]interface{}, _ map[string]json.RawMessage) (any, error) {
	name, err := getDatabaseName(request)

	if err != nil {
		return nil, err
	}

	return a.idb.CreateDatabase(name)
}

func deleteDatabaseHandler(a *Api, _ *websocket.Conn, request map[string]interface{}, _ map[string]json.RawMessage) (any, error) {
	name, err := getDatabaseName(request)

	if err != nil {
		return nil, err
	}

	return a.idb.DeleteDatabase(name)
}

func getDatabaseHandler(a *Api, _ *websocket.Conn, request map[string]interface{}, _ map[string]json.RawMessage) (any, error) {
	name, err := getDatabaseName(request)

	if err != nil {
		return nil, err
	}

	return a.idb.GetDatabase(name)
}

func getDatabaseTableHandler(a *Api, _ *websocket.Conn, request map[string]interface{}, _ map[string]json.RawMessage) (any, error) {
	name, err := getDatabaseName(request)

	if err != nil {
		return nil, err
	}

	tableName, err := getTableName(request)

	if err != nil {
		return nil, err
	}

	return a.idb.GetDatabaseTable(name, tableName)
}

func createTableInDatabaseHandler(a *Api, _ *websocket.Conn, request map[string]interface{}, _ map[string]json.RawMessage) (any, error) {
	name, err := getDatabaseName(request)

	if err != nil {
		return nil, err
	}

	tableName, err := getTableName(request)

	if err != nil {
		return nil, err
	}

	fields, isMap := request["fields"].(map[string]interface{})

	if !isMap {
		return nil, e.IsNotAMap("fields")
	}

	var o models.TableOptions

	options, isMap := request["options"].(map[string]interface{})

	if !isMap {
		o = models.TableOptions{}
	} else {
		err = util.ToStruct(options, &o)

		if err != nil {
			return nil, err
		}
	}

	var f map[string]models.Field
	err = util.ToStruct(fields, &f)

	if err != nil {
		return nil, err
	}

	parsedFields, err := parse.Fields(f)

	if err != nil {
		return nil, err
	}

	return a.idb.CreateTableInDatabase(name, tableName, parsedFields, o)
}

func deleteTableInDatabaseHandler(a *Api, _ *websocket.Conn, request map[string]interface{}, _ map[string]json.RawMessage) (any, error) {
	name, err := getDatabaseName(request)

	if err != nil {
		return nil, err
	}

	tableName, err := getTableName(request)

	if err != nil {
		return nil, err
	}

	return a.idb.DeleteTableInDatabase(name, tableName)
}

func getFromDatabaseTableHandler(a *Api, _ *websocket.Conn, request map[string]interface{}, _ map[string]json.RawMessage) (any, error) {
	name, err := getDatabaseName(request)

	if err != nil {
		return nil, err
	}

	tableName, err := getTableName(request)

	if err != nil {
		return nil, err
	}

	var req models.Request
	err = util.ToStruct(request["request"], &req)

	if err != nil {
		return nil, err
	}

	parsedRequest, err := parse.Request(req)

	if err != nil {
		return nil, err
	}

	return a.idb.GetFromDatabaseTable(name, tableName, *parsedRequest)
}

func insertToDatabaseTableHandler(a *Api, _ *websocket.Conn, request map[string]interface{}, _ map[string]json.RawMessage) (any, error) {
	name, err := getDatabaseName(request)

	if err != nil {
		return nil, err
	}

	tableName, err := getTableName(request)

	if err != nil {
		return nil, err
	}

	var o map[string]json.RawMessage
	err = util.ToStruct(request["object"], &o)

	if err != nil {
		return nil, err
	}

	return a.idb.InsertToDatabaseTable(name, tableName, o)
}

func removeFromDatabaseTableHandler(a *Api, _ *websocket.Conn, request map[string]interface{}, _ map[string]json.RawMessage) (any, error) {
	name, err := getDatabaseName(request)

	if err != nil {
		return nil, err
	}

	tableName, err := getTableName(request)

	if err != nil {
		return nil, err
	}

	var req models.Request
	err = util.ToStruct(request["request"], &req)

	if err != nil {
		return nil, err
	}

	parsedRequest, err := parse.Request(req)

	if err != nil {
		return nil, err
	}

	return a.idb.RemoveFromDatabaseTable(name, tableName, *parsedRequest)
}

func updateInDatabaseTableHandler(a *Api, _ *websocket.Conn, request map[string]interface{}, _ map[string]json.RawMessage) (any, error) {
	name, err := getDatabaseName(request)

	if err != nil {
		return nil, err
	}

	tableName, err := getTableName(request)

	if err != nil {
		return nil, err
	}

	var o map[string]json.RawMessage
	err = util.ToStruct(request["object"], &o)

	if err != nil {
		return nil, err
	}

	return a.idb.UpdateInDatabaseTable(name, tableName, o)
}

func subscribeToMetricUpdates(a *Api, conn *websocket.Conn, _ map[string]interface{}, _ map[string]json.RawMessage) (any, error) {
	a.subscribedToMetricUpdates = append(a.subscribedToMetricUpdates, conn)

	return response.SubscribeToMetricUpdatesResponse{}, nil
}

func unsubscribeFromMetricUpdates(a *Api, conn *websocket.Conn, _ map[string]interface{}, _ map[string]json.RawMessage) (any, error) {
	removedSubscribed := a.subscribedToMetricUpdates

	for _, subscribedConn := range a.subscribedToMetricUpdates {
		if subscribedConn != conn {
			removedSubscribed = append(removedSubscribed, subscribedConn)
		}
	}

	a.subscribedToMetricUpdates = removedSubscribed

	return response.UnsubscribedFromMetricUpdatesResponse{}, nil
}
