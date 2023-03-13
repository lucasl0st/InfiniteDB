/*
 * Copyright (c) 2023 Lucas Pape
 */

package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	e "github.com/lucasl0st/InfiniteDB/errors"
	"github.com/lucasl0st/InfiniteDB/request"
	"github.com/lucasl0st/InfiniteDB/response"
	"math/rand"
	"net/http"
	"nhooyr.io/websocket"
	"sync"
	"time"
)

const VERSION = "1.0"

type RequestResult struct {
	M   map[string]interface{}
	Err error
}

type Client struct {
	hostname      string
	port          uint
	ssl           bool
	skipSSLVerify bool
	authKey       *string
	timeout       time.Duration
	readLimit     int64
	connected     bool
	ws            *websocket.Conn
	ctx           context.Context

	channels sync.Map
}

func New(options Options) *Client {
	if options.SSL == nil {
		options.SSL = ptr(false)
	}

	if options.SkipSSLVerify == nil {
		options.SkipSSLVerify = ptr(false)
	}

	if options.Timeout == nil {
		options.Timeout = ptr(time.Second * 10)
	}

	if options.ReadLimit == nil {
		options.ReadLimit = ptr(int64(1024 * 1000 * 1000))
	}

	return &Client{
		hostname:      options.Hostname,
		port:          options.Port,
		ssl:           *options.SSL,
		skipSSLVerify: *options.SkipSSLVerify,
		authKey:       options.AuthKey,
		timeout:       *options.Timeout,
		readLimit:     *options.ReadLimit,
		connected:     false,
		ctx:           context.Background(),
		channels:      sync.Map{},
	}
}

func (c *Client) Connect() error {
	header := http.Header{}

	if c.authKey != nil {
		header.Add("Authorization", *c.authKey)
	}

	protocol := "ws"

	if c.ssl {
		protocol = "wss"
	}

	address := fmt.Sprintf("%s://%s:%v/ws", protocol, c.hostname, c.port)

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: c.skipSSLVerify,
			},
		},
	}

	c.channels.Store(int64(0), make(chan RequestResult))

	defer func() {
		close(c.getChannel(int64(0)))
		c.channels.Delete(int64(0))
	}()

	ws, _, err := websocket.Dial(c.ctx, address, &websocket.DialOptions{
		HTTPHeader: header,
		HTTPClient: httpClient,
	})

	if err != nil {
		return err
	}

	ws.SetReadLimit(c.readLimit)
	c.ws = ws

	go c.read()

	r, err := c.getResponse(int64(0))

	if err != nil {
		return err
	}

	version, isString := r["database_version"].(string)

	if !isString {
		return errors.New("did not receive database version")
	}

	if version != VERSION {
		return errors.New("this client is not compatible with the database version")
	}

	c.connected = true

	return nil
}

func (c *Client) read() {
	for {
		_, data, err := c.ws.Read(c.ctx)

		if err != nil {
			return
		}

		var msg map[string]interface{}

		err = json.Unmarshal(data, &msg)

		if err != nil {
			return
		}

		requestId := int64(msg["requestId"].(float64))

		status, ok := msg["status"].(float64)

		if ok {
			r := RequestResult{}

			if status == http.StatusOK {
				r.M = msg
			} else {
				r.Err = errors.New(msg["message"].(string))
			}

			if c.getChannel(requestId) != nil {
				c.getChannel(requestId) <- r
			} else {
				panic(r.Err)
			}
		}
	}
}

func (c *Client) sendRequest(request map[string]interface{}) (map[string]interface{}, error) {
	if !c.connected {
		return nil, errors.New("client is not connected")
	}

	requestId := int64(float64(rand.Int()))

	request["requestId"] = requestId

	data, err := json.Marshal(request)

	if err != nil {
		return nil, err
	}

	c.channels.Store(requestId, make(chan RequestResult))

	err = c.ws.Write(c.ctx, websocket.MessageText, data)

	if err != nil {
		return nil, err
	}

	defer func() {
		close(c.getChannel(requestId))
		c.channels.Delete(requestId)
	}()

	return c.getResponse(requestId)
}

func (c *Client) getResponse(requestId int64) (map[string]interface{}, error) {
	select {
	case res := <-c.getChannel(requestId):
		if res.Err != nil {
			return nil, res.Err
		}

		return res.M, nil
	case <-time.After(c.timeout):
		return nil, e.TimeoutReceivingDatabaseResult(requestId)
	}
}

func (c *Client) getChannel(requestId int64) chan RequestResult {
	channel, have := c.channels.Load(requestId)

	if !have {
		return nil
	}

	ch, ok := channel.(chan RequestResult)

	if !ok {
		return nil
	}

	return ch
}

func (c *Client) GetDatabases() (response.GetDatabasesResponse, error) {
	r := make(map[string]interface{})

	r["method"] = "getDatabases"

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

	r["method"] = "createDatabase"
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

	r["method"] = "deleteDatabase"
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

	r["method"] = "getDatabase"
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

func (c *Client) GetDatabaseTables(name string) (response.GetDatabaseTablesResponse, error) {
	r := make(map[string]interface{})

	r["method"] = "getDatabaseTables"
	r["name"] = name

	res, err := c.sendRequest(r)

	if err != nil {
		return response.GetDatabaseTablesResponse{}, err
	}

	var getDatabaseTablesResponse response.GetDatabaseTablesResponse

	err = mapToStruct(res, &getDatabaseTablesResponse)

	if err != nil {
		return response.GetDatabaseTablesResponse{}, err
	}

	return getDatabaseTablesResponse, nil
}

func (c *Client) CreateTableInDatabase(name string, tableName string, fields map[string]request.Field, options *request.TableOptions) (response.CreateTableInDatabaseResponse, error) {
	r := make(map[string]interface{})

	r["method"] = "createTableInDatabase"
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

	r["method"] = "deleteTableInDatabase"
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

	r["method"] = "getFromDatabaseTable"
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

func (c *Client) InsertToDatabaseTable(name string, tableName string, object map[string]interface{}) (response.InsertToDatabaseTableResponse, error) {
	r := make(map[string]interface{})

	r["method"] = "insertToDatabaseTable"
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

	r["method"] = "removeFromDatabaseTable"
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

	r["method"] = "updateInDatabaseTable"
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
