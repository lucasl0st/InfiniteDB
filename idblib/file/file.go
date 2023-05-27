/*
 * Copyright (c) 2023 Lucas Pape
 */

package file

import (
	"bufio"
	"os"
	"sort"
)

type File struct {
	path string
}

func New(path string) (*File, error) {
	exists, err := fileExists(path)

	if err != nil {
		return nil, err
	}

	if !exists {
		_, err := os.Create(path)

		if err != nil {
			return nil, err
		}
	}

	return &File{
		path: path,
	}, nil
}

func (f *File) Append(lines []string) error {
	file, err := openWritableFile(f.path)

	if err != nil {
		return err
	}

	defer func() {
		err = file.Close()
	}()

	for _, line := range lines {
		_, err = file.WriteString(line + "\n")

		if err != nil {
			return err
		}
	}

	return err
}

func (f *File) Read(lineNumbers []int64) ([]string, error) {
	if len(lineNumbers) == 0 {
		return nil, nil
	}

	sort.Slice(lineNumbers, func(i, j int) bool {
		return lineNumbers[i] < lineNumbers[j]
	})

	file, err := openReadOnlyFile(f.path)

	if err != nil {
		return nil, err
	}

	defer func() {
		err = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var lines []string

	var i = 0
	var line int64 = 0

	for scanner.Scan() {
		if line == lineNumbers[i] {
			text := scanner.Text()

			for {
				lines = append(lines, text)

				i++

				if i >= len(lineNumbers) {
					break
				}

				if lineNumbers[i] != line {
					break
				}
			}

			if i >= len(lineNumbers) {
				break
			}
		}

		line++
	}

	return lines, err
}

func (f *File) ReadAtStartLine(start int64, readLine func(line string)) error {
	file, err := openReadOnlyFile(f.path)

	if err != nil {
		return err
	}

	defer func() {
		err = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var line int64 = 0

	for scanner.Scan() {
		if line >= start {
			readLine(scanner.Text())
		}

		line++
	}

	return err
}
