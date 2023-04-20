/*
 * Copyright (c) 2023 Lucas Pape
 */

package methods

import "github.com/lucasl0st/InfiniteDB/client"

type Method struct {
	Name         string
	Arguments    []Argument
	RawArguments []Argument
	Run          func(c *client.Client, args []string) error
}

var Methods []Method
