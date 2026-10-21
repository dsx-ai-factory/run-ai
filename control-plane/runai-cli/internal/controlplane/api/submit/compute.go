/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"errors"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/spf13/pflag"
)

type Compute struct {
	// GPU
	GpuDeviceRequest        openapi.NullableInt32
	GpuRequestType          string
	GpuPortionRequest       openapi.NullableFloat64
	GpuPortionLimit         openapi.NullableFloat64
	GpuMemoryPortionRequest openapi.NullableFloat64
	GpuMemoryPortionLimit   openapi.NullableFloat64
	GpuComputeMode          string
	GpuMemoryRequest        string
	GpuMemoryLimit          string
	// CPU
	CpuCoreRequest   openapi.NullableFloat64
	CpuCoreLimit     openapi.NullableFloat64
	CpuMemoryRequest string
	CpuMemoryLimit   string
	// Shared Memory
	LargeShmRequest bool
	// Extended
	ExtendedResources []string
	// Distributed Training (Master)
	MasterGpuDeviceRequest        openapi.NullableInt32
	MasterGpuPortionRequest       openapi.NullableFloat64
	MasterGpuPortionLimit         openapi.NullableFloat64
	MasterGpuMemoryPortionRequest openapi.NullableFloat64
	MasterGpuMemoryPortionLimit   openapi.NullableFloat64
	MasterGpuComputeMode          string
	MasterExtendedResources       []string
	MasterCpuCoreRequest          openapi.NullableFloat64
	MasterCpuCoreLimit            openapi.NullableFloat64
	MasterCpuMemoryRequest        string
	MasterCpuMemoryLimit          string
	// Distributed Inference (Worker)
	WorkerGpuDevicesRequest       openapi.NullableInt32
	WorkerGpuPortionRequest       openapi.NullableFloat64
	WorkerGpuPortionLimit         openapi.NullableFloat64
	WorkerGpuMemoryPortionRequest openapi.NullableFloat64
	WorkerGpuMemoryPortionLimit   openapi.NullableFloat64
	WorkerGpuComputeMode          string
	WorkerGpuMemoryRequest        string
	WorkerGpuMemoryLimit          string
	WorkerCpuCoreRequest          openapi.NullableFloat64
	WorkerCpuCoreLimit            openapi.NullableFloat64
	WorkerCpuMemoryRequest        string
	WorkerCpuMemoryLimit          string
	WorkerExtendedResources       []string
}

func (c *Compute) GenerateComputeSpec() (*openapi.WorkspaceSpecSpecAllOfCompute, error) {
	compute := openapi.NewWorkspaceSpecSpecAllOfCompute()
	fieldsSet := false

	if c.GpuDeviceRequest.IsSet() {
		compute.GpuDevicesRequest.Set(c.GpuDeviceRequest.Get())
		fieldsSet = true
	}

	err := specset.SetGenericType[openapi.GpuRequestType, openapi.NullableGpuRequestType](
		c.GpuRequestType, openapi.NewGpuRequestTypeFromValue,
		&compute.GpuRequestType, openapi.NewNullableGpuRequestType)
	if err != nil && !errors.Is(err, specset.ErrEmptyValue) {
		return nil, err
	}
	fieldsSet = fieldsSet || compute.GpuRequestType.IsSet()

	err = specset.SetGenericType[openapi.GpuComputeMode, openapi.NullableGpuComputeMode](
		c.GpuComputeMode, openapi.NewGpuComputeModeFromValue,
		&compute.GpuComputeMode, openapi.NewNullableGpuComputeMode)
	if err != nil && !errors.Is(err, specset.ErrEmptyValue) {
		return nil, err
	}
	fieldsSet = fieldsSet || compute.GpuComputeMode.IsSet()

	if c.GpuMemoryPortionRequest.IsSet() {
		compute.GpuMemoryPortionRequest = c.GpuMemoryPortionRequest
		fieldsSet = true
	}
	if c.GpuMemoryPortionLimit.IsSet() {
		compute.GpuMemoryPortionLimit = c.GpuMemoryPortionLimit
		fieldsSet = true
	}

	if c.GpuPortionRequest.IsSet() {
		specset.SetNullableFloat64(*c.GpuPortionRequest.Get(), &compute.GpuPortionRequest)
		fieldsSet = true
	}

	if c.GpuPortionLimit.IsSet() {
		specset.SetNullableFloat64(*c.GpuPortionLimit.Get(), &compute.GpuPortionLimit)
		fieldsSet = true
	}

	specset.SetNullableString(c.GpuMemoryRequest, &compute.GpuMemoryRequest)
	fieldsSet = fieldsSet || compute.GpuMemoryRequest.IsSet()

	specset.SetNullableString(c.GpuMemoryLimit, &compute.GpuMemoryLimit)
	fieldsSet = fieldsSet || compute.GpuMemoryLimit.IsSet()

	if c.CpuCoreRequest.IsSet() {
		specset.SetNullableFloat64(*c.CpuCoreRequest.Get(), &compute.CpuCoreRequest)
		fieldsSet = true
	}

	if c.CpuCoreLimit.IsSet() {
		specset.SetNullableFloat64(*c.CpuCoreLimit.Get(), &compute.CpuCoreLimit)
		fieldsSet = true
	}

	specset.SetNullableString(c.CpuMemoryRequest, &compute.CpuMemoryRequest)
	fieldsSet = fieldsSet || compute.CpuMemoryRequest.IsSet()

	specset.SetNullableString(c.CpuMemoryLimit, &compute.CpuMemoryLimit)
	fieldsSet = fieldsSet || compute.CpuMemoryLimit.IsSet()

	specset.SetNullableBool(c.LargeShmRequest, &compute.LargeShmRequest)
	fieldsSet = fieldsSet || compute.LargeShmRequest.IsSet()

	if len(c.ExtendedResources) > 0 {
		err = c.setExtendedResources(&compute.ExtendedResources)
		if err != nil {
			return nil, err
		}
		fieldsSet = fieldsSet || (len(compute.ExtendedResources) > 0)
	}

	if !fieldsSet {
		return nil, nil
	}

	return compute, nil
}

func (c *Compute) setExtendedResources(extendedResource *[]openapi.ExtendedResource) error {
	if len(c.ExtendedResources) < 1 {
		return nil
	}

	for _, extended := range c.ExtendedResources {
		split, err := specset.SplitEqualSepValidateTwoSection(extended)
		if err != nil {
			return err
		}
		*extendedResource = append(*extendedResource, openapi.ExtendedResource{
			Resource: *openapi.NewNullableString(&split[0]),
			Quantity: *openapi.NewNullableString(&split[1]),
		})
	}
	return nil
}

func ComputerFlagSet(flagsCompute *pflag.FlagSet, specCompute *Compute, getType spectype.Workload, framework api.Framework, exist bool) {
	if getType == spectype.Submit {
		flagsCompute.VarP(cli.NewNullableNonNegativeFloat64Flag(&specCompute.GpuPortionRequest), consts.AttributeBCNameGpu, "g", consts.AttributeBCUsageGpu)
		flagsCompute.StringVar(&specCompute.GpuMemoryRequest, consts.AttributeBCNameGpuMemory, "", consts.AttributeUsageGpuMemoryRequest)
		flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&specCompute.CpuCoreRequest), consts.AttributeBCNameCpu, consts.AttributeUsageCpuCoreRequest)
		flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&specCompute.CpuCoreLimit), consts.AttributeBCNameCpuLimit, consts.AttributeUsageCpuCoreLimit)
		flagsCompute.StringVar(&specCompute.CpuMemoryRequest, consts.AttributeBCNameMemory, "", consts.AttributeUsageCpuMemoryRequest)
		flagsCompute.StringVar(&specCompute.CpuMemoryLimit, consts.AttributeBCNameMemoryLimit, "", consts.AttributeUsageCpuMemoryLimit)
	} else {
		if exist {
			flagsCompute.Var(cli.NewNullableNonNegativeInt32Flag(&specCompute.GpuDeviceRequest), consts.AttributeNameGpuDeviceRequest, consts.AttributeUsageGpuDeviceRequest)
		} else {
			flagsCompute.VarP(cli.NewNullableNonNegativeInt32Flag(&specCompute.GpuDeviceRequest), consts.AttributeNameGpuDeviceRequest, "g", consts.AttributeUsageGpuDeviceRequest)
		}
		flagsCompute.StringVar(&specCompute.GpuRequestType, consts.AttributeNameGpuRequestType, "", consts.AttributeUsageGpuRequestType)
		flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&specCompute.GpuPortionRequest), consts.AttributeNameGpuPortionRequest, consts.AttributeUsageGpuPortionRequest)
		flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&specCompute.GpuPortionLimit), consts.AttributeNameGpuPortionLimit, consts.AttributeUsageGpuPortionLimit)
		flagsCompute.StringVar(&specCompute.GpuMemoryRequest, consts.AttributeNameGpuMemoryRequest, "", consts.AttributeUsageGpuMemoryRequest)
		flagsCompute.StringVar(&specCompute.GpuMemoryLimit, consts.AttributeNameGpuMemoryLimit, "", consts.AttributeUsageGpuMemoryLimit)
		flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&specCompute.CpuCoreRequest), consts.AttributeNameCpuCoreRequest, consts.AttributeUsageCpuCoreRequest)
		flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&specCompute.CpuCoreLimit), consts.AttributeNameCpuCoreLimit, consts.AttributeUsageCpuCoreLimit)
		flagsCompute.StringVar(&specCompute.CpuMemoryRequest, consts.AttributeNameCpuMemoryRequest, "", consts.AttributeUsageCpuMemoryRequest)
		flagsCompute.StringVar(&specCompute.CpuMemoryLimit, consts.AttributeNameCpuMemoryLimit, "", consts.AttributeUsageCpuMemoryLimit)
	}

	GpuMemoryAndModeFlagSet(flagsCompute, specCompute)

	if getType != spectype.NimService {
		flagsCompute.BoolVar(&specCompute.LargeShmRequest, consts.AttributeNameLargeShm, false, consts.AttributeUsageLargeShm)
		flagsCompute.StringArrayVar(&specCompute.ExtendedResources, consts.AttributeNameExtendedResource,
			[]string{}, consts.AttributeUsageExtendedResource)
	}

	if getType == spectype.InferenceDistributed {
		// Distributed Inference worker-specific compute flags
		DistributedInferenceComputeFlagSet(flagsCompute, specCompute)
	}

	if getType == spectype.TrainingDistributed && framework != api.Jax {
		// Distributed Training master-specific compute flags
		flagsCompute.Var(cli.NewNullableNonNegativeInt32Flag(&specCompute.MasterGpuDeviceRequest), consts.AttributeNameMasterGpuDeviceRequest, consts.AttributeUsageMasterGpuDeviceRequest)
		flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&specCompute.MasterGpuPortionRequest), consts.AttributeNameMasterGpuPortionRequest, consts.AttributeUsageMasterGpuPortionRequest)
		flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&specCompute.MasterGpuPortionLimit), consts.AttributeNameMasterGpuPortionLimit, consts.AttributeUsageMasterGpuPortionLimit)
		flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&specCompute.MasterGpuMemoryPortionRequest), consts.AttributeNameMasterGpuMemoryPortionRequest, consts.AttributeUsageMasterGpuMemoryPortionRequest)
		flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&specCompute.MasterGpuMemoryPortionLimit), consts.AttributeNameMasterGpuMemoryPortionLimit, consts.AttributeUsageMasterGpuMemoryPortionLimit)
		flagsCompute.StringVar(&specCompute.MasterGpuComputeMode, consts.AttributeNameMasterGpuComputeMode, "", consts.AttributeUsageGpuComputeMode)
		if err := flagsCompute.MarkDeprecated(consts.AttributeNameMasterGpuPortionRequest,
			"use --"+consts.AttributeNameMasterGpuMemoryPortionRequest+" for GPU memory fractions, or --"+
				consts.AttributeNameMasterGpuDeviceRequest+" for whole-GPU counts"); err != nil {
			panic(err)
		}
		if err := flagsCompute.MarkDeprecated(consts.AttributeNameMasterGpuPortionLimit, "use --"+consts.AttributeNameMasterGpuMemoryPortionLimit+" instead"); err != nil {
			panic(err)
		}
		if framework == api.MPI {
			if err := flagsCompute.MarkDeprecated(consts.AttributeNameMasterGpuMemoryPortionRequest, consts.WarningSubmitMpuWithGpu); err != nil {
				panic(err)
			}
			if err := flagsCompute.MarkDeprecated(consts.AttributeNameMasterGpuMemoryPortionLimit, consts.WarningSubmitMpuWithGpu); err != nil {
				panic(err)
			}
			if err := flagsCompute.MarkDeprecated(consts.AttributeNameMasterGpuComputeMode, consts.WarningSubmitMpuWithGpu); err != nil {
				panic(err)
			}
			_ = flagsCompute.MarkDeprecated(consts.AttributeNameMasterGpuDeviceRequest, consts.WarningSubmitMpuWithGpu)
			_ = flagsCompute.MarkDeprecated(consts.AttributeNameMasterGpuPortionRequest, consts.WarningSubmitMpuWithGpu)
			_ = flagsCompute.MarkDeprecated(consts.AttributeNameMasterGpuPortionLimit, consts.WarningSubmitMpuWithGpu)
			_ = flagsCompute.MarkDeprecated(consts.AttributeBCMasterNameGpu, consts.WarningSubmitMpuWithGpu)
		}
		flagsCompute.StringArrayVar(&specCompute.MasterExtendedResources, consts.AttributeNameMasterExtendedResource,
			[]string{}, consts.AttributeUsageExtendedResource)
		flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&specCompute.MasterCpuCoreRequest), consts.AttributeNameMasterCpuCoreRequest, consts.AttributeUsageMasterCpuCoreRequest)
		flagsCompute.Var(cli.NewNullableNonNegativeFloat64Flag(&specCompute.MasterCpuCoreLimit), consts.AttributeNameMasterCpuCoreLimit, consts.AttributeUsageMasterCpuCoreLimit)
		flagsCompute.StringVar(&specCompute.MasterCpuMemoryRequest, consts.AttributeNameMasterCpuMemoryRequest, "", consts.AttributeUsageMasterCpuMemoryRequest)
		flagsCompute.StringVar(&specCompute.MasterCpuMemoryLimit, consts.AttributeNameMasterCpuMemoryLimit, "", consts.AttributeUsageMasterCpuMemoryLimit)
	}
}

func GpuMemoryAndModeFlagSet(flags *pflag.FlagSet, compute *Compute) {
	flags.Var(cli.NewNullableNonNegativeFloat64Flag(&compute.GpuMemoryPortionRequest), consts.AttributeNameGpuMemoryPortionRequest, consts.AttributeUsageGpuMemoryPortionRequest)
	flags.Var(cli.NewNullableNonNegativeFloat64Flag(&compute.GpuMemoryPortionLimit), consts.AttributeNameGpuMemoryPortionLimit, consts.AttributeUsageGpuMemoryPortionLimit)
	flags.StringVar(&compute.GpuComputeMode, consts.AttributeNameGpuComputeMode, "", consts.AttributeUsageGpuComputeMode)
	if flags.Lookup(consts.AttributeNameGpuPortionRequest) != nil {
		if err := flags.MarkDeprecated(consts.AttributeNameGpuPortionRequest,
			"use --"+consts.AttributeNameGpuMemoryPortionRequest+" for GPU memory fractions, or --"+
				consts.AttributeNameGpuDeviceRequest+" for whole-GPU counts"); err != nil {
			panic(err)
		}
		if err := flags.MarkDeprecated(consts.AttributeNameGpuPortionLimit, "use --"+consts.AttributeNameGpuMemoryPortionLimit+" instead"); err != nil {
			panic(err)
		}
	}
}
