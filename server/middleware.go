/*
 * Copyright (c) 2023 Lucas Pape
 */

package server

import (
	e "github.com/lucasl0st/InfiniteDB/models/errors"
	"github.com/lucasl0st/InfiniteDB/server/internal_database"
	"strings"
)

func CreateDatabaseMiddleware(name string) (bool, func() error) {
	return strings.HasPrefix(name, internal_database.InternalDatabase), func() error {
		return e.DatabaseNameIsReservedForInternalUse()
	}
}
