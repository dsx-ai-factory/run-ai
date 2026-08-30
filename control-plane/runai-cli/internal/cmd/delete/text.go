package delete

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
)

const (
	ExampleMsg = `# Delete a workload with a default project
runai {{.CommandType}} delete <workload-name>

# Delete a workload with a specific project
runai {{.CommandType}} delete <workload-name> -p <project-name>

# Delete a workload by UUID
runai {{.CommandType}} delete --uuid=<workload-id> -p <project-name>`

	SuccessMsg    = "Deleting workload '{{.WorkloadName}}'...\n"
	ShortMsg      = "delete {{.AOrAn}} {{.Type}} workload"
	BaseErrorMsg  = "failed to delete workload"
	NoResultsMsg  = "The requested workload not found"
	SearchableMsg = "failed to search for requested workload"
)

var (
	aliases = []string{"del"}
)

func NewFrameworkTexts(workloadType api.WorkloadType, framework api.Framework) *cmd.BaseSearchableCommandTexts {
	replacementTokens := cmd.GetFrameworkWorkloadReplacementTokens(workloadType, framework)

	return cmd.NewBaseSearchableCommandTexts(
		ShortMsg,
		"",
		ExampleMsg,
		BaseErrorMsg,
		SuccessMsg,
		NoResultsMsg,
		SearchableMsg,
		replacementTokens,
		aliases,
	)
}

func NewWorkloadOnlyTexts(workloadType api.WorkloadType) *cmd.BaseSearchableCommandTexts {
	replacementTokens := cmd.GetWorkloadOnlyReplacementTokens(workloadType)

	return cmd.NewBaseSearchableCommandTexts(
		ShortMsg,
		"",
		ExampleMsg,
		BaseErrorMsg,
		SuccessMsg,
		NoResultsMsg,
		SearchableMsg,
		replacementTokens,
		aliases,
	)
}

func NewInferenceTexts() *cmd.BaseSearchableCommandTexts {
	replacementTokens := cmd.GetInferenceReplacementTokens(api.Inference)
	return cmd.NewBaseSearchableCommandTexts(
		ShortMsg,
		"",
		ExampleMsg,
		BaseErrorMsg,
		SuccessMsg,
		NoResultsMsg,
		SearchableMsg,
		replacementTokens,
		aliases,
	)
}

func NewDistributedInferenceTexts() *cmd.BaseSearchableCommandTexts {
	replacementTokens := cmd.GetInferenceReplacementTokens(api.DistributedInference)
	return cmd.NewBaseSearchableCommandTexts(
		ShortMsg,
		"",
		ExampleMsg,
		BaseErrorMsg,
		SuccessMsg,
		NoResultsMsg,
		SearchableMsg,
		replacementTokens,
		aliases,
	)
}

func NewNimInferenceDeleteTexts() *cmd.BaseSearchableCommandTexts {
	replacementTokens := cmd.GetNimReplacementTokens()
	return cmd.NewBaseSearchableCommandTexts(
		ShortMsg,
		"",
		ExampleMsg,
		BaseErrorMsg,
		SuccessMsg,
		NoResultsMsg,
		SearchableMsg,
		replacementTokens,
		aliases,
	)
}
