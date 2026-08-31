/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package spec

import (
	"fmt"
	"math"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/submit"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	AttributeNameFlag = "name"
)

type BCFlags struct {
	WorkloadName string
}

type DistBCSpec struct {
	GeneralSpec
	BcFlags BCFlags
}

func (dbc *DistBCSpec) SetFlagsBc(cmd *cobra.Command) {
	dbc.getGrouper(cmd)

	general := dbc.grouper.GetOrAddFlagSet(consts.GroupNameGeneral)
	GeneralBcFlagSet(general, &dbc.BcFlags, dbc.framework)

	compute := dbc.grouper.GetOrAddFlagSet(consts.GroupNameCompute)
	ComputerBCFlagSet(compute, &dbc.flags.Compute, dbc.framework)
}

func (dbc *DistBCSpec) Mutate() error {
	// Convert worker GPU
	if dbc.flags.Compute.GpuPortionRequest.IsSet() {
		gpuPortion := dbc.flags.Compute.GpuPortionRequest.Get()
		if *gpuPortion > 1 {
			// It's a device count, not a portion
			if *gpuPortion != math.Trunc(*gpuPortion) {
				return fmt.Errorf("gpu value cannot be a fraction in whole numbers: %.2f", *gpuPortion)
			}
			gpuInt32 := int32(*gpuPortion)
			dbc.flags.Compute.GpuDeviceRequest.Set(&gpuInt32)
			dbc.flags.Compute.GpuPortionRequest.Unset()
		}
	}

	// Convert master GPU
	if dbc.flags.Compute.MasterGpuPortionRequest.IsSet() {
		masterGpuPortion := dbc.flags.Compute.MasterGpuPortionRequest.Get()
		if *masterGpuPortion > 1 {
			// It's a device count, not a portion
			if *masterGpuPortion != math.Trunc(*masterGpuPortion) {
				return fmt.Errorf("master gpu value cannot be a fraction in whole numbers: %.2f", *masterGpuPortion)
			}
			masterGpuInt32 := int32(*masterGpuPortion)
			dbc.flags.Compute.MasterGpuDeviceRequest.Set(&masterGpuInt32)
			dbc.flags.Compute.MasterGpuPortionRequest.Unset()
		}
	}

	return nil
}

func (dbc *DistBCSpec) Clear() {
	dbc.BcFlags.WorkloadName = ""
	dbc.GeneralSpec.Clear()
}

func GeneralBcFlagSet(general *pflag.FlagSet, bcFlag *BCFlags, _ api.Framework) {
	general.StringVar(&bcFlag.WorkloadName, AttributeNameFlag, "", "Job name")
}

func ComputerBCFlagSet(compute *pflag.FlagSet, specCompute *submit.Compute, _ api.Framework) {
	compute.VarP(cli.NewNullableNonNegativeFloat64Flag(&specCompute.GpuPortionRequest), consts.AttributeBCNameGpu, "g", consts.AttributeBCUsageGpu)
	compute.Var(cli.NewNullableNonNegativeFloat64Flag(&specCompute.MasterGpuPortionRequest), consts.AttributeBCMasterNameGpu, consts.AttributeBCMasterNameGpu)
	compute.Var(cli.NewNullableNonNegativeFloat64Flag(&specCompute.CpuCoreRequest), consts.AttributeBCNameCpu, consts.AttributeUsageCpuCoreRequest)
	compute.StringVar(&specCompute.CpuMemoryRequest, consts.AttributeBCNameMemory, "", consts.AttributeUsageCpuMemoryRequest)
}

func NewDistributedBCSpec(framework api.Framework) *DistBCSpec {
	return &DistBCSpec{
		GeneralSpec: GeneralSpec{workloadType: spectype.TrainingDistributed,
			framework: framework},
		BcFlags: BCFlags{},
	}
}
