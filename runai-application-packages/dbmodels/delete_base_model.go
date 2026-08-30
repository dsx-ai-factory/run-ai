package dbmodels

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeleteBaseModel struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy string
	UpdatedBy string
}

func (m *DeleteBaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	emptyId := uuid.UUID{}
	if m.ID == emptyId {
		m.ID = uuid.New()
	}
	return
}
