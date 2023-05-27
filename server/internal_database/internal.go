/*
 * Copyright (c) 2023 Lucas Pape
 */

package internal_database

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/lucasl0st/InfiniteDB/idblib"
	"github.com/lucasl0st/InfiniteDB/idblib/dbtype"
	"github.com/lucasl0st/InfiniteDB/idblib/field"
	"github.com/lucasl0st/InfiniteDB/idblib/table"
	idbutil "github.com/lucasl0st/InfiniteDB/idblib/util"
	"github.com/lucasl0st/InfiniteDB/models/request"
	"github.com/lucasl0st/InfiniteDB/util"
)

const InternalDatabase = "internal"

const AuthenticationTable = "authentication"
const AuthenticationTableFieldKeyId = "keyID"
const AuthenticationTableFieldKeyValue = "keyValue"

const AuthenticationKeyMain = "main"

var l idbutil.Logger

func SetLogger(logger idbutil.Logger) {
	l = logger
}

func SetupInternalDatabase(idb *idblib.IDB) error {
	r, err := idb.GetDatabases()

	if err != nil {
		return err
	}

	for _, database := range r.Databases {
		if database == InternalDatabase {
			return nil
		}
	}

	_, err = idb.CreateDatabase(InternalDatabase)

	if err != nil {
		return err
	}

	return nil
}

func SetupAuthenticationTable(idb *idblib.IDB) error {
	r, err := idb.GetDatabase(InternalDatabase)

	if err != nil {
		return err
	}

	createTable := true

	for _, t := range r.Tables {
		if t == AuthenticationTable {
			createTable = false
			break
		}
	}

	if createTable {
		fields := make(map[string]field.Field)
		fields[AuthenticationTableFieldKeyId] = field.Field{
			Name:    AuthenticationTableFieldKeyId,
			Indexed: true,
			Unique:  true,
			Type:    dbtype.TEXT,
		}

		fields[AuthenticationTableFieldKeyValue] = field.Field{
			Name:    AuthenticationTableFieldKeyValue,
			Indexed: true,
			Unique:  false,
			Type:    dbtype.TEXT,
		}

		_, err = idb.CreateTableInDatabase(InternalDatabase, AuthenticationTable, fields, request.TableOptions{})

		if err != nil {
			return err
		}
	}

	res, err := idb.GetFromDatabaseTable(InternalDatabase, AuthenticationTable, table.Request{
		Query: &table.Query{
			Where: &request.Where{
				Field:    AuthenticationTableFieldKeyId,
				Operator: request.EQUALS,
				Value:    util.StringToJsonRaw(AuthenticationKeyMain),
			},
		},
	})

	if err != nil {
		return err
	}

	if len(res.Results) == 0 {
		mainKey := uuid.New().String()

		mainKeyObject := make(map[string]json.RawMessage)
		mainKeyObject[AuthenticationTableFieldKeyId] = util.StringToJsonRaw(AuthenticationKeyMain)
		mainKeyObject[AuthenticationTableFieldKeyValue] = util.StringToJsonRaw(mainKey)

		_, err = idb.InsertToDatabaseTable(InternalDatabase, AuthenticationTable, mainKeyObject)

		if err != nil {
			return err
		}

		l.Println("generated main authentication key (will only be shown once): " + mainKey)
	}

	return nil
}

func Authenticated(idb *idblib.IDB, key string) (bool, error) {
	res, err := idb.GetFromDatabaseTable(InternalDatabase, AuthenticationTable, table.Request{
		Query: &table.Query{
			Where: &request.Where{
				Field:    AuthenticationTableFieldKeyValue,
				Operator: request.EQUALS,
				Value:    util.StringToJsonRaw(key),
			},
		},
	})

	if err != nil {
		return false, err
	}

	for _, result := range res.Results {
		s, err := util.JsonRawToString(result[AuthenticationTableFieldKeyValue])

		if err != nil {
			return false, err
		}

		if *s == key {
			return true, nil
		}
	}

	return false, nil
}
