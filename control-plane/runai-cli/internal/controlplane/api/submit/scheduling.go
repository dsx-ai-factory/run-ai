/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/node"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/completion"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/schedulig/toleration"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/spf13/pflag"
)

type Scheduling struct {
	NodePools     []string
	ExcludeNodes  []string
	NodeType      string
	Toleration    []string
	PriorityClass string
}

func (s *Scheduling) GenerateTolerations() ([]openapi.Toleration, error) {
	tolerations, err := toleration.NewList(s.Toleration)
	return tolerations, err
}

func SchedulingFlagSet(flagsScheduling *pflag.FlagSet, specScheduling *Scheduling, workloadType spectype.Workload, completionFnsToRegister map[string]completion.CompletionFn) {
	if workloadType == spectype.NimService {
		flagsScheduling.StringArrayVar(&specScheduling.Toleration, consts.AttributeNameToleration, []string{}, consts.AttributeUsageToleration)
		return
	}

	flagsScheduling.StringArrayVar(&specScheduling.NodePools, consts.AttributeNameNodePools, []string{}, consts.AttributeUsageNodePools)
	flagsScheduling.StringArrayVar(&specScheduling.Toleration, consts.AttributeNameToleration, []string{}, consts.AttributeUsageToleration)

	flagsScheduling.StringVar(&specScheduling.NodeType, consts.AttributeNameNodeType, "", consts.AttributeUsageNodeType)
	flagsScheduling.StringVar(&specScheduling.PriorityClass, consts.AttributeNamePriorityClass, "", consts.AttributeUsagePriorityClass)

	flagsScheduling.StringArrayVar(&specScheduling.ExcludeNodes, consts.AttributeNameExcludeNode, []string{}, consts.AttributeUsageExcludeNode)
	completionFnsToRegister[consts.AttributeNameExcludeNode] = node.GetNodeNamesForFlag

	// deprecated flags - should be removed in the future
	nodePool := "node-pool"
	flagsScheduling.StringArrayVar(&specScheduling.NodePools, nodePool, []string{}, consts.AttributeUsageNodePools)
	_ = flagsScheduling.MarkDeprecated(nodePool, "use --node-pools instead")
}
