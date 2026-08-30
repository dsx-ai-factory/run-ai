package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/query_types"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

var AutoDeleteGraceTimeSec = (-120) // allow two minutes before deleting an auto delete asset

func AutoDeletePvcAssets(ctx context.Context, deletionType DeletionType) *api.Error {
	db := runaidb.DB().WithContext(ctx)

	//
	//   start asset metadata search query with auto deletion = true
	//
	result, err := GetAssetMeta(ctx, db, dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			TenantId:   convutils.Ref(auth.TenantIdOf(ctx)), // ignoring the case of system scoped asset, which is irrelevant here
			Kind:       api.ASSETKIND_PVC,                   // workload volumes are always PVCs, no need to browse other kinds
			AutoDelete: convutils.Ref(true),
		},
	}, &query_types.QueryParams{IncludeAutoDeleteAssets: true})
	if err != nil {
		return err
	}

	//
	//   do not delete assets that just have been created, it is possible that they will be in use soon
	//
	graceTime := time.Now().Add(time.Duration(AutoDeleteGraceTimeSec) * time.Second)

	//
	//   and add to it where conditions to identify assets no longer used
	//
	records := make([]dbmodel.AssetMetaRec, 0)

	_, scanErr := result.Where(fmt.Sprintf("%s NOT in (select contained_asset_id from %s)", dbtables.AssetMetaColumnId, dbtables.TableNameAssetsTemplateContent)).
		Where(fmt.Sprintf("%s NOT in (select asset_id from %s)", dbtables.AssetMetaColumnId, dbtables.TableNameWorkloadsAssets)).
		Where(fmt.Sprintf("%s.%s < ?", dbtables.TableNameAssetsMeta, dbtables.AssetMetaColumnCreationTime), graceTime).
		Scan(&records)
	if scanErr != nil {
		return database.Error(database.LIST, api.ASSETKIND_PVC).WithDetails(scanErr)
	}

	if len(records) == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("no auto delete assets found for tenant %d", auth.TenantIdOf(ctx))
		return nil
	}

	//
	//   now force-delete those assets
	//
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	idsToDelete := make([]string, 0, len(records))
	for _, record := range records {
		idsToDelete = append(idsToDelete, record.Id)
	}

	if deletionType == SoftDelete {
		delRows, delErr := tx.Model(&dbmodel.AssetMetaRec{}).Where("id in ?", idsToDelete).Delete(&dbmodel.AssetMetaRec{})
		if delErr != nil {
			return database.Error(database.DELETE, api.ASSETKIND_PVC).WithDetailsFromDbResult(delRows, delErr)
		}
	} else {
		delRows, delErr := tx.Where("asset_id in ?", idsToDelete).Delete(&dbmodel.PVCAssetRec{})
		if delErr != nil {
			return database.Error(database.DELETE, api.ASSETKIND_PVC).WithDetailsFromDbResult(delRows, delErr)
		}
		delRows, delErr = tx.Where("asset_id in ?", idsToDelete).Delete(&dbmodel.PolicyComplianceRecord{})
		if delErr != nil {
			return database.Error(database.DELETE, api.ASSETKIND_PVC).WithDetailsFromDbResult(delRows, delErr)
		}
		delRows, delErr = tx.Where("id in ?", idsToDelete).Unscoped().Delete(&dbmodel.AssetMetaRec{})
		if delErr != nil {
			return database.Error(database.DELETE, api.ASSETKIND_PVC).WithDetailsFromDbResult(delRows, delErr)
		}
	}

	if err := tx.Commit(); err != nil {
		return database.Error(database.DELETE, api.ASSETKIND_PVC).WithDetails(err)
	}
	committed = true

	return nil
}
