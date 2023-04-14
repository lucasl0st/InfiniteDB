/*
 * Copyright (c) 2023 Lucas Pape
 */

package util

type Logger interface {
	Println(a ...any)
	Print(a ...any)
	Printf(format string, a ...any)
	Fatal(a ...any)
	Fatalf(format string, a ...any)
}
