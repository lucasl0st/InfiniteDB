/*
 * Copyright (c) 2023 Lucas Pape
 */

package integration_tests

import (
	"archive/zip"
	"bufio"
	"errors"
	"fmt"
	"github.com/dimchansky/utfbom"
	"github.com/lucasl0st/InfiniteDB/client"
	idbutil "github.com/lucasl0st/InfiniteDB/idblib/util"
	"github.com/lucasl0st/InfiniteDB/models/request"
	"github.com/lucasl0st/InfiniteDB/util"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"vimagination.zapto.org/dos2unix"
)

const covid19GermanyUrl = "https://files.l0stnet.xyz/covid19-germany.zip"

const databaseName = "covid19"

type Data struct {
	State     string
	County    string
	AgeGroup  string
	Gender    string
	Date      string
	Cases     int64
	Deaths    int64
	Recovered int64
}

func runCovid19Tests(c *client.Client) error {
	f, size, err := download(covid19GermanyUrl)

	if err != nil {
		return err
	}

	dir, err := unpack(f, size)

	if err != nil {
		return err
	}

	err = f.Close()

	if err != nil {
		return err
	}

	csvFile, err := os.OpenFile(dir+"/covid_de.csv", os.O_RDONLY, os.ModePerm)

	if err != nil {
		return err
	}

	defer csvFile.Close()

	scanner := bufio.NewScanner(csvFile)

	var d []Data

	first := true

	for scanner.Scan() {
		if first {
			first = false
			continue
		}

		fields := strings.Split(scanner.Text(), ",")

		state := fields[0]
		county := fields[1]
		ageGroup := fields[2]
		gender := fields[3]
		date := fields[4]

		cases, err := strconv.ParseInt(fields[5], 10, 64)

		if err != nil {
			return err
		}

		deaths, err := strconv.ParseInt(fields[6], 10, 64)

		if err != nil {
			return err
		}

		recovered, err := strconv.ParseInt(fields[7], 10, 64)

		if err != nil {
			return err
		}

		d = append(d, Data{
			State:     state,
			County:    county,
			AgeGroup:  ageGroup,
			Gender:    gender,
			Date:      date,
			Cases:     cases,
			Deaths:    deaths,
			Recovered: recovered,
		})
	}

	err = scanner.Err()

	if err != nil {
		return err
	}

	err = createDatabase(c)

	if err != nil {
		return err
	}

	err = runObjectTests(c, d)

	if err != nil {
		return err
	}

	return nil
}

func createDatabase(c *client.Client) error {
	res, err := c.CreateDatabase(databaseName)

	if err != nil {
		return err
	}

	if res.Name != databaseName {
		return errors.New(fmt.Sprintf("database responded with database name %s, expected %s", res.Name, databaseName))
	}

	return nil
}

func runObjectTests(c *client.Client, objects []Data) error {
	fmt.Println("running covid19 test max-cases")

	err := maxCases(c, objects)

	if err != nil {
		return err
	}

	fmt.Println("test covid19 max-cases successful")

	return nil
}

func maxCases(c *client.Client, objects []Data) error {
	cases := map[string]int64{}

	for _, o := range objects {
		m, ok := cases[o.State]

		if !ok {
			cases[o.State] = o.Cases
			continue
		}

		if o.Cases > m {
			cases[o.State] = o.Cases
		}
	}

	tableName := "max_cases"
	fieldNameState := "state"
	fieldNameMaxCases := "max_cases"

	_, err := c.CreateTableInDatabase(databaseName, tableName, map[string]request.Field{
		fieldNameState: {
			Type:    "text",
			Indexed: util.Ptr(true),
		},
		fieldNameMaxCases: {
			Type:    "number",
			Indexed: util.Ptr(true),
		},
	}, nil)

	if err != nil {
		return err
	}

	for state, max := range cases {
		_, err = c.InsertToDatabaseTable(databaseName, tableName, idbutil.InterfaceMapToJsonRawMap(map[string]interface{}{
			fieldNameState:    state,
			fieldNameMaxCases: max,
		}))

		if err != nil {
			return err
		}
	}

	res, err := c.GetFromDatabaseTable(databaseName, tableName, request.Request{
		Query: &request.Query{
			Where: &request.Where{
				Field:    fieldNameState,
				Operator: request.NOT,
				Value:    nil,
			},
		},
	})

	if err != nil {
		return err
	}

	for _, r := range res.Results {
		m := idbutil.JsonRawMapToInterfaceMap(r)

		if m[fieldNameState] != cases[fieldNameState] {
			return errors.New(fmt.Sprintf("expected %v, got %v", cases[fieldNameState], m[fieldNameState]))
		}
	}

	return nil
}

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

	defer r.Body.Close()

	size, err := io.Copy(f, r.Body)

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

		df.Close()
		af.Close()
	}

	err = os.Remove(f.Name())

	return d, err
}
