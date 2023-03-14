/*
 * Copyright (c) 2023 Lucas Pape
 */

package util

import (
	"math/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandomStringWithCharset(length int, charset string) string {
	b := make([]byte, length)

	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(b)
}

func RandomString(length int) string {
	return RandomStringWithCharset(length, charset)
}

func RandomFloat() float64 {
	return seededRand.Float64()
}

func RandomBoolean() bool {
	return seededRand.Int63()%2 != 0
}
