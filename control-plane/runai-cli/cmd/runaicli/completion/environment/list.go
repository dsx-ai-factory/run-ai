package environment

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli/completion"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/environment/list/manipulate"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

// For outputting the autocomplete options in pascal case
var lowercaseToPascalCase = map[manipulate.EnvironmentColumn]string{
	manipulate.ColumnKeyName:          "Name",
	manipulate.ColumnKeyScope:         "Scope",
	manipulate.ColumnKeyImage:         "Image",
	manipulate.ColumnKeyTools:         "Tools",
	manipulate.ColumnKeyWorkloadTypes: "WorkloadTypes",
}

// ColumnsFlagCompletion provides autocomplete for the --columns flag with multi-value support
func ColumnsFlagCompletion(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	currentColumns, previousColumns, allColumnsSplit := completion.ParseColumnsForCompletion(toComplete)
	selectedColumnsMap := completion.CreateSelectedColumnsMap(allColumnsSplit)
	sortedCompletion := completion.CreateSortedColumnsCompletions(currentColumns, previousColumns, selectedColumnsMap, lo.Keys(manipulate.ColumnMap), lowercaseToPascalCase)
	return sortedCompletion, cobra.ShellCompDirectiveNoSpace
}
