/*
 * Copyright (c) 2023 Lucas Pape
 */

package dump

import (
	"fmt"
)

type ConsoleReceiver struct {
}

func (r ConsoleReceiver) WriteStruct(s interface{}) error {
	str, err := formatStruct(s)

	if err != nil {
		return err
	}

	fmt.Println(*str)

	return nil
}

func (r ConsoleReceiver) ProgressStart(title string, max int64) {
}

func (r ConsoleReceiver) ProgressUpdate(add int64) {
}

func (r ConsoleReceiver) ProgressEnd() {
}
