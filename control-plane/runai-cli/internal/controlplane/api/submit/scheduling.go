/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"errors"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/node"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/project"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/completion"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/schedulig/toleration"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/spf13/pflag"
)

type Scheduling struct {
	NodePools      []string
	ExcludeNodes   []string
	NodeType       string
	GpuType        string
	Accelerator    string
	DeviceLabelKey string
	Toleration     []string
	PriorityClass  string
}

func (s *Scheduling) GenerateTolerations() ([]openapi.Toleration, error) {
	tolerations, err := toleration.NewList(s.Toleration)
	return tolerations, err
}

// AcceleratorFlagSet adds the placement options shared by typed submissions and inference updates.
func AcceleratorFlagSet(flags *pflag.FlagSet, scheduling *Scheduling, completions map[string]completion.CompletionFn) {
	flags.StringVar(&scheduling.GpuType, "gpu-type", "", "Exact GPU product label value, or any to clear the model restriction")
	flags.StringVar(&scheduling.Accelerator, "accelerator", "", "Kubernetes accelerator resource name (default nvidia.com/gpu)")
	flags.StringVar(&scheduling.DeviceLabelKey, "device-label-key", "", "Node label key identifying the GPU model (required with a non-NVIDIA resource and --gpu-type)")
	for _, name := range []string{"gpu-type", "accelerator", "device-label-key"} {
		flag := flags.Lookup(name)
		flag.Value = nonEmptyPlacementValue{flag.Value}
	}
	completions["gpu-type"] = project.GPUTypeCompletion()
}

type nonEmptyPlacementValue struct{ pflag.Value }

func (v nonEmptyPlacementValue) Set(value string) error {
	if err := v.Value.Set(value); err != nil {
		return err
	}
	if strings.TrimSpace(value) == "" {
		return errors.New("value must not be empty")
	}
	return nil
}

func SchedulingFlagSet(flagsScheduling *pflag.FlagSet, specScheduling *Scheduling, workloadType spectype.Workload, completionFnsToRegister map[string]completion.CompletionFn) {
	if workloadType == spectype.NimService {
		flagsScheduling.StringArrayVar(&specScheduling.Toleration, consts.AttributeNameToleration, []string{}, consts.AttributeUsageToleration)
		return
	}

	AcceleratorFlagSet(flagsScheduling, specScheduling, completionFnsToRegister)
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
