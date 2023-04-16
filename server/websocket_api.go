/*
 * Copyright (c) 2023 Lucas Pape
 */

package server

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	e "github.com/lucasl0st/InfiniteDB/errors"
	"github.com/lucasl0st/InfiniteDB/idblib"
	"github.com/lucasl0st/InfiniteDB/idblib/util"
	"github.com/lucasl0st/InfiniteDB/request"
	"net/http"
	"time"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebsocketApi struct {
	idb       *idblib.IDB
	logging   bool
	readLimit int64
}

func (w *WebsocketApi) Run(r *gin.Engine) {
	w.registerHandler(r)
}

func (w *WebsocketApi) registerHandler(r *gin.Engine) {
	r.GET("/ws", func(c *gin.Context) {
		w.handler(c, c.Writer, c.Request)
	})
}

func (w *WebsocketApi) handler(c *gin.Context, rw http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(rw, r, nil)

	if err != nil {
		l.Printf("failed to set websocket upgrade: %+v\n", err)
		return
	}

	conn.SetReadLimit(w.readLimit)

	w.send(conn, 0, util.InterfaceMapToJsonRawMap(gin.H{
		"message":          "HELO",
		"status":           http.StatusOK,
		"database_version": VERSION,
	}))

	for {
		_, bytes, err := conn.ReadMessage()

		if err != nil {
			closed := w.send(conn, 0, util.InterfaceMapToJsonRawMap(gin.H{
				"status":  http.StatusInternalServerError,
				"message": "failed to read message",
			}))

			if closed {
				return
			}
		}

		body, closed := w.getBody(conn, bytes)

		if closed {
			return
		}

		if body != nil {
			m := util.JsonRawMapToInterfaceMap(*body)

			requestId := m["requestId"]

			if requestId != nil {
				if w.methodHandler(conn, c.ClientIP(), int64(requestId.(float64)), m, *body) {
					return
				}
			} else {
				if w.send(conn, 0, util.InterfaceMapToJsonRawMap(gin.H{
					"status":  http.StatusInternalServerError,
					"message": "every request must have a requestId",
				})) {
					return
				}
			}
		}
	}
}

func (w *WebsocketApi) getBody(conn *websocket.Conn, bytes []byte) (*map[string]json.RawMessage, bool) {
	var r map[string]json.RawMessage
	err := json.Unmarshal(bytes, &r)

	if err != nil {
		return nil, w.send(conn, 0, util.InterfaceMapToJsonRawMap(gin.H{
			"status":  http.StatusInternalServerError,
			"message": "failed to parse JSON",
		}))
	}

	return &r, false
}

func (w *WebsocketApi) send(conn *websocket.Conn, requestId int64, m map[string]json.RawMessage) bool {
	m["requestId"] = util.Int64ToJsonRaw(requestId)

	err := conn.WriteJSON(m)

	if err != nil {
		l.Println(err)
		err = conn.Close()

		if err != nil {
			l.Println(err)
		}

		return true
	}

	return false
}

func (w *WebsocketApi) methodHandler(
	conn *websocket.Conn,
	clientIp string,
	requestId int64,
	m map[string]interface{},
	r map[string]json.RawMessage,
) bool {
	closed := false
	status := 0
	since := time.Now()

	method := m["method"]

	if method != nil {
		switch method.(string) {
		case "getDatabases":
			closed, status = w.getDatabasesHandler(conn, requestId)
		case "createDatabase":
			closed, status = w.createDatabaseHandler(conn, requestId, m)
		case "deleteDatabase":
			closed, status = w.deleteDatabaseHandler(conn, requestId, m)
		case "getDatabase":
			closed, status = w.getDatabaseHandler(conn, requestId, m)
		case "getDatabaseTables":
			closed, status = w.getDatabaseTablesHandler(conn, requestId, m)
		case "createTableInDatabase":
			closed, status = w.createTableInDatabaseHandler(conn, requestId, m)
		case "deleteTableInDatabase":
			closed, status = w.deleteTableInDatabaseHandler(conn, requestId, m)
		case "getFromDatabaseTable":
			closed, status = w.getFromDatabaseTableHandler(conn, requestId, m, r)
		case "insertToDatabaseTable":
			closed, status = w.insertToDatabaseTableHandler(conn, requestId, m, r)
		case "removeFromDatabaseTable":
			closed, status = w.removeFromDatabaseTableHandler(conn, requestId, m, r)
		case "updateInDatabaseTable":
			closed, status = w.updateInDatabaseTableHandler(conn, requestId, m, r)
		default:
			closed = w.send(conn, requestId, util.InterfaceMapToJsonRawMap(gin.H{
				"status":  http.StatusInternalServerError,
				"message": "method not found",
			}))
			status = http.StatusInternalServerError
		}

		if w.logging {
			w.logHandler(method.(string), status, since, clientIp, requestId)
		}
	} else {
		closed = w.send(conn, requestId, util.InterfaceMapToJsonRawMap(gin.H{
			"status":  http.StatusInternalServerError,
			"message": "no method specified",
		}))
	}

	return closed
}

func (w *WebsocketApi) getDatabasesHandler(conn *websocket.Conn, requestId int64) (bool, int) {
	results, err := w.idb.GetDatabases()

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	m, err := toMap(results)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	return w.sendResults(conn, requestId, &m, err)
}

func (w *WebsocketApi) createDatabaseHandler(conn *websocket.Conn, requestId int64, r map[string]interface{}) (bool, int) {
	name, isString := r["name"].(string)

	if !isString {
		return w.sendResults(conn, requestId, nil, e.IsNotAString("name"))
	}

	err := validateName(name)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	results, err := w.idb.CreateDatabase(name)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	m, err := toMap(results)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	return w.sendResults(conn, requestId, &m, err)
}

func (w *WebsocketApi) deleteDatabaseHandler(conn *websocket.Conn, requestId int64, r map[string]interface{}) (bool, int) {
	name, isString := r["name"].(string)

	if !isString {
		return w.sendResults(conn, requestId, nil, e.IsNotAString("name"))
	}

	err := validateName(name)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	results, err := w.idb.DeleteDatabase(name)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	m, err := toMap(results)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	return w.sendResults(conn, requestId, &m, err)
}

func (w *WebsocketApi) getDatabaseHandler(conn *websocket.Conn, requestId int64, r map[string]interface{}) (bool, int) {
	name, isString := r["name"].(string)

	if !isString {
		return w.sendResults(conn, requestId, nil, e.IsNotAString("name"))
	}

	err := validateName(name)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	results, err := w.idb.GetDatabase(name)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	m, err := toMap(results)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	return w.sendResults(conn, requestId, &m, err)
}

func (w *WebsocketApi) getDatabaseTablesHandler(conn *websocket.Conn, requestId int64, r map[string]interface{}) (bool, int) {
	name, isString := r["name"].(string)

	if !isString {
		return w.sendResults(conn, requestId, nil, e.IsNotAString("name"))
	}

	err := validateName(name)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	results, err := w.idb.GetDatabaseTables(name)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	m, err := toMap(results)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	return w.sendResults(conn, requestId, &m, err)
}

func (w *WebsocketApi) createTableInDatabaseHandler(conn *websocket.Conn, requestId int64, r map[string]interface{}) (bool, int) {
	name, isString := r["name"].(string)

	if !isString {
		return w.sendResults(conn, requestId, nil, e.IsNotAString("name"))
	}

	err := validateName(name)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	tableName, isString := r["tableName"].(string)

	err = validateName(tableName)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	if !isString {
		return w.sendResults(conn, requestId, nil, e.IsNotAString("tableName"))
	}

	fields, isMap := r["fields"].(map[string]interface{})

	if !isMap {
		return w.sendResults(conn, requestId, nil, e.IsNotAMap("fields"))
	}

	var o request.TableOptions
	options, isMap := r["options"].(map[string]interface{})

	if !isMap {
		o = request.TableOptions{}
	} else {
		err := toStruct(options, &o)

		if err != nil {
			return w.sendResults(conn, requestId, nil, err)
		}
	}

	var f map[string]request.Field
	err = toStruct(fields, &f)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	parsedFields, err := parseFields(f)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	results, err := w.idb.CreateTableInDatabase(name, tableName, parsedFields, o)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	m, err := toMap(results)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	return w.sendResults(conn, requestId, &m, err)
}

func (w *WebsocketApi) deleteTableInDatabaseHandler(conn *websocket.Conn, requestId int64, r map[string]interface{}) (bool, int) {
	name, isString := r["name"].(string)

	if !isString {
		return w.sendResults(conn, requestId, nil, e.IsNotAString("name"))
	}

	err := validateName(name)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	tableName, isString := r["tableName"].(string)

	if !isString {
		return w.sendResults(conn, requestId, nil, e.IsNotAString("tableName"))
	}

	err = validateName(tableName)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	results, err := w.idb.DeleteTableInDatabase(name, tableName)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	m, err := toMap(results)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	return w.sendResults(conn, requestId, &m, err)
}

func (w *WebsocketApi) getFromDatabaseTableHandler(
	conn *websocket.Conn,
	requestId int64,
	m map[string]interface{},
	r map[string]json.RawMessage,
) (bool, int) {
	name, isString := m["name"].(string)

	if !isString {
		return w.sendResults(conn, requestId, nil, e.IsNotAString("name"))
	}

	err := validateName(name)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	tableName, isString := m["tableName"].(string)

	if !isString {
		return w.sendResults(conn, requestId, nil, e.IsNotAString("tableName"))
	}

	err = validateName(tableName)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	var req request.Request
	err = toStruct(r["request"], &req)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	parsedRequest, err := parseRequest(req)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	results, err := w.idb.GetFromDatabaseTable(name, tableName, *parsedRequest)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	rm, err := toMap(results)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	return w.sendResults(conn, requestId, &rm, err)
}

func (w *WebsocketApi) insertToDatabaseTableHandler(
	conn *websocket.Conn,
	requestId int64,
	m map[string]interface{},
	r map[string]json.RawMessage,
) (bool, int) {
	name, isString := m["name"].(string)

	if !isString {
		return w.sendResults(conn, requestId, nil, e.IsNotAString("name"))
	}

	err := validateName(name)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	tableName, isString := m["tableName"].(string)

	if !isString {
		return w.sendResults(conn, requestId, nil, e.IsNotAString("tableName"))
	}

	err = validateName(tableName)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	var o map[string]json.RawMessage
	err = toStruct(r["object"], &o)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	results, err := w.idb.InsertToDatabaseTable(name, tableName, o)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	rm, err := toMap(results)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	return w.sendResults(conn, requestId, &rm, err)
}

func (w *WebsocketApi) removeFromDatabaseTableHandler(
	conn *websocket.Conn,
	requestId int64,
	m map[string]interface{},
	r map[string]json.RawMessage,
) (bool, int) {
	name, isString := m["name"].(string)

	if !isString {
		return w.sendResults(conn, requestId, nil, e.IsNotAString("name"))
	}

	err := validateName(name)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	tableName, isString := m["tableName"].(string)

	if !isString {
		return w.sendResults(conn, requestId, nil, e.IsNotAString("tableName"))
	}

	err = validateName(tableName)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	var req request.Request
	err = toStruct(r["request"], &req)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	parsedRequest, err := parseRequest(req)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	results, err := w.idb.RemoveFromDatabaseTable(name, tableName, *parsedRequest)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	rm, err := toMap(results)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	return w.sendResults(conn, requestId, &rm, err)
}

func (w *WebsocketApi) updateInDatabaseTableHandler(
	conn *websocket.Conn,
	requestId int64,
	m map[string]interface{},
	r map[string]json.RawMessage,
) (bool, int) {
	name, isString := m["name"].(string)

	if !isString {
		return w.sendResults(conn, requestId, nil, e.IsNotAString("name"))
	}

	err := validateName(name)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	tableName, isString := m["tableName"].(string)

	if !isString {
		return w.sendResults(conn, requestId, nil, e.IsNotAString("tableName"))
	}

	err = validateName(tableName)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	var o map[string]json.RawMessage
	err = toStruct(r["object"], &o)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	results, err := w.idb.UpdateInDatabaseTable(name, tableName, o)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	rm, err := toMap(results)

	if err != nil {
		return w.sendResults(conn, requestId, nil, err)
	}

	return w.sendResults(conn, requestId, &rm, err)
}

func (w *WebsocketApi) sendResults(conn *websocket.Conn, requestId int64, results *map[string]json.RawMessage, err error) (bool, int) {
	if err == nil && results != nil {
		r := *results
		r["status"] = util.InterfaceToJsonRaw(http.StatusOK)

		return w.send(conn, requestId, r), http.StatusOK
	} else {
		return w.send(conn, requestId, util.InterfaceMapToJsonRawMap(gin.H{
			"status":  http.StatusInternalServerError,
			"message": fmt.Sprint(err),
		})), http.StatusInternalServerError
	}
}

func (w *WebsocketApi) logHandler(method string, statusCode int, since time.Time, clientIp string, requestId int64) {
	param := new(gin.LogFormatterParams)
	param.Path = method
	param.Method = http.MethodGet
	param.ClientIP = clientIp
	param.Latency = time.Since(since)
	param.StatusCode = statusCode
	param.TimeStamp = time.Now()

	statusColor := param.StatusCodeColor()
	methodColor := param.MethodColor()
	resetColor := param.ResetColor()

	param.Method = "Websocket"

	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}

	l.Print(fmt.Sprintf("[GIN-WS] %v |%s %3d %s| %13v | %15s | %v |%s %-7s %s %#v\n%s",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		param.ClientIP,
		requestId,
		methodColor, param.Method, resetColor,
		param.Path,
		param.ErrorMessage,
	))
}
