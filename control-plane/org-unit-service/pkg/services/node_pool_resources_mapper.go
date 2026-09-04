/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	cluster_openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"k8s.io/utils/ptr"
)

func mapModelNodePoolResources(nodePoolResources models.ResourcesPerNodePool, nodePoolId string, nodePoolsMetadata map[string]cluster_openapi.NodepoolMinimal) (api.ResourcesPerNodePool, error) {
	capabilities, err := GetSchedulingCapabilitiesPerNodePoolId(nodePoolsMetadata, nodePoolId)
	if err != nil {
		return api.ResourcesPerNodePool{}, err
	}
	adjustableWeightsFlag := capabilities.AdjustableWeights
	resources := api.ResourcesPerNodePool{
		Gpu: api.ResourcesPerNodePoolGpu{
			Deserved:        nodePoolResources.GpuDeserved,
			Limit:           nodePoolResources.GpuLimit,
			OverQuotaWeight: nodePoolResources.GpuOverQuotaWeight,
		},
	}
	if !adjustableWeightsFlag {
		resources.Gpu.OverQuotaWeight = PtrOrNil(nodePoolResources.GpuDeserved)
	}
	if nodePoolResources.CpuDeserved != nil || nodePoolResources.CpuLimit != nil || nodePoolResources.CpuOverQuotaWeight != nil {
		resources.Cpu = &api.ResourcesCpu{
			Deserved:        nodePoolResources.CpuDeserved,
			Limit:           nodePoolResources.CpuLimit,
			OverQuotaWeight: nodePoolResources.CpuOverQuotaWeight,
		}
		if !adjustableWeightsFlag {
			resources.Cpu.OverQuotaWeight = PtrOrNil(nodePoolResources.CpuDeserved)
		}
	}
	if nodePoolResources.MemoryDeserved != nil || nodePoolResources.MemoryLimit != nil || nodePoolResources.MemoryOverQuotaWeight != nil {
		resources.Memory = &api.ResourcesMemory{
			Deserved:        nodePoolResources.MemoryDeserved,
			Limit:           nodePoolResources.MemoryLimit,
			OverQuotaWeight: nodePoolResources.MemoryOverQuotaWeight,
			Units:           ModelUnitsToApi(nodePoolResources.MemoryUnits),
		}
		if !adjustableWeightsFlag {
			resources.Memory.OverQuotaWeight = PtrOrNil(nodePoolResources.MemoryDeserved)
		}
	}
	return resources, nil
}

func MapModelProjectNodePoolResourcesToApi(nodePoolResources models.ProjectNodePoolResources, nodePoolsMetadata map[string]cluster_openapi.NodepoolMinimal) *api.ProjectNodePoolResources {
	resources, _ := mapModelNodePoolResources(nodePoolResources.ResourcesPerNodePool, nodePoolResources.NodePoolId, nodePoolsMetadata)
	response := &api.ProjectNodePoolResources{
		Id:           nodePoolResources.ID.String(),
		OrgUnitId:    nodePoolResources.OrgUnitId,
		OrgUnitName:  nodePoolResources.OrgUnitName,
		ParentId:     nodePoolResources.ParentId,
		ClusterId:    nodePoolResources.ClusterId.String(),
		NodePoolId:   nodePoolResources.NodePoolId,
		NodePoolName: nodePoolResources.NodePoolName,
		Rank:         nodePoolResources.Rank,
		Resources:    resources,
		QuotaStatus:  resourcesStatusToQuotaStatus(nodePoolResources.ResourcesStatus),
	}
	return response
}

func resourcesStatusToQuotaStatus(status models.ResourcesStatus) api.QuotaStatus {
	return api.QuotaStatus{
		Allocated: api.QuotaStatusResource{
			Gpu:    ptr.Deref(status.AllocatedGpu, 0),
			Cpu:    ptr.Deref(status.AllocatedCpu, 0),
			Memory: ptr.Deref(status.AllocatedMemory, 0),
		},
		AllocatedNonPreemptible: api.QuotaStatusResource{
			Gpu:    ptr.Deref(status.AllocatedNonPreemptibleGpu, 0),
			Cpu:    ptr.Deref(status.AllocatedNonPreemptibleCpu, 0),
			Memory: ptr.Deref(status.AllocatedNonPreemptibleMemory, 0),
		},
		Requested: api.QuotaStatusResource{
			Gpu:    ptr.Deref(status.RequestedGpu, 0),
			Cpu:    ptr.Deref(status.RequestedCpu, 0),
			Memory: ptr.Deref(status.RequestedMemory, 0),
		},
	}
}

func MapModelDepartmentNodePoolResourcesToApi(nodePoolResources models.DepartmentNodePoolResources, nodePoolsMetadata map[string]cluster_openapi.NodepoolMinimal) *api.DepartmentNodePoolResources {
	resources, _ := mapModelNodePoolResources(nodePoolResources.ResourcesPerNodePool, nodePoolResources.NodePoolId, nodePoolsMetadata)
	response := &api.DepartmentNodePoolResources{
		Id:           nodePoolResources.ID.String(),
		OrgUnitId:    nodePoolResources.OrgUnitId,
		OrgUnitName:  nodePoolResources.OrgUnitName,
		ClusterId:    nodePoolResources.ClusterId.String(),
		NodePoolId:   nodePoolResources.NodePoolId,
		NodePoolName: nodePoolResources.NodePoolName,
		Rank:         nodePoolResources.Rank,
		Resources:    resources,
	}
	return response
}
