/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
)

func createClaimNameLock(ctx context.Context, claimName string) (*dbmodel.ClaimNameLockRecord, error) {
	db := runaidb.DB().WithContext(ctx)

	claimNameLockRec := dbmodel.ClaimNameLockRecord{
		Id:        uuid.New().String(),
		TenantId:  auth.TenantIdOf(ctx),
		ClaimName: claimName,
		CreatedBy: auth.SubjectIdOf(ctx),
	}

	if _, err := db.Model(claimNameLockRec).Create(&claimNameLockRec); err != nil {
		return nil, err
	}

	return &claimNameLockRec, nil
}

func deleteClaimNameLock(ctx context.Context, id string) error {
	db := runaidb.DB().WithContext(ctx)

	tenantId := auth.TenantIdOf(ctx)
	claimNameLockRec := dbmodel.ClaimNameLockRecord{
		Id:       id,
		TenantId: tenantId,
	}

	_, err := db.Where(&claimNameLockRec).Delete(&dbmodel.ClaimNameLockRecord{})
	return err
}
