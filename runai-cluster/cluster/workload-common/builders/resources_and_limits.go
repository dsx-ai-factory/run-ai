/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

const toMillis float64 = 1e3

func GetResourcesReqAndLimits(ctx context.Context, resourceFields *field_groups.ResourceAllocationFields) (*corev1.ResourceRequirements, []corev1.EnvVar, error) {
	var err error

	result := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{},
		Limits:   corev1.ResourceList{},
	}
	envVars := make([]corev1.EnvVar, 0)

	result.Requests, err = getCpuResourcesReq(ctx, resourceFields)
	if err != nil {
		return nil, nil, err
	}

	result.Limits, err = getCpuResourcesLimit(ctx, resourceFields)
	if err != nil {
		return nil, nil, err
	}

	updateEnvVarsWithGpuReqAndLimit(ctx, resourceFields, &result.Limits, &envVars)

	//
	//   add requests and limits for customer extended resources
	//
	extendedResources := resourceFields.ExtendedResources.Deref()

	for _, nick := range field_types.SortedExtendedResourcesMapKeys(extendedResources) {
		extendedResource := extendedResources[nick]
		if extendedResource.Value == nil {
			continue
		}
		if extendedResource.Value.Quantity != nil {
			quantity, err := resource.ParseQuantity(*extendedResource.Value.Quantity)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid quantity for extended resource %s: %s", ptr.Deref(extendedResource.Value.Name, ""), err.Error())
			}
			name := ptr.Deref(extendedResource.Value.Name, "")
			result.Requests[corev1.ResourceName(name)] = quantity
			result.Limits[corev1.ResourceName(name)] = quantity
		}
	}

	return &result, envVars, nil
}

func getCpuResourcesReq(_ context.Context, resourceFields *field_groups.ResourceAllocationFields) (corev1.ResourceList, error) {

	var err error

	requests := corev1.ResourceList{}

	if resourceFields.Cpu.Deref() > 0 {
		requests[corev1.ResourceCPU] = *resource.NewMilliQuantity(int64(resourceFields.Cpu.Deref()*toMillis), resource.DecimalSI)
	}

	if resourceFields.Memory.Deref() != "" {
		requests[corev1.ResourceMemory], err = resource.ParseQuantity(resourceFields.Memory.Deref())
		if err != nil {
			return nil, fmt.Errorf("invalid memory size %s: %s", resourceFields.Memory.Deref(), err.Error())
		}
	}

	return requests, nil
}

func getCpuResourcesLimit(_ context.Context, resourceFields *field_groups.ResourceAllocationFields) (corev1.ResourceList, error) {
	limits := corev1.ResourceList{}

	if resourceFields.CpuLimit.Deref() > 0 {
		limits[corev1.ResourceCPU] = *resource.NewMilliQuantity(int64(resourceFields.CpuLimit.Deref()*toMillis), resource.DecimalSI)
	}

	if resourceFields.MemoryLimit.Deref() != "" {
		var err error
		limits[corev1.ResourceMemory], err = resource.ParseQuantity(resourceFields.MemoryLimit.Deref())
		if err != nil {
			return nil, fmt.Errorf("invalid memory limit %s: %s", resourceFields.MemoryLimit.Deref(), err.Error())
		}
	}

	return limits, nil
}

func updateEnvVarsWithGpuReqAndLimit(ctx context.Context, resourceFields *field_groups.ResourceAllocationFields,
	limits *corev1.ResourceList, envVars *[]corev1.EnvVar) {
	//
	//  number of GPU devices is expected using gpuDevices fields, but for backward compatibility
	//  we also allow the gpu field to report it
	//
	var gpuDevicesRequest int

	if resourceFields.Gpu.Deref() > 0 && !IsFractionalGpu(ctx, resourceFields.Gpu.DerefStr()) {
		gpuDevicesRequest = int(resourceFields.Gpu.Deref())
	} else if resourceFields.GpuDevices.Deref() > 0 {
		gpuDevicesRequest = resourceFields.GpuDevices.Deref()
	}

	if gpuDevicesRequest > 0 {
		// only if the gpu is fractional
		if resourceFields.Gpu.Deref() >= 1 || resourceFields.Gpu.Deref() == 0 {
			if resourceFields.GpuMemory.Deref() == "" {
				(*limits)[constants.NvidiaGpuResourceName] = *resource.NewQuantity(int64(gpuDevicesRequest), resource.DecimalSI)
			}
		}
	}

	//
	//   either memory limit or gpu limit can be specified, depending on the unit in which the user
	//   have chosen to specify the gpu request
	//
	if resourceFields.GpuMemory.IsValueInitialized() && resourceFields.GpuMemoryLimit.IsValueInitialized() {
		*envVars = append(*envVars, corev1.EnvVar{
			Name:  constants.GpuMemoryLimitEnvVar,
			Value: resourceFields.GpuMemoryLimit.Deref(),
		})
	}

	if resourceFields.Gpu.IsValueInitialized() && resourceFields.GpuLimit.IsValueInitialized() {
		*envVars = append(*envVars, corev1.EnvVar{
			Name:  constants.GpuMemoryLimitEnvVar,
			Value: resourceFields.GpuLimit.DerefStr(),
		})
	}
}
