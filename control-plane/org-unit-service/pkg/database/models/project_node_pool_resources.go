/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"github.com/google/uuid"
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils/validators"
)

type ProjectNodePoolResources struct {
	dao_utils.DeleteBaseModel
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

func (ProjectNodePoolResources) FilterByFieldsConfigurationForSum() validators.FieldsValidation {
	return map[string]validators.FieldValidator{
		"orgUnitName": validators.NewAlwaysValidFieldValidation("org_unit_name"),
		"orgUnitId":   validators.NewAlwaysValidFieldValidation("org_unit_id"),
		"parentId":    validators.NewAlwaysValidFieldValidation("parent_id"),
		"rank":        validators.NewAlwaysValidFieldValidation("rank"),
		"clusterId":   validators.NewUUIDFieldValidation("cluster_id"),
	}
}

func (ProjectNodePoolResources) FilterByFieldsConfiguration() validators.FieldsValidation {
	return map[string]validators.FieldValidator{
		"orgUnitName":  validators.NewAlwaysValidFieldValidation("org_unit_name"),
		"orgUnitId":    validators.NewAlwaysValidFieldValidation("org_unit_id"),
		"nodePoolName": validators.NewAlwaysValidFieldValidation(ResourcesNodePoolNameColumn),
		"nodePoolId":   validators.NewAlwaysValidFieldValidation(ResourcesNodePoolIdColumn),
		"parentId":     validators.NewAlwaysValidFieldValidation("parent_id"),
		"rank":         validators.NewAlwaysValidFieldValidation("rank"),
		"clusterId":    validators.NewUUIDFieldValidation("cluster_id"),
	}
}

func (ProjectNodePoolResources) SortByFieldsConfiguration() validators.FieldsValidation {
	filterByValidation := ProjectNodePoolResources{}.FilterByFieldsConfiguration()
	result := make(validators.FieldsValidation)
	for k, v := range filterByValidation {
		result[k] = v
	}

	result["gpuDeserved"] = validators.NewAlwaysValidFieldValidation("gpu_deserved")
	result["gpuOverQuotaWeight"] = validators.NewAlwaysValidFieldValidation("gpu_over_quota_weight")
	return result
}
