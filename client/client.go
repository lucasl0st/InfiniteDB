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
	e "github.com/lucasl0st/InfiniteDB/models/errors"
	"github.com/lucasl0st/InfiniteDB/models/method"
	"github.com/lucasl0st/InfiniteDB/models/metric"
	"github.com/lucasl0st/InfiniteDB/util"
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
	hostname string
	port     uint

	tls           bool
	skipTLSVerify bool

	authKey *string

	timeout   time.Duration
	readLimit int64

	panicOnConnectionError bool
	connected              bool

	ws  *websocket.Conn
	ctx context.Context

	channels sync.Map

	MetricsReceiver metric.Receiver
}

func New(options Options) *Client {
	if options.TLS == nil {
		options.TLS = util.Ptr(false)
	}

	if options.SkipTLSVerify == nil {
		options.SkipTLSVerify = util.Ptr(false)
	}

	if options.Timeout == nil {
		options.Timeout = util.Ptr(time.Second * 10)
	}

	if options.ReadLimit == nil {
		options.ReadLimit = util.Ptr(int64(1000 * 1000 * 1000))
	}

	if options.PanicOnConnectionError == nil {
		options.PanicOnConnectionError = util.Ptr(true)
	}

	return &Client{
		hostname:               options.Hostname,
		port:                   options.Port,
		tls:                    *options.TLS,
		skipTLSVerify:          *options.SkipTLSVerify,
		authKey:                options.AuthKey,
		timeout:                *options.Timeout,
		readLimit:              *options.ReadLimit,
		panicOnConnectionError: *options.PanicOnConnectionError,
		connected:              false,
		ctx:                    context.Background(),
		channels:               sync.Map{},
	}
}

func (c *Client) Connect() error {
	header := http.Header{}

	if c.authKey != nil {
		header.Add("Authorization", *c.authKey)
	}

	protocol := "ws"

	if c.tls {
		protocol = "wss"
	}

	address := fmt.Sprintf("%s://%s:%v/ws", protocol, c.hostname, c.port)

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: c.skipTLSVerify,
			},
		},
	}

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

	for !c.connected {
		//TODO get error here
	}

	return nil
}

func (c *Client) read() {
	for {
		_, data, err := c.ws.Read(c.ctx)

		if err != nil {
			if c.panicOnConnectionError && c.connected {
				panic(err.Error())
			}

			return
		}

		var msg map[string]interface{}

		err = json.Unmarshal(data, &msg)

		if err != nil {
			return
		}

		c.handleResponse(msg)
	}
}

func (c *Client) handleResponse(msg map[string]interface{}) {
	m, ok := msg["method"].(string)

	if !ok {
		panic("server did not respond with method")
	}

	switch m {
	case fmt.Sprint(method.HeloMethod):
		c.handleHeloMethod(msg)
	case fmt.Sprint(method.GenericErrorMethod):
		c.handleGenericErrorMethod(msg)
	case fmt.Sprint(method.RequestResponseMethod):
		c.handleRequestResultResponseMethod(msg)
	case fmt.Sprint(method.MetricsUpdateMethod):
		c.handleMetricsUpdateMethod(msg)
	}
}

func (c *Client) handleHeloMethod(msg map[string]interface{}) {
	version, isString := msg["database_version"].(string)

	if !isString {
		panic(e.DidNotReceiveDatabaseVersion())
	}

	if version != VERSION {
		panic(e.ClientNotCompatibleWithDatabaseServer(version, VERSION))
	}

	fmt.Println("connected")

	c.connected = true
}

func (c *Client) handleGenericErrorMethod(msg map[string]interface{}) {
	panic(msg["message"].(string))
}

func (c *Client) handleRequestResultResponseMethod(msg map[string]interface{}) {
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
		} else if c.connected {
			panic(r.Err)
		}
	}
}

func (c *Client) handleMetricsUpdateMethod(msg map[string]interface{}) {
	if c.MetricsReceiver != nil {
		m := msg["metric"]

		switch m {
		case fmt.Sprint(metric.DatabaseMetric):
			var databaseMetricsResponse metric.DatabaseMetricResponse

			err := mapToStruct(msg["value"].(map[string]interface{}), &databaseMetricsResponse)

			if err != nil {
				panic(err.Error())
			}

			c.MetricsReceiver.DatabaseMetrics(databaseMetricsResponse.Database, databaseMetricsResponse.Metrics)

		case fmt.Sprint(metric.PerformanceMetric):
			var performanceMetricsResponse metric.PerformanceMetricResponse

			err := mapToStruct(msg["value"].(map[string]interface{}), &performanceMetricsResponse)

			if err != nil {
				panic(err.Error())
			}

			c.MetricsReceiver.PerformanceMetrics(performanceMetricsResponse.Metrics)
		}
	}
}

func (c *Client) sendRequest(request map[string]interface{}) (map[string]interface{}, error) {
	if !c.connected {
		return nil, e.ClientNotConnected()
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
		if c.panicOnConnectionError && c.connected {
			panic(err.Error())
		}

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
		err := e.TimeoutReceivingDatabaseResult(requestId)

		if c.panicOnConnectionError && c.connected {
			panic(err.Error())
		}

		return nil, err
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
