package workload

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli/completion"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/list/manipulate"
	openapi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
	"github.com/samber/lo"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli"
	openapi2 "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/spf13/cobra"
)

// For outputting the autocomplete options in pascal case.
var lowercaseToPascalCase = map[manipulate.WorkloadColumn]string{
	manipulate.ColumnKeyName:                "Name",
	manipulate.ColumnKeyType:                "Type",
	manipulate.ColumnKeyFramework:           "Framework",
	manipulate.ColumnKeyStatus:              "Status",
	manipulate.ColumnKeyProject:             "Project",
	manipulate.ColumnKeyCategory:            "Category",
	manipulate.ColumnKeyPriority:            "Priority",
	manipulate.ColumnKeyDepartment:          "Department",
	manipulate.ColumnKeySubmittedBy:         "SubmittedBy",
	manipulate.ColumnKeyCreatedAt:           "CreatedAt",
	manipulate.ColumnKeyTotalPendingTime:    "TotalPendingTimeSeconds",
	manipulate.ColumnKeyAllocatedGPUMemory:  "AllocatedGPUMemory",
	manipulate.ColumnKeyAllocatedCPUCompute: "AllocatedCPUCompute",
	manipulate.ColumnKeyAllocatedCPUMemory:  "AllocatedCPUMemory",
	manipulate.ColumnKeyCompletedAt:         "CompletedAt",
	manipulate.ColumnKeyPreemptible:         "Preemptible",
	manipulate.ColumnKeyPods:                "Pods",
	manipulate.ColumnKeyGPUAlloc:            "AllocatedGPU",
	manipulate.ColumnKeyUrls:                "URLs",
	manipulate.ColumnKeyImage:               "Image",
	manipulate.ColumnKeyCurrentNodepool:     "CurrentNodepool",
	manipulate.ColumnKeyRequestedNodepool:   "RequestedNodepool",
}

// StatusFlagCompletion auto completion support for log level
func StatusFlagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	var phases []string
	for _, phase := range openapi2.AllowedPhaseEnumValues {
		phases = append(phases, string(phase))
	}
	return phases, cobra.ShellCompDirectiveNoFileComp
}

func TypeFlagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return api.GetWorkloadTypes(), cobra.ShellCompDirectiveNoFileComp
}

func FrameworkFlagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return api.GetWorkloadFramework(), cobra.ShellCompDirectiveNoFileComp
}

func CleanPodPolicyFlagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	apiAllowedOptions := openapi2.AllowedCleanPodPolicyEnumValues
	options := make([]string, len(apiAllowedOptions))

	for i, apiOption := range apiAllowedOptions {
		options[i] = string(apiOption)
	}

	return options, cobra.ShellCompDirectiveNoFileComp
}

func RestartPolicyFlagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{
		string(openapi2.RESTARTPOLICY_NEVER),
		string(openapi2.RESTARTPOLICY_ALWAYS),
		string(openapi2.RESTARTPOLICY_ON_FAILURE),
	}, cobra.ShellCompDirectiveNoFileComp
}

func LauncherCreationPolicyFlagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{
		cli.MPILauncherCreationPolicyAtStartup,
		cli.MPILauncherCreationPolicyWaitForWorkerReady,
	}, cobra.ShellCompDirectiveNoFileComp
}

func PreemptibilityFlagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{
		string(openapi2.PREEMPTIBILITY_PREEMPTIBLE),
		string(openapi2.PREEMPTIBILITY_NON_PREEMPTIBLE),
	}, cobra.ShellCompDirectiveNoFileComp
}

func MnnvlFlagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	stringArr := make([]string, len(openapi.AllowedMnnvlModeEnumValues))
	for i, v := range openapi.AllowedMnnvlModeEnumValues {
		stringArr[i] = string(v) // Cast individual element
	}
	return stringArr, cobra.ShellCompDirectiveNoFileComp
}

// ColumnsFlagCompletion provides autocomplete for the --columns flag with multi-value support
func ColumnsFlagCompletion(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	currentColumns, previousColumns, allColumnsSplit := completion.ParseColumnsForCompletion(toComplete)
	selectedColumnsMap := completion.CreateSelectedColumnsMap(allColumnsSplit)
	sortedCompletion := completion.CreateSortedColumnsCompletions(currentColumns, previousColumns, selectedColumnsMap, lo.Keys(manipulate.ColumnMap), lowercaseToPascalCase)
	return sortedCompletion, cobra.ShellCompDirectiveNoSpace
}
