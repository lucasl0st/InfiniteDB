/*
 * Copyright (c) 2023 Lucas Pape
 */

package database

import (
	"encoding/json"
	e "github.com/lucasl0st/InfiniteDB/errors"
	"github.com/lucasl0st/InfiniteDB/idblib/object"
	"github.com/lucasl0st/InfiniteDB/idblib/table"
	"github.com/lucasl0st/InfiniteDB/request"
)

func (d *Database) implement(t *table.Table, implement request.Implement, objects []object.Object) (map[int64]json.RawMessage, *string, error) {
	fromTable := d.tables[implement.From.Table]

	if fromTable == nil {
		return nil, nil, e.TableDoesNotExist()
	}

	implementObjectsMap := map[int64]json.RawMessage{}

	for _, o := range objects {
		i := o.M[implement.Field].ToJsonRaw()

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

			var i json.RawMessage

			var a []map[string]json.RawMessage

			for _, id := range queryObjects {
				o := fromTable.Storage.GetObject(id)

				if o == nil {
					continue
				}

				io, err := fromTable.ObjectToJsonRawMap(*o)

				if err != nil {
					return nil, nil, err
				}

				a = append(a, io)
			}

			var b []byte

			if len(a) > 1 || forceArray {
				b, err = json.Marshal(a)
			} else {
				b, err = json.Marshal(a[0])
			}

			if err != nil {
				return nil, nil, err
			}

			i = b

			implementObjectsMap[o.Id] = i
		}
	}

	as := implement.From.Table

	if implement.As != nil {
		as = *implement.As
	}

	return implementObjectsMap, &as, nil
}
