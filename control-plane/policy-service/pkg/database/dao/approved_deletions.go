/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbtables"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

func ListApprovedDeletionsByTenant(ctx context.Context) (map[string]*api.SyncConfig, *api.Error) {
	tenantId := auth.TenantIdOf(ctx)

	key := dbmodel.ApprovedDeletionsKey{
		TenantId: tenantId,
	}

	records := make([]dbmodel.ApprovedDeletions, 0)

	result := dao_utils.DB().WithContext(ctx).Model(&dbmodel.ApprovedDeletions{}).Where(&key).Scan(&records)
	if result.Error != nil {
		dbErr := database.Error(database.LIST, database.ApprovedDeletions).WithDetailsFromDb(result)
		logging.Errorf("failed to list approved deletions: %s: %v", dbErr.Message, result.Error)
		return nil, dbErr
	}

	clusters := make(map[string]*api.SyncConfig)

	for _, record := range records {
		if record.ClusterId != nil {
			clusterId := *record.ClusterId

			clusters[clusterId] = &api.SyncConfig{
				ClusterId:                        clusterId,
				AvoidDeletionsOfObsoletePolicies: record.AvoidDeletionsOfObsoletePolicies,
			}
		}
	}

	return clusters, nil
}

func GetApprovedDeletionsInfoOfCluster(ctx context.Context, clusterId string) (*dbmodel.ApprovedDeletions, *api.Error) {
	tenantId := auth.TenantIdOf(ctx)

	key := dbmodel.ApprovedDeletionsKey{
		TenantId:  tenantId,
		ClusterId: &clusterId,
	}

	var record dbmodel.ApprovedDeletions

	result := dao_utils.DB().WithContext(ctx).Model(&record).Where(&key).Find(&record)
	if result.Error != nil {
		dbErr := database.Error(database.GET, database.ApprovedDeletions).WithDetailsFromDb(result)
		logging.Errorf("failed to get approved deletions info: %s: %v", dbErr.Message, result.Error)
		return nil, dbErr
	}

	if result.RowsAffected == 0 {
		return nil, nil // no info found
	}

	return &record, nil
}

func UpdateApprovedDeletionOfCluster(ctx context.Context, syncConfig api.SyncConfig) (*dbmodel.ApprovedDeletions, *api.Error) {
	tenantId := auth.TenantIdOf(ctx)

	db := dao_utils.DB().WithContext(ctx)

	key := dbmodel.ApprovedDeletionsKey{
		TenantId:  tenantId,
		ClusterId: &syncConfig.ClusterId,
	}

	values := dbmodel.ApprovedDeletions{
		AvoidDeletionsOfObsoletePolicies: syncConfig.AvoidDeletionsOfObsoletePolicies,
	}

	var record dbmodel.ApprovedDeletions

	result := db.Model(&record).Where(&key).Assign(values).FirstOrCreate(&record)

	if result.Error != nil {
		dbErr := database.Error(database.UPSERT, database.ApprovedDeletions).WithDetailsFromDb(result)
		logging.Errorf("failed to upsert approved deletions: %s: %v", dbErr.Message, result.Error)
		return nil, dbErr
	}

	return &record, nil
}

func UpdateApprovedDeletionByIds(ctx context.Context, ids []string, avoidDeletionsOfObsoletePolicies bool) {
	db := dao_utils.DB().WithContext(ctx)

	for _, id := range ids {
		record := dbmodel.ApprovedDeletions{
			ApprovedDeletionsKey: dbmodel.ApprovedDeletionsKey{
				TenantId:  auth.TenantIdOf(ctx),
				ClusterId: &id,
			},
			AvoidDeletionsOfObsoletePolicies: avoidDeletionsOfObsoletePolicies,
		}

		result := db.Model(&record).Create(&record)
		if result.Error != nil {
			// in this case we only want to log the error and not fail the API, because the implication of this
			// error is that the user will have to keep approving but at least his original request will not fail
			err := database.Error(database.CREATE, database.ApprovedDeletions).WithDetailsFromDb(result)
			if err.Code != http.StatusConflict {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("[%d] failed to update approved deletions for cluster %v: %s: %v",
					record.TenantId, record.ClusterId, err.Message, result.Error)
			}
		}
	}
}

func DeleteApprovedDeletionByIds(ctx context.Context, ids []string) {
	db := dao_utils.DB().WithContext(ctx)

	tenantKey := dbmodel.ApprovedDeletionsKey{
		TenantId: auth.TenantIdOf(ctx),
	}

	result := db.Model(&dbmodel.ApprovedDeletions{}).Where(&tenantKey).Where(fmt.Sprintf("%s in ?", dbtables.ApprovedDeletionsColumnClusterId), ids).Delete(dao_utils.All)
	if result.Error != nil {
		// in this case we only want to log the error and not fail the API, because the implication of this
		// error is that the cluster will continue to get deletions instead of applies.
		err := database.Error(database.CREATE, database.ApprovedDeletions).WithDetailsFromDb(result)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("[%d] failed to update approved deletions for clusters %s: %s: %v",
			tenantKey.TenantId, strings.Join(ids, ","), err.Message, result.Error)
	}
}
