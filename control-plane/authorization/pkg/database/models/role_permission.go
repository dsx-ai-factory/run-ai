package models

import (
	"gorm.io/gorm"
	"time"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
)

type RolePermission struct {
	ID                     uuid.UUID
	CreatedAt              time.Time
	UpdatedAt              time.Time
	ResourceType           api.ResourceType `gorm:"primarykey;index"`
	PermissionSetID        uuid.UUID        `gorm:"primarykey"`
	Create                 bool
	Read                   bool
	Update                 bool
	Delete                 bool
	Sync                   bool
	ReadOnlyForDGXCManaged bool `gorm:"-"`
	ReadOnlyForCSP         bool `gorm:"-"`
}

func (rp *RolePermission) BeforeCreate(tx *gorm.DB) (err error) {
	rp.ID = uuid.New()
	return
}
