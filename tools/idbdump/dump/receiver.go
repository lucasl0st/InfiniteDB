/*
 * Copyright (c) 2023 Lucas Pape
 */

package dump

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type Receiver interface {
	WriteStruct(s interface{}) error

	ProgressStart(title string, max int64)
	ProgressUpdate(add int64)
	ProgressEnd()
}

func formatStruct(s interface{}) (*string, error) {
	b, err := json.Marshal(s)

	if err != nil {
		return nil, err
	}

	str := fmt.Sprintf("//%s:%s", getStructName(s), string(b))
	return &str, nil
}

func getStructName(s interface{}) string {
	if t := reflect.TypeOf(s); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}
