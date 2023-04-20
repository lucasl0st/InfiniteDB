/*
 * Copyright (c) 2023 Lucas Pape
 */

package server

import (
	e "github.com/lucasl0st/InfiniteDB/models/errors"
	"regexp"
)

func validateName(name string) error {
	r, err := regexp.Compile("^[a-zA-Z\\d-_]+$")

	if err != nil {
		return err
	}

	if !r.MatchString(name) {
		return e.NameDoesNotMatchAllowedPattern(name)
	}

	return nil
}
