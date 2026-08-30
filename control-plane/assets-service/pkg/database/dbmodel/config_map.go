package dbmodel

import (
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
)

// ConfigMapAssetRec represents a ConfigMap asset record in the table
type ConfigMapAssetRec struct {
	AssetId                string
	api.ConfigMapAssetSpec `gorm:"embedded"`
}

// TableName implements gorm tabler interface
func (ConfigMapAssetRec) TableName() string {
	return dbtables.TableNameAssetsConfigMap
}

// ConfigMapAssetUpdate used as a db model for updating ConfigMap record
type ConfigMapAssetUpdate struct {
	api.ConfigMapAssetSpec `gorm:"embedded"`
}

// TableName implements gorm tabler interface
func (ConfigMapAssetUpdate) TableName() string {
	return dbtables.TableNameAssetsConfigMap
}

// ConfigMapCombinedRec holds result of join assets_metadata and assets_configmap queries
type ConfigMapCombinedRec struct {
	api.AssetMeta              `gorm:"embedded"`
	api.ConfigMapAssetSpec     `gorm:"embedded"`
	api.WorkloadSupportedTypes `gorm:"embedded"`
}

// GetApiAssetMeta implements the AssetRecord interface
func (cm *ConfigMapCombinedRec) GetApiAssetMeta() api.AssetMeta {
	return cm.AssetMeta
}

// GetApiWorkloadSupportedTypes implements the AssetRecord interface
func (cm *ConfigMapCombinedRec) GetApiWorkloadSupportedTypes() *api.WorkloadSupportedTypes {
	return &cm.WorkloadSupportedTypes
}
