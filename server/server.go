/*
 * Copyright (c) 2023 Lucas Pape
 */

package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lucasl0st/InfiniteDB/idblib"
	"github.com/lucasl0st/InfiniteDB/idblib/metrics"
	"github.com/lucasl0st/InfiniteDB/idblib/table"
	"github.com/lucasl0st/InfiniteDB/idblib/util"
	e "github.com/lucasl0st/InfiniteDB/models/errors"
	"github.com/lucasl0st/InfiniteDB/server/http"
	"github.com/lucasl0st/InfiniteDB/server/internal_database"
	"github.com/lucasl0st/InfiniteDB/server/websocket"
)

var l util.Logger

type Server struct {
	c   Config
	idb *idblib.IDB
	r   *gin.Engine
}

func New(
	logger util.Logger,
	idbLogger util.Logger,
	metricsReceiver *metrics.Receiver,
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

	idb, err := idblib.New(config.DatabasePath, idbLogger, metricsReceiver, config.CacheSize)

	if err != nil {
		return nil, err
	}

	err = internal_database.SetupInternalDatabase(idb)

	if err != nil {
		return nil, e.FailedToSetupInternalDatabase(err)
	}

	if config.Authentication {
		err = internal_database.SetupAuthenticationTable(idb)

		if err != nil {
			return nil, e.FailedToSetupInternalAuthenticationTable(err)
		}
	}

	table.CreateDatabaseMiddleware = CreateDatabaseMiddleware

	r := gin.Default()

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

	return &Server{
		c:   *config,
		idb: idb,
		r:   r,
	}, nil
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
