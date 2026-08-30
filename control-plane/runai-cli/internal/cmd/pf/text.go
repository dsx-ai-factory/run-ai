package pf

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
)

const (
	ShortMsg     = "forward one or more local ports to {{.AOrAn}} {{.Type}} workload"
	BaseErrorMsg = "failed to set up port forwarding for {{.AOrAn}} {{.Type}} workload"
	ExampleMsg   = `# Forward connections from localhost:8080 to a workload on port 8090:
runai {{.CommandType}} port-forward <workload-name> --port 8080:8090 --address localhost

# Forward connections from 0.0.0.0:8080 to a workload on port 8080:
runai {{.CommandType}} port-forward <workload-name> --port 8080 --address 0.0.0.0 [requires privileges]

# Forward multiple connections from localhost:8080 to a workload on port 8090 and from localhost:6443 to a workload on port 443:
runai {{.CommandType}} port-forward <workload-name> --port 8080:8090 --port 6443:443 --address localhost`
)

var (
	aliases = []string{"pf"}
)

func NewFrameworkTexts(workloadType api.WorkloadType, framework api.Framework) *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetFrameworkWorkloadReplacementTokens(workloadType, framework)

	return cmd.NewBaseCommandTexts(ShortMsg, "", ExampleMsg, BaseErrorMsg, "", aliases, replacementTokens)
}

func NewWorkloadOnlyTexts(workloadType api.WorkloadType) *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetWorkloadOnlyReplacementTokens(workloadType)

	return cmd.NewBaseCommandTexts(ShortMsg, "", ExampleMsg, BaseErrorMsg, "", aliases, replacementTokens)
}

func NewInferenceTexts() *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetInferenceReplacementTokens(api.Inference)
	return cmd.NewBaseCommandTexts(ShortMsg, cmd.GetInferencePodDescription(), ExampleMsg, BaseErrorMsg, "", aliases, replacementTokens)
}

func NewDistributedInferenceTexts() *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetInferenceReplacementTokens(api.DistributedInference)
	return cmd.NewBaseCommandTexts(ShortMsg, cmd.GetDistributedInferencePodDescription(), ExampleMsg, BaseErrorMsg, "", aliases, replacementTokens)
}

func NewNimInferenceTexts() *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetNimReplacementTokens()
	return cmd.NewBaseCommandTexts(ShortMsg, cmd.GetNimPodDescription(), ExampleMsg, BaseErrorMsg, "", aliases, replacementTokens)
}
