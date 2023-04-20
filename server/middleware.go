/*
 * Copyright (c) 2023 Lucas Pape
 */

package main

import (
	e "github.com/lucasl0st/InfiniteDB/models/errors"
	"strings"
)

func CreateDatabaseMiddleware(name string) (bool, func() error) {
	return strings.HasPrefix(name, InternalDatabase), func() error {
		return e.DatabaseNameIsReservedForInternalUse()
	}
}
