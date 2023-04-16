/*
 * Copyright (c) 2023 Lucas Pape
 */

package cmd

import (
	"errors"
	"fmt"
	"github.com/lucasl0st/InfiniteDB/client"
	"github.com/lucasl0st/InfiniteDB/tools/idbdump/dump"
	"github.com/lucasl0st/InfiniteDB/util"
	"github.com/spf13/cobra"
	"os"
	"time"
)

// fileCmd represents the file command
var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Dump idb database into a single file",
	Long:  "Dump idb database into a single file",
	Run: func(cmd *cobra.Command, args []string) {
		err := dumpToFile()

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

var (
	outputFile string
)

func init() {
	fileCmd.Flags().SortFlags = false

	fileCmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "output file for dump")

	_ = fileCmd.MarkFlagFilename("output-file")
	_ = fileCmd.MarkFlagRequired("output-file")

	rootCmd.AddCommand(fileCmd)
}

func dumpToFile() error {
	c := client.New(client.Options{
		Hostname:      databaseHostname,
		Port:          databasePort,
		TLS:           util.Ptr(databaseTls),
		SkipTLSVerify: util.Ptr(databaseTlsVerifyDisable),
		AuthKey:       util.Ptr(databaseKey),
		Timeout:       util.Ptr(time.Second * time.Duration(databaseTimeout)),
		ReadLimit:     util.Ptr(databaseReadLimit),
	})

	if _, err := os.Stat(outputFile); err == nil {
		return errors.New(fmt.Sprintf("the output file %s already exists, not overwriting", outputFile))
	}

	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)

	defer func() {
		err = f.Close()
	}()

	if err != nil {
		return err
	}

	r := dump.FileReceiver{File: f}

	d := dump.New(c, r)

	err = d.Dump()

	if err != nil {
		return err
	}

	return err
}
