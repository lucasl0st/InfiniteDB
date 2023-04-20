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
		Name:      "get_databases",
		Arguments: []Argument{},
		Run:       runGetDatabases,
	})
}

func runGetDatabases(c *client.Client, _ []string) error {
	res, err := c.GetDatabases()

	if err != nil {
		return err
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	t.AppendHeader(table.Row{"#", "Database Name"})

	for i, database := range res.Databases {
		t.AppendRow(table.Row{
			i,
			database,
		})
	}

	t.Render()

	return nil
}
