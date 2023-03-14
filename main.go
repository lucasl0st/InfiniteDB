/*
 * Copyright (c) 2023 Lucas Pape
 */

package main

import (
	"github.com/lucasl0st/InfiniteDB/idblib/metrics"
	"github.com/lucasl0st/InfiniteDB/server"
	"github.com/lucasl0st/InfiniteDB/util"
	"log"
)

func main() {
	var metricsReceiver metrics.Receiver = &util.MetricsReceiver{}

	s, err := server.New(
		util.LoggerWithPrefix{Prefix: "[InfiniteDB]"},
		util.LoggerWithPrefix{Prefix: "[idblib]"},
		&metricsReceiver,
	)

	if err != nil {
		log.Fatal(err.Error())
		return
	}

	err = s.Run()

	if err != nil {
		log.Fatal(err.Error())
	}
}
