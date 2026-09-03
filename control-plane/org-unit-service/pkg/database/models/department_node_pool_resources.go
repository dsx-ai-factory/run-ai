/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/google/uuid"
)

type DepartmentNodePoolResources struct {
	dbmodels.DeleteBaseModel
	TenantId     string
	ClusterId    uuid.UUID
	OrgUnitId    string
	OrgUnitName  string
	NodePoolId   string
	NodePoolName string
	Rank         string
	ResourcesPerNodePool
}

func (DepartmentNodePoolResources) FilterByFieldsToColumnForSum() queryapi.APIFieldsMapping {
	return queryapi.APIFieldsMapping{
		"orgUnitName": queryapi.NewFieldMapping("org_unit_name"),
		"orgUnitId":   queryapi.NewFieldMapping("org_unit_id"),
		"rank":        queryapi.NewFieldMapping("rank"),
		"clusterId":   queryapi.NewFieldMapping("cluster_id", queryapi.UUIDValidator{}),
	}
}

func (DepartmentNodePoolResources) FilterByFieldsToColumn() queryapi.APIFieldsMapping {
	return queryapi.APIFieldsMapping{
		"orgUnitName":  queryapi.NewFieldMapping("org_unit_name"),
		"orgUnitId":    queryapi.NewFieldMapping("org_unit_id"),
		"nodePoolName": queryapi.NewFieldMapping(ResourcesNodePoolNameColumn),
		"nodePoolId":   queryapi.NewFieldMapping(ResourcesNodePoolIdColumn),
		"rank":         queryapi.NewFieldMapping("rank"),
		"clusterId":    queryapi.NewFieldMapping("cluster_id", queryapi.UUIDValidator{}),
	}
}

func (DepartmentNodePoolResources) SortByFieldsToColumn() queryapi.APIFieldsMapping {
	mapping := DepartmentNodePoolResources{}.FilterByFieldsToColumn()
	mapping["gpuDeserved"] = queryapi.NewFieldMapping("gpu_deserved")
	mapping["gpuOverQuotaWeight"] = queryapi.NewFieldMapping("gpu_over_quota_weight")
	return mapping
}

func (p *DepartmentNodePoolResources) ToResources() Resources {
	return Resources{
		NodePoolId:            p.NodePoolId,
		NodePoolName:          p.NodePoolName,
		Rank:                  &p.Rank,
		GpuDeserved:           p.GpuDeserved,
		GpuLimit:              p.GpuLimit,
		GpuOverQuotaWeight:    p.GpuOverQuotaWeight,
		CpuDeserved:           p.CpuDeserved,
		CpuLimit:              p.CpuLimit,
		CpuOverQuotaWeight:    p.CpuOverQuotaWeight,
		MemoryDeserved:        p.MemoryDeserved,
		MemoryLimit:           p.MemoryLimit,
		MemoryOverQuotaWeight: p.MemoryOverQuotaWeight,
		MemoryUnits:           p.MemoryUnits,
	}
}
