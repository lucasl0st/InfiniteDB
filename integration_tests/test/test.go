/*
 * Copyright (c) 2023 Lucas Pape
 */

package test

import (
	"github.com/lucasl0st/InfiniteDB/client"
	"github.com/lucasl0st/InfiniteDB/integration_tests/covid19"
)

var Tests []Test

type Test struct {
	Name string
	Run  func(c *client.Client) error
}

func init() {
	Tests = append(Tests, Test{
		Name: "covid-19",
		Run:  covid19.Run,
	})
}
