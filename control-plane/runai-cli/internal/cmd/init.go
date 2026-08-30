package cmd

import "github.com/spf13/cobra"

func NewEmptyCobraCommand() *cobra.Command {
	return &cobra.Command{}
}
