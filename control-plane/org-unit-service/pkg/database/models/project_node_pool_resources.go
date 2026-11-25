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

// A missing limit allows over-quota GPUs. A missing department pool grants no
// access, while legacy projects without a department use their own limit.
const gpuAccessibleColumn = `(COALESCE(gpu_limit, -1) <> 0 AND (parent_id IS NULL OR EXISTS (
 SELECT 1 FROM department_node_pool_resources_view d
 WHERE d.tenant_id = project_node_pool_resources_view.tenant_id
   AND d.cluster_id = project_node_pool_resources_view.cluster_id
   AND d.org_unit_id = project_node_pool_resources_view.parent_id
   AND d.node_pool_id = project_node_pool_resources_view.node_pool_id
   AND COALESCE(d.gpu_limit, -1) <> 0)))::text`

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
		"orgUnitName":   queryapi.NewFieldMapping("org_unit_name"),
		"orgUnitId":     queryapi.NewFieldMapping("org_unit_id"),
		"nodePoolName":  queryapi.NewFieldMapping(ResourcesNodePoolNameColumn),
		"nodePoolId":    queryapi.NewFieldMapping(ResourcesNodePoolIdColumn),
		"parentId":      queryapi.NewFieldMapping("parent_id"),
		"rank":          queryapi.NewFieldMapping("rank"),
		"clusterId":     queryapi.NewFieldMapping("cluster_id", queryapi.UUIDValidator{}),
		"gpuLimit":      queryapi.NewFieldMapping(gpuLimitOrUnlimitedColumn),
		"gpuAccessible": queryapi.NewFieldMapping(gpuAccessibleColumn),
	}
}

func (ProjectNodePoolResources) SortByFieldsToColumn() queryapi.APIFieldsMapping {
	mapping := ProjectNodePoolResources{}.FilterByFieldsToColumn()
	delete(mapping, "gpuAccessible")
	mapping["gpuDeserved"] = queryapi.NewFieldMapping("gpu_deserved")
	mapping["gpuOverQuotaWeight"] = queryapi.NewFieldMapping("gpu_over_quota_weight")
	return mapping
}
