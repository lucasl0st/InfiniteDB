/*
 * Copyright (c) 2023 Lucas Pape
 */

package client

import (
	"errors"
	"net/http"
)

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
