package dbmodel

import (
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
)

type RegistryRecord struct {
	AssetId        string
	Url            string
	CredentialKind api.CredentialKind
	CredentialId   string
}

func (RegistryRecord) TableName() string {
	return dbtables.TableNameAssetsRegistry
}

type RegistryUpdate struct {
	Url            string
	CredentialKind api.CredentialKind
	CredentialId   string
}

func (RegistryUpdate) TableName() string {
	return dbtables.TableNameAssetsRegistry
}

// RegistryFullRecord is used for result of joining assets_registry_password and assets_metadata
type RegistryFullRecord struct {
	api.AssetMeta              `gorm:"embedded"`
	api.WorkloadSupportedTypes `gorm:"embedded"`
	Url                        string
	CredentialKind             api.CredentialKind
	CredentialId               string
}

func (p RegistryFullRecord) GetApiAssetMeta() api.AssetMeta {
	return p.AssetMeta
}

func (p RegistryFullRecord) GetApiWorkloadSupportedTypes() *api.WorkloadSupportedTypes {
	return &p.WorkloadSupportedTypes
}
