package datavolume

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli/completion"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/datavolume/list/manipulate"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var lowercaseToPascalCase = map[manipulate.DatavolumeColumn]string{
	manipulate.ColumnKeyName:      "Name",
	manipulate.ColumnKeyScope:     "Scope",
	manipulate.ColumnKeyOriginPvc: "OriginPvc",
}

func ColumnsFlagCompletion(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	currentColumns, previousColumns, allColumnsSplit := completion.ParseColumnsForCompletion(toComplete)
	selectedColumnsMap := completion.CreateSelectedColumnsMap(allColumnsSplit)
	sortedCompletion := completion.CreateSortedColumnsCompletions(currentColumns, previousColumns, selectedColumnsMap, lo.Keys(manipulate.ColumnMap), lowercaseToPascalCase)
	return sortedCompletion, cobra.ShellCompDirectiveNoSpace
}
