/*
 * Copyright (c) 2023 Lucas Pape
 */

package dump

import (
	"errors"
	"fmt"
	"github.com/lucasl0st/InfiniteDB/client"
	"github.com/lucasl0st/InfiniteDB/idblib/field"
	"github.com/lucasl0st/InfiniteDB/request"
	"github.com/lucasl0st/InfiniteDB/response"
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
	err := dump.r.WriteStruct(Database{Name: d})

	if err != nil {
		return err
	}

	r, err := dump.c.GetDatabaseTables(d)

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

func (dump *Dump) table(d string, t response.GetDatabaseTablesResponseTable) error {
	err := dump.r.WriteStruct(Table{
		Name:   t.Name,
		Fields: t.Fields,
	})

	if err != nil {
		return err
	}

	max, err := dump.getMaxObjectId(d, t.Name)

	if err != nil {
		return errors.New(fmt.Sprintf("failed to retrieve max object id for table %s in database %s: %s", t.Name, d, err.Error()))
	}

	if *max >= 0 {
		dump.r.ProgressStart(fmt.Sprintf("[%s] %s", d, t.Name), (*max)+1)

		var step int64 = 10000

		for i := int64(0); i <= *max; i += step {
			err = dump.objects(d, t.Name, i, step)

			if err != nil {
				return err
			}

			dump.r.ProgressUpdate(step)
		}

		dump.r.ProgressEnd()
	} else {
		fmt.Println(fmt.Sprintf("table %s in database %s has no objects", t.Name, d))
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
					Parameters: util.Ptr(map[string]interface{}{
						"fieldName": field.InternalObjectIdField,
						"as":        as,
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
		max, ok := r[as].(float64)

		if ok {
			return util.Ptr(int64(max)), nil
		}
	}

	return util.Ptr(int64(-1)), nil
}

func (dump *Dump) getObjects(d string, t string, start int64, count int64) ([]Object, error) {
	res, err := dump.c.GetFromDatabaseTable(d, t, request.Request{
		Query: &request.Query{
			Where: &request.Where{
				Field:    field.InternalObjectIdField,
				Operator: request.BETWEEN,
				Value:    fmt.Sprintf("%v_%v", start-1, start+count),
			},
		},
	})

	var objects []Object

	for _, r := range res.Results {
		objects = append(objects, r)
	}

	return objects, err
}