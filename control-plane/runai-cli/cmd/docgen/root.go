/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"os"

	"github.com/spf13/cobra"
)

const outputDirFlag = "output-dir"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate documentation for runai CLI.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(NewMarkDownCommand())
	rootCmd.PersistentFlags().StringP(outputDirFlag, "o", "./", "output directory")
}
