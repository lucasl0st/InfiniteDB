package util

import (
	"fmt"
	"log"
)

type LoggerWithPrefix struct {
	Prefix string
}

func (l LoggerWithPrefix) Println(a ...any) {
	log.Println(l.Prefix + " " + fmt.Sprint(a...))
}

func (l LoggerWithPrefix) Print(a ...any) {
	log.Print(l.Prefix + " " + fmt.Sprint(a...))
}

func (l LoggerWithPrefix) Printf(format string, a ...any) {
	log.Println(l.Prefix + " " + fmt.Sprintf(format, a...))
}

func (l LoggerWithPrefix) Fatal(a ...any) {
	log.Fatal(l.Prefix + " " + fmt.Sprint(a...))
}

func (l LoggerWithPrefix) Fatalf(format string, a ...any) {
	log.Fatal(l.Prefix + " " + fmt.Sprintf(format, a...))
}
