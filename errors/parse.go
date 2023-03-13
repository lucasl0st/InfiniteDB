/*
 * Copyright (c) 2023 Lucas Pape
 */

package errors

import (
	"errors"
	"fmt"
)

func CannotHaveAndANDOrInOneQuery() error {
	return errors.New("cannot have AND and OR in one query")
}

func NotAValidFunction() error {
	return errors.New("not a valid function")
}

func TypeNotSupported() error {
	return errors.New("type not supported")
}

func FieldCannotBeUniqueWithoutBeingIndexed() error {
	return errors.New("field cannot be unique without being indexed")
}

func IsNotAString(param string) error {
	return errors.New(fmt.Sprintf("%s is not a string", param))
}

func IsNotAMap(param string) error {
	return errors.New(fmt.Sprintf("%s is not a map", param))
}

func IsNotANumber(param string) error {
	return errors.New(fmt.Sprintf("%s is not a number", param))
}
