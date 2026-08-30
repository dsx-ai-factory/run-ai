package models

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/queryapi"
	"gorm.io/gorm"
	"k8s.io/utils/ptr"
)

const (
	StorageClassConfigurationsTable          = "storage_class_configurations"
	StorageClassConfigurationIDColumn        = "id"
	StorageClassConfigurationTenantIdColumn  = "tenant_id"
	StorageClassConfigurationScopeTypeColumn = "scope_type"
	StorageClassConfigurationScopeIdColumn   = "scope_id"
	StorageClassConfigurationNameColumn      = "name"
)

type StorageClassConfigurationFilters struct {
	TenantId  int32
	ScopeType *string
	ScopeId   *string
	Name      *string
}

type StorageClassConfiguration struct {
	ID                               uuid.UUID `gorm:"primaryKey"`
	StorageClassConfigurationFilters `gorm:"embedded"`
	Permissions                      *api.StorageClassPermissions   `gorm:"serializer:json; type:jsonb"`
	Customization                    *api.StorageClassCustomization `gorm:"serializer:json; type:jsonb"`
}

func (sccf *StorageClassConfigurationFilters) TableName() string {
	return StorageClassConfigurationsTable
}

func (scc *StorageClassConfiguration) TableName() string {
	return StorageClassConfigurationsTable
}

func (scc *StorageClassConfiguration) TenantIdColumn() string {
	return StorageClassConfigurationTenantIdColumn
}

func (scc *StorageClassConfiguration) ScopeTypeColumn() string {
	return StorageClassConfigurationScopeTypeColumn
}

func (scc *StorageClassConfiguration) ScopeIdColumn() string {
	return StorageClassConfigurationScopeIdColumn
}

func (scc *StorageClassConfiguration) SortAndFilterFieldsToColumn() queryapi.APIFieldsMapping {
	return queryapi.APIFieldsMapping{
		"tenantId":  queryapi.NewFieldMapping(StorageClassConfigurationTenantIdColumn),
		"scopeType": queryapi.NewFieldMapping(StorageClassConfigurationScopeTypeColumn),
		"scopeId":   queryapi.NewFieldMapping(StorageClassConfigurationScopeIdColumn),
		"name":      queryapi.NewFieldMapping(StorageClassConfigurationNameColumn),
	}
}

func NewStorageClassConfiguration(filters *StorageClassConfigurationFilters) *StorageClassConfiguration {
	var result StorageClassConfiguration
	if filters != nil {
		result.StorageClassConfigurationFilters = *filters
	}

	return &result
}

func (scc *StorageClassConfiguration) BeforeCreate(_ *gorm.DB) error {
	scc.ID = uuid.New()
	return nil
}

func (scc *StorageClassConfiguration) ErrorMessage(title string) string {
	return fmt.Sprintf("failed to %s storage class configuration %s", title, scc.details())
}

func (scc *StorageClassConfiguration) details() string {
	return fmt.Sprintf("tenant[%d] %s[%s] storage class configuration [%s]",
		scc.TenantId, ptr.Deref(scc.ScopeType, ""), ptr.Deref(scc.ScopeId, ""), ptr.Deref(scc.Name, ""))
}

type StorageClassWithOptionalConfig struct {
	Object
	*StorageClassConfiguration
}
