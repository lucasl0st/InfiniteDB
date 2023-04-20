/*
 * Copyright (c) 2023 Lucas Pape
 */

package cmd

import (
	"bufio"
	"fmt"
	"github.com/lucasl0st/InfiniteDB/client"
	_import "github.com/lucasl0st/InfiniteDB/tools/idbimport/import"
	"github.com/lucasl0st/InfiniteDB/util"
	"github.com/spf13/cobra"
	"os"
	"time"
)

// consoleCmd represents the console command
var consoleCmd = &cobra.Command{
	Use:   "console",
	Short: "Import idb database from stdin",
	Long:  "Import idb database from stdin",
	Run: func(cmd *cobra.Command, args []string) {
		err := importFromConsole()

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(consoleCmd)
}

func importFromConsole() error {
	c := client.New(client.Options{
		Hostname:      databaseHostname,
		Port:          databasePort,
		TLS:           util.Ptr(databaseTls),
		SkipTLSVerify: util.Ptr(databaseTlsVerifyDisable),
		AuthKey:       util.Ptr(databaseKey),
		Timeout:       util.Ptr(time.Second * time.Duration(databaseTimeout)),
		ReadLimit:     util.Ptr(databaseReadLimit),
	})

	s := *bufio.NewScanner(os.Stdin)
	i := _import.New(c, s)

	err := i.Import()

	if err != nil {
		return err
	}

	return nil
}
