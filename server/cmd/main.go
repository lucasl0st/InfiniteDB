/*
 * Copyright (c) 2023 Lucas Pape
 */

package main

import (
	"github.com/lucasl0st/InfiniteDB/idblib/metrics"
	"github.com/lucasl0st/InfiniteDB/server"
	"github.com/lucasl0st/InfiniteDB/server/util"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var s *server.Server

func main() {
	var err error
	var metricsReceiver metrics.Receiver = &util.MetricsReceiver{}

	s, err = server.New(
		util.LoggerWithPrefix{Prefix: "[InfiniteDB]"},
		util.LoggerWithPrefix{Prefix: "[idblib]"},
		&metricsReceiver,
		shutdown,
	)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		_ = <-sigChan

		shutdown()
	}()

	if err != nil {
		log.Fatal(err.Error())
		return
	}

	err = s.Run()

	if err != nil {
		log.Fatal(err.Error())
	}
}

func shutdown() {
	s.Kill()

	os.Exit(0)
}
