package models

import (
	"github.com/google/uuid"
)

type RevisionChild struct {
	RevisionID uuid.UUID
	Revision   Revision `gorm:"foreignKey:RevisionID"`
	ChildID    uuid.UUID
	ChildType  string
}

func (RevisionChild) TableName() string {
	return "revisions_children"
}
