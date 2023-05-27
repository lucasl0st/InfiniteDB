/*
 * Copyright (c) 2023 Lucas Pape
 */

package file

import "os"

func openWritableFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

func openReadOnlyFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE, 0644)
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}
