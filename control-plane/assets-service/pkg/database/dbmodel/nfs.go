package dbmodel

import (
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
)

type NFSAssetRec struct {
	AssetId          string
	api.NfsAssetSpec `gorm:"embedded"`
}

func (NFSAssetRec) TableName() string {
	return dbtables.TableNameAssetsNFS
}

func (n NFSAssetRec) GetAssetId() string {
	return n.AssetId
}

func ReplaceNFSSpecIds(specs []NFSAssetRec, translationMap map[string]string) {
	for idx := range specs {
		specs[idx].AssetId = translationMap[specs[idx].AssetId]
	}
}

type NFSAssetUpdate struct {
	api.NfsAssetSpec `gorm:"embedded"`
}

func (NFSAssetUpdate) TableName() string {
	return dbtables.TableNameAssetsNFS
}

// NFSCombinedRec holds result of join assets_metadata and assets_nfs queries
type NFSCombinedRec struct {
	api.AssetMeta              `gorm:"embedded"`
	api.NfsAssetSpec           `gorm:"embedded"`
	api.WorkloadSupportedTypes `gorm:"embedded"`
}

func (p NFSCombinedRec) GetApiAssetMeta() api.AssetMeta {
	return p.AssetMeta
}

func (p NFSCombinedRec) GetApiWorkloadSupportedTypes() *api.WorkloadSupportedTypes {
	return &p.WorkloadSupportedTypes
}
