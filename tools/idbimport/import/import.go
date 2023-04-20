/*
 * Copyright (c) 2023 Lucas Pape
 */

package _import

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lucasl0st/InfiniteDB/client"
	"github.com/lucasl0st/InfiniteDB/tools/util"
	"strings"
)

type Import struct {
	c      *client.Client
	reader bufio.Scanner

	currentDatabase string
	currentTable    string
}

func New(c *client.Client, reader bufio.Scanner) *Import {
	return &Import{
		c:               c,
		reader:          reader,
		currentDatabase: "",
		currentTable:    "",
	}
}

func (i *Import) Import() error {
	err := i.c.Connect()

	if err != nil {
		return errors.New(fmt.Sprintf("failed to connect to database: %s", err.Error()))
	}

	for i.reader.Scan() {
		l := i.reader.Text()
		err = i.processLine(l)

		if err != nil {
			return err
		}
	}

	return nil
}

const (
	databasePrefix = "//Database:"
	tablePrefix    = "//Table:"
	objectPrefix   = "//Object:"
)

func (i *Import) processLine(l string) error {

	if strings.HasPrefix(l, databasePrefix) {
		return i.processDatabase(strings.ReplaceAll(l, databasePrefix, ""))
	} else if strings.HasPrefix(l, tablePrefix) {
		return i.processTable(strings.ReplaceAll(l, tablePrefix, ""))
	} else if strings.HasPrefix(l, objectPrefix) {
		return i.processObject(strings.ReplaceAll(l, objectPrefix, ""))
	} else {
		return errors.New("unknown input")
	}
}

func (i *Import) processDatabase(s string) error {
	var database util.Database

	err := json.Unmarshal([]byte(s), &database)

	if err != nil {
		return err
	}

	_, err = i.c.CreateDatabase(database.Name)

	if err != nil {
		return errors.New(fmt.Sprintf("error creating database %s: %s", database.Name, err.Error()))
	}

	i.currentDatabase = database.Name

	return nil
}

func (i *Import) processTable(s string) error {
	if len(i.currentDatabase) == 0 {
		return errors.New("database needs to be defined before defining a table")
	}

	var table util.Table

	err := json.Unmarshal([]byte(s), &table)

	if err != nil {
		return err
	}

	_, err = i.c.CreateTableInDatabase(i.currentDatabase, table.Name, table.Fields, &table.Options)

	if err != nil {
		return errors.New(fmt.Sprintf("error creating table %s: %s", table.Name, err.Error()))
	}

	i.currentTable = table.Name

	return nil
}

func (i *Import) processObject(s string) error {
	if len(i.currentDatabase) == 0 {
		return errors.New("database needs to be defined before defining an object")
	}

	if len(i.currentTable) == 0 {
		return errors.New("table needs to be defined before defining an object")
	}

	var object util.Object

	err := json.Unmarshal([]byte(s), &object)

	if err != nil {
		return err
	}

	_, err = i.c.InsertToDatabaseTable(i.currentDatabase, i.currentTable, object)

	if err != nil {
		return err
	}

	return nil
}
