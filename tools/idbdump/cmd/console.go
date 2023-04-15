/*
 * Copyright (c) 2023 Lucas Pape
 */

package cmd

import (
	"fmt"
	"github.com/lucasl0st/InfiniteDB/client"
	"github.com/lucasl0st/InfiniteDB/tools/idbdump/dump"
	"github.com/lucasl0st/InfiniteDB/util"
	"github.com/spf13/cobra"
	"os"
)

// consoleCmd represents the console command
var consoleCmd = &cobra.Command{
	Use:   "console",
	Short: "Dump idb database into stdout",
	Long:  "Dump idb database into stdout",
	Run: func(cmd *cobra.Command, args []string) {
		err := dumpToConsole()

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(consoleCmd)
}

func dumpToConsole() error {
	c := client.New(client.Options{
		Hostname:      databaseHostname,
		Port:          databasePort,
		TLS:           util.Ptr(databaseTls),
		SkipTLSVerify: util.Ptr(databaseTlsVerifyDisable),
		AuthKey:       util.Ptr(databaseKey),
	})

	r := dump.ConsoleReceiver{}

	d := dump.New(c, r)

	err := d.Dump()

	if err != nil {
		return err
	}

	return nil
}
