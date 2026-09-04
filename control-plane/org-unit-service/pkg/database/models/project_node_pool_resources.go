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

type ProjectNodePoolResources struct {
	dbmodels.DeleteBaseModel
	TenantId     string
	ClusterId    uuid.UUID
	OrgUnitId    string
	OrgUnitName  string
	ParentId     string
	NodePoolId   string
	NodePoolName string
	Rank         string
	ResourcesPerNodePool
	ResourcesStatus
}

func (ProjectNodePoolResources) FilterByFieldsToColumnForSum() queryapi.APIFieldsMapping {
	return queryapi.APIFieldsMapping{
		"orgUnitName": queryapi.NewFieldMapping("org_unit_name"),
		"orgUnitId":   queryapi.NewFieldMapping("org_unit_id"),
		"parentId":    queryapi.NewFieldMapping("parent_id"),
		"rank":        queryapi.NewFieldMapping("rank"),
		"clusterId":   queryapi.NewFieldMapping("cluster_id", queryapi.UUIDValidator{}),
	}
}

func (ProjectNodePoolResources) FilterByFieldsToColumn() queryapi.APIFieldsMapping {
	return queryapi.APIFieldsMapping{
		"orgUnitName":  queryapi.NewFieldMapping("org_unit_name"),
		"orgUnitId":    queryapi.NewFieldMapping("org_unit_id"),
		"nodePoolName": queryapi.NewFieldMapping(ResourcesNodePoolNameColumn),
		"nodePoolId":   queryapi.NewFieldMapping(ResourcesNodePoolIdColumn),
		"parentId":     queryapi.NewFieldMapping("parent_id"),
		"rank":         queryapi.NewFieldMapping("rank"),
		"clusterId":    queryapi.NewFieldMapping("cluster_id", queryapi.UUIDValidator{}),
	}
}

func (ProjectNodePoolResources) SortByFieldsToColumn() queryapi.APIFieldsMapping {
	mapping := ProjectNodePoolResources{}.FilterByFieldsToColumn()
	mapping["gpuDeserved"] = queryapi.NewFieldMapping("gpu_deserved")
	mapping["gpuOverQuotaWeight"] = queryapi.NewFieldMapping("gpu_over_quota_weight")
	return mapping
}
