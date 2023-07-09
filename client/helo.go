/*
 * Copyright (c) 2023 Lucas Pape
 */

package client

import e "github.com/lucasl0st/InfiniteDB/models/errors"

func (c *Client) handleHeloMethod(msg map[string]interface{}) {
	version, isString := msg["database_version"].(string)

	if !isString {
		panic(e.DidNotReceiveDatabaseVersion())
	}

	if version != VERSION {
		panic(e.ClientNotCompatibleWithDatabaseServer(version, VERSION))
	}

	c.connected = true
}
