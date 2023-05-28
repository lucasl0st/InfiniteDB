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
	"github.com/lucasl0st/InfiniteDB/models/method"
	"github.com/lucasl0st/InfiniteDB/models/metric"
	models "github.com/lucasl0st/InfiniteDB/models/response"
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
	idb *idblib.IDB

	logging bool
	l       idbutil.Logger

	subscribedToMetricUpdates []*websocket.Conn

	readLimit int64

	shutdown func()
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

	a.send(conn, infinitedbutil.InterfaceMapToJsonRawMap(gin.H{
		"message":          "HELO",
		"status":           http.StatusOK,
		"database_version": util.SERVER_VERSION,
		"method":           method.HeloMethod,
	}))

	a.read(ctx, conn)
}

func (a *Api) read(ctx *gin.Context, conn *websocket.Conn) {
	for {
		_, bytes, err := conn.ReadMessage()

		if err != nil {
			if a.sendGenericError(conn, http.StatusInternalServerError, "failed to read message") {
				return
			}

			continue
		}

		body, err := a.parseBody(bytes)

		if err != nil {
			if a.sendGenericError(conn, http.StatusBadRequest, err.Error()) {
				return
			}

			continue
		}

		request := infinitedbutil.JsonRawMapToInterfaceMap(body)
		requestIdFloat, ok := request["requestId"].(float64)

		if !ok {
			if a.sendGenericError(conn, http.StatusBadRequest, "requestId is not a number") {
				return
			}

			continue
		}

		requestId := int64(requestIdFloat)

		since := time.Now()

		response, m, err := a.handleRequest(conn, request, body)

		if err != nil {
			if a.logging && m != nil {
				a.log(*m, http.StatusInternalServerError, since, ctx.ClientIP(), requestId)
			}

			if a.sendRequestError(conn, requestId, http.StatusInternalServerError, err.Error()) {
				return
			}

			continue
		}

		if a.logging && m != nil {
			a.log(*m, http.StatusOK, since, ctx.ClientIP(), requestId)
		}

		if a.sendRequestResponse(conn, requestId, response) {
			return
		}
	}
}

func (a *Api) send(conn *websocket.Conn, msg any) bool {
	err := conn.WriteJSON(msg)

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

func (a *Api) sendRequestResponse(conn *websocket.Conn, requestId int64, m map[string]json.RawMessage) bool {
	m["method"] = infinitedbutil.StringToJsonRaw(fmt.Sprint(method.RequestResponseMethod))
	m["requestId"] = infinitedbutil.Int64ToJsonRaw(requestId)

	_, ok := m["status"]

	if !ok {
		m["status"] = infinitedbutil.InterfaceToJsonRaw(http.StatusOK)
	}

	return a.send(conn, m)
}

func (a *Api) sendRequestError(conn *websocket.Conn, requestId int64, status int, message string) bool {
	m := infinitedbutil.InterfaceMapToJsonRawMap(gin.H{
		"status":  status,
		"message": message,
	})

	return a.sendRequestResponse(conn, requestId, m)
}

func (a *Api) sendGenericError(conn *websocket.Conn, status int, message string) bool {
	m := infinitedbutil.InterfaceMapToJsonRawMap(gin.H{
		"method":  method.GenericErrorMethod,
		"status":  status,
		"message": message,
	})

	return a.send(conn, m)
}

func (a *Api) parseBody(bytes []byte) (map[string]json.RawMessage, error) {
	var body map[string]json.RawMessage
	err := json.Unmarshal(bytes, &body)

	if err != nil {
		return nil, errors.New("failed to parse JSON")
	}

	return body, nil
}

func (a *Api) handleRequest(conn *websocket.Conn, request map[string]interface{}, rawRequest map[string]json.RawMessage) (map[string]json.RawMessage, *method.ServerMethod, error) {
	m, ok := request["method"].(string)

	if !ok {
		return nil, nil, errors.New("method is not a string")
	}

	for _, handler := range MethodHandlers {
		if string(handler.Method) == m {
			results, err := handler.Handler(a, conn, request, rawRequest)

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

func (a *Api) log(method method.ServerMethod, statusCode int, since time.Time, clientIp string, requestId int64) {
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

func (a *Api) SubmitMetricUpdate(metric metric.Metric, value any) {
	if a == nil {
		return
	}

	response, err := util.ToMap(models.MetricUpdateResponse{
		Metric: fmt.Sprint(metric),
		Value:  value,
	})

	if err != nil {
		a.l.Fatal(err.Error())
	}

	for _, conn := range a.subscribedToMetricUpdates {
		response["method"] = infinitedbutil.StringToJsonRaw(fmt.Sprint(method.MetricsUpdateMethod))

		a.send(conn, response)
	}
}
