/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package update

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
)

const (
	ShortMsg       = "update {{.AOrAn}} {{.Type}} workload"
	FailMsg        = "{{.AOrAn}} {{.Type}} update failed"
	NoResultsMsg   = "No workloads found"
	SearchableMsg  = "failed to search for requested workload"
	BaseExampleMsg = `# Update a workload with a new image
runai {{.CommandType}} update <workload-name> -p <project-name> -i runai.jfrog.io/demo/quickstart-demo

# Update a workload with a new autoscaling configuration
runai {{.CommandType}} update <workload-name> -p <project-name> --max-replicas=5 --min-replicas=3 --metric=concurrency --metric-threshold=10
`

	NimUpdateLongMsg = `IMPORTANT: Update operations OVERWRITE entire field sections, they do not merge with existing values.
For example, if a workload has 10 environment variables and you update with -e NEW_VAR=value,
the workload will have only 1 environment variable after the update.`

	NimUpdateExampleMsg = `# Update image and GPU count
runai {{.CommandType}} update <workload-name> -i nvcr.io/nim/meta/llama3-8b-instruct:v2 -g 2

# Update compute resources (replaces entire compute section)
runai {{.CommandType}} update <workload-name> --cpu-core-request 4 --cpu-memory-request 16Gi -g 2

# Update environment variables (replaces all existing env vars)
runai {{.CommandType}} update <workload-name> -e LOG_LEVEL=debug

# Switch from autoscaling to fixed replicas (explicit removal required)
runai {{.CommandType}} update <workload-name> --remove autoscaling --replicas 5

# Update autoscaling configuration
runai {{.CommandType}} update <workload-name> --min-replicas 1 --max-replicas 5 --metric concurrency --metric-threshold 10
`
)

func NewNimUpdateText() *cmd.BaseSearchableCommandTexts {
	return cmd.NewBaseSearchableCommandTexts(
		ShortMsg,
		NimUpdateLongMsg,
		NimUpdateExampleMsg,
		FailMsg,
		"",
		NoResultsMsg,
		SearchableMsg,
		cmd.GetNimReplacementTokens(),
		nil,
	)
}

func NewWorkloadOnlyText(workloadType api.WorkloadType) *cmd.BaseSearchableCommandTexts {
	var replacementTokens []cmd.TokenPair

	if workloadType == api.Inference || workloadType == api.DistributedInference {
		replacementTokens = cmd.GetInferenceReplacementTokens(workloadType)
	} else {
		replacementTokens = cmd.GetWorkloadOnlyReplacementTokens(workloadType)
	}

	var example string
	var longMsg string

	switch workloadType {
	case api.Inference, api.DistributedInference:
		example = BaseExampleMsg
	case api.Distributed:
		example = BaseExampleMsg
	default:
		example = BaseExampleMsg
	}

	return cmd.NewBaseSearchableCommandTexts(
		ShortMsg,
		longMsg,
		example,
		FailMsg,
		"",
		NoResultsMsg,
		SearchableMsg,
		replacementTokens,
		nil,
	)
}
