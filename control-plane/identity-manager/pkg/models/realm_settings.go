package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RealmSettings struct {
	ID                uuid.UUID              `gorm:"primaryKey"`
	TenantID          string                 `gorm:"not null"`
	SettingKey        string                 `gorm:"not null"`
	SettingValue      map[string]interface{} `gorm:"serializer:json;type:jsonb"`
	UpdatedAt         time.Time
	UpdatedByCustomer bool `gorm:"not null"`
}

func (m *RealmSettings) BeforeCreate(tx *gorm.DB) (err error) {
	m.ID = uuid.New()
	return
}
