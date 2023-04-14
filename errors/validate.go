/*
 * Copyright (c) 2023 Lucas Pape
 */

package errors

import (
	"errors"
	"fmt"
)

func NameDoesNotMatchAllowedPattern(name string) error {
	return errors.New(fmt.Sprintf("the name \"%s\" does not match the allowed pattern", name))
}
