/*
 * Copyright (c) 2023 Lucas Pape
 */

package server

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lucasl0st/InfiniteDB/idblib"
	"github.com/lucasl0st/InfiniteDB/request"
	"io"
	"net/http"
)

const apiPrefix = ""

type HttpApi struct {
	idb            *idblib.IDB
	authentication bool
}

func (h *HttpApi) Run(r *gin.Engine) {
	h.registerHandlers(r)
}

func (h *HttpApi) registerHandlers(r *gin.Engine) {
	r.Use(h.authenticationHandler())

	r.GET(apiPrefix+"/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello world"})
	})

	r.GET(apiPrefix+"/health", h.healthHandler)
	r.GET(apiPrefix+"/version", h.versionHandler)

	r.GET(apiPrefix+"/databases", h.getDatabasesHandler)
	r.POST(apiPrefix+"/database", h.createDatabaseHandler)
	r.DELETE(apiPrefix+"/database/:name", h.deleteDatabaseHandler)
	r.GET(apiPrefix+"/database/:name", h.getDatabaseHandler)
	r.GET(apiPrefix+"/database/:name/tables", h.getDatabaseTablesHandler)
	r.POST(apiPrefix+"/database/:name/table", h.createTableInDatabaseHandler)
	r.DELETE(apiPrefix+"/database/:name/table/:tableName", h.deleteTableInDatabaseHandler)
	r.POST(apiPrefix+"/database/:name/table/:tableName/get", h.getFromDatabaseTableHandler)
	r.POST(apiPrefix+"/database/:name/table/:tableName/insert", h.insertToDatabaseTableHandler)
	r.POST(apiPrefix+"/database/:name/table/:tableName/remove", h.removeFromDatabaseTableHandler)
	r.POST(apiPrefix+"/database/:name/table/:tableName/update", h.updateInDatabaseTableHandler)
}

func (h *HttpApi) authenticationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !h.authentication {
			return
		}

		if c.Request.URL.Path == "/health" {
			return
		}

		a, err := authenticated(h.idb, c.GetHeader("Authorization"))

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			c.Abort()
			return
		}

		if !a {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			c.Abort()
			return
		}
	}
}

func (h *HttpApi) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *HttpApi) versionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"database_version": VERSION})
}

func (h *HttpApi) getDatabasesHandler(c *gin.Context) {
	results, err := h.idb.GetDatabases()

	if err == nil {
		c.JSON(http.StatusOK, results)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
	}
}

func (h *HttpApi) createDatabaseHandler(c *gin.Context) {
	body := h.getBody(c)

	if body != nil {
		name, isString := (*body)["name"].(string)

		if !isString {
			c.JSON(http.StatusBadRequest, gin.H{"message": "name is not a string"})
			return
		}

		results, err := h.idb.CreateDatabase(name)

		if err == nil {
			c.JSON(http.StatusOK, results)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
		}
	}
}

func (h *HttpApi) deleteDatabaseHandler(c *gin.Context) {
	name := c.Param("name")

	results, err := h.idb.DeleteDatabase(name)

	if err == nil {
		c.JSON(http.StatusOK, results)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
	}
}

func (h *HttpApi) getDatabaseHandler(c *gin.Context) {
	name := c.Param("name")

	results, err := h.idb.GetDatabase(name)

	if err == nil {
		c.JSON(http.StatusOK, results)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
	}
}

func (h *HttpApi) getDatabaseTablesHandler(c *gin.Context) {
	name := c.Param("name")

	results, err := h.idb.GetDatabaseTables(name)

	if err == nil {
		c.JSON(http.StatusOK, results)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
	}
}

func (h *HttpApi) createTableInDatabaseHandler(c *gin.Context) {
	body := h.getBody(c)

	if body != nil {
		name := c.Param("name")

		tableName, isString := (*body)["name"].(string)

		if !isString {
			c.JSON(http.StatusBadRequest, gin.H{"message": "tableName is not a string"})
			return
		}

		fields, isMap := (*body)["fields"].(map[string]interface{})

		if !isMap {
			c.JSON(http.StatusBadRequest, gin.H{"message": "fields is not a map"})
			return
		}

		var o request.TableOptions
		options, isMap := (*body)["options"].(map[string]interface{})

		if !isMap {
			o = request.TableOptions{}
		} else {
			err := toStruct(options, &o)

			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
				return
			}
		}

		var f map[string]request.Field
		err := toStruct(fields, &f)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
			return
		}

		parsedFields, err := parseFields(f)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
			return
		}

		results, err := h.idb.CreateTableInDatabase(name, tableName, parsedFields, o)

		if err == nil {
			c.JSON(http.StatusOK, results)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
		}
	}
}

func (h *HttpApi) deleteTableInDatabaseHandler(c *gin.Context) {
	name := c.Param("name")
	tableName := c.Param("tableName")

	results, err := h.idb.DeleteTableInDatabase(name, tableName)

	if err == nil {
		c.JSON(http.StatusOK, results)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
	}
}

func (h *HttpApi) getFromDatabaseTableHandler(c *gin.Context) {
	r := h.getRequest(c)

	if r != nil {
		name := c.Param("name")
		tableName := c.Param("tableName")

		parsedRequest, err := parseRequest(*r)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
			return
		}

		results, err := h.idb.GetFromDatabaseTable(name, tableName, *parsedRequest)

		if err == nil {
			c.JSON(http.StatusOK, results)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
		}
	}
}

// TODO check if table exists
func (h *HttpApi) insertToDatabaseTableHandler(c *gin.Context) {
	body := h.getBody(c)

	if body != nil {
		name := c.Param("name")
		tableName := c.Param("tableName")

		results, err := h.idb.InsertToDatabaseTable(name, tableName, *body)

		if err == nil {
			c.JSON(http.StatusOK, results)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
		}
	}
}

func (h *HttpApi) removeFromDatabaseTableHandler(c *gin.Context) {
	r := h.getRequest(c)

	if r != nil {
		name := c.Param("name")
		tableName := c.Param("tableName")

		parsedRequest, err := parseRequest(*r)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
			return
		}

		results, err := h.idb.RemoveFromDatabaseTable(name, tableName, *parsedRequest)

		if err == nil {
			c.JSON(http.StatusOK, results)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
		}
	}
}

func (h *HttpApi) updateInDatabaseTableHandler(c *gin.Context) {
	body := h.getBody(c)

	if body != nil {
		name := c.Param("name")
		tableName := c.Param("tableName")

		results, err := h.idb.UpdateInDatabaseTable(name, tableName, *body)

		if err == nil {
			c.JSON(http.StatusOK, results)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
		}
	}
}

func (h *HttpApi) getBody(c *gin.Context) *map[string]interface{} {
	bytes, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to read body"})
		return nil
	}

	var m map[string]interface{}
	err = json.Unmarshal(bytes, &m)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to parse JSON", "error": err.Error()})
		return nil
	}

	return &m
}

func (h *HttpApi) getRequest(c *gin.Context) *request.Request {
	bytes, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to read body"})
		return nil
	}

	var r request.Request
	err = json.Unmarshal(bytes, &r)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
		return nil
	}

	return &r
}
