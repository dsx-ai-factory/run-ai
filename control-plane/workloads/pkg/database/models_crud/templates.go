/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models_crud

import (
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/google/uuid"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
)

const (
	TemplatesTableName = "crud_templates"

	TemplatesColumnId = "id"

	TemplatesColumnWorkloadType = "workload_type"
	TemplatesColumnName         = "name"
	TemplatesColumnWorkloadDesc = "description"

	TemplatesColumnScopeType    = "scope_type"
	TemplatesColumnScopeId      = "scope_id"
	TemplatesColumnTenantId     = "tenant_id"
	TemplatesColumnClusterId    = "cluster_id"
	TemplatesColumnDepartmentId = "department_id"
	TemplatesColumnProjectId    = "project_id"

	TemplatesColumnCreatedBy = "created_by"
	TemplatesColumnCreatedAt = "created_at"
	TemplatesColumnUpdatedBy = "updated_by"
	TemplatesColumnUpdatedAt = "updated_at"

	TemplatesColumnsSearchVector = "search_vector"

	TemplatesColumnRoutingMethod         = "routing_method"
	TemplatesColumnRoutingMethodHostBase = "host-base"
	TemplatesColumnRoutingMethodPathBase = "path-base"
	TemplatesColumnRoutingMethodAgnostic = "agnostic"
)

const (
	AssetCategoryEnvironment = string(assetsApi.ASSETKIND_ENVIRONMENT)
	AssetCategoryCompute     = string(assetsApi.ASSETKIND_COMPUTE)
	AssetCategoryDatasource  = "datasource"
)

type TemplateMeta struct {
	CreatedAt time.Time
	CreatedBy string

	// I want to control the value of this field and prefer GORM will not attempt to stamp it with its own time of update.
	UpdatedAt time.Time `gorm:"autoUpdateTime:false"`
	UpdatedBy string

	WorkloadType string
	Name         string
	Description  *string

	TenantId  uint
	ScopeType string
	ScopeId   string

	ClusterId    *string
	DepartmentId *string
	ProjectId    *string

	RoutingMethod string
}

type Template struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	TemplateMeta `gorm:"embedded"`

	Spec               dbmodels.JsonBMap
	MasterSpec         dbmodels.JsonBMap
	MasterSameAsWorker *bool

	CombinedSpec       dbmodels.JsonBMap
	CombinedMasterSpec dbmodels.JsonBMap

	Runnable      bool
	RunnableIssue *string

	Assets []TemplateAssets `gorm:"foreignKey:TemplateId;references:ID"`
}

func (Template) TableName() string {
	return TemplatesTableName
}

func (t Template) DebugInfo() string {
	result := fmt.Sprintf("[%d] %s/%s %s:%s)", t.TenantId, t.WorkloadType, t.Name, t.ScopeType, t.ScopeId)
	return result
}

type LoadedTemplate struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	TemplateMeta `gorm:"embedded"`

	Spec               api.SupersetSpec        `gorm:"serializer:json; type:jsonb"`
	MasterSpec         *api.MasterTemplateSpec `gorm:"serializer:json; type:jsonb"`
	MasterSameAsWorker *bool

	CombinedSpec       api.SupersetSpec        `gorm:"serializer:json; type:jsonb"`
	CombinedMasterSpec *api.MasterTemplateSpec `gorm:"serializer:json; type:jsonb"`

	Runnable      bool
	RunnableIssue *string

	Assets []TemplateAssets `gorm:"foreignKey:TemplateId;references:ID"`
}

func (LoadedTemplate) TableName() string {
	return TemplatesTableName
}

func (t LoadedTemplate) DebugInfo() string {
	result := fmt.Sprintf("[%d] %s/%s %s:%s)", t.TenantId, t.WorkloadType, t.Name, t.ScopeType, t.ScopeId)
	return result
}

func (Template) FilterFieldsToColumn() queryapi.APIFieldsMapping {
	return queryapi.APIFieldsMapping{
		"name":          queryapi.NewFieldMapping("name"),
		"workloadType":  queryapi.NewFieldMapping("workload_type"),
		"runnable":      queryapi.NewFieldMapping("runnable"),
		"scopeType":     queryapi.NewFieldMapping("scope_type"),
		"clusterId":     queryapi.NewFieldMapping("cluster_id"),
		"projectId":     queryapi.NewFieldMapping("project_id"),
		"departmentId":  queryapi.NewFieldMapping("department_id"),
		"createdAt":     queryapi.NewFieldMapping("created_at"),
		"updatedAt":     queryapi.NewFieldMapping("updated_at"),
		"createdBy":     queryapi.NewFieldMapping("created_by"),
		"updatedBy":     queryapi.NewFieldMapping("updated_by"),
		"routingMethod": queryapi.NewFieldMapping("routing_method"),
	}
}

func (t Template) SortFieldsToColumn() queryapi.APIFieldsMapping {
	return t.FilterFieldsToColumn()
}

func (t LoadedTemplate) ContainedAssetsOfReplica(replicaType api.ReplicaType) (bool, *api.TemplateAssetsIdsOptional) {
	return ContainedAssetsOfReplica(t.Assets, replicaType)
}

func ContainedAssetsOfReplica(templateAssets []TemplateAssets, replicaType api.ReplicaType) (bool, *api.TemplateAssetsIdsOptional) {
	var result api.TemplateAssetsIdsOptional
	var found bool

	for _, asset := range templateAssets {
		if asset.ReplicaType != string(replicaType) {
			continue
		}

		found = true

		switch asset.Category {
		case AssetCategoryEnvironment:
			result.Environment = convutils.Ref(asset.AssetId)
		case AssetCategoryCompute:
			result.Compute = convutils.Ref(asset.AssetId)
		default:
			result.Datasources = append(result.Datasources, asset.AssetId)
		}
	}

	return found, &result
}
