/*
 * Copyright (c) 2023 Lucas Pape
 */

package util

import "encoding/json"

func ToMap(i interface{}) (map[string]json.RawMessage, error) {
	b, err := json.Marshal(i)

	if err != nil {
		return make(map[string]json.RawMessage), err
	}

	var m map[string]json.RawMessage

	err = json.Unmarshal(b, &m)

	if err != nil {
		return make(map[string]json.RawMessage), err
	}

	return m, nil
}

func ToStruct(i interface{}, r any) error {
	b, err := json.Marshal(i)

	if err != nil {
		return err
	}

	return json.Unmarshal(b, r)
}
