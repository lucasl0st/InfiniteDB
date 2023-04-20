/*
 * Copyright (c) 2023 Lucas Pape
 */

package errors

import "errors"

func DatabaseAlreadyExists() error {
	return errors.New("database already exists")
}

func DatabaseDoesNotExist() error {
	return errors.New("database does not exist")
}

func TableAlreadyExists() error {
	return errors.New("table already exists")
}

func TableDoesNotExist() error {
	return errors.New("table does not exist")
}
