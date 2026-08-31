/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package scale

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
)

const (
	ShortScaleMsg   = "scale {{.AOrAn}} {{.Type}} workload"
	ExampleScaleMsg = `# Scale a workload (replicas flag is required)
runai {{.CommandType}} scale <workload-name> --replicas 5`
	NimExampleScaleMsg = `# Scale a workload (replicas flag is required)
runai {{.CommandType}} scale <workload-name> --replicas 5

# Scale with --force to skip autoscaling confirmation
runai {{.CommandType}} scale <workload-name> --replicas 3 --force`
	SuccessMsg          = "Scaling workload '{{.WorkloadName}}'...\n"
	ErrorScaleMsg       = "failed to scale"
	ErrorSearchScaleMsg = "failed to search for workload"
	ErrorNoResultsMsg   = "workload not found"
)

func NewScaleText(workloadType api.WorkloadType, framework api.Framework) cmd.SearchableCommandTexts {
	var replacementTokens []cmd.TokenPair
	var example string
	if framework == api.NimService {
		replacementTokens = cmd.GetNimReplacementTokens()
		example = NimExampleScaleMsg
	} else {
		replacementTokens = cmd.GetInferenceReplacementTokens(workloadType)
		example = ExampleScaleMsg
	}

	return cmd.NewBaseSearchableCommandTexts(
		ShortScaleMsg,
		"",
		example,
		ErrorScaleMsg,
		SuccessMsg,
		ErrorNoResultsMsg,
		ErrorSearchScaleMsg,
		replacementTokens,
		[]string{},
	)
}
