package dbmodel

import (
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
)

type PasswordRecord struct {
	AssetId          string
	PrivateUseOnly   bool
	api.PasswordSpec `gorm:"embedded"`
}

func (PasswordRecord) TableName() string {
	return dbtables.TableNameAssetsPassword
}

type PasswordUpdate struct {
	api.PasswordSpec `gorm:"embedded"`
}

func (PasswordUpdate) TableName() string {
	return dbtables.TableNameAssetsPassword
}

// PasswordFullRecord is used for result of joining assets_password and assets_metadata
type PasswordFullRecord struct {
	api.AssetMeta              `gorm:"embedded"`
	PrivateUseOnly             bool
	PrevName                   *string
	api.PasswordSpec           `gorm:"embedded"`
	api.WorkloadSupportedTypes `gorm:"embedded"`
}

func (p PasswordFullRecord) GetApiAssetMeta() api.AssetMeta {
	return p.AssetMeta
}

func (p PasswordFullRecord) GetApiWorkloadSupportedTypes() *api.WorkloadSupportedTypes {
	return &p.WorkloadSupportedTypes
}
