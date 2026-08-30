package models

import (
	"time"

	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
)

type Permission struct {
	CreatedAt              time.Time
	UpdatedAt              time.Time
	RoleID                 uint             `gorm:"primarykey"`
	ResourceType           api.ResourceType `gorm:"primarykey;index"`
	Create                 bool
	Read                   bool
	Update                 bool
	Delete                 bool
	Sync                   bool
	ReadOnlyForDGXCManaged bool `gorm:"-"`
	ReadOnlyForCSP         bool `gorm:"-"`
}
