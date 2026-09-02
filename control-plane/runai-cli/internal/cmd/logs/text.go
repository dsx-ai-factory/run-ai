/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logs

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
)

const (
	ShortMsg     = "view logs of {{.AOrAn}} {{.Type}} workload"
	BaseErrorMsg = "failed to get logs for {{.AOrAn}} {{.Type}} workload"
	ExampleMsg   = `# Get logs for a workload
runai {{.CommandType}} logs <workload-name>

# Get logs for a specific pod in a workload
runai {{.CommandType}} logs <workload-name> --pod=<pod-name>

# Get logs for a specific container in a workload
runai {{.CommandType}} logs <workload-name> --container=<container-name>

# Get the last 100 lines of logs
runai {{.CommandType}} logs <workload-name> --tail=100

# Get logs with timestamps
runai {{.CommandType}} logs <workload-name> --timestamps

# Follow the logs
runai {{.CommandType}} logs <workload-name> --follow

# Get logs for the previous instance of the workload
runai {{.CommandType}} logs <workload-name> --previous

# Limit the logs to 1024 bytes
runai {{.CommandType}} logs <workload-name> --limit-bytes=1024

# Get logs since the last 5 minutes
runai {{.CommandType}} logs <workload-name> --since=300s

# Get logs since a specific timestamp
runai {{.CommandType}} logs <workload-name> --since-time=2023-05-30T10:00:00Z

# Wait up to 30 seconds for a workload to be ready for logs
runai {{.CommandType}} logs <workload-name> --wait-timeout=30s`
)

func NewFrameworkLogsTexts(workloadType api.WorkloadType, framework api.Framework) *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetFrameworkWorkloadReplacementTokens(workloadType, framework)

	return cmd.NewBaseCommandTexts(ShortMsg, "", ExampleMsg, BaseErrorMsg, "", nil, replacementTokens)
}

func NewWorkloadOnlyLogsTexts(workloadType api.WorkloadType) *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetWorkloadOnlyReplacementTokens(workloadType)

	return cmd.NewBaseCommandTexts(ShortMsg, "", ExampleMsg, BaseErrorMsg, "", nil, replacementTokens)
}

func NewInferenceLogsTexts() *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetInferenceReplacementTokens(api.Inference)
	return cmd.NewBaseCommandTexts(ShortMsg, cmd.GetInferencePodDescription(), ExampleMsg, BaseErrorMsg, "", nil, replacementTokens)
}

func NewDistributedInferenceLogsTexts() *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetInferenceReplacementTokens(api.DistributedInference)
	return cmd.NewBaseCommandTexts(ShortMsg, cmd.GetDistributedInferencePodDescription(), ExampleMsg, BaseErrorMsg, "", nil, replacementTokens)
}

func NewNimInferenceLogsTexts() *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetNimReplacementTokens()
	return cmd.NewBaseCommandTexts(ShortMsg, cmd.GetNimPodDescription(), ExampleMsg, BaseErrorMsg, "", nil, replacementTokens)
}
