/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package update

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/node"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/completion"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/submit"
	"github.com/spf13/pflag"
)

func InferenceUpdateCommonFlagSet(flagsInference *pflag.FlagSet, specInference *submit.Inference) {
	flagsInference.Var(cli.NewNullableInt32Flag(&specInference.MinReplicas), consts.AttributeNameMinReplicas, consts.AttributeUsageMinReplicas)
	flagsInference.Var(cli.NewNullableInt32Flag(&specInference.MaxReplicas), consts.AttributeNameMaxReplicas, consts.AttributeUsageMaxReplicas)
	flagsInference.Var(cli.NewNullableInt32Flag(&specInference.InitialReplicas), consts.AttributeNameInitialReplicas, consts.AttributeUsageInitialReplicas)
	flagsInference.Var(cli.NewNullableInt32Flag(&specInference.ActivationReplicas), consts.AttributeNameActivationReplicas, consts.AttributeUsageActivationReplicas)
	flagsInference.Var(cli.NewNullableInt32Flag(&specInference.ConcurrencyHardLimit), consts.AttributeNameConcurrencyHardLimit, consts.AttributeUsageConcurrencyHardLimit)
	flagsInference.Var(cli.NewNullableInt32Flag(&specInference.ScaleToZeroRetentionSeconds), consts.AttributeNameScaleToZeroRetentionSeconds, consts.AttributeUsageScaleToZeroRetentionSeconds)
	flagsInference.Var(cli.NewNullableInt32Flag(&specInference.ScaleDownDelaySeconds), consts.AttributeNameScaleDownDelaySeconds, consts.AttributeUsageScaleDownDelaySeconds)
	flagsInference.Var(cli.NewNullableInt32Flag(&specInference.InitializationTimeoutSeconds), consts.AttributeNameInitializationTimeoutSeconds, consts.AttributeUsageInitializationTimeoutSeconds)
	flagsInference.Var(cli.NewNullableInt32Flag(&specInference.RequestTimeoutSeconds), consts.AttributeNameRequestTimeoutSeconds, consts.AttributeUsageRequestTimeoutSeconds)
	flagsInference.StringVar(&specInference.Metric, consts.AttributeNameMetric, "", consts.AttributeUsageMetric)
	flagsInference.Var(cli.NewNullableInt32Flag(&specInference.MetricThreshold), consts.AttributeNameMetricThreshold, consts.AttributeUsageMetricThreshold)
	flagsInference.Var(cli.NewNullableFloat32Flag(&specInference.MetricThresholdPercentage), consts.AttributeNameMetricThresholdPercentage, consts.AttributeUsageMetricThresholdPercentage)
}

func InferenceUpdateComputeFlagSet(flagsCompute *pflag.FlagSet, specCompute *submit.Compute) {
	// GPU related flags
	flagsCompute.Var(cli.NewNullableNonNegativeInt32Flag(&specCompute.GpuDeviceRequest), consts.AttributeNameGpuDeviceRequest, consts.AttributeUsageGpuDeviceRequest)
	flagsCompute.StringVar(&specCompute.GpuRequestType, consts.AttributeNameGpuRequestType, "", consts.AttributeUsageGpuRequestType)
	flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&specCompute.GpuPortionRequest), consts.AttributeNameGpuPortionRequest, consts.AttributeUsageGpuPortionRequest)
	flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&specCompute.GpuPortionLimit), consts.AttributeNameGpuPortionLimit, consts.AttributeUsageGpuPortionLimit)
	flagsCompute.StringVar(&specCompute.GpuMemoryRequest, consts.AttributeNameGpuMemoryRequest, "", consts.AttributeUsageGpuMemoryRequest)
	flagsCompute.StringVar(&specCompute.GpuMemoryLimit, consts.AttributeNameGpuMemoryLimit, "", consts.AttributeUsageGpuMemoryLimit)
	submit.GpuMemoryAndModeFlagSet(flagsCompute, specCompute)

	// CPU related flags
	flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&specCompute.CpuCoreRequest), consts.AttributeNameCpuCoreRequest, consts.AttributeUsageCpuCoreRequest)
	flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&specCompute.CpuCoreLimit), consts.AttributeNameCpuCoreLimit, consts.AttributeUsageCpuCoreLimit)
	flagsCompute.StringVar(&specCompute.CpuMemoryRequest, consts.AttributeNameCpuMemoryRequest, "", consts.AttributeUsageCpuMemoryRequest)
	flagsCompute.StringVar(&specCompute.CpuMemoryLimit, consts.AttributeNameCpuMemoryLimit, "", consts.AttributeUsageCpuMemoryLimit)

	// Other compute related flags
	flagsCompute.BoolVar(&specCompute.LargeShmRequest, consts.AttributeNameLargeShm, false, consts.AttributeUsageLargeShm)
	flagsCompute.StringArrayVar(&specCompute.ExtendedResources, consts.AttributeNameExtendedResource, []string{}, consts.AttributeUsageExtendedResource)
}

func InferenceUpdateNodeSelectionFlagSet(flagsScheduling *pflag.FlagSet, specScheduling *submit.Scheduling, completionFnsToRegister map[string]completion.CompletionFn) {
	flagsScheduling.StringArrayVar(&specScheduling.NodePools, consts.AttributeNameNodePools, []string{}, consts.AttributeUsageNodePools)
	flagsScheduling.StringArrayVar(&specScheduling.ExcludeNodes, consts.AttributeNameExcludeNode, []string{}, consts.AttributeUsageExcludeNode)
	completionFnsToRegister[consts.AttributeNameExcludeNode] = node.GetNodeNamesForFlag
}

func NimUpdateComputeFlagSet(flagsCompute *pflag.FlagSet, specCompute *submit.Compute) {
	flagsCompute.Var(cli.NewNullableNonNegativeInt32Flag(&specCompute.GpuDeviceRequest), consts.AttributeNameGpuDeviceRequest, consts.AttributeUsageGpuDeviceRequest)
	flagsCompute.StringVar(&specCompute.GpuRequestType, consts.AttributeNameGpuRequestType, "", consts.AttributeUsageGpuRequestType)
	flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&specCompute.GpuPortionRequest), consts.AttributeNameGpuPortionRequest, consts.AttributeUsageGpuPortionRequest)
	flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&specCompute.GpuPortionLimit), consts.AttributeNameGpuPortionLimit, consts.AttributeUsageGpuPortionLimit)
	flagsCompute.StringVar(&specCompute.GpuMemoryRequest, consts.AttributeNameGpuMemoryRequest, "", consts.AttributeUsageGpuMemoryRequest)
	flagsCompute.StringVar(&specCompute.GpuMemoryLimit, consts.AttributeNameGpuMemoryLimit, "", consts.AttributeUsageGpuMemoryLimit)
	submit.GpuMemoryAndModeFlagSet(flagsCompute, specCompute)

	flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&specCompute.CpuCoreRequest), consts.AttributeNameCpuCoreRequest, consts.AttributeUsageCpuCoreRequest)
	flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&specCompute.CpuCoreLimit), consts.AttributeNameCpuCoreLimit, consts.AttributeUsageCpuCoreLimit)
	flagsCompute.StringVar(&specCompute.CpuMemoryRequest, consts.AttributeNameCpuMemoryRequest, "", consts.AttributeUsageCpuMemoryRequest)
	flagsCompute.StringVar(&specCompute.CpuMemoryLimit, consts.AttributeNameCpuMemoryLimit, "", consts.AttributeUsageCpuMemoryLimit)
}

func NimUpdateSchedulingFlagSet(flagsScheduling *pflag.FlagSet, specScheduling *submit.Scheduling) {
	flagsScheduling.StringArrayVar(&specScheduling.Toleration, consts.AttributeNameToleration, []string{}, consts.AttributeUsageToleration)
}
