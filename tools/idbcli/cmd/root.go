/*
 * Copyright (c) 2023 Lucas Pape
 */

package cmd

import (
	"bufio"
	"fmt"
	"github.com/lucasl0st/InfiniteDB/client"
	m "github.com/lucasl0st/InfiniteDB/idbcli/methods"
	"github.com/lucasl0st/InfiniteDB/util"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "idbcli",
	Short: "Connect to InfiniteDB Server interactively",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		err := runRootCmd()

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

var (
	databaseHostname         string
	databasePort             uint
	databaseTls              bool
	databaseTlsVerifyDisable bool
	databaseKey              string
	databaseTimeout          int
	databaseReadLimit        int64
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().SortFlags = false
	rootCmd.PersistentFlags().SortFlags = false

	rootCmd.PersistentFlags().StringVarP(&databaseHostname, "database-hostname", "a", "127.0.0.1", "hostname of database, for example 127.0.0.1")
	rootCmd.PersistentFlags().UintVarP(&databasePort, "database-port", "p", 8080, "port of database, for example 8080")
	rootCmd.PersistentFlags().BoolVarP(&databaseTls, "database-tls", "s", false, "connect to database with SSL")
	rootCmd.PersistentFlags().BoolVar(&databaseTlsVerifyDisable, "database-disable-tls-verify", false, "disable TLS certificate verification for database")
	rootCmd.PersistentFlags().StringVarP(&databaseKey, "database-key", "k", "", "key for database if authentication is enabled")
	rootCmd.PersistentFlags().IntVar(&databaseTimeout, "database-timeout", 10, "timeout for receiving database results in seconds")
	rootCmd.PersistentFlags().Int64Var(&databaseReadLimit, "database-read-limit", int64(1000*1000*1000), "read limit for database results in bytes")
}

func runRootCmd() error {
	c := client.New(client.Options{
		Hostname:      databaseHostname,
		Port:          databasePort,
		TLS:           util.Ptr(databaseTls),
		SkipTLSVerify: util.Ptr(databaseTlsVerifyDisable),
		AuthKey:       util.Ptr(databaseKey),
		Timeout:       util.Ptr(time.Second * time.Duration(databaseTimeout)),
		ReadLimit:     util.Ptr(databaseReadLimit),
	})

	err := c.Connect()

	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")

		input, err := reader.ReadString('\n')

		if err != nil {
			return err
		}

		err = runInput(strings.ReplaceAll(input, "\n", ""), reader, c)

		if err != nil {
			return err
		}
	}
}

func runInput(i string, reader *bufio.Reader, c *client.Client) error {
	arguments := strings.Split(i, " ")

	if len(arguments) == 0 {
		fmt.Println("no command given")
		return nil
	}

	for _, method := range m.Methods {
		if method.Name == arguments[0] {
			if len(arguments)-1 != len(method.Arguments) {
				fmt.Printf("not enough aruments for command %s\n", method.Name)
				return nil
			}

			for _, rawArgument := range method.RawArguments {
				fmt.Printf("raw argument %s required, type in and end with \\n.\\n\n", rawArgument.Name)

				rawInput := ""
				next := ""

				for next != ".\n" {
					rawInput += next

					s, err := reader.ReadString('\n')

					if err != nil {
						return err
					}

					next = s
				}

				arguments = append(arguments, rawInput)
			}

			err := method.Run(c, arguments[1:])

			if err != nil {
				fmt.Printf("error running request: %s\n", err.Error())
			}

			return nil
		}
	}

	fmt.Printf("method %s not found\n", arguments[0])

	return nil
}
