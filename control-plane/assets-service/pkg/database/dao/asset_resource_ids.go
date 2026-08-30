package dao

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
	"gorm.io/gorm/clause"
)

func UpsertAssetResourceIds(ctx context.Context, assetResourcesIds []dbmodel.AssetResourceIdsRecord) *api.Error {
	db := runaidb.DB().WithContext(ctx)

	model := dbmodel.AssetResourceIdsRecord{}

	// If the asset resource id already exists, we update the lastSyncedAt time
	_, resErr := db.Model(model).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: dbtables.AssetIdColumnName},
			{Name: dbtables.ResourceIdColumnName},
			{Name: dbtables.AssetMetaColumnTenantId},
		},
		DoUpdates: clause.AssignmentColumns([]string{dbtables.LastSyncedAtColumnName}),
	}).Create(&assetResourcesIds)
	if resErr != nil {
		return database.Error(database.UPSERT, dbtables.TableNameAssetsResourcesIds).WithDetails(resErr)
	}

	return nil
}

func DeleteAssetResourceIds(ctx context.Context, tx orm.DB, assetId string) *api.Error {
	return DeleteAssetResourceIdsByIds(ctx, tx, []string{assetId})
}

func DeleteAssetResourceIdsByIds(ctx context.Context, tx orm.DB, assetIds []string) *api.Error {
	if len(assetIds) == 0 {
		return nil
	}

	tenantId := auth.TenantIdOf(ctx)
	rowsAffected, err := tx.Model(&dbmodel.AssetResourceIdsRecord{}).
		Where(fmt.Sprintf("%s = ?", dbtables.AssetMetaColumnTenantId), tenantId).
		Where(fmt.Sprintf("%s IN ?", dbtables.AssetIdColumnName), assetIds).
		Delete(&dbmodel.AssetResourceIdsRecord{})
	if err != nil {
		return database.Error(database.DELETE, dbtables.TableNameAssetsResourcesIds).WithDetailsFromDbResult(rowsAffected, err)
	}

	return nil
}
