/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package attach

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
)

const (
	ShortMsg   = "attach to a running container in {{.AOrAn}} {{.Type}} workload"
	FailedMsg  = "failed to attach to {{.AOrAn}} {{.Type}} workload"
	ExampleMsg = `# Attaching to the main worker of a workload
runai {{.CommandType}} attach <workload-name> --tty --stdin

# Attaching to a specific pod of a workload
runai {{.CommandType}} attach <workload-name> --pod <pod-name> --tty --stdin`
)

func NewFrameworkTexts(workloadType api.WorkloadType, framework api.Framework) *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetFrameworkWorkloadReplacementTokens(workloadType, framework)

	return cmd.NewBaseCommandTexts(ShortMsg, "", ExampleMsg, FailedMsg, "", nil, replacementTokens)
}

func NewWorkloadOnlyTexts(workloadType api.WorkloadType) *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetWorkloadOnlyReplacementTokens(workloadType)

	return cmd.NewBaseCommandTexts(ShortMsg, "", ExampleMsg, FailedMsg, "", nil, replacementTokens)
}
