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

func DidNotReceiveDatabaseVersion() error {
	return errors.New("did not receive database version")
}

func ClientNotCompatibleWithDatabaseServer(serverVersion string, clientVersion string) error {
	return errors.New(fmt.Sprintf("this client is not compatible with the database server, server version: %s client version: %s", serverVersion, clientVersion))
}

func ClientNotConnected() error {
	return errors.New("client is not connected")
}
