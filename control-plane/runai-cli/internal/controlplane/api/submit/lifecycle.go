/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config/attributes"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/spf13/pflag"
)

const (
	AttributeNameTerminationGracePeriod  = "termination-grace-period"
	AttributeUsageTerminationGracePeriod = "The length of time (like 5s or 2m, higher than zero) the workload's pod is expected to terminate gracefully upon probe failure. " +
		"In case value is not specified, kubernetes default of 30 seconds applies"
)

type Lifecycle struct {
	AutoDeletionTimeAfterCompletion attributes.Duration
	TerminationGracePeriod          attributes.Duration
	BackoffLimit                    openapi.NullableInt32
	Preemptible                     bool
	Preemptibility                  openapi.NullablePreemptibility
	Completions                     int32
	Parallelism                     int32
	RestartPolicy                   openapi.NullableRestartPolicy
	MasterRestartPolicy             openapi.NullableRestartPolicy
	DistInferenceRestartPolicy      openapi.NullableDistributedInferenceRestartPolicy
}

func LifecycleFlagSet(flagsLC *pflag.FlagSet, specLC *Lifecycle, workloadType spectype.Workload, framework api.Framework) {
	// Add preemptibility flag for all workload types
	flagsLC.Var(cli.NewNullablePreemptibilityFlag(&specLC.Preemptibility),
		consts.AttributeNamePreemptibility,
		consts.AttributeUsagePreemptibility)

	if workloadType == spectype.InferenceDistributed || workloadType == spectype.Inference || workloadType == spectype.NimService {
		return
	}

	flagsLC.Var(&specLC.AutoDeletionTimeAfterCompletion, consts.AttributeNameADTAC, consts.AttributeUsageADTAC)
	flagsLC.Var(&specLC.TerminationGracePeriod,
		AttributeNameTerminationGracePeriod,
		AttributeUsageTerminationGracePeriod)

	flagsLC.Var(cli.NewNullableRestartPolicyFlag(&specLC.RestartPolicy), consts.AttributeNameRestartPolicy, consts.AttributeUsageRestartPolicy)
	flagsLC.Var(cli.NewNullableInt32Flag(&specLC.BackoffLimit), consts.AttributeNameBackoffLimit, consts.AttributeUsageBackoffLimit)

	switch workloadType {
	case spectype.Workspace:
		flagsLC.BoolVar(&specLC.Preemptible, consts.AttributeNamePreemptible, false, consts.AttributeUsagePreemptible)
		_ = flagsLC.MarkDeprecated(consts.AttributeNamePreemptible, "use --preemptibility instead")
	case spectype.Training:
		flagsLC.Int32Var(&specLC.Completions, consts.AttributeNameCompletions, 0, consts.AttributeUsageCompletions)
		_ = flagsLC.MarkDeprecated(consts.AttributeNameCompletions, consts.AttributeUsageCompletions)
		flagsLC.Int32Var(&specLC.Completions, consts.AttributeNameRuns, 0, consts.AttributeUsageRuns)
		flagsLC.Int32Var(&specLC.Parallelism, consts.AttributeNameParallelism, 0, consts.AttributeUsageParallelism)
	case spectype.TrainingDistributed:
		// Skip master-restart-policy for Jax
		if framework != api.Jax {
			flagsLC.Var(cli.NewNullableRestartPolicyFlag(&specLC.MasterRestartPolicy), consts.AttributeNameMasterRestartPolicy, consts.AttributeUsageMasterRestartPolicy)
		}
	default:
		// for backward compatibility (submit flag)
		flagsLC.Int32Var(&specLC.Completions, consts.AttributeNameCompletions, 0, consts.AttributeUsageCompletions)
		flagsLC.Int32Var(&specLC.Completions, consts.AttributeNameRuns, 0, consts.AttributeUsageRuns)
		flagsLC.Int32Var(&specLC.Parallelism, consts.AttributeNameParallelism, 0, consts.AttributeUsageParallelism)
		flagsLC.BoolVar(&specLC.Preemptible, consts.AttributeNamePreemptible, false, consts.AttributeUsagePreemptible)
	}
}

// DistInferenceLifecycleFlagSet adds lifecycle flags for Distributed Inference (not the common Lifecycle Flags )
func DistInferenceLifecycleFlagSet(flagsLC *pflag.FlagSet, specLC *Lifecycle) {
	flagsLC.Var(cli.NewDistributedInferenceRestartPolicyFlag(&specLC.DistInferenceRestartPolicy), consts.AttributeNameRestartPolicy, consts.AttributeUsageDistInferenceRestartPolicy)
	flagsLC.Var(&specLC.TerminationGracePeriod,
		AttributeNameTerminationGracePeriod,
		AttributeUsageTerminationGracePeriod)
}
