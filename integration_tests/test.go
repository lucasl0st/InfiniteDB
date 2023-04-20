/*
 * Copyright (c) 2023 Lucas Pape
 */

package integration_tests

import "github.com/lucasl0st/InfiniteDB/client"

var Tests []Test

type Test struct {
	Name string
	Run  func(c *client.Client) error
}
