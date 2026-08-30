package models

import "github.com/google/uuid"

const PodElementsTableName = "pod_elements"

type PodElement struct {
	PodID     uuid.UUID `gorm:"column:pod_id;primaryKey"`
	ElementID string    `gorm:"column:element_id"`
	Level     int       `gorm:"column:level;primaryKey"`
}

func (PodElement) TableName() string {
	return PodElementsTableName
}
