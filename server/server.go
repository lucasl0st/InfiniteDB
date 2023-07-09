/*
 * Copyright (c) 2023 Lucas Pape
 */

package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lucasl0st/InfiniteDB/idblib"
	"github.com/lucasl0st/InfiniteDB/idblib/table"
	"github.com/lucasl0st/InfiniteDB/idblib/util"
	e "github.com/lucasl0st/InfiniteDB/models/errors"
	"github.com/lucasl0st/InfiniteDB/models/metric"
	"github.com/lucasl0st/InfiniteDB/server/http"
	"github.com/lucasl0st/InfiniteDB/server/internal_database"
	serverutil "github.com/lucasl0st/InfiniteDB/server/util"
	"github.com/lucasl0st/InfiniteDB/server/websocket"
)

var l util.Logger

type Server struct {
	c   Config
	idb *idblib.IDB
	r   *gin.Engine

	websocketApi *websocket.Api
}

func New(
	logger util.Logger,
	idbLogger util.Logger,
	shutdown func(),
) (*Server, error) {
	l = logger
	internal_database.SetLogger(l)

	config, err := LoadConfig()

	if err != nil {
		return nil, err
	}

	l.Println("using database path: " + config.DatabasePath)
	l.Println("authentication enabled: " + fmt.Sprint(config.Authentication))

	s := &Server{}

	var metricsReceiver metric.Receiver = &serverutil.MetricsReceiver{
		SubmitMetric: s.submitMetric,
	}

	l.Println("starting up idb")

	idb, err := idblib.New(config.DatabasePath, idbLogger, &metricsReceiver, config.CacheSize, func() {
		err = internal_database.SetupInternalDatabase(s.idb)

		if err != nil {
			l.Fatal(e.FailedToSetupInternalDatabase(err))
		}

		if config.Authentication {
			err = internal_database.SetupAuthenticationTable(s.idb)

			if err != nil {
				l.Fatal(e.FailedToSetupInternalAuthenticationTable(err))
			}
		}

		table.CreateDatabaseMiddleware = CreateDatabaseMiddleware

		l.Println("idb is ready")
	})

	if err != nil {
		return nil, err
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{
			"/health",
		},
	}))

	httpApi := http.New(
		idb,
		config.Authentication,
		shutdown,
	)

	httpApi.Run(r)

	websocketApi := websocket.New(
		idb,
		config.RequestLogging,
		l,
		config.WebsocketReadLimit,
		shutdown,
	)
	websocketApi.Run(r)

	s.c = *config
	s.idb = idb
	s.r = r
	s.websocketApi = websocketApi

	return s, nil
}

func (s *Server) Run() error {
	if s.c.TLS {
		l.Println("listening on port " + fmt.Sprint(s.c.Port) + " with TLS")

		return s.r.RunTLS(":"+fmt.Sprint(s.c.Port), s.c.TLSCert, s.c.TLSKey)
	} else {
		l.Println("listening on port " + fmt.Sprint(s.c.Port))

		return s.r.Run(":" + fmt.Sprint(s.c.Port))
	}
}

func (s *Server) Kill() {
	s.idb.Kill()
}

func (s *Server) submitMetric(metric metric.Metric, value any) {
	s.websocketApi.SubmitMetricUpdate(metric, value)
}
