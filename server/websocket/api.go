/*
 * Copyright (c) 2023 Lucas Pape
 */

package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/lucasl0st/InfiniteDB/idblib"
	idbutil "github.com/lucasl0st/InfiniteDB/idblib/util"
	"github.com/lucasl0st/InfiniteDB/server/util"
	infinitedbutil "github.com/lucasl0st/InfiniteDB/util"
	"net/http"
	"time"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Api struct {
	idb       *idblib.IDB
	logging   bool
	l         idbutil.Logger
	readLimit int64
	shutdown  func()
}

func New(idb *idblib.IDB, logging bool, logger idbutil.Logger, readLimit int64, shutdown func()) *Api {
	return &Api{
		idb:       idb,
		logging:   logging,
		l:         logger,
		readLimit: readLimit,
		shutdown:  shutdown,
	}
}

func (a *Api) Run(r *gin.Engine) {
	a.registerHandler(r)
}

func (a *Api) registerHandler(r *gin.Engine) {
	r.GET("/ws", func(c *gin.Context) {
		a.handler(c, c.Writer, c.Request)
	})
}

func (a *Api) handler(ctx *gin.Context, rw http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(rw, r, nil)

	if err != nil {
		a.l.Printf("failed to set websocket upgrade: %+v\n", err)
		return
	}

	conn.SetReadLimit(a.readLimit)

	a.sendResponse(conn, 0, infinitedbutil.InterfaceMapToJsonRawMap(gin.H{
		"message":          "HELO",
		"status":           http.StatusOK,
		"database_version": util.SERVER_VERSION,
	}))

	a.read(ctx, conn)
}

func (a *Api) read(ctx *gin.Context, conn *websocket.Conn) {
	for {
		_, bytes, err := conn.ReadMessage()

		if err != nil {
			if a.sendError(conn, 0, http.StatusInternalServerError, "failed to read message") {
				return
			}

			continue
		}

		body, err := a.parseBody(bytes)

		if err != nil {
			if a.sendError(conn, 0, http.StatusBadRequest, err.Error()) {
				return
			}

			continue
		}

		request := infinitedbutil.JsonRawMapToInterfaceMap(body)
		requestId, ok := request["requestId"].(int64)

		if !ok {
			if a.sendError(conn, 0, http.StatusBadRequest, "requestId is not a number") {
				return
			}

			continue
		}

		since := time.Now()

		response, method, err := a.handleRequest(request, body)

		if err != nil {
			if a.logging && method != nil {
				a.log(*method, http.StatusInternalServerError, since, ctx.ClientIP(), requestId)
			}

			if a.sendError(conn, requestId, http.StatusInternalServerError, err.Error()) {
				return
			}

			continue
		}

		if a.logging && method != nil {
			a.log(*method, http.StatusOK, since, ctx.ClientIP(), requestId)
		}

		if a.sendResponse(conn, requestId, response) {
			return
		}
	}
}

func (a *Api) sendResponse(conn *websocket.Conn, requestId int64, m map[string]json.RawMessage) bool {
	m["requestId"] = infinitedbutil.Int64ToJsonRaw(requestId)

	_, ok := m["status"]

	if !ok {
		m["status"] = infinitedbutil.InterfaceToJsonRaw(http.StatusOK)
	}

	err := conn.WriteJSON(m)

	if err != nil {
		a.l.Println(err)
		err = conn.Close()

		if err != nil {
			a.l.Println(err)
		}

		return true
	}

	return false
}

func (a *Api) sendError(conn *websocket.Conn, requestId int64, status int, message string) bool {
	m := infinitedbutil.InterfaceMapToJsonRawMap(gin.H{
		"status":  status,
		"message": message,
	})

	return a.sendResponse(conn, requestId, m)
}

func (a *Api) parseBody(bytes []byte) (map[string]json.RawMessage, error) {
	var body map[string]json.RawMessage
	err := json.Unmarshal(bytes, &body)

	if err != nil {
		return nil, errors.New("failed to parse JSON")
	}

	return body, nil
}

func (a *Api) handleRequest(request map[string]interface{}, rawRequest map[string]json.RawMessage) (map[string]json.RawMessage, *Method, error) {
	method, ok := request["method"].(string)

	if !ok {
		return nil, nil, errors.New("method is not a string")
	}

	for _, handler := range MethodHandlers {
		if string(handler.Method) == method {
			results, err := handler.Handler(a, request, rawRequest)

			if err != nil {
				return nil, &handler.Method, err
			}

			m, err := util.ToMap(results)

			if err != nil {
				return nil, &handler.Method, err
			}

			return m, &handler.Method, nil
		}
	}

	return nil, nil, errors.New("method not found")
}

func (a *Api) log(method Method, statusCode int, since time.Time, clientIp string, requestId int64) {
	param := new(gin.LogFormatterParams)
	param.Path = fmt.Sprint(method)
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

	a.l.Print(fmt.Sprintf("[GIN-WS] %v |%s %3d %s| %13v | %15s | %v |%s %-7s %s %#v\n%s",
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
