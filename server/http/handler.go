/*
 * Copyright (c) 2023 Lucas Pape
 */

package http

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lucasl0st/InfiniteDB/models/request"
	"github.com/lucasl0st/InfiniteDB/server/internal_database"
	"github.com/lucasl0st/InfiniteDB/server/parse"
	"github.com/lucasl0st/InfiniteDB/server/util"
	"io"
	"net/http"
)

func (a *Api) authenticationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !a.authentication {
			return
		}

		if c.Request.URL.Path == "/health" {
			return
		}

		a, err := internal_database.Authenticated(a.idb, c.GetHeader("Authorization"))

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

func (a *Api) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (a *Api) versionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"database_version": util.SERVER_VERSION})
}

func (a *Api) shutdownHandler(_ *gin.Context) {
	a.shutdown()
}

func (a *Api) getDatabasesHandler(c *gin.Context) {
	results, err := a.idb.GetDatabases()

	if err == nil {
		c.JSON(http.StatusOK, results)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
	}
}

func (a *Api) createDatabaseHandler(c *gin.Context) {
	body := a.getBody(c)

	if body != nil {
		name, isString := (*body)["name"].(string)

		if !isString {
			c.JSON(http.StatusBadRequest, gin.H{"message": "name is not a string"})
			return
		}

		err := util.ValidateName(name)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
			return
		}

		results, err := a.idb.CreateDatabase(name)

		if err == nil {
			c.JSON(http.StatusOK, results)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
		}
	}
}

func (a *Api) deleteDatabaseHandler(c *gin.Context) {
	name := c.Param("name")

	err := util.ValidateName(name)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
		return
	}

	results, err := a.idb.DeleteDatabase(name)

	if err == nil {
		c.JSON(http.StatusOK, results)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
	}
}

func (a *Api) getDatabaseHandler(c *gin.Context) {
	name := c.Param("name")

	err := util.ValidateName(name)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
		return
	}

	results, err := a.idb.GetDatabase(name)

	if err == nil {
		c.JSON(http.StatusOK, results)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
	}
}

func (a *Api) getDatabaseTableHandler(c *gin.Context) {
	name := c.Param("name")

	err := util.ValidateName(name)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
		return
	}

	tableName := c.Param("tableName")

	err = util.ValidateName(tableName)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
		return
	}

	results, err := a.idb.GetDatabaseTable(name, tableName)

	if err == nil {
		c.JSON(http.StatusOK, results)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
	}
}

func (a *Api) createTableInDatabaseHandler(c *gin.Context) {
	body := a.getBody(c)

	if body != nil {
		name := c.Param("name")

		err := util.ValidateName(name)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
			return
		}

		tableName, isString := (*body)["name"].(string)

		err = util.ValidateName(tableName)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
			return
		}

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
			err := util.ToStruct(options, &o)

			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
				return
			}
		}

		var f map[string]request.Field
		err = util.ToStruct(fields, &f)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
			return
		}

		parsedFields, err := parse.Fields(f)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
			return
		}

		results, err := a.idb.CreateTableInDatabase(name, tableName, parsedFields, o)

		if err == nil {
			c.JSON(http.StatusOK, results)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
		}
	}
}

func (a *Api) deleteTableInDatabaseHandler(c *gin.Context) {
	name := c.Param("name")

	err := util.ValidateName(name)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
		return
	}

	tableName := c.Param("tableName")

	err = util.ValidateName(tableName)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
		return
	}

	results, err := a.idb.DeleteTableInDatabase(name, tableName)

	if err == nil {
		c.JSON(http.StatusOK, results)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
	}
}

func (a *Api) getFromDatabaseTableHandler(c *gin.Context) {
	r := a.getRequest(c)

	if r != nil {
		name := c.Param("name")

		err := util.ValidateName(name)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
			return
		}

		tableName := c.Param("tableName")

		err = util.ValidateName(tableName)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
			return
		}

		parsedRequest, err := parse.Request(*r)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
			return
		}

		results, err := a.idb.GetFromDatabaseTable(name, tableName, *parsedRequest)

		if err == nil {
			c.JSON(http.StatusOK, results)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
		}
	}
}

func (a *Api) insertToDatabaseTableHandler(c *gin.Context) {
	body := a.getJsonRawBody(c)

	if body != nil {
		name := c.Param("name")

		err := util.ValidateName(name)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
			return
		}

		tableName := c.Param("tableName")

		err = util.ValidateName(tableName)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
			return
		}

		results, err := a.idb.InsertToDatabaseTable(name, tableName, *body)

		if err == nil {
			c.JSON(http.StatusOK, results)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
		}
	}
}

func (a *Api) removeFromDatabaseTableHandler(c *gin.Context) {
	r := a.getRequest(c)

	if r != nil {
		name := c.Param("name")

		err := util.ValidateName(name)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
			return
		}

		tableName := c.Param("tableName")

		err = util.ValidateName(tableName)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
			return
		}

		parsedRequest, err := parse.Request(*r)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
			return
		}

		results, err := a.idb.RemoveFromDatabaseTable(name, tableName, *parsedRequest)

		if err == nil {
			c.JSON(http.StatusOK, results)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
		}
	}
}

func (a *Api) updateInDatabaseTableHandler(c *gin.Context) {
	body := a.getJsonRawBody(c)

	if body != nil {
		name := c.Param("name")

		err := util.ValidateName(name)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
			return
		}

		tableName := c.Param("tableName")

		err = util.ValidateName(tableName)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprint(err)})
			return
		}

		results, err := a.idb.UpdateInDatabaseTable(name, tableName, *body)

		if err == nil {
			c.JSON(http.StatusOK, results)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprint(err)})
		}
	}
}

func (a *Api) getBody(c *gin.Context) *map[string]interface{} {
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

func (a *Api) getJsonRawBody(c *gin.Context) *map[string]json.RawMessage {
	bytes, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to read body"})
		return nil
	}

	var m map[string]json.RawMessage

	err = json.Unmarshal(bytes, &m)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to parse JSON", "error": err.Error()})
		return nil
	}

	return &m
}

func (a *Api) getRequest(c *gin.Context) *request.Request {
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
