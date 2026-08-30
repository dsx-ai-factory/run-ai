package cli

import "github.com/spf13/cobra"

func newAirgapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "airgap",
		Short: "Air-gapped bundle tooling",
	}
	cmd.AddCommand(newAirgapPackCmd())
	cmd.AddCommand(newAirgapLoadCmd())
	return cmd
}
