/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submitdistbc

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
)

const (
	ShortMsg = "submit-dist {{.Framework}}"
	FailMsg  = "{{.Framework}} submission failed"

	BaseExampleMsg = `# Submit the workload
runai submit-dist {{.Framework}} <workload-name> -p <project-name> -i ubuntu

# Submit the workload with arguments
runai submit-dist {{.Framework}} <workload-name> -p <project-name> -i ubuntu -- ls -la

# Submit the workload with a custom command
runai submit-dist {{.Framework}} <workload-name> -p <project-name> -i ubuntu --command -- echo "Hello, World"`

	MasterArgsAndCommandExampleMsg = `

# Submit the workload master args with worker args
runai submit-dist {{.Framework}} <workload-name> -p <project-name> -i ubuntu --master-args "-a master_arg_a -b master-arg_b'" -- '-a worker_arg_a'

# Submit the workload master command with worker args
runai submit-dist {{.Framework}} <workload-name> -p <project-name> -i ubuntu --master-command "echo -e 'master command'" -- '-a worker_arg_a'

# Submit the workload master command with worker command
runai submit-dist {{.Framework}} <workload-name> -p <project-name> -i ubuntu --master-command "echo -e 'master command'" --command -- echo -e 'worker command'`
)

var baseWIthMaster = BaseExampleMsg + MasterArgsAndCommandExampleMsg

func NewFrameworkText(workloadType api.WorkloadType, framework api.Framework) *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetFrameworkWorkloadReplacementTokens(workloadType, framework)
	if framework == api.Jax {
		return cmd.NewBaseCommandTexts(ShortMsg, "", BaseExampleMsg, FailMsg, "", nil, replacementTokens)
	}
	return cmd.NewBaseCommandTexts(ShortMsg, "", baseWIthMaster, FailMsg, "", nil, replacementTokens)
}

func NewWorkloadOnlyText(workloadType api.WorkloadType) *cmd.BaseCommandTexts {
	replacementTokens := cmd.GetWorkloadOnlyReplacementTokens(workloadType)
	return cmd.NewBaseCommandTexts(ShortMsg, "", BaseExampleMsg, FailMsg, "", nil, replacementTokens)
}
