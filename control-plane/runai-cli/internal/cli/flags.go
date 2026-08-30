package cli

import "github.com/spf13/cobra"

const (
	JsonOutput  = "json"
	YamlOutput  = "yaml"
	TableOutput = "table"
)

// GetOutputTypeFromFlags returns "json", "yaml", or "table" based on mutually exclusive flags.
// Defaults to "table" if none are set.
func GetOutputTypeFromFlags(cmd *cobra.Command) string {
	outputFlags := []string{JsonOutput, YamlOutput, TableOutput}
	for _, flag := range outputFlags {
		if cmd.Flags().Changed(flag) {
			isSet, _ := cmd.Flags().GetBool(flag)
			if isSet {
				return flag
			}
		}
	}
	return TableOutput
}
