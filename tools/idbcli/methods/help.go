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
		Name: "help",
		Run:  runHelp,
	})
}

func runHelp(c *client.Client, args []string) error {
	fmt.Print("available commands are: ")

	for _, method := range Methods {
		fmt.Print(method.Name + ", ")
	}

	fmt.Print("\n")

	return nil
}
