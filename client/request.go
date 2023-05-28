/*
 * Copyright (c) 2023 Lucas Pape
 */

package client

import (
	"encoding/json"
	e "github.com/lucasl0st/InfiniteDB/models/errors"
	"github.com/lucasl0st/InfiniteDB/models/method"
	"github.com/lucasl0st/InfiniteDB/models/request"
	"github.com/lucasl0st/InfiniteDB/models/response"
	"math/rand"
	"nhooyr.io/websocket"
)

func (c *Client) ShutdownServer() error {
	if !c.connected {
		return e.ClientNotConnected()
	}

	r := make(map[string]interface{})

	r["method"] = method.ShutdownMethod

	requestId := int64(float64(rand.Int()))

	r["requestId"] = requestId

	data, err := json.Marshal(r)

	if err != nil {
		return err
	}

	err = c.ws.Write(c.ctx, websocket.MessageText, data)

	if err != nil {
		return err
	}

	c.connected = false

	return nil
}

func (c *Client) GetDatabases() (response.GetDatabasesResponse, error) {
	r := make(map[string]interface{})

	r["method"] = method.GetDatabasesMethod

	res, err := c.sendRequest(r)

	if err != nil {
		return response.GetDatabasesResponse{}, err
	}

	var getDatabasesResponse response.GetDatabasesResponse

	err = mapToStruct(res, &getDatabasesResponse)

	if err != nil {
		return response.GetDatabasesResponse{}, err
	}

	return getDatabasesResponse, nil
}

func (c *Client) CreateDatabase(name string) (response.CreateDatabaseResponse, error) {
	r := make(map[string]interface{})

	r["method"] = method.CreateDatabaseMethod
	r["name"] = name

	res, err := c.sendRequest(r)

	if err != nil {
		return response.CreateDatabaseResponse{}, err
	}

	var createDatabaseResponse response.CreateDatabaseResponse

	err = mapToStruct(res, &createDatabaseResponse)

	if err != nil {
		return response.CreateDatabaseResponse{}, err
	}

	return createDatabaseResponse, nil
}

func (c *Client) DeleteDatabase(name string) (response.DeleteDatabaseResponse, error) {
	r := make(map[string]interface{})

	r["method"] = method.DeleteDatabaseMethod
	r["name"] = name

	res, err := c.sendRequest(r)

	if err != nil {
		return response.DeleteDatabaseResponse{}, err
	}

	var deleteDatabaseResponse response.DeleteDatabaseResponse

	err = mapToStruct(res, &deleteDatabaseResponse)

	if err != nil {
		return response.DeleteDatabaseResponse{}, err
	}

	return deleteDatabaseResponse, nil
}

func (c *Client) GetDatabase(name string) (response.GetDatabaseResponse, error) {
	r := make(map[string]interface{})

	r["method"] = method.GetDatabaseMethod
	r["name"] = name

	res, err := c.sendRequest(r)

	if err != nil {
		return response.GetDatabaseResponse{}, err
	}

	var getDatabaseResponse response.GetDatabaseResponse

	err = mapToStruct(res, &getDatabaseResponse)

	if err != nil {
		return response.GetDatabaseResponse{}, err
	}

	return getDatabaseResponse, nil
}

func (c *Client) GetDatabaseTable(name string, tableName string) (response.GetDatabaseTableResponse, error) {
	r := make(map[string]interface{})

	r["method"] = method.GetDatabaseTableMethod
	r["name"] = name
	r["tableName"] = tableName

	res, err := c.sendRequest(r)

	if err != nil {
		return response.GetDatabaseTableResponse{}, err
	}

	var getDatabaseTablesResponse response.GetDatabaseTableResponse

	err = mapToStruct(res, &getDatabaseTablesResponse)

	if err != nil {
		return response.GetDatabaseTableResponse{}, err
	}

	return getDatabaseTablesResponse, nil
}

func (c *Client) CreateTableInDatabase(name string, tableName string, fields map[string]request.Field, options *request.TableOptions) (response.CreateTableInDatabaseResponse, error) {
	r := make(map[string]interface{})

	r["method"] = method.CreateTableInDatabaseMethod
	r["name"] = name
	r["tableName"] = tableName
	r["fields"] = fields
	r["options"] = options

	res, err := c.sendRequest(r)

	if err != nil {
		return response.CreateTableInDatabaseResponse{}, err
	}

	var createTableInDatabaseResponse response.CreateTableInDatabaseResponse

	err = mapToStruct(res, &createTableInDatabaseResponse)

	if err != nil {
		return response.CreateTableInDatabaseResponse{}, err
	}

	return createTableInDatabaseResponse, nil
}

func (c *Client) DeleteTableInDatabase(name string, tableName string) (response.DeleteTableInDatabaseResponse, error) {
	r := make(map[string]interface{})

	r["method"] = method.DeleteTableInDatabaseMethod
	r["name"] = name
	r["tableName"] = tableName

	res, err := c.sendRequest(r)

	if err != nil {
		return response.DeleteTableInDatabaseResponse{}, err
	}

	var deleteTableInDatabaseResponse response.DeleteTableInDatabaseResponse

	err = mapToStruct(res, &deleteTableInDatabaseResponse)

	if err != nil {
		return response.DeleteTableInDatabaseResponse{}, err
	}

	return deleteTableInDatabaseResponse, nil
}

func (c *Client) GetFromDatabaseTable(name string, tableName string, request request.Request) (response.GetFromDatabaseTableResponse, error) {
	r := make(map[string]interface{})

	r["method"] = method.GetFromDatabaseTableMethod
	r["name"] = name
	r["tableName"] = tableName
	r["request"] = request

	res, err := c.sendRequest(r)

	if err != nil {
		return response.GetFromDatabaseTableResponse{}, err
	}

	var getFromDatabaseTableResponse response.GetFromDatabaseTableResponse

	err = mapToStruct(res, &getFromDatabaseTableResponse)

	if err != nil {
		return response.GetFromDatabaseTableResponse{}, err
	}

	return getFromDatabaseTableResponse, nil
}

func (c *Client) InsertToDatabaseTable(name string, tableName string, object map[string]json.RawMessage) (response.InsertToDatabaseTableResponse, error) {
	r := make(map[string]interface{})

	r["method"] = method.InsertToDatabaseTableMethod
	r["name"] = name
	r["tableName"] = tableName
	r["object"] = object

	res, err := c.sendRequest(r)

	if err != nil {
		return response.InsertToDatabaseTableResponse{}, err
	}

	var insertToDatabaseTableResponse response.InsertToDatabaseTableResponse

	err = mapToStruct(res, &insertToDatabaseTableResponse)

	if err != nil {
		return response.InsertToDatabaseTableResponse{}, err
	}

	return insertToDatabaseTableResponse, nil
}

func (c *Client) RemoveFromDatabaseTable(name string, tableName string, request request.Request) (response.RemoveFromDatabaseTableResponse, error) {
	r := make(map[string]interface{})

	r["method"] = method.RemoveFromDatabaseTableMethod
	r["name"] = name
	r["tableName"] = tableName
	r["request"] = request

	res, err := c.sendRequest(r)

	if err != nil {
		return response.RemoveFromDatabaseTableResponse{}, err
	}

	var removeFromDatabaseTableResponse response.RemoveFromDatabaseTableResponse

	err = mapToStruct(res, &removeFromDatabaseTableResponse)

	if err != nil {
		return response.RemoveFromDatabaseTableResponse{}, err
	}

	return removeFromDatabaseTableResponse, nil
}

func (c *Client) UpdateInDatabaseTable(name string, tableName string, object map[string]interface{}) (response.UpdateInDatabaseTableResponse, error) {
	r := make(map[string]interface{})

	r["method"] = method.UpdateInDatabaseTableMethod
	r["name"] = name
	r["tableName"] = tableName
	r["object"] = object

	res, err := c.sendRequest(r)

	if err != nil {
		return response.UpdateInDatabaseTableResponse{}, err
	}

	var updateInDatabaseTableResponse response.UpdateInDatabaseTableResponse

	err = mapToStruct(res, &updateInDatabaseTableResponse)

	if err != nil {
		return response.UpdateInDatabaseTableResponse{}, err
	}

	return updateInDatabaseTableResponse, nil
}

func (c *Client) SubscribeToMetricUpdates() (response.SubscribeToMetricUpdatesResponse, error) {
	r := make(map[string]interface{})

	r["method"] = method.SubscribeToMetricUpdates

	res, err := c.sendRequest(r)

	if err != nil {
		return response.SubscribeToMetricUpdatesResponse{}, err
	}

	var subscribeToMetricUpdatesResponse response.SubscribeToMetricUpdatesResponse

	err = mapToStruct(res, &subscribeToMetricUpdatesResponse)

	if err != nil {
		return response.SubscribeToMetricUpdatesResponse{}, err
	}

	return subscribeToMetricUpdatesResponse, nil
}

func (c *Client) UnsubscribeFromMetricUpdates() (response.UnsubscribedFromMetricUpdatesResponse, error) {
	r := make(map[string]interface{})

	r["method"] = method.UnsubscribeFromMetricUpdates

	res, err := c.sendRequest(r)

	if err != nil {
		return response.UnsubscribedFromMetricUpdatesResponse{}, err
	}

	var unsubscribeFromMetricUpdatesResponse response.UnsubscribedFromMetricUpdatesResponse

	err = mapToStruct(res, &unsubscribeFromMetricUpdatesResponse)

	if err != nil {
		return response.UnsubscribedFromMetricUpdatesResponse{}, err
	}

	return unsubscribeFromMetricUpdatesResponse, nil
}
