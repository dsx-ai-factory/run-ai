/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"k8s.io/utils/ptr"
)

type Units string

const (
	MB  Units = "MB"
	GB  Units = "GB"
	Mib Units = "Mib"

	DefaultResourceRank    = "MediumLow"
	DefaultOverQuotaWeight = float64(2)

	ResourcesNodePoolIdColumn   = "node_pool_id"
	ResourcesNodePoolNameColumn = "node_pool_name"
)

type Resources struct {
	dbmodels.BaseModel
	NodePoolId            string
	NodePoolName          string
	GpuDeserved           *float64
	GpuLimit              *float64
	GpuOverQuotaWeight    *float64
	CpuDeserved           *float64
	CpuLimit              *float64
	CpuOverQuotaWeight    *float64
	MemoryDeserved        *float64
	MemoryLimit           *float64
	MemoryOverQuotaWeight *float64
	MemoryUnits           Units
	OrgUnitDataID         uuid.UUID
	Rank                  *string `gorm:"column:priority"`
	ResourcesStatus
}

type ResourcesStatus struct {
	AllocatedGpu                  *float64
	AllocatedCpu                  *float64
	AllocatedMemory               *float64
	AllocatedNonPreemptibleGpu    *float64
	AllocatedNonPreemptibleCpu    *float64
	AllocatedNonPreemptibleMemory *float64
	RequestedGpu                  *float64
	RequestedCpu                  *float64
	RequestedMemory               *float64
}

func (Resources) TableName() string {
	return ResourcesTable
}

func (r Resources) Clone() Resources {
	return Resources{
		BaseModel: dbmodels.BaseModel{
			ID:        r.ID,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
			DeletedAt: gorm.DeletedAt{
				Time:  r.DeletedAt.Time,
				Valid: r.DeletedAt.Valid,
			},
			CreatedBy: r.CreatedBy,
			UpdatedBy: r.UpdatedBy,
			DeletedBy: r.DeletedBy,
		},
		NodePoolId:            r.NodePoolId,
		NodePoolName:          r.NodePoolName,
		GpuDeserved:           CopyPtr(r.GpuDeserved),
		GpuLimit:              CopyPtr(r.GpuLimit),
		GpuOverQuotaWeight:    CopyPtr(r.GpuOverQuotaWeight),
		CpuDeserved:           CopyPtr(r.CpuDeserved),
		CpuLimit:              CopyPtr(r.CpuLimit),
		CpuOverQuotaWeight:    CopyPtr(r.CpuOverQuotaWeight),
		MemoryDeserved:        CopyPtr(r.MemoryDeserved),
		MemoryLimit:           CopyPtr(r.MemoryLimit),
		MemoryOverQuotaWeight: CopyPtr(r.MemoryOverQuotaWeight),
		MemoryUnits:           r.MemoryUnits,
		OrgUnitDataID:         r.OrgUnitDataID,
	}
}

func ResourcesArrayEqual(r1, r2 []Resources) bool {
	if len(r1) != len(r2) {
		return false
	}

	r1Map := map[string]Resources{}

	for _, resource := range r1 {
		r1Map[resource.NodePoolId] = resource
	}

	for _, r2Resource := range r2 {
		r1Resource, found := r1Map[r2Resource.NodePoolId]
		if !found {
			return false
		}

		if !ResourcesEqual(r1Resource, r2Resource) {
			return false
		}
	}

	return true
}

func ResourcesEqual(r1, r2 Resources) bool {
	if r1.NodePoolId != r2.NodePoolId ||
		r1.NodePoolName != r2.NodePoolName ||
		r1.OrgUnitDataID != r2.OrgUnitDataID ||
		r1.MemoryUnits != r2.MemoryUnits {
		return false
	}

	// Compare pointer fields
	if !ptr.Equal(r1.GpuDeserved, r2.GpuDeserved) ||
		!ptr.Equal(r1.GpuLimit, r2.GpuLimit) ||
		!ptr.Equal(r1.GpuOverQuotaWeight, r2.GpuOverQuotaWeight) ||
		!ptr.Equal(r1.CpuDeserved, r2.CpuDeserved) ||
		!ptr.Equal(r1.CpuLimit, r2.CpuLimit) ||
		!ptr.Equal(r1.CpuOverQuotaWeight, r2.CpuOverQuotaWeight) ||
		!ptr.Equal(r1.MemoryDeserved, r2.MemoryDeserved) ||
		!ptr.Equal(r1.MemoryLimit, r2.MemoryLimit) ||
		!ptr.Equal(r1.MemoryOverQuotaWeight, r2.MemoryOverQuotaWeight) {
		return false
	}

	return true
}

func convertRawToResources(resourceMap map[string]interface{}) Resources {
	resource := Resources{}
	resource.BaseModel = convertRawToBaseModel(resourceMap)

	if resourceMap[ResourcesNodePoolIdColumn] != nil {
		resource.NodePoolId = resourceMap[ResourcesNodePoolIdColumn].(string)
	}
	if resourceMap[ResourcesNodePoolNameColumn] != nil {
		resource.NodePoolName = resourceMap[ResourcesNodePoolNameColumn].(string)
	}
	if resourceMap["gpu_deserved"] != nil {
		gpuDeserved := (resourceMap["gpu_deserved"].(float64))
		resource.GpuDeserved = &gpuDeserved
	}
	if resourceMap["gpu_limit"] != nil {
		gpuLimit := resourceMap["gpu_limit"].(float64)
		resource.GpuLimit = &gpuLimit
	}
	if resourceMap["gpu_over_quota_weight"] != nil {
		gpuOverQuotaWeight := resourceMap["gpu_over_quota_weight"].(float64)
		resource.GpuOverQuotaWeight = &gpuOverQuotaWeight
	}
	if resourceMap["cpu_deserved"] != nil {
		cpuDeserved := resourceMap["cpu_deserved"].(float64)
		resource.CpuDeserved = &cpuDeserved
	}
	if resourceMap["cpu_limit"] != nil {
		cpuLimit := resourceMap["cpu_limit"].(float64)
		resource.CpuLimit = &cpuLimit
	}
	if resourceMap["cpu_over_quota_weight"] != nil {
		cpuOverQuotaWeight := resourceMap["cpu_over_quota_weight"].(float64)
		resource.CpuOverQuotaWeight = &cpuOverQuotaWeight
	}
	if resourceMap["memory_deserved"] != nil {
		memoryDeserved := resourceMap["memory_deserved"].(float64)
		resource.MemoryDeserved = &memoryDeserved
	}
	if resourceMap["memory_limit"] != nil {
		memoryLimit := resourceMap["memory_limit"].(float64)
		resource.MemoryLimit = &memoryLimit
	}
	if resourceMap["memory_over_quota_weight"] != nil {
		memoryOverQuotaWeight := resourceMap["memory_over_quota_weight"].(float64)
		resource.MemoryOverQuotaWeight = &memoryOverQuotaWeight
	}
	if resourceMap["allocated_gpu"] != nil {
		allocatedGpu := resourceMap["allocated_gpu"].(float64)
		resource.AllocatedGpu = &allocatedGpu
	}
	if resourceMap["allocated_cpu"] != nil {
		allocatedCpu := resourceMap["allocated_cpu"].(float64)
		resource.AllocatedCpu = &allocatedCpu
	}
	if resourceMap["allocated_memory"] != nil {
		allocatedMemory := resourceMap["allocated_memory"].(float64)
		resource.AllocatedMemory = &allocatedMemory
	}
	if resourceMap["allocated_non_preemptible_gpu"] != nil {
		allocatedNonPreemptibleGpu := resourceMap["allocated_non_preemptible_gpu"].(float64)
		resource.AllocatedNonPreemptibleGpu = &allocatedNonPreemptibleGpu
	}
	if resourceMap["allocated_non_preemptible_cpu"] != nil {
		allocatedNonPreemptibleCpu := resourceMap["allocated_non_preemptible_cpu"].(float64)
		resource.AllocatedNonPreemptibleCpu = &allocatedNonPreemptibleCpu
	}
	if resourceMap["allocated_non_preemptible_memory"] != nil {
		allocatedNonPreemptibleMemory := resourceMap["allocated_non_preemptible_memory"].(float64)
		resource.AllocatedNonPreemptibleMemory = &allocatedNonPreemptibleMemory
	}
	if resourceMap["requested_gpu"] != nil {
		requestedGpu := resourceMap["requested_gpu"].(float64)
		resource.RequestedGpu = &requestedGpu
	}
	if resourceMap["requested_cpu"] != nil {
		requestedCpu := resourceMap["requested_cpu"].(float64)
		resource.RequestedCpu = &requestedCpu
	}
	if resourceMap["requested_memory"] != nil {
		requestedMemory := resourceMap["requested_memory"].(float64)
		resource.RequestedMemory = &requestedMemory
	}
	if resourceMap["memory_units"] != nil {
		resource.MemoryUnits = Units(resourceMap["memory_units"].(string))
	}
	if resourceMap["org_unit_data_id"] != nil {
		resource.OrgUnitDataID = uuid.MustParse(resourceMap["org_unit_data_id"].(string))
	}
	if resourceMap["priority"] != nil {
		resource.Rank = ptr.To(resourceMap["priority"].(string))
	}

	return resource
}
