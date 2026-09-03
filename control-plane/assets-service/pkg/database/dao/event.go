/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

func UpsertEvents(ctx context.Context, events []*dbmodel.Event) *api.Error {
	db := runaidb.DB().WithContext(ctx)

	_, err := db.Table(dbtables.TableNameEvents).Clauses(clause.OnConflict{DoNothing: true}).Create(events)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create the events: %v", err)
		return database.Error(database.UPSERT, dbtables.TableNameEvents).WithDetails(err)
	}

	return nil
}

func DeleteAssetEvents(ctx context.Context, tx orm.DB, assetId string) *api.Error {
	return DeleteAssetEventsByIds(ctx, tx, []string{assetId})
}

func DeleteAssetEventsByIds(ctx context.Context, tx orm.DB, assetIds []string) *api.Error {
	if len(assetIds) == 0 {
		return nil
	}

	tenantId := auth.TenantIdOf(ctx)

	// Convert string IDs to UUIDs
	assetUUIDs := make([]uuid.UUID, len(assetIds))
	for i, assetId := range assetIds {
		assetUUIDs[i] = uuid.MustParse(assetId)
	}

	_, err := tx.Model(&dbmodel.Event{}).
		Where(fmt.Sprintf("%s = ?", dbtables.AssetMetaColumnTenantId), uint(tenantId)).
		Where(fmt.Sprintf("%s IN ?", dbtables.AssetIdColumnName), assetUUIDs).
		Delete(&dbmodel.Event{})
	if err != nil {
		return database.Error(database.DELETE, dbtables.TableNameEvents).WithDetails(err)
	}

	return nil
}

func GetEventsByAssetId(ctx context.Context, assetId string, offset int32, limit int32) ([]dbmodel.Event, *api.Error) {
	db := runaidb.DB().WithContext(ctx)

	events := make([]dbmodel.Event, 0)

	filters := dbmodel.Event{
		AssetId: uuid.MustParse(assetId),
	}

	err := db.Table(dbtables.TableNameEvents).
		Where(filters).
		Offset(int(offset)).
		Limit(int(limit)).
		Find(&events)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get events by asset id: %v", err)
		return events, database.Error(database.GET, dbtables.TableNameEvents).WithDetails(err)
	}

	return events, nil
}
