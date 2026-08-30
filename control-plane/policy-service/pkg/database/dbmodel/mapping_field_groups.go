package dbmodel

import (
	"time"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbtables"
)

// MappingFieldGroups is one scope-specific row (system or tenant). The same logical name may exist
// twice (system + tenant); the service merges them for API list/get.
type MappingFieldGroups struct {
	ID          uuid.UUID `gorm:"primaryKey"`
	Name        string
	Description *string

	ScopeType api.MappingScope
	TenantId  int32

	Fields        map[string]api.FieldGroupFieldsWithScopeValue `gorm:"serializer:json;type:jsonb;not null"`
	ItemSelectors map[string]api.ItemSelectorsWithScopeValue    `gorm:"serializer:json;type:jsonb;not null"`

	CreatedBy string
	CreatedAt time.Time
	UpdatedBy string
	UpdatedAt time.Time
}

func (MappingFieldGroups) TableName() string {
	return dbtables.MappingFieldGroupsTable
}

func (pf MappingFieldGroups) ToApiFieldGroup() *api.FieldGroup {
	return &api.FieldGroup{
		Id: pf.ID.String(),

		Name:        pf.Name,
		Description: pf.Description,

		Scope: pf.ScopeType,

		Fields:        pf.Fields,
		ItemSelectors: pf.ItemSelectors,

		CreatedBy: pf.CreatedBy,
		CreatedAt: pf.CreatedAt,
		UpdatedBy: pf.UpdatedBy,
		UpdatedAt: pf.UpdatedAt,
	}
}
