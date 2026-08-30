package models

import (
	"time"

	"github.com/google/uuid"
)

const WorkloadElementsTableName = "workload_elements"

type WorkloadElement struct {
	ID          string    `gorm:"column:id;primaryKey"`
	WorkloadID  uuid.UUID `gorm:"column:workload_id;primaryKey"`
	TenantID    uint      `gorm:"column:tenant_id"`
	ClusterID   uuid.UUID `gorm:"column:cluster_id"`
	Name        string    `gorm:"column:name"`
	Level       int       `gorm:"column:level"`
	ParentID    *string   `gorm:"column:parent_id"`
	InstanceKey *string   `gorm:"column:instance_key"`
	ReplicaKey  *string   `gorm:"column:replica_key"`
	K8sRefUID   *string   `gorm:"column:k8s_ref_uid"`
	K8sRefName  *string   `gorm:"column:k8s_ref_name"`
	GVKGroup    *string   `gorm:"column:gvk_group"`
	GVKVersion  *string   `gorm:"column:gvk_version"`
	GVKKind     *string   `gorm:"column:gvk_kind"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (WorkloadElement) TableName() string {
	return WorkloadElementsTableName
}
