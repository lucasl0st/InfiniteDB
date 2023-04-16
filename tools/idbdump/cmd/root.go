/*
 * Copyright (c) 2023 Lucas Pape
 */

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "idbdump",
	Short: "Dump infinitedb database",
	Long:  "",
}

var (
	databaseHostname         string
	databasePort             uint
	databaseTls              bool
	databaseTlsVerifyDisable bool
	databaseKey              string
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
}
