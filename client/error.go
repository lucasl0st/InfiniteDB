/*
 * Copyright (c) 2023 Lucas Pape
 */

package client

func (c *Client) handleGenericErrorMethod(msg map[string]interface{}) {
	panic(msg["message"].(string))
}
