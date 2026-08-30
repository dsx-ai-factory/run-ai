package config

import "github.com/spf13/cobra"

// updateShellFlagCompletion provides auto-completion for update-shell flag
func updateShellFlagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"bash\tuse bash shell for updating CLI",
		"dash\tuse dash shell for updating CLI",
		"sh\tuse sh shell for updating CLI",
		"fish\tuse fish shell for updating CLI",
		"zsh\tuse zsh shell for updating CLI",
	}, cobra.ShellCompDirectiveNoFileComp
}

// autoUpdateFlagCompletion provides auto-completion for auto-update flag
func autoUpdateFlagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"enabled\tenable automatic CLI updates",
		"disabled\tdisable automatic CLI updates",
	}, cobra.ShellCompDirectiveNoFileComp
}

// listAlignFlagCompletion provides auto-completion for lift-format-align flag
func listAlignFlagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"enabled\tenable aligned output formatting",
		"disabled\tdisable aligned output formatting",
	}, cobra.ShellCompDirectiveNoFileComp
}
