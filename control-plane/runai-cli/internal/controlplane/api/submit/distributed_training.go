/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/spf13/pflag"
)

// DistributedTraining holds configuration for distributed training workloads
type DistributedTraining struct {
	NoMaster               bool
	Workers                openapi.NullableInt32
	SegmentSize            openapi.NullableInt32
	MinReplica             int32
	MaxReplica             int32
	SlotsPerWorker         int32
	DistributedFramework   string
	CleanPodPolicy         string
	SshAuthMountPath       string
	LauncherCreationPolicy openapi.NullableMpiLauncherCreationPolicy
}

func (s *DistributedTraining) GenerateToDistributedFramework() (*openapi.DistributedFramework, error) {
	framework, err := openapi.NewDistributedFrameworkFromValue(s.DistributedFramework)
	return framework, err
}

// DistributedFlagSet deprecated - remove after distributed command is removed
func DistributedFlagSet(flagsDistributed *pflag.FlagSet, specDistributed *DistributedTraining) {
	flagsDistributed.BoolVar(&specDistributed.NoMaster, consts.AttributeNameNoMaster, false, consts.AttributeUsageNoMaster)
	flagsDistributed.StringVarP(&specDistributed.DistributedFramework, consts.AttributeNameDistributedFramework,
		consts.AttributeShortDistributedFramework, "", consts.AttributeUsageDistributedFramework)
	flagsDistributed.Int32Var(&specDistributed.MinReplica, consts.AttributeNameMinReplica, 0, consts.AttributeUsageMinReplica)
	flagsDistributed.Int32Var(&specDistributed.MaxReplica, consts.AttributeNameMaxReplica, 0, consts.AttributeUsageMaxReplica)
	flagsDistributed.Int32Var(&specDistributed.SlotsPerWorker, consts.AttributeNameSlotsPerWorker, 0, consts.AttributeUsageSlotsPerWorker)
}

func DistributedCommonFlagSet(flagsDistributed *pflag.FlagSet, specDistributed *DistributedTraining) {
	flagsDistributed.Var(cli.NewNullableNonNegativeInt32Flag(&specDistributed.Workers), consts.AttributeNameWorkers, consts.AttributeUsageWorkers)
	flagsDistributed.StringVar(&specDistributed.CleanPodPolicy, consts.AttributeNameCleanPodPolicy, "", consts.AttributeUsageCleanPodPolicy)
}

func MPIFlagSet(flagsDistributed *pflag.FlagSet, specDistributed *DistributedTraining) {
	flagsDistributed.Int32Var(&specDistributed.SlotsPerWorker, consts.AttributeNameSlotsPerWorker, 0, consts.AttributeUsageSlotsPerWorker)
	flagsDistributed.StringVar(&specDistributed.SshAuthMountPath, consts.AttributeNameSshAuthMountPath, "", consts.AttributeUsageSshAuthMountPath)
	flagsDistributed.Var(cli.NewNullableLauncherCreationPolicyFlag(&specDistributed.LauncherCreationPolicy), consts.AttributeNameLauncherCreationPolicy, consts.AttributeUsageLauncherCreationPolicy)
	specDistributed.DistributedFramework = string(openapi.DISTRIBUTEDFRAMEWORK_MPI)
}

func TensorFlowFlagSet(flagsDistributed *pflag.FlagSet, specDistributed *DistributedTraining) {
	pytorchTensorFlowFlagSet(flagsDistributed, specDistributed)
	specDistributed.DistributedFramework = string(openapi.DISTRIBUTEDFRAMEWORK_TF)
}

func PyTorchFlagSet(flagsDistributed *pflag.FlagSet, specDistributed *DistributedTraining) {
	pytorchTensorFlowFlagSet(flagsDistributed, specDistributed)
	flagsDistributed.Var(cli.NewNullableNonNegativeInt32Flag(&specDistributed.SegmentSize),
		consts.AttributeNameSegmentSize, consts.AttributeUsageSegmentSize)
	flagsDistributed.Int32Var(&specDistributed.MinReplica, consts.AttributeNameMinReplica, 0, consts.AttributeUsageMinReplica)
	flagsDistributed.Int32Var(&specDistributed.MaxReplica, consts.AttributeNameMaxReplica, 0, consts.AttributeUsageMaxReplica)
	specDistributed.DistributedFramework = string(openapi.DISTRIBUTEDFRAMEWORK_PY_TORCH)
}

func JaxFlagSet(_ *pflag.FlagSet, specDistributed *DistributedTraining) {
	specDistributed.DistributedFramework = string(openapi.DISTRIBUTEDFRAMEWORK_JAX)
	specDistributed.NoMaster = true
}

func XgboostFlagSet(_ *pflag.FlagSet, specDistributed *DistributedTraining) {
	specDistributed.DistributedFramework = string(openapi.DISTRIBUTEDFRAMEWORK_XG_BOOST)
}

func pytorchTensorFlowFlagSet(flagsDistributed *pflag.FlagSet, specDistributed *DistributedTraining) {
	flagsDistributed.BoolVar(&specDistributed.NoMaster, consts.AttributeNameNoMaster, false, consts.AttributeUsageNoMaster)
}
