/*
 * Copyright (c) 2023 Lucas Pape
 */

package dump

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lucasl0st/InfiniteDB/client"
	"github.com/lucasl0st/InfiniteDB/idblib/field"
	idbutil "github.com/lucasl0st/InfiniteDB/idblib/util"
	"github.com/lucasl0st/InfiniteDB/models/request"
	"github.com/lucasl0st/InfiniteDB/server"
	toolsutil "github.com/lucasl0st/InfiniteDB/tools/util"
	"github.com/lucasl0st/InfiniteDB/util"
)

type Dump struct {
	c *client.Client
	r Receiver
}

func New(c *client.Client, r Receiver) *Dump {
	return &Dump{
		c: c,
		r: r,
	}
}

func (dump *Dump) Dump() error {
	err := dump.c.Connect()

	if err != nil {
		return errors.New(fmt.Sprintf("failed to connect to database: %s", err.Error()))
	}

	r, err := dump.c.GetDatabases()

	if err != nil {
		return errors.New(fmt.Sprintf("failed to retrieve databases: %s", err.Error()))
	}

	for _, d := range r.Databases {
		err = dump.database(d)

		if err != nil {
			return err
		}
	}

	return nil
}

func (dump *Dump) database(d string) error {
	if d == server.InternalDatabase {
		return nil
	}

	err := dump.r.WriteStruct(toolsutil.Database{Name: d})

	if err != nil {
		return err
	}

	r, err := dump.c.GetDatabase(d)

	if err != nil {
		return errors.New(fmt.Sprintf("failed to retrieve tables for database %s: %s", d, err.Error()))
	}

	for _, t := range r.Tables {
		err = dump.table(d, t)

		if err != nil {
			return err
		}
	}

	return nil
}

func (dump *Dump) table(d string, t string) error {
	r, err := dump.c.GetDatabaseTable(d, t)

	if err != nil {
		return err
	}

	err = dump.r.WriteStruct(toolsutil.Table{
		Name:   r.TableName,
		Fields: r.Fields,
	})

	if err != nil {
		return err
	}

	max, err := dump.getMaxObjectId(d, r.TableName)

	if err != nil {
		return errors.New(fmt.Sprintf("failed to retrieve max object id for table %s in database %s: %s", r.TableName, d, err.Error()))
	}

	if *max >= 0 {
		dump.r.ProgressStart(fmt.Sprintf("[%s] %s", d, r.TableName), (*max)+1)

		var step int64 = 10000

		for i := int64(0); i <= *max; i += step {
			err = dump.objects(d, r.TableName, i, step)

			if err != nil {
				return err
			}

			dump.r.ProgressUpdate(step)
		}

		dump.r.ProgressEnd()
	}

	return nil
}

func (dump *Dump) objects(d string, t string, start int64, count int64) error {
	objects, err := dump.getObjects(d, t, start, count)

	if err != nil {
		return errors.New(fmt.Sprintf("failed to retrieve objects in table %s of database %s", t, d))
	}

	for _, o := range objects {
		err = dump.r.WriteStruct(o)

		if err != nil {
			return err
		}
	}

	return nil
}

func (dump *Dump) getMaxObjectId(d string, t string) (*int64, error) {
	as := "maxInternalObjectId"

	res, err := dump.c.GetFromDatabaseTable(d, t, request.Request{
		Query: &request.Query{
			Where: &request.Where{
				Field:    field.InternalObjectIdField,
				Operator: request.NOT,
				Value:    nil,
			},
			Functions: []request.Function{
				{
					Function: "max",
					Parameters: util.Ptr(map[string]json.RawMessage{
						"fieldName": idbutil.StringToJsonRaw(field.InternalObjectIdField),
						"as":        idbutil.StringToJsonRaw(as),
					}),
				},
			},
		},
		Limit: util.Ptr(int64(1)),
	})

	if err != nil {
		return nil, err
	}

	for _, r := range res.Results {
		m := idbutil.JsonRawMapToInterfaceMap(r)

		max, ok := m[as].(float64)

		if ok {
			return util.Ptr(int64(max)), nil
		}
	}

	return util.Ptr(int64(-1)), nil
}

func (dump *Dump) getObjects(d string, t string, start int64, count int64) ([]toolsutil.Object, error) {
	res, err := dump.c.GetFromDatabaseTable(d, t, request.Request{
		Query: &request.Query{
			Where: &request.Where{
				Field:    field.InternalObjectIdField,
				Operator: request.BETWEEN,
				Value:    idbutil.StringToJsonRaw(fmt.Sprintf("%v_%v", start-1, start+count)),
			},
		},
	})

	var objects []toolsutil.Object

	for _, r := range res.Results {
		objects = append(objects, r)
	}

	return objects, err
}
