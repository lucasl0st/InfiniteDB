/*
 * Copyright (c) 2023 Lucas Pape
 */

package server

import (
	"errors"
	"regexp"
)

func validateName(name string) error {
	r, err := regexp.Compile("^[a-zA-Z\\d-_]+$")

	if err != nil {
		return err
	}

	if !r.MatchString(name) {
		return errors.New("name does not match allowed pattern")
	}

	return nil
}
