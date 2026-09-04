/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
)

const (
	ShortMsg     = "describe {{.AOrAn}} {{.Type}} workload"
	BaseErrorMsg = "failed to describe {{.AOrAn}} {{.Type}}"
	ExampleMsg   = `# Describe a workload with a default project
runai {{.CommandType}} describe <workload-name>

# Describe a workload in a specific project
runai {{.CommandType}} describe <workload-name> -p <project-name>

# Describe a workload by UUID
runai {{.CommandType}} describe --uuid=<workload-id>

# Describe a workload with specific output format
runai {{.CommandType}} describe <workload-name> -o json

# Describe a workload with specific sections
runai {{.CommandType}} describe <workload-name> --general --compute --pods --events --networks

# Describe a workload with container details and custom limits
runai {{.CommandType}} describe <workload-name> --containers --pod-limit 20 --event-limit 100`
)

func NewFrameworkTexts(workloadType api.WorkloadType, framework api.Framework) *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetFrameworkWorkloadReplacementTokens(workloadType, framework)

	return cmd.NewBaseCommandTexts(ShortMsg, "", ExampleMsg, BaseErrorMsg, "", nil, replacementTokens)
}

func NewWorkloadOnlyTexts(workloadType api.WorkloadType) *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetWorkloadOnlyReplacementTokens(workloadType)
	return cmd.NewBaseCommandTexts(ShortMsg, "", ExampleMsg, BaseErrorMsg, "", nil, replacementTokens)
}

func NewInferenceTexts() *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetInferenceReplacementTokens(api.Inference)
	return cmd.NewBaseCommandTexts(ShortMsg, "", ExampleMsg, BaseErrorMsg, "", nil, replacementTokens)
}

func NewDistributedInferenceTexts() *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetInferenceReplacementTokens(api.DistributedInference)
	return cmd.NewBaseCommandTexts(ShortMsg, "", ExampleMsg, BaseErrorMsg, "", nil, replacementTokens)
}

func NewNimInferenceTexts() *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetNimReplacementTokens()
	return cmd.NewBaseCommandTexts(ShortMsg, "", ExampleMsg, BaseErrorMsg, "", nil, replacementTokens)
}
