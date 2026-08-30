package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Pod struct {
	ID                         uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
	CompletedAt                *time.Time
	DeletedAt                  gorm.DeletedAt
	WorkloadID                 uuid.UUID
	RevisionID                 *string
	Containers                 []*Container
	Name                       string
	PriorityClassName          string
	TenantID                   uint
	ClusterID                  uuid.UUID
	ProjectID                  string
	NodeName                   *string
	CurrentNodepool            *string
	RequestedNodepools         pq.StringArray `gorm:"type:text[]"`
	RequestedGPULimit          *float64
	RequestedGPURequest        *float64
	RequestedGPUMemoryLimit    *string
	RequestedGPUMemoryRequest  *string
	RequestedCPULimit          *float64
	RequestedCPURequest        *float64
	RequestedCPUMemoryLimit    *string
	RequestedCPUMemoryRequest  *string
	RequestedMigDevice         *string
	AllocatedGPU               *float64
	AllocatedGPUMemory         *string
	AllocatedCPU               *float64
	AllocatedCPUMemory         *string
	AllocatedMigDevice         *string
	Phase                      string
	PhaseUpdatedAt             time.Time
	Tolerations                TolerationArray `gorm:"type:jsonb"`
	Ip                         *string
	Task                       *string
	NodeAffinity               NodeAffinity    `gorm:"type:jsonb"`
	NetworkTopologyLabelValues *pq.StringArray `gorm:"type:text[]"`
	Elements                   []*PodElement
}

type PodFull struct {
	Pod
	ProjectName  *string `gorm:"column:project_name"`
	WorkloadName *string `gorm:"column:workload_name"`
}

func (Pod) TableName() string {
	return "pods"
}

func (p Pod) SortAndFilterFieldsToColumn() map[string]string {
	return map[string]string{
		"name":            "pods.name",
		"clusterId":       "pods.cluster_id",
		"workloadId":      "pods.workload_id",
		"revisionId":      "pods.revision_id",
		"nodeName":        "node_name",
		"currentNodepool": "current_nodepool",
		"createdAt":       "pods.created_at",
		"completedAt":     "completed_at",
		"deletedAt":       "pods.deleted_at",
		"task":            "task",
		"elementId":       "pod_elements.element_id",
	}
}
