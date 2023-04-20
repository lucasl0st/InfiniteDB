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
		Name: "create_database",
		Arguments: []Argument{
			{
				Name:        "name",
				Description: "Name of the database",
			},
		},
		Run: runCreateDatabase,
	})
}

func runCreateDatabase(c *client.Client, args []string) error {
	name := args[0]

	res, err := c.CreateDatabase(name)

	if err != nil {
		return err
	}

	fmt.Println(res.Message)

	return nil
}
