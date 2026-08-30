package models

import (
	"github.com/run-ai/runai/runai-application-packages/dbmodels"
)

func (PermissionSet) TableName() string {
	return "permission_sets"
}

type PermissionSet struct {
	dbmodels.DeleteBaseModel
	CreatedBy   string
	UpdatedBy   string
	Name        string
	Description string
	Internal    bool
	Permissions []RolePermission
}
