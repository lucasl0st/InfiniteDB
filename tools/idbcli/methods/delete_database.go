/*
 * Copyright (c) 2023 Lucas Pape
 */

package methods

import (
	"fmt"
	"github.com/lucasl0st/InfiniteDB/client"
)

func init() {
	Methods = append(Methods, Method{
		Name: "delete_database",
		Arguments: []Argument{
			{
				Name:        "name",
				Description: "Name of the database",
			},
		},
		Run: runDeleteDatabase,
	})
}

func runDeleteDatabase(c *client.Client, args []string) error {
	name := args[0]

	res, err := c.DeleteDatabase(name)

	if err != nil {
		return err
	}

	fmt.Println(res.Message)

	return nil
}
