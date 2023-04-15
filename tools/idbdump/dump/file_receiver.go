/*
 * Copyright (c) 2023 Lucas Pape
 */

package dump

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
	"os"
)

var bar *progressbar.ProgressBar

type FileReceiver struct {
	File *os.File
}

func (r FileReceiver) WriteStruct(s interface{}) error {
	str, err := formatStruct(s)

	if err != nil {
		return err
	}

	_, err = r.File.WriteString(*str + "\n")
	return err
}

func (r FileReceiver) ProgressStart(title string, max int64) {
	bar = progressbar.Default(max, fmt.Sprintf(title))
}

func (r FileReceiver) ProgressUpdate(add int64) {
	_ = bar.Add64(add)
}

func (r FileReceiver) ProgressEnd() {
	_ = bar.Finish()
	_ = bar.Exit()
}
