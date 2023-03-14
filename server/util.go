/*
 * Copyright (c) 2023 Lucas Pape
 */

package server

import "encoding/json"

func toMap(i interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(i)

	if err != nil {
		return make(map[string]interface{}), err
	}

	var m map[string]interface{}

	err = json.Unmarshal(b, &m)

	if err != nil {
		return make(map[string]interface{}), err
	}

	return m, nil
}

func toStruct(i interface{}, r any) error {
	b, err := json.Marshal(i)

	if err != nil {
		return err
	}

	return json.Unmarshal(b, r)
}
