/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package exec

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
)

const (
	// Exec command strings
	execShortMsg     = "execute a command in {{.AOrAn}} {{.Type}} workload"
	execBaseErrorMsg = "failed to execute command in {{.AOrAn}} {{.Type}} workload"
	execExampleMsg   = `# Execute bash in the workload's main worker
runai {{.CommandType}} exec <workload-name> --tty --stdin -- /bin/bash 

# Execute ls command in the workload's main worker
runai {{.CommandType}} exec <workload-name> -- ls

# Execute a command in a specific worker of the workload
runai {{.CommandType}} exec <workload-name> --pod <pod-name> -- nvidia-smi`

	execExampleMsgInference = `# Execute bash in the workload
runai {{.CommandType}} exec <workload-name> --tty --stdin -- /bin/bash 

# Execute ls command in the workload
runai {{.CommandType}} exec <workload-name> -- ls

# Execute a command in a specific pod of the workload
runai {{.CommandType}} exec <workload-name> --pod <pod-name> -- nvidia-smi`

	// Bash command strings
	bashShortMsg     = "open a bash shell in {{.AOrAn}} {{.Type}} workload"
	bashBaseErrorMsg = "failed to open bash shell in {{.AOrAn}} {{.Type}} workload"
	bashExampleMsg   = `# Open a bash shell in the workload's main worker
runai {{.CommandType}} bash <workload-name>

# Open a bash shell in a specific worker of the workload
runai {{.CommandType}} bash <workload-name> --pod <pod-name>`

	bashExampleMsgInference = `# Open a bash shell in the workload
runai {{.CommandType}} bash <workload-name>

# Open a bash shell in a specific pod of the workload
runai {{.CommandType}} bash <workload-name> --pod <pod-name>`
)

func NewFrameworkExecTexts(workloadType api.WorkloadType, framework api.Framework) *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetFrameworkWorkloadReplacementTokens(workloadType, framework)

	return cmd.NewBaseCommandTexts(execShortMsg, "", execExampleMsg, execBaseErrorMsg, "", nil, replacementTokens)
}

func NewWorkloadOnlyExecTexts(workloadType api.WorkloadType) *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetWorkloadOnlyReplacementTokens(workloadType)
	return cmd.NewBaseCommandTexts(execShortMsg, "", execExampleMsg, execBaseErrorMsg, "", nil, replacementTokens)
}

func NewInferenceExecTexts() *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetInferenceReplacementTokens(api.Inference)
	return cmd.NewBaseCommandTexts(execShortMsg, cmd.GetInferencePodDescription(), execExampleMsgInference, execBaseErrorMsg, "", nil, replacementTokens)
}

func NewDistributedInferenceExecTexts() *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetInferenceReplacementTokens(api.DistributedInference)
	return cmd.NewBaseCommandTexts(execShortMsg, cmd.GetDistributedInferencePodDescription(), execExampleMsgInference, execBaseErrorMsg, "", nil, replacementTokens)
}

func NewFrameworkBashTexts(workloadType api.WorkloadType, framework api.Framework) *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetFrameworkWorkloadReplacementTokens(workloadType, framework)
	return cmd.NewBaseCommandTexts(bashShortMsg, "", bashExampleMsg, bashBaseErrorMsg, "", nil, replacementTokens)
}

func NewWorkloadOnlyBashTexts(workloadType api.WorkloadType) *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetWorkloadOnlyReplacementTokens(workloadType)
	return cmd.NewBaseCommandTexts(bashShortMsg, "", bashExampleMsg, bashBaseErrorMsg, "", nil, replacementTokens)
}

func NewInferenceBashTexts() *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetInferenceReplacementTokens(api.Inference)
	return cmd.NewBaseCommandTexts(bashShortMsg, cmd.GetInferencePodDescription(), bashExampleMsgInference, bashBaseErrorMsg, "", nil, replacementTokens)
}

func NewNimInferenceExecTexts() *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetNimReplacementTokens()
	return cmd.NewBaseCommandTexts(execShortMsg, cmd.GetNimPodDescription(), execExampleMsgInference, execBaseErrorMsg, "", nil, replacementTokens)
}

func NewNimInferenceBashTexts() *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetNimReplacementTokens()
	return cmd.NewBaseCommandTexts(bashShortMsg, cmd.GetNimPodDescription(), bashExampleMsgInference, bashBaseErrorMsg, "", nil, replacementTokens)
}

func NewDistributedInferenceBashTexts() *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetInferenceReplacementTokens(api.DistributedInference)
	return cmd.NewBaseCommandTexts(bashShortMsg, cmd.GetDistributedInferencePodDescription(), bashExampleMsgInference, bashBaseErrorMsg, "", nil, replacementTokens)
}
