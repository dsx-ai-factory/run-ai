package node

import (
	listCompletion "github.com/run-ai/runai/control-plane/runai-cli/internal/cli/completion"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/node/list/manipulate"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var nodeLowercaseToPascalCase = map[manipulate.NodeColumn]string{
	manipulate.ColumnKeyName:          "Name",
	manipulate.ColumnKeyStatus:        "Status",
	manipulate.ColumnKeyGpuType:       "GPUType",
	manipulate.ColumnKeyGpuAllocation: "GPUAllocation",
	manipulate.ColumnKeyFreeGpu:       "FreeGPU",
	manipulate.ColumnKeyNodepool:      "Nodepool",
}

func ColumnsFlagCompletion(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	currentColumns, previousColumns, allColumnsSplit := listCompletion.ParseColumnsForCompletion(toComplete)
	selectedColumnsMap := listCompletion.CreateSelectedColumnsMap(allColumnsSplit)
	sortedCompletion := listCompletion.CreateSortedColumnsCompletions(currentColumns, previousColumns, selectedColumnsMap, lo.Keys(manipulate.ColumnMap), nodeLowercaseToPascalCase)
	return sortedCompletion, cobra.ShellCompDirectiveNoSpace
}
