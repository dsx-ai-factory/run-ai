/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gpu_validator

import (
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
)

var (
	// NvidiaGpuExtendedResourceNamePrefixes is a list of strings, that if exist as prefix of an extended resource name, indicate that it's a nvidia GPU resource
	NvidiaGpuExtendedResourceNamePrefixes = []string{
		"nvidia.com/gpu",
		"nvidia.com/mig",
	}
)

// HasRequestedGpus returns true if the provided ResourceAllocationFields has requested GPU (except for request via ExtendedResources)
func HasRequestedGpus(resourceFields *field_groups.ResourceAllocationFields) bool {
	if resourceFields == nil {
		return false
	}
	return resourceFields.GpuDevices.Deref() > 0 ||
		resourceFields.Gpu.Deref() > 0 || resourceFields.GpuLimit.Deref() > 0 ||
		isPositiveQuantity(resourceFields.GpuMemory) || isPositiveQuantity(resourceFields.GpuMemoryLimit) ||
		hasNvidiaGpuInExtendedResources(resourceFields.ExtendedResources)
}

func isPositiveQuantity(qf *flds.QuantityField) bool {
	if qf == nil {
		return false
	}

	q, err := qf.DerefQuantity(0.0)
	if err != nil {
		return false
	}

	return !q.IsZero()
}

// hasNvidiaGpuInExtendedResources returns true if there is a GPU resource request in the workload's ExtendedResources
func hasNvidiaGpuInExtendedResources(extendedResourcesField *flds.ExtendedResourcesField) bool {
	if extendedResourcesField == nil {
		return false
	}

	extendedResources := extendedResourcesField.Deref()

	for _, resource := range extendedResources {
		if resource.Value == nil || resource.Value.Name == nil || resource.Value.Quantity == nil {
			continue
		}

		if !isNvidiaGpuExtendedResource(*resource.Value.Name) {
			continue
		}

		quantity, err := k8sresource.ParseQuantity(*resource.Value.Quantity)
		if err == nil && !quantity.IsZero() {
			return true
		}
	}

	return false
}

func isNvidiaGpuExtendedResource(resourceName string) bool {
	for _, gpuPrefix := range NvidiaGpuExtendedResourceNamePrefixes {
		if strings.HasPrefix(resourceName, gpuPrefix) {
			return true
		}
	}

	return false
}
