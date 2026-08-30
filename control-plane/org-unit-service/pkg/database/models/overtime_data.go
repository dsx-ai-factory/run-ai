package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OvertimeData struct {
	ID            uuid.UUID
	DeletedAt     gorm.DeletedAt
	OrgUnitDataID uuid.UUID

	AvgGpuAllocation24h        *float64
	AvgGpuUtilization24h       *float64
	AvgGpuMemoryUtilization24h *float64
	Avg24hRangeUpdatedAt       *time.Time
	AvgGpuAllocation7d         *float64
	AvgGpuUtilization7d        *float64
	AvgGpuMemoryUtilization7d  *float64
	Avg7dRangeUpdatedAt        *time.Time
	AvgGpuAllocation30d        *float64
	AvgGpuUtilization30d       *float64
	AvgGpuMemoryUtilization30d *float64
	Avg30dRangeUpdatedAt       *time.Time
}

func (OvertimeData) TableName() string {
	return OvertimeDataTable
}

func (m *OvertimeData) BeforeCreate(tx *gorm.DB) (err error) {
	emptyId := uuid.UUID{}
	if m.ID == emptyId {
		m.ID = uuid.New()
	}
	return
}
