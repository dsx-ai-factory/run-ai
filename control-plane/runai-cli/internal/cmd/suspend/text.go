package suspend

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
)

const (
	ShortMsg   = "suspend {{.AOrAn}} {{.Type}} workload"
	ExampleMsg = `# Suspend a workload
runai {{.CommandType}} suspend <{{.ExampleType}}-name>

# Suspend a workload in a specific project
runai {{.CommandType}} suspend <{{.ExampleType}}-name> -p <project-name>

# Suspend a workload by UUID
runai {{.CommandType}} suspend --uuid=<{{.ExampleType}}_uuid>`
	BaseErrorMsg = "failed to suspend workload"
	SuccessMsg   = "Stopping workload '{{.WorkloadName}}'...\n{{.Following}}"

	ResumeShortMsg   = "resume {{.Type}}"
	ResumeExampleMsg = `# Resume a workload
runai {{.CommandType}} resume <{{.ExampleType}}-name>

# Resume a workload in a specific project
runai {{.CommandType}} resume <{{.ExampleType}}-name> -p <project-name>

# Resume a workload by UUID
runai {{.CommandType}} resume --uuid=<{{.ExampleType}}_uuid>`
	ResumeBaseErrorMsg = "failed to resume workload"
	ResumeSuccessMsg   = "Resuming workload '{{.WorkloadName}}'...\n{{.Following}}"

	NoResultsErrorMsg  = "The requested workload not found"
	SearchableErrorMsg = "failed to search for requested workload"
)

func NewFrameworkSuspendTexts(workloadType api.WorkloadType, framework api.Framework) *cmd.BaseSearchableCommandTexts {
	replacementTokens := cmd.GetFrameworkWorkloadReplacementTokens(workloadType, framework)

	return cmd.NewBaseSearchableCommandTexts(
		ShortMsg,
		"",
		ExampleMsg,
		BaseErrorMsg,
		SuccessMsg,
		NoResultsErrorMsg,
		SearchableErrorMsg,
		replacementTokens,
		nil,
	)
}

func NewWorkloadOnlySuspendTexts(workloadType api.WorkloadType) *cmd.BaseSearchableCommandTexts {
	replacementTokens := cmd.GetWorkloadOnlyReplacementTokens(workloadType)

	return cmd.NewBaseSearchableCommandTexts(
		ShortMsg,
		"",
		ExampleMsg,
		BaseErrorMsg,
		SuccessMsg,
		NoResultsErrorMsg,
		SearchableErrorMsg,
		replacementTokens,
		nil,
	)
}

func NewFrameworkResumeTexts(workloadType api.WorkloadType, framework api.Framework) *cmd.BaseSearchableCommandTexts {
	replacementTokens := cmd.GetFrameworkWorkloadReplacementTokens(workloadType, framework)

	return cmd.NewBaseSearchableCommandTexts(
		ResumeShortMsg,
		"",
		ResumeExampleMsg,
		ResumeBaseErrorMsg,
		ResumeSuccessMsg,
		NoResultsErrorMsg,
		SearchableErrorMsg,
		replacementTokens,
		nil,
	)
}

func NewWorkloadOnlyResumeTexts(workloadType api.WorkloadType) *cmd.BaseSearchableCommandTexts {
	replacementTokens := cmd.GetWorkloadOnlyReplacementTokens(workloadType)

	return cmd.NewBaseSearchableCommandTexts(
		ResumeShortMsg,
		"",
		ResumeExampleMsg,
		ResumeBaseErrorMsg,
		ResumeSuccessMsg,
		NoResultsErrorMsg,
		SearchableErrorMsg,
		replacementTokens,
		nil,
	)
}

func NewInferenceSuspendTexts(workloadType api.WorkloadType) *cmd.BaseSearchableCommandTexts {
	replacementTokens := cmd.GetInferenceReplacementTokens(workloadType)

	return cmd.NewBaseSearchableCommandTexts(
		ShortMsg,
		"",
		ExampleMsg,
		BaseErrorMsg,
		SuccessMsg,
		NoResultsErrorMsg,
		SearchableErrorMsg,
		replacementTokens,
		nil,
	)
}

func NewInferenceResumeTexts(workloadType api.WorkloadType) *cmd.BaseSearchableCommandTexts {
	replacementTokens := cmd.GetInferenceReplacementTokens(workloadType)

	return cmd.NewBaseSearchableCommandTexts(
		ResumeShortMsg,
		"",
		ResumeExampleMsg,
		ResumeBaseErrorMsg,
		ResumeSuccessMsg,
		NoResultsErrorMsg,
		SearchableErrorMsg,
		replacementTokens,
		nil,
	)
}
