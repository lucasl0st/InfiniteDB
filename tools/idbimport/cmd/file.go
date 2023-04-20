/*
 * Copyright (c) 2023 Lucas Pape
 */

package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/lucasl0st/InfiniteDB/client"
	_import "github.com/lucasl0st/InfiniteDB/tools/idbimport/import"
	"github.com/lucasl0st/InfiniteDB/util"
	"github.com/spf13/cobra"
	"os"
	"time"
)

// fileCmd represents the file command
var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Import idb database from a single file",
	Long:  "Import idb database from a single file",
	Run: func(cmd *cobra.Command, args []string) {
		err := importFromFile()

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

var (
	inputFile string
)

func init() {
	fileCmd.Flags().SortFlags = false

	fileCmd.Flags().StringVarP(&inputFile, "input-file", "i", "", "input dump file")

	_ = fileCmd.MarkFlagFilename("input-file")
	_ = fileCmd.MarkFlagRequired("input-file")

	rootCmd.AddCommand(fileCmd)
}

func importFromFile() error {
	c := client.New(client.Options{
		Hostname:      databaseHostname,
		Port:          databasePort,
		TLS:           util.Ptr(databaseTls),
		SkipTLSVerify: util.Ptr(databaseTlsVerifyDisable),
		AuthKey:       util.Ptr(databaseKey),
		Timeout:       util.Ptr(time.Second * time.Duration(databaseTimeout)),
		ReadLimit:     util.Ptr(databaseReadLimit),
	})

	if _, err := os.Stat(inputFile); err != nil {
		return errors.New(fmt.Sprintf("the input file %s does not exist", inputFile))
	}

	f, err := os.OpenFile(inputFile, os.O_RDONLY, os.ModePerm)

	if err != nil {
		return err
	}

	defer func() {
		err = f.Close()
	}()

	s := *bufio.NewScanner(f)
	i := _import.New(c, s)

	err = i.Import()

	if err != nil {
		return err
	}

	return err
}
