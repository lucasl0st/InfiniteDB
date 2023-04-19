/*
 * Copyright (c) 2023 Lucas Pape
 */

package methods

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/lucasl0st/InfiniteDB/client"
	"os"
)

func init() {
	Methods = append(Methods, Method{
		Name: "get_database_table",
		Arguments: []Argument{
			{
				Name:        "name",
				Description: "Name of the database",
			},
			{
				Name:        "table-name",
				Description: "Name of the table",
			},
		},
		Run: runGetDatabaseTable,
	})
}

func runGetDatabaseTable(c *client.Client, args []string) error {
	name := args[0]
	tableName := args[1]

	res, err := c.GetDatabaseTable(name, tableName)

	if err != nil {
		return err
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	t.AppendHeader(table.Row{"Field", "Type", "Indexed", "Unique", "Null"})

	for fieldName, field := range res.Fields {
		indexed := false

		if field.Indexed != nil && *field.Indexed {
			indexed = true
		}

		unique := false

		if field.Unique != nil && *field.Unique {
			unique = true
		}

		null := false

		if field.Null != nil && *field.Null {
			null = true
		}

		t.AppendRow(table.Row{
			fieldName,
			field.Type,
			indexed,
			unique,
			null,
		})
	}

	t.Render()

	return nil
}
