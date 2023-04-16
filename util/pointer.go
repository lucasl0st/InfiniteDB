/*
 * Copyright (c) 2023 Lucas Pape
 */

package util

func Ptr[T any](v T) *T {
	return &v
}
