/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

func ListApprovedDeletionsByTenant(ctx context.Context) (map[string]*api.SyncConfig, *api.Error) {
	tenantId := auth.TenantIdOf(ctx)

	key := dbmodel.ApprovedDeletionsKey{
		TenantId: tenantId,
	}

	records := make([]dbmodel.ApprovedDeletions, 0)

	if _, err := runaidb.DB().WithContext(ctx).Model(&dbmodel.ApprovedDeletions{}).Where(&key).Scan(&records); err != nil {
		dbErr := database.Error(database.LIST, database.ApprovedDeletions).WithDetails(err)
		logging.Errorf("failed to list approved deletions: %s: %v", dbErr.Message, err)
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

	if err := runaidb.DB().WithContext(ctx).Model(&record).Where(&key).Take(&record); err != nil {
		if errors.Is(err, orm.ErrRecordNotFound) {
			return nil, nil
		}
		dbErr := database.Error(database.GET, database.ApprovedDeletions).WithDetails(err)
		logging.Errorf("failed to get approved deletions info: %s: %v", dbErr.Message, err)
		return nil, dbErr
	}

	return &record, nil
}

func UpdateApprovedDeletionOfCluster(ctx context.Context, syncConfig api.SyncConfig) (*dbmodel.ApprovedDeletions, *api.Error) {
	tenantId := auth.TenantIdOf(ctx)

	db := runaidb.DB().WithContext(ctx)

	key := dbmodel.ApprovedDeletionsKey{
		TenantId:  tenantId,
		ClusterId: &syncConfig.ClusterId,
	}

	values := dbmodel.ApprovedDeletions{
		AvoidDeletionsOfObsoletePolicies: syncConfig.AvoidDeletionsOfObsoletePolicies,
	}

	var record dbmodel.ApprovedDeletions

	if _, err := db.Model(&record).Where(&key).Assign(values).FirstOrCreate(&record); err != nil {
		dbErr := database.Error(database.UPSERT, database.ApprovedDeletions).WithDetails(err)
		logging.Errorf("failed to upsert approved deletions: %s: %v", dbErr.Message, err)
		return nil, dbErr
	}

	return &record, nil
}

func UpdateApprovedDeletionByIds(ctx context.Context, ids []string, avoidDeletionsOfObsoletePolicies bool) {
	db := runaidb.DB().WithContext(ctx)

	for _, id := range ids {
		record := dbmodel.ApprovedDeletions{
			ApprovedDeletionsKey: dbmodel.ApprovedDeletionsKey{
				TenantId:  auth.TenantIdOf(ctx),
				ClusterId: &id,
			},
			AvoidDeletionsOfObsoletePolicies: avoidDeletionsOfObsoletePolicies,
		}

		if _, createErr := db.Model(&record).Create(&record); createErr != nil {
			// in this case we only want to log the error and not fail the API, because the implication of this
			// error is that the user will have to keep approving but at least his original request will not fail
			err := database.Error(database.CREATE, database.ApprovedDeletions).WithDetails(createErr)
			if err.Code != http.StatusConflict {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("[%d] failed to update approved deletions for cluster %v: %s: %v",
					record.TenantId, record.ClusterId, err.Message, createErr)
			}
		}
	}
}

func DeleteApprovedDeletionByIds(ctx context.Context, ids []string) {
	db := runaidb.DB().WithContext(ctx)

	tenantKey := dbmodel.ApprovedDeletionsKey{
		TenantId: auth.TenantIdOf(ctx),
	}

	rowsAffected, deleteErr := db.Model(&dbmodel.ApprovedDeletions{}).Where(&tenantKey).Where(fmt.Sprintf("%s in ?", dbtables.ApprovedDeletionsColumnClusterId), ids).Delete(&dbmodel.ApprovedDeletions{})
	if deleteErr != nil {
		// in this case we only want to log the error and not fail the API, because the implication of this
		// error is that the cluster will continue to get deletions instead of applies.
		err := database.Error(database.CREATE, database.ApprovedDeletions).WithDetailsFromDbResult(rowsAffected, deleteErr)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("[%d] failed to update approved deletions for clusters %s: %s: %v",
			tenantKey.TenantId, strings.Join(ids, ","), err.Message, deleteErr)
	}
}
