package credential

import (
	"github.com/spf13/cobra"
)

func TypeFlagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{"accessKey", "dockerRegistry", "password", "genericSecret", "ngcApiKey"}, cobra.ShellCompDirectiveNoFileComp
}
