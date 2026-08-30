package models

import (
	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	"gorm.io/gorm"
)

const NodeNumaTable = "node_numa"

type NodeNuma struct {
	Common                  `gorm:"embedded"`
	ID                      uuid.UUID `gorm:"primaryKey"`
	TenantId                int32     `gorm:"not null"`
	ClusterId               uuid.UUID `gorm:"not null"`
	NodeId                  uuid.UUID `gorm:"not null"`
	Name                    string    `gorm:"not null"`
	TotalCpuCores           float64   `gorm:"type:decimal(12,3);not null"`
	AllocatedCpuCores       float64   `gorm:"type:decimal(12,3);not null"`
	TotalCpuMemoryBytes     int64     `gorm:"not null"`
	AllocatedCpuMemoryBytes int64     `gorm:"not null"`
	TotalGpus               int32     `gorm:"not null"`
	AllocatedGpus           int32     `gorm:"not null"`
}

func (n *NodeNuma) BeforeCreate(_ *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

func (NodeNuma) TableName() string {
	return NodeNumaTable
}

func (n *NodeNuma) ToApiModel() api.NumaNodeDetails {
	return api.NumaNodeDetails{
		Name:                    n.Name,
		TotalCpuCores:           n.TotalCpuCores,
		AllocatedCpuCores:       n.AllocatedCpuCores,
		TotalCpuMemoryBytes:     n.TotalCpuMemoryBytes,
		AllocatedCpuMemoryBytes: n.AllocatedCpuMemoryBytes,
		TotalGpus:               n.TotalGpus,
		AllocatedGpus:           n.AllocatedGpus,
	}
}
