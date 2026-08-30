package credential

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli/completion"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/credential/list/manipulate"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var lowercaseToPascalCase = map[manipulate.CredentialColumn]string{
	manipulate.ColumnKeyName:  "Name",
	manipulate.ColumnKeyType:  "Type",
	manipulate.ColumnKeyScope: "Scope",
}

func ColumnsFlagCompletion(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	currentColumns, previousColumns, allColumnsSplit := completion.ParseColumnsForCompletion(toComplete)
	selectedColumnsMap := completion.CreateSelectedColumnsMap(allColumnsSplit)
	sortedCompletion := completion.CreateSortedColumnsCompletions(currentColumns, previousColumns, selectedColumnsMap, lo.Keys(manipulate.ColumnMap), lowercaseToPascalCase)
	return sortedCompletion, cobra.ShellCompDirectiveNoSpace
}
