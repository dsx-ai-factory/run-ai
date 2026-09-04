/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
)

func UpsertAssetLastUsedByWorkload(ctx context.Context, assetId string, workloadType string, workloadId string, projectId int32, workloadCreationTime time.Time) *api.Error {
	tx := runaidb.DB().WithContext(ctx)

	record := dbmodel.LastUsedByWorkloadRecord{
		TenantId:             auth.TenantIdOf(ctx),
		ProjectId:            projectId,
		AssetId:              assetId,
		WorkloadType:         workloadType,
		WorkloadId:           workloadId,
		WorkloadCreationTime: workloadCreationTime,
	}

	searchParams := dbmodel.LastUsedByWorkloadRecord{
		TenantId:     auth.TenantIdOf(ctx),
		ProjectId:    projectId,
		AssetId:      assetId,
		WorkloadType: workloadType,
	}

	_, err := tx.Model(record).Assign(record).Where(searchParams).FirstOrCreate(&record)
	if err != nil {
		return database.Error(database.UPSERT, dbtables.TableNameLastUsedByWorkload).WithDetails(err)
	}

	return nil
}

func ListUsageTimes(ctx context.Context, workloadType string, projectId int32, assetIdsToCheck []string) ([]dbmodel.LastUsedByWorkloadRecord, *api.Error) {
	searchParams := dbmodel.LastUsedByWorkloadRecord{
		TenantId:     auth.TenantIdOf(ctx),
		WorkloadType: workloadType,
		ProjectId:    projectId,
	}

	usageRecords := make([]dbmodel.LastUsedByWorkloadRecord, 0)

	_, err := runaidb.DB().WithContext(ctx).
		Model(&searchParams).
		Where(&searchParams).
		Where(fmt.Sprintf("%s in ?", dbtables.LastUsedByWorkloadColumnAssetId), assetIdsToCheck).
		Scan(&usageRecords)
	if err != nil {
		return nil, database.Error(database.LIST, dbtables.LastUsedByWorkloadColumnAssetId).WithDetails(err)
	}

	return usageRecords, nil
}
