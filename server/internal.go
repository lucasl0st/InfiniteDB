/*
 * Copyright (c) 2023 Lucas Pape
 */

package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/lucasl0st/InfiniteDB/idblib"
	"github.com/lucasl0st/InfiniteDB/idblib/field"
	"github.com/lucasl0st/InfiniteDB/idblib/table"
	"github.com/lucasl0st/InfiniteDB/idblib/util"
	"github.com/lucasl0st/InfiniteDB/models/request"
)

const InternalDatabase = "internal"

const AuthenticationTable = "authentication"
const AuthenticationTableFieldKeyId = "keyID"
const AuthenticationTableFieldKeyValue = "keyValue"

const AuthenticationKeyMain = "main"

func SetupAuthenticationTable(idb *idblib.IDB) error {
	r, err := idb.GetDatabase(InternalDatabase)

	if err != nil {
		return err
	}

	for _, t := range r.Tables {
		if t == AuthenticationTable {
			return nil
		}
	}

	fields := make(map[string]field.Field)
	fields[AuthenticationTableFieldKeyId] = field.Field{
		Name:    AuthenticationTableFieldKeyId,
		Indexed: true,
		Unique:  true,
		Type:    field.TEXT,
	}

	fields[AuthenticationTableFieldKeyValue] = field.Field{
		Name:    AuthenticationTableFieldKeyValue,
		Indexed: true,
		Unique:  false,
		Type:    field.TEXT,
	}

	_, err = idb.CreateTableInDatabase(InternalDatabase, AuthenticationTable, fields, request.TableOptions{})

	if err != nil {
		return err
	}

	mainKey := uuid.New().String()

	mainKeyObject := make(map[string]json.RawMessage)
	mainKeyObject[AuthenticationTableFieldKeyId] = util.StringToJsonRaw(AuthenticationKeyMain)
	mainKeyObject[AuthenticationTableFieldKeyValue] = util.StringToJsonRaw(mainKey)

	_, err = idb.InsertToDatabaseTable(InternalDatabase, AuthenticationTable, mainKeyObject)

	if err != nil {
		return err
	}

	l.Println("generated main authentication key (will only be shown once): " + mainKey)

	return nil
}

func authenticated(idb *idblib.IDB, key string) (bool, error) {
	r := table.Request{
		Query: &table.Query{
			Where: &request.Where{
				Field:    AuthenticationTableFieldKeyValue,
				Operator: request.EQUALS,
				Value:    util.StringToJsonRaw(key),
			},
		},
	}

	res, err := idb.GetFromDatabaseTable(InternalDatabase, AuthenticationTable, r)

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
