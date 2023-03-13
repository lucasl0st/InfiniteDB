/*
 * Copyright (c) 2023 Lucas Pape
 */

package client

import "time"

type Options struct {
	Hostname      string
	Port          uint
	SSL           *bool
	SkipSSLVerify *bool
	AuthKey       *string
	Timeout       *time.Duration
	ReadLimit     *int64
}
