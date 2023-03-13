/*
 * Copyright (c) 2023 Lucas Pape
 */

package errors

import (
	"errors"
	"fmt"
)

func TimeoutReceivingDatabaseResult(requestId int64) error {
	return errors.New(fmt.Sprintf("timeout receiving database result for requestId %v", requestId))
}
