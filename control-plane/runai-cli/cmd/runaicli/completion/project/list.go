package project

import (
	listCompletion "github.com/run-ai/runai/control-plane/runai-cli/internal/cli/completion"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project/list/manipulate"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var projectLowercaseToPascalCase = map[manipulate.ProjectColumn]string{
	manipulate.ColumnKeyName:            "Name",
	manipulate.ColumnKeyStatus:          "Status",
	manipulate.ColumnKeyDepartment:      "Department",
	manipulate.ColumnKeyGPUQuota:        "GPUQuota",
	manipulate.ColumnKeyAllocatedGPUs:   "AllocatedGPUs",
	manipulate.ColumnKeyAllocationRatio: "AllocationRatio",
}

func ColumnsFlagCompletion(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	currentColumns, previousColumns, allColumnsSplit := listCompletion.ParseColumnsForCompletion(toComplete)
	selectedColumnsMap := listCompletion.CreateSelectedColumnsMap(allColumnsSplit)
	sortedCompletion := listCompletion.CreateSortedColumnsCompletions(currentColumns, previousColumns, selectedColumnsMap, lo.Keys(manipulate.ColumnMap), projectLowercaseToPascalCase)
	return sortedCompletion, cobra.ShellCompDirectiveNoSpace
}
