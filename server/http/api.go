/*
 * Copyright (c) 2023 Lucas Pape
 */

package http

import (
	"github.com/gin-gonic/gin"
	"github.com/lucasl0st/InfiniteDB/idblib"
	"net/http"
)

const apiPrefix = ""

type Api struct {
	idb            *idblib.IDB
	authentication bool
	shutdown       func()
}

func New(idb *idblib.IDB, authentication bool, shutdown func()) *Api {
	return &Api{
		idb:            idb,
		authentication: authentication,
		shutdown:       shutdown,
	}
}

func (a *Api) Run(r *gin.Engine) {
	a.registerHandlers(r)
}

func (a *Api) registerHandlers(r *gin.Engine) {
	r.Use(a.authenticationHandler())

	r.GET(apiPrefix+"/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello world"})
	})

	r.GET(apiPrefix+"/health", a.healthHandler)
	r.GET(apiPrefix+"/version", a.versionHandler)
	r.GET(apiPrefix+"/shutdown", a.shutdownHandler)

	r.GET(apiPrefix+"/databases", a.getDatabasesHandler)
	r.POST(apiPrefix+"/database", a.createDatabaseHandler)
	r.DELETE(apiPrefix+"/database/:name", a.deleteDatabaseHandler)
	r.GET(apiPrefix+"/database/:name", a.getDatabaseHandler)
	r.GET(apiPrefix+"/database/:name/table/:tableName", a.getDatabaseTableHandler)
	r.POST(apiPrefix+"/database/:name/table", a.createTableInDatabaseHandler)
	r.DELETE(apiPrefix+"/database/:name/table/:tableName", a.deleteTableInDatabaseHandler)
	r.POST(apiPrefix+"/database/:name/table/:tableName/get", a.getFromDatabaseTableHandler)
	r.POST(apiPrefix+"/database/:name/table/:tableName/insert", a.insertToDatabaseTableHandler)
	r.POST(apiPrefix+"/database/:name/table/:tableName/remove", a.removeFromDatabaseTableHandler)
	r.POST(apiPrefix+"/database/:name/table/:tableName/update", a.updateInDatabaseTableHandler)
}
