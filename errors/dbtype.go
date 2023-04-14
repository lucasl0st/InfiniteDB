/*
 * Copyright (c) 2023 Lucas Pape
 */

package errors

import "errors"

func UnknownDBTypeError() error {
	return errors.New("unknown DBType")
}

func ValueIsNotText() error {
	return errors.New("value is not text")
}

func ValueIsNotNumber() error {
	return errors.New("value is not number")
}

func ValueIsNotBool() error {
	return errors.New("value is not bool")
}
