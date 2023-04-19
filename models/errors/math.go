/*
 * Copyright (c) 2023 Lucas Pape
 */

package errors

import "errors"

func CouldNotParseFormula() error {
	return errors.New("could not parse formula")
}

func CannotDivideByZero() error {
	return errors.New("cannot divide by zero")
}
