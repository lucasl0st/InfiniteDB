/*
 * Copyright (c) 2023 Lucas Pape
 */

package errors

import (
	"errors"
	"fmt"
)

func FailedToSetupInternalDatabase(err error) error {
	return errors.New(fmt.Sprintf("error setting up internal database: %s", err.Error()))
}

func FailedToSetupInternalAuthenticationTable(err error) error {
	return errors.New(fmt.Sprintf("failed to setup internal authentication table: %s", err.Error()))
}

func DatabaseNameIsReservedForInternalUse() error {
	return errors.New("database name is reserved for internal use")
}
