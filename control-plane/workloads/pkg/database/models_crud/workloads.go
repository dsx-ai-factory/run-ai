package models_crud

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"gorm.io/gorm"
)

const (
	WorkloadsTableName          = "crud_workloads"
	WorkloadsColumnTenantId     = "tenant_id"
	WorkloadsColumnClusterId    = "cluster_id"
	WorkloadsColumnId           = "id"
	WorkloadsColumnType         = "workload_type"
	WorkloadsColumnDesiredPhase = "desired_phase"
	WorkloadsColumnUpdatedBy    = "updated_by"
	WorkloadsColumnUpdatedAt    = "updated_at"
	WorkloadsColumnUpdateCount  = "update_count"
	WorkloadsColumnProjectId    = "project_id"
)

type Workload struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	CreatedAt time.Time
	CreatedBy string

	// I want the DB updated_at to be identical to the value I provide in this field and do not want GORM to
	// attempt to stamp it with its own time of update. this is important since we're using the time to sync
	// the desired phase with workloads table.
	UpdatedAt time.Time `gorm:"autoUpdateTime:false"`
	UpdatedBy string

	DeletedAt gorm.DeletedAt
	DeletedBy string

	UpdateCount int64

	TenantId     int32
	ClusterId    string
	DepartmentId string
	ProjectId    string

	WorkloadType string

	RequestedName string

	Name      string
	Namespace string

	Spec       *workloadsApi.SupersetSpec `gorm:"serializer:json; type:jsonb"`
	MasterSpec *workloadsApi.MasterSpec   `gorm:"serializer:json; type:jsonb"`

	MasterSameAsWorker *bool

	DesiredPhase *workloadsApi.WorkloadDesiredPhase

	AssetsMapping []workloadsApi.AssetsMapping `gorm:"serializer:json; type:jsonb"`
}

func (Workload) TableName() string {
	return WorkloadsTableName
}

func (w Workload) DebugInfo() string {
	result := fmt.Sprintf("[%d] %s (cluster %s, project %s)", w.TenantId, w.Name, w.ClusterId, w.ProjectId)
	return result
}

// WorkloadSyncUpdate is the information about the workload that we send to the agent.
// Apart from the workload itself it contains also the update counter.
type WorkloadSyncUpdate struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	CreatedAt time.Time
	CreatedBy string

	UpdateCount int64

	TenantId     int32
	ClusterId    string
	DepartmentId string
	ProjectId    string

	WorkloadType string

	RequestedName string

	Name      string
	Namespace string

	Spec       api.SupersetSpec  `gorm:"serializer:json; type:jsonb"`
	MasterSpec *api.SupersetSpec `gorm:"serializer:json; type:jsonb"`

	MasterSameAsWorker *bool

	DesiredPhase *api.WorkloadDesiredPhase

	AssetsMapping []api.AssetsMapping `gorm:"serializer:json; type:jsonb"`
}

func (WorkloadSyncUpdate) TableName() string {
	return WorkloadsTableName
}
