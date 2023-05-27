/*
 * Copyright (c) 2023 Lucas Pape
 */

package covid19

import (
	"archive/zip"
	"errors"
	"github.com/dimchansky/utfbom"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"vimagination.zapto.org/dos2unix"
)

func download(url string) (*os.File, int64, error) {
	f, err := os.CreateTemp("", "covid-19")

	if err != nil {
		return nil, 0, err
	}

	c := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	r, err := c.Get(url)

	if err != nil {
		return nil, 0, err
	}

	defer func() {
		err = r.Body.Close()
	}()

	size, err := io.Copy(f, r.Body)

	if err != nil {
		return f, size, err
	}

	return f, size, err
}

func unpack(f *os.File, size int64) (string, error) {
	d, err := os.MkdirTemp("", "covid-19")

	z, err := zip.NewReader(f, size)

	if err != nil {
		return d, err
	}

	for _, f := range z.File {
		p := filepath.Join(d, f.Name)

		if !strings.HasPrefix(p, filepath.Clean(d)+string(os.PathSeparator)) {
			return d, errors.New("invalid file path in zip")
		}

		if f.FileInfo().IsDir() {
			err = os.MkdirAll(p, os.ModePerm)

			if err != nil {
				return d, err
			}

			continue
		}

		err = os.MkdirAll(filepath.Dir(p), os.ModePerm)

		if err != nil {
			return "", err
		}

		df, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())

		if err != nil {
			return d, err
		}

		af, err := f.Open()

		if err != nil {
			return d, err
		}

		_, err = io.Copy(df, utfbom.SkipOnly(dos2unix.DOS2Unix(af)))

		if err != nil {
			return d, err
		}

		err = df.Close()

		if err != nil {
			return d, err
		}

		err = af.Close()

		if err != nil {
			return d, err
		}
	}

	err = os.Remove(f.Name())

	return d, err
}
