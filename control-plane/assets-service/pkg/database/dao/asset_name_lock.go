/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/google/uuid"
)

func createAssetNameLock(ctx context.Context, name string, kind api.AssetKind) (*dbmodel.AssetNameLockRecord, error) {
	db := runaidb.DB().WithContext(ctx)

	assetNameLockRec := dbmodel.AssetNameLockRecord{
		Id:        uuid.New().String(),
		TenantId:  auth.TenantIdOf(ctx),
		Name:      name,
		Kind:      kind,
		CreatedBy: auth.SubjectIdOf(ctx),
	}

	if _, err := db.Model(assetNameLockRec).Create(&assetNameLockRec); err != nil {
		return nil, err
	}

	return &assetNameLockRec, nil
}

func deleteAssetNameLock(ctx context.Context, id string) error {
	db := runaidb.DB().WithContext(ctx)

	tenantId := auth.TenantIdOf(ctx)
	assetNameLockRec := dbmodel.AssetNameLockRecord{
		Id:       id,
		TenantId: tenantId,
	}

	_, err := db.Where(&assetNameLockRec).Delete(&dbmodel.AssetNameLockRecord{})
	return err
}
