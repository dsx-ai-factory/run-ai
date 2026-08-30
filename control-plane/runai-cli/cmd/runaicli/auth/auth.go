package auth

import (
	"github.com/spf13/cobra"
)

// Cmd represents the auth command group
var Cmd = &cobra.Command{
	Use:   AuthUse,
	Short: AuthShortMsg,
	Args:  cobra.NoArgs,
}

func init() {
}
