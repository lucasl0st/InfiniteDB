/*
 * Copyright (c) 2023 Lucas Pape
 */

package main

import (
	"errors"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	DatabasePath       string `env:"DATABASE_PATH" envDefault:"/var/lib/infinitedb/"`
	Authentication     bool   `env:"AUTHENTICATION" envDefault:"true"`
	Port               uint   `env:"PORT" envDefault:"8080"`
	RequestLogging     bool   `env:"REQUEST_LOGGING" envDefault:"false"`
	CacheSize          uint   `env:"CACHE_SIZE" envDefault:"1000"`
	TLS                bool   `env:"TLS" envDefault:"false"`
	TLSCert            string `env:"TLS_CERT"`
	TLSKey             string `env:"TLS_KEY"`
	WebsocketReadLimit int64  `env:"WEBSOCKET_READ_LIMIT" envDefault:"10000000"`
}

func LoadConfig() (*Config, error) {
	c := Config{}

	err := env.Parse(&c)

	if c.TLS {
		if len(c.TLSCert) == 0 {
			return nil, errors.New("TLS enabled but no TLS_CERT provided")
		}

		if len(c.TLSKey) == 0 {
			return nil, errors.New("TLS enabled but no TLS_KEY provided")
		}
	}

	return &c, err
}
