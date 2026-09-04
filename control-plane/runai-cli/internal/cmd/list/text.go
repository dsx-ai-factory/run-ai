/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
)

const (
	ShortMsg   = "list {{.Type}} workloads"
	ExampleMsg = `# List all {{.Type}} workloads
runai {{.CommandType}} list -A

# List workloads with default project
runai {{.CommandType}} list

# List workloads in a specific project
runai {{.CommandType}} list -p <project-name>

# List all workloads with a specific output format
runai {{.CommandType}} list --yaml

# List workloads with pagination
runai {{.CommandType}} list --limit 20 --offset 40`

	BaseErrorMsg = "failed to list {{.AOrAn}} {{.Type}}"
)

func NewFrameworkListTexts(workloadType api.WorkloadType, framework api.Framework) *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetFrameworkWorkloadReplacementTokens(workloadType, framework)

	return cmd.NewBaseCommandTexts(ShortMsg, "", ExampleMsg, BaseErrorMsg, "", nil, replacementTokens)
}

func NewWorkloadOnlyListTexts(workloadType api.WorkloadType) *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetWorkloadOnlyReplacementTokens(workloadType)

	return cmd.NewBaseCommandTexts(ShortMsg, "", ExampleMsg, BaseErrorMsg, "", nil, replacementTokens)
}

func NewInferenceDistributedListTexts(workloadType api.WorkloadType) *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetInferenceReplacementTokens(workloadType)

	return cmd.NewBaseCommandTexts(ShortMsg, "", ExampleMsg, BaseErrorMsg, "", nil, replacementTokens)
}

func NewInferenceStandardListTexts(workloadType api.WorkloadType) *cmd.BaseCommandTexts {
	replacementTokens := []cmd.TokenPair{
		{Old: cmd.IndefiniteArticleReplacer, New: cmd.IndefiniteArticleAn},
		{Old: cmd.TypeReplacer, New: workloadType.CLIType()},
		{Old: cmd.CommandTypeReplacer, New: "inference standard"},
		{Old: cmd.WorkloadTypeReplacer, New: workloadType.CLIType()},
		{Old: cmd.ExampleReplacer, New: "inference"},
		{Old: cmd.FollowingReplacer, New: cmd.FollowingFrameworkLess},
	}

	return cmd.NewBaseCommandTexts(ShortMsg, "", ExampleMsg, BaseErrorMsg, "", nil, replacementTokens)
}

func NewInferenceNimListTexts(workloadType api.WorkloadType) *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetNimReplacementTokens()
	return cmd.NewBaseCommandTexts(ShortMsg, "", ExampleMsg, BaseErrorMsg, "", nil, replacementTokens)
}
