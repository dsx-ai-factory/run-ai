/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"

	"github.com/google/uuid"
	"github.com/iancoleman/strcase"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/config"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	SequenceLastValue  = "last_value"
	SequenceAssetsMeta = "assets_metadata_update_count"
)

var (
	assetNotFoundPattern = "\".*\" not found"
	assetNotFoundRegexp  = regexp.MustCompile(assetNotFoundPattern)
)

type SyncedSpecs interface {
	api.PvcAssetSpec | api.AccessKeySpec | api.GenericSecretSpec | api.DockerRegistrySpec | api.PasswordSpec
}

type EmptySpec struct{}

// ListAssetsForSync traverses the metadata and relevant specs to find records that need to be synced to the cluster
// NOTE! The query does not filter existing pvc/credentials, that do not need to be synced. it leaves it to the service
// layer to filter. The implication is that the query will return 50 records, 30 of them of existing assets that the
// service will filter, so the bulk will only contain 20 records after all.
// We CAN improve the query to filter out existing assets, but it will require to add the following clauses:
//
//	Where("id NOT IN (?)",
//	 db.Table("assets_pvc").  <- same for all the 4 credential tables
//	     Select("asset_id").Where("is_existing = true"))
//
// which will make the query much heavier. so for the time being, I believe the right place to filter is in the service
// layer. we can also consider an increased bulk size to have more chance that the bulk will contain closer to 50 real updates
func ListAssetsForSync(ctx context.Context, tenantId int32, clusterId string,
	assetKinds []api.AssetKind, updateCount int64, bulkSize int, includePrivateCredentials bool) ([]dbmodel.AssetsToSync, *api.Error,
) {
	result := runaidb.DB().WithContext(ctx).Model(dbmodel.AssetMetaRec{})

	selectedColumns := fmt.Sprintf("%s.*", dbtables.TableNameAssetsMeta)
	result = result.Select(selectedColumns)

	result = result.Order(dbtables.AssetMetaColumnUpdateCount)

	// restrict the search to the assets relevant to this sync (datasource or credentials)
	result = result.Where(fmt.Sprintf("%s in ?", dbtables.AssetMetaColumnKind), assetKinds)

	// we want to filter by tenant id so Coke assets will not be synced to Pepsi cluster
	result = result.Where(fmt.Sprintf("%s = ?", dbtables.AssetMetaColumnTenantId), tenantId)

	// we want either (tenant scoped) OR (cluster scope of the same cluster)
	result = result.Where(fmt.Sprintf("%s is null or %s = ?", dbtables.AssetMetaColumnClusterId, dbtables.AssetMetaColumnClusterId), clusterId)

	// we want to filter out private assets that are not relevant to the cluster if the cluster doesn't support user credentials
	if !includePrivateCredentials {
		result = result.Where(fmt.Sprintf("COALESCE(%v.%v, FALSE) = false", dbtables.TableNameAssetsMeta, dbtables.AssetMetaColumnPrivate))
	}

	// agent only receives updates that he did not yet process
	result = result.Where(fmt.Sprintf("%s > ?", dbtables.SyncStatusColumnUpdateCount), updateCount)

	// in first time sync (either cluster installed for the first time or was running asset-sync so far)
	// we only take the updates from the last few days
	if updateCount == 0 {
		daysLimitFirstSync := (-1) * runai_config.Get[*config.ServerConfig]().DaysLimitForFirstSync
		takeFrom := time.Now().AddDate(0, 0, daysLimitFirstSync)

		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("cluster %s -> first time sync: take events starting %s", clusterId, takeFrom.String())

		result = result.Where(fmt.Sprintf("(%s >= ? or %s >= ?)", dbtables.AssetMetaColumnUpdatedAt, dbtables.AssetMetaColumnDeletedAt), takeFrom, takeFrom)
	}

	// we want soft-deleted assets in order to synchronize deletion of assets to the cluster
	result = result.Unscoped()

	// join with the relevant specs
	for _, tableName := range dbtables.TablesOfSyncedAssets {
		result = result.Preload(strcase.ToCamel(tableName))
	}

	assets := make([]dbmodel.AssetsToSync, 0, bulkSize)
	if err := result.Find(&assets); err != nil {
		dbErr := database.Error(database.LIST, "assets for sync").WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("%s: %v", dbErr.Message, err)
		return nil, dbErr
	}

	return assets, nil
}

func GetSyncStatusByClusterAndAssetId(ctx context.Context, tx orm.DB, tenantId int32, clusterId string, assetId string) (*dbmodel.SyncStatusRec, *api.Error) {
	query := dbmodel.SyncStatusRec{
		TenantId:  tenantId,
		ClusterId: clusterId,
		AssetId:   assetId,
	}

	var syncStatusRecs []dbmodel.SyncStatusRec

	err := tx.Model(dbmodel.SyncStatusRec{}).Where(&query).Find(&syncStatusRecs)
	if err != nil || len(syncStatusRecs) == 0 {
		dbErr := database.Error(database.GET, dbtables.TableNameSyncStatus).WithDetailsFromDbResult(int64(len(syncStatusRecs)), err)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("%s: %v", dbErr.Message, err)
		}
		return nil, dbErr
	}

	return &syncStatusRecs[0], nil
}

func UpdateSyncStatusSendToAgent(ctx context.Context, tenantId int32, clusterId string, updates []dbmodel.AssetsToSync) *api.Error {
	//
	//  we prefer to update sync status for all the updates in the bulk, not one by one, so it will have more
	//  chance of being accurate. it is still possible that agent will not get the update, and we updated for
	//  nothing, but the sync mechanism is not affected by it (same updated will be sent in the next request
	//  from the agent with the same update count)
	//
	tx := runaidb.DB().WithContext(ctx).Begin()

	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	for _, update := range updates {
		syncStatus, err := GetSyncStatusByClusterAndAssetId(ctx, tx, tenantId, clusterId, update.Id)
		if err != nil {
			if err.Code != http.StatusNotFound {
				return err
			}

			syncStatus = &dbmodel.SyncStatusRec{
				Id: uuid.NewString(),

				TenantId:  tenantId,
				ClusterId: clusterId,

				AssetId: update.Id,
			}
		}

		syncStatus.Spec, err = update.SpecAsJsonBmap()
		if err != nil {
			return err
		}

		//
		//   set the update count depending on the type of update: apply or delete
		//
		if update.DeletedAt.Valid {
			syncStatus.SentToAgentDelete = update.UpdateCount
		} else {
			syncStatus.SentToAgentApply = update.UpdateCount
		}

		_, saveErr := tx.Save(syncStatus)
		if saveErr != nil {
			dbErr := database.Error(database.UPSERT, dbtables.TableNameSyncStatus).WithDetails(saveErr)
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, saveErr)
			return dbErr
		}
	}

	//
	//   now that all the bulk is updated in sync status, we can commit the update
	//
	if commitErr := tx.Commit(); commitErr != nil {
		dbErr := database.Error(database.COMMIT_UPDATES, dbtables.TableNameSyncStatus).WithDetails(commitErr)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr.Message, commitErr)
		return dbErr
	}
	committed = true

	return nil
}

func HandleAgentSyncUpdates(ctx context.Context, tenantId int32, clusterId string, updates []api.ClusterSyncReportUpdate) {
	db := runaidb.DB()

	// maintain a list of assets that can be deleted as their deletion was synced to all the clusters
	toDelete := make([]string, 0, len(updates))

	for _, update := range updates {
		errorPrefix := fmt.Sprintf("sync update %v from cluster %s[%d]", update.UpdateCount, clusterId, tenantId)

		syncStatusRec, err := GetSyncStatusByClusterAndAssetId(ctx, db, tenantId, clusterId, update.Id)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("%s: no sync info, update ignored, probably lagged update", errorPrefix)
			continue
		}

		//
		//   update the processedByAgent counter, unless this is an outdated update
		//
		if update.UpdateCount < syncStatusRec.ProcessedByAgent {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("%s: update %v already processed by agent, so this update is outdated, and can be ignored", errorPrefix, syncStatusRec.ProcessedByAgent)
			continue
		}

		syncStatusRec.ProcessedByAgent = update.UpdateCount

		//
		//   check if the update is for apply or delete
		//
		var isDelete bool

		switch update.UpdateCount {
		case syncStatusRec.SentToAgentApply:
			isDelete = false
		case syncStatusRec.SentToAgentDelete:
			isDelete = true
		default:
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("%s: recent updates sent to agent %v/%v, update no longer relevant, ignored",
				errorPrefix, syncStatusRec.SentToAgentApply, syncStatusRec.SentToAgentDelete)
			continue
		}

		if isDelete {
			//
			//  handle successful delete in one of two cases:
			//  (1) delete succeeded in the cluster
			//  (2) asset was not found in the cluster
			//
			if update.Message == nil || assetNotFoundRegexp.MatchString(*update.Message) {
				queryCurrentCluster := dbmodel.SyncStatusRec{
					TenantId:  tenantId,
					ClusterId: clusterId,
					AssetId:   update.Id,
				}

				_, err := db.Model(&syncStatusRec).Where(&queryCurrentCluster).Delete(&dbmodel.SyncStatusRec{})
				if err != nil {
					logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: failed to delete sync status record: %v", errorPrefix, err)
					continue
				}

				//
				//   check if there are any remaining clusters still updating this asset (relevant only for tenant assets
				//
				queryRemainingClusters := dbmodel.SyncStatusRec{
					TenantId: tenantId,
					AssetId:  update.Id,
				}

				var numRemainingClusters int64

				err = db.Model(&syncStatusRec).Where(&queryRemainingClusters).Select("count(*) as count").Count(&numRemainingClusters)
				if err != nil {
					logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("%s: failed to count remaining clusters: %v", errorPrefix, err)
					continue
				}

				if numRemainingClusters == 0 {
					toDelete = append(toDelete, update.Id)
				} else {
					logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("%s: remaining %d clusters, cannot clean asset yet", errorPrefix, numRemainingClusters)
				}

				continue
			}
		}

		syncStatusRec.Message = convutils.Deref(update.Message, "")

		//
		//  if deletion failed, or in case of apply, just update the table
		//
		if _, saveErr := db.Save(syncStatusRec); saveErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: failed to update sync status: %v", errorPrefix, saveErr)
			continue
		}
	}

	if len(toDelete) == 0 {
		return
	}

	//
	//  finally we need to permanently delete assets that were deleted from all clusters
	//
	cleanupPrefix := fmt.Sprintf("post-sync cleanup of %d assets of tenant %d", len(toDelete), tenantId)

	tx := runaidb.DB().WithContext(ctx).Begin()

	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	for _, table := range append(dbtables.TablesOfSyncedAssets, dbtables.TableNameAssetsMeta) {
		idColumn := "asset_id"
		if table == dbtables.TableNameAssetsMeta {
			_, _, delErr := deleteAssetMetaByIds(ctx, tx, toDelete, HardDelete)
			if delErr != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: failed to delete asset meta: %v", cleanupPrefix, delErr)
				return
			}
		} else {
			if _, delErr := db.Table(table).Where(fmt.Sprintf("%s in ?", idColumn), toDelete).Unscoped().Delete(nil); delErr != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: failed to post-delete asset from table %s: %v", cleanupPrefix, table, delErr)
				return
			}
		}
	}

	if commitErr := tx.Commit(); commitErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: commit post-sync deletion %v", cleanupPrefix, commitErr)
	}

	committed = true
}

func GetHighestUpdateCount(ctx context.Context) (int64, *api.Error) {
	var currentValue int64
	rowsAffected, err := runaidb.DB().WithContext(ctx).Table(SequenceAssetsMeta).Select(SequenceLastValue).Scan(&currentValue)
	if err != nil {
		return 0, api.NewError(http.StatusServiceUnavailable, "failed to read highest update count", err).WithDetailsFromDbResult(rowsAffected, err)
	}
	return currentValue, nil
}

func ListSyncStatusByAssetIds(ctx context.Context, tenantId int32, assetIds []string) (map[string]*dbmodel.SyncStatusRec, *api.Error) {
	result := runaidb.DB().WithContext(ctx).Model(dbmodel.SyncStatusRec{})

	result = result.Where(fmt.Sprintf("%s = ?", dbtables.SyncStatusColumnTenantId), tenantId)

	result = result.Where(fmt.Sprintf("%s in ?", dbtables.SyncStatusColumnAssetId), assetIds)

	records := make([]dbmodel.SyncStatusRec, 0)
	err := result.Find(&records)

	if err != nil {
		dbErr := database.Error(database.LIST, dbtables.TableNameSyncStatus).WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("%s: %v", dbErr.Message, err)
		return nil, dbErr
	}

	answer := make(map[string]*dbmodel.SyncStatusRec)

	for _, record := range records {
		answer[record.AssetId] = &record
	}

	return answer, nil
}
