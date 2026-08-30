package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy string
	UpdatedBy string
}

func (m *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	m.ID = uuid.New()
	return
}
