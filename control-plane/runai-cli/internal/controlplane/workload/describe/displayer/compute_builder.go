/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer

import (
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

// Builder
type ComputeBuilder struct {
	compute Compute
}

func NewComputeBuilder() *ComputeBuilder {
	return &ComputeBuilder{compute: Compute{}}
}

func (b *ComputeBuilder) RequestedResources(resources ComputeResources) *ComputeBuilder {
	b.compute.RequestedResources = resources
	return b
}

func (b *ComputeBuilder) GPUComputeMode(mode string) *ComputeBuilder {
	b.compute.GPUComputeMode = &mode
	return b
}

func (b *ComputeBuilder) AllocatedResources(resources AllocatedResources) *ComputeBuilder {
	b.compute.AllocatedResources = resources
	return b
}

func (b *ComputeBuilder) RequestedNodePools(pools []string) *ComputeBuilder {
	b.compute.RequestedNodePools = pools
	return b
}

func (b *ComputeBuilder) CurrentNodePools(pools []string) *ComputeBuilder {
	b.compute.CurrentNodePools = pools
	return b
}

func (b *ComputeBuilder) Set(compute Compute) *ComputeBuilder {
	b.compute = compute
	return b
}

func (b *ComputeBuilder) Build() Compute {
	return b.compute
}

type ComputeResourcesBuilder struct {
	cr ComputeResources
}

func NewComputeResourcesBuilder() *ComputeResourcesBuilder {
	return &ComputeResourcesBuilder{cr: ComputeResources{
		GPUCompute: RequestedLimitedResources{},
		CPUCores:   RequestedLimitedResources{},
		GPUMemory:  RequestedLimitQuantity{},
		CPUMemory:  RequestedLimitQuantity{},
	}}
}

func (b *ComputeResourcesBuilder) GPUComputeRequest(requested float64) *ComputeResourcesBuilder {
	b.cr.GPUCompute.Requested = requested
	return b
}

func (b *ComputeResourcesBuilder) CPUCoresRequest(requested float64) *ComputeResourcesBuilder {
	b.cr.CPUCores.Requested = requested
	return b
}

func (b *ComputeResourcesBuilder) GPUMemoryRequest(requested string) *ComputeResourcesBuilder {
	b.cr.GPUMemory.Requested = requested
	return b
}

func (b *ComputeResourcesBuilder) CPUMemoryRequest(requested string) *ComputeResourcesBuilder {
	b.cr.CPUMemory.Requested = requested
	return b
}

func (b *ComputeResourcesBuilder) GPUComputeLimit(limit float64) *ComputeResourcesBuilder {
	b.cr.GPUCompute.Limit = &limit
	return b
}

func (b *ComputeResourcesBuilder) CPUCoresLimit(limit float64) *ComputeResourcesBuilder {
	b.cr.CPUCores.Limit = &limit
	return b
}

func (b *ComputeResourcesBuilder) GPUMemoryLimit(limit string) *ComputeResourcesBuilder {
	b.cr.GPUMemory.Limit = limit
	return b
}

func (b *ComputeResourcesBuilder) CPUMemoryLimit(limit string) *ComputeResourcesBuilder {
	b.cr.CPUMemory.Limit = limit
	return b
}

func (b *ComputeResourcesBuilder) Build() ComputeResources {
	return b.cr
}

type AllocatedResourcesBuilder struct {
	ar AllocatedResources
}

func NewAllocatedResourcesBuilder() *AllocatedResourcesBuilder {
	return &AllocatedResourcesBuilder{ar: AllocatedResources{}}
}

func (b *AllocatedResourcesBuilder) GPUCompute(gpuCompute float64) *AllocatedResourcesBuilder {
	b.ar.GPUCompute = gpuCompute
	return b
}

func (b *AllocatedResourcesBuilder) CPUCores(cpuCores float64) *AllocatedResourcesBuilder {
	b.ar.CPUCores = cpuCores
	return b
}

func (b *AllocatedResourcesBuilder) GPUMemory(gpuMemory string) *AllocatedResourcesBuilder {
	b.ar.GPUMemory = gpuMemory
	return b
}

func (b *AllocatedResourcesBuilder) CPUMemory(cpuMemory string) *AllocatedResourcesBuilder {
	b.ar.CPUMemory = cpuMemory
	return b
}

func (b *AllocatedResourcesBuilder) Build() AllocatedResources {
	return b.ar
}

// Helper functions
func BuildComputeFromSpec(spec *openapi.WorkspaceSpecSpecAllOfCompute) *Compute {
	if spec == nil {
		return nil
	}

	resourcesBuilder := NewComputeResourcesBuilder()

	gpuPortionRequest := spec.GpuPortionRequest
	if spec.GpuMemoryPortionRequest.Get() != nil {
		gpuPortionRequest = spec.GpuMemoryPortionRequest
	}
	if gpuPortionRequest.Get() != nil {
		gpuRequest := *gpuPortionRequest.Get()
		if spec.GpuDevicesRequest.IsSet() {
			gpuDevices := float64(spec.GetGpuDevicesRequest())
			if gpuRequest > 0 && gpuRequest <= 1 {
				gpuRequest *= gpuDevices
			} else {
				gpuRequest = gpuDevices
			}
		}
		resourcesBuilder.GPUComputeRequest(gpuRequest)
	} else if spec.GpuDevicesRequest.IsSet() {
		resourcesBuilder.GPUComputeRequest(float64(spec.GetGpuDevicesRequest()))
	}
	gpuPortionLimit := spec.GpuPortionLimit
	if spec.GpuMemoryPortionLimit.Get() != nil {
		gpuPortionLimit = spec.GpuMemoryPortionLimit
	}
	if gpuPortionLimit.Get() != nil {
		gpuLimit := *gpuPortionLimit.Get()
		if spec.GpuDevicesRequest.IsSet() && gpuLimit > 0 && gpuLimit <= 1 {
			gpuLimit *= float64(spec.GetGpuDevicesRequest())
		}
		resourcesBuilder.GPUComputeLimit(gpuLimit)
	}

	if spec.CpuCoreRequest.IsSet() {
		resourcesBuilder.CPUCoresRequest(spec.GetCpuCoreRequest())
	}

	if spec.CpuCoreLimit.IsSet() {
		resourcesBuilder.CPUCoresLimit(spec.GetCpuCoreLimit())
	}

	if spec.CpuMemoryRequest.IsSet() {
		resourcesBuilder.CPUMemoryRequest(spec.GetCpuMemoryRequest())
	}

	if spec.CpuMemoryLimit.IsSet() {
		resourcesBuilder.CPUMemoryLimit(spec.GetCpuMemoryLimit())
	}

	if spec.GpuMemoryRequest.IsSet() {
		resourcesBuilder.GPUMemoryRequest(spec.GetGpuMemoryRequest())
	}

	if spec.GpuMemoryLimit.IsSet() {
		resourcesBuilder.GPUMemoryLimit(spec.GetGpuMemoryLimit())
	}

	computeBuilder := NewComputeBuilder().RequestedResources(resourcesBuilder.Build())
	if mode, _ := spec.GetGpuComputeModeOk(); mode != nil {
		computeBuilder.GPUComputeMode(string(*mode))
	}
	compute := computeBuilder.Build()

	return &compute
}
