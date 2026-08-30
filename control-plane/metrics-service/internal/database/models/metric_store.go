package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const MetricStoreTable = "metric_stores"

type MetricStore struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	TenantId           int32              `gorm:"not null;index" json:"tenant_id"`
	DatasourceId       string             `gorm:"not null" json:"datasource_id"`
	MetricStoreDetails MetricStoreDetails `gorm:"type:jsonb;not null" json:"metric_store_details"`
}

func (ms *MetricStore) TableName() string {
	return MetricStoreTable
}

func (ms *MetricStore) BeforeCreate(_ *gorm.DB) (err error) {
	ms.ID = uuid.New()
	return nil
}
