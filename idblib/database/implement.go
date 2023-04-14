/*
 * Copyright (c) 2023 Lucas Pape
 */

package database

import (
	e "github.com/lucasl0st/InfiniteDB/errors"
	"github.com/lucasl0st/InfiniteDB/idblib/object"
	"github.com/lucasl0st/InfiniteDB/idblib/table"
	"github.com/lucasl0st/InfiniteDB/idblib/util"
	"github.com/lucasl0st/InfiniteDB/request"
)

func (d *Database) implement(t *table.Table, implement request.Implement, objects []object.Object) (map[int64]interface{}, *string, error) {
	fromTable := d.tables[implement.From.Table]

	if fromTable == nil {
		return nil, nil, e.TableDoesNotExist()
	}

	implementObjectsMap := map[int64]interface{}{}

	for _, o := range objects {
		i, err := util.DBTypeToInterface(o.M[implement.Field], t.Config.Fields[implement.Field])

		if err != nil {
			return nil, nil, err
		}

		queryObjects, _, err := fromTable.Query(table.Query{
			Where: &request.Where{
				Field:    implement.From.Field,
				Operator: request.EQUALS,
				Value:    i,
			},
		}, nil, nil)

		if err != nil {
			return nil, nil, err
		}

		if len(queryObjects) > 0 {
			forceArray := false

			if implement.ForceArray != nil && *implement.ForceArray {
				forceArray = *implement.ForceArray
			}

			var i interface{}

			var a []map[string]interface{}

			for _, id := range queryObjects {
				o := fromTable.Storage.GetObject(id)

				if o == nil {
					continue
				}

				io, err := fromTable.ObjectToInterfaceMap(*o)

				if err != nil {
					return nil, nil, err
				}

				a = append(a, io)
			}

			if len(a) > 1 || forceArray {
				i = a
			} else {
				i = a[0]
			}

			implementObjectsMap[o.Id] = i
		}
	}

	as := implement.From.Table

	if implement.As != nil {
		as = *implement.As
	}

	return implementObjectsMap, &as, nil
}
