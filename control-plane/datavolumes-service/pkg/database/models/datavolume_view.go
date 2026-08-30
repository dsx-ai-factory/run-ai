package models

import (
	"time"

	"github.com/run-ai/runai/control-plane/datavolumes-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/queryapi"
)

const (
	DatavolumeViewTableName               = "datavolumes_view"
	DatavolumeViewSharedScopeFieldsPrefix = "shared_scope_"
)

type DatavolumeView struct {
	Datavolume

	ScopeId   string
	ScopeType api.ScopeType
	ScopeName string

	SharedScopeDepartmentId string
	SharedScopeProjectId    string
	SharedScopeCreatedAt    time.Time
	SharedScopeCreatedBy    string
}

func (DatavolumeView) TableName() string {
	return DatavolumeViewTableName
}

func (DatavolumeView) DistinctDatavolumeColumns() []string {
	return []string{
		"id",
		"created_at",
		"updated_at",
		"created_by",
		"updated_by",
		"name",
		"description",
		"project_id",
		"origin_pvc_name",
		"project_name",
		"department_id",
		"tenant_id",
		"cluster_id",
		"phase",
		"phase_message",
		"conditions",
		"datavolume_pvc_name",
		"datavolume_pv_name",
		"should_delete_original_volume",
		"should_use_original_volume",
	}
}

func (DatavolumeView) SortAndFilterFieldsToColumn() queryapi.APIFieldsMapping {
	return queryapi.APIFieldsMapping{
		"name":         queryapi.NewFieldMapping("name"),
		"projectId":    queryapi.NewFieldMapping("project_id"),
		"clusterId":    queryapi.NewFieldMapping("cluster_id"),
		"departmentId": queryapi.NewFieldMapping("department_id"),
		"projectName":  queryapi.NewFieldMapping("project_name"),
		"createdAt":    queryapi.NewFieldMapping("created_at"),
		"createdBy":    queryapi.NewFieldMapping("created_by"),
		"phase":        queryapi.NewFieldMapping("phase"),
	}
}
