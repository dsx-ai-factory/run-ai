package models

import (
	"gorm.io/gorm"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
)

type Revision struct {
	ID                        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt                 time.Time
	DeletedAt                 gorm.DeletedAt
	Name                      string
	Type                      string
	WorkloadID                uuid.UUID
	TenantID                  uint
	ClusterID                 uuid.UUID
	ProjectID                 string
	DepartmentID              string
	RequestedGPULimit         *float64
	RequestedGPURequest       *float64
	RequestedGPUMemoryLimit   *string
	RequestedGPUMemoryRequest *string
	RequestedCPULimit         *float64
	RequestedCPURequest       *float64
	RequestedCPUMemoryLimit   *string
	RequestedCPUMemoryRequest *string
	RequestedMigDevice        pq.StringArray `gorm:"type:text[]"`
	Phase                     api.Phase
	PhaseMessage              string
	PhaseUpdatedAt            time.Time
	RequestedPodsNumber       *int32
	RequestedPodsMin          *int32
	RequestedPodsMax          *int32
	RequestedPodsParallelism  *int32
	RequestedPodsCompletions  *int32
	RequestedNodepools        pq.StringArray `gorm:"type:text[]"`
	Images                    pq.StringArray `gorm:"type:text[]"`
	Command                   string
	Arguments                 string
	EnvironmentVariables      EnvironmentVariables `gorm:"type:jsonb"`
	Conditions                ConditionArray       `gorm:"type:jsonb"`
	AdditionalFields          AdditionalFields     `gorm:"type:jsonb"`
}

func (Revision) TableName() string {
	return "revisions"
}
