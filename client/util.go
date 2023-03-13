/*
 * Copyright (c) 2023 Lucas Pape
 */

package client

import "encoding/json"

func mapToStruct[Type any](m map[string]interface{}, r *Type) error {
	b, err := json.Marshal(m)

	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &r)

	if err != nil {
		return err
	}

	return nil
}

func ptr[T any](v T) *T {
	return &v
}
