/*
 * Copyright (c) 2023 Lucas Pape
 */

package covid19

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/lucasl0st/InfiniteDB/client"
	"github.com/lucasl0st/InfiniteDB/models/request"
	"github.com/lucasl0st/InfiniteDB/util"
	"os"
	"strconv"
	"strings"
)

const covid19GermanyUrl = "https://files.l0stnet.xyz/covid19-germany.zip"
const databaseName = "covid19"

func Run(c *client.Client) error {
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

	defer func() {
		err = csvFile.Close()
	}()

	scanner := bufio.NewScanner(csvFile)

	var d []Data

	first := true

	for scanner.Scan() {
		if first {
			first = false
			continue
		}

		data, err := parseLine(scanner.Text())

		if err != nil {
			return err
		}

		d = append(d, *data)
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

	return err
}

func parseLine(line string) (*Data, error) {
	fields := strings.Split(line, ",")

	state := fields[0]
	county := fields[1]
	ageGroup := fields[2]
	gender := fields[3]
	date := fields[4]

	cases, err := strconv.ParseInt(fields[5], 10, 64)

	if err != nil {
		return nil, err
	}

	deaths, err := strconv.ParseInt(fields[6], 10, 64)

	if err != nil {
		return nil, err
	}

	recovered, err := strconv.ParseInt(fields[7], 10, 64)

	if err != nil {
		return nil, err
	}

	return &Data{
		State:     state,
		County:    county,
		AgeGroup:  ageGroup,
		Gender:    gender,
		Date:      date,
		Cases:     cases,
		Deaths:    deaths,
		Recovered: recovered,
	}, nil
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
		_, err = c.InsertToDatabaseTable(databaseName, tableName, util.InterfaceMapToJsonRawMap(map[string]interface{}{
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
		m := util.JsonRawMapToInterfaceMap(r)

		if int64(m[fieldNameMaxCases].(float64)) != cases[m[fieldNameState].(string)] {
			return errors.New(fmt.Sprintf("expected %v, got %v", cases[m[fieldNameState].(string)], m[fieldNameMaxCases]))
		}
	}

	return nil
}
