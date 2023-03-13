/*
 * Copyright (c) 2023 Lucas Pape
 */

package errors

import "errors"

func DontHaveLock() error {
	return errors.New("don't have lock on file")
}
