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
		Name: "get_database",
		Arguments: []Argument{
			{
				Name:        "name",
				Description: "Name of the database",
			},
		},
		Run: runGetDatabase,
	})
}

func runGetDatabase(c *client.Client, args []string) error {
	name := args[0]

	res, err := c.GetDatabase(name)

	if err != nil {
		return err
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	t.AppendHeader(table.Row{"#", "Table Name"})

	for i, dt := range res.Tables {
		t.AppendRow(table.Row{
			i,
			dt,
		})
	}

	t.Render()

	return nil
}
