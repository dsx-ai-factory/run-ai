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
	"time"

	"github.com/lib/pq"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbtables"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"gorm.io/gorm"
)

const (
	tenantMappingLockStaleDuration = 5 * time.Second
	postgresDuplicateKey           = "23505"
	mappingLockBusyMessage         = "mapping is being updated, please retry your update later"
)

// AcquireTenantMappingLock enforces at most one tenant lock. return
// a function to release the lock
func AcquireTenantMappingLock(ctx context.Context) (release func(), apiErr *api.Error) {
	tenantID := auth.TenantIdOf(ctx)

	now := time.Now().UTC()
	staleCutoff := now.Add(-tenantMappingLockStaleDuration)

	db := dao_utils.DB().WithContext(ctx)

	// delete obsolete entries from a possibly panicked service (5 seconds grace time)
	delRes := db.Model(&dbmodel.MappingLocking{}).
		Where(fmt.Sprintf("%s = ? AND %s <= ?", dbtables.MappingLockingColumnTenantId, dbtables.MappingLockingColumnLockedAt), tenantID, staleCutoff).
		Delete(&dbmodel.MappingLocking{})
	if delRes.Error != nil {
		dbErr := database.Error(database.DELETE, database.MappingLocking).WithDetailsFromDb(delRes)
		logging.Errorf("failed to delete stale tenant mapping locks: %s: %v", dbErr.Message, delRes.Error)
		return nil, dbErr
	}

	// Attempt to add a record for the given tenant. if another action just added one, we will be rejected
	// for duplicate record.
	row := dbmodel.MappingLocking{
		TenantId: tenantID,
		LockedAt: now,
	}

	res := db.Model(&row).Create(&row)
	if res.Error != nil {
		if isUniqueViolation(res.Error) {
			return nil, api.NewError(http.StatusConflict, mappingLockBusyMessage, nil)
		}
		dbErr := database.Error(database.CREATE, database.MappingLocking).WithDetailsFromDb(res)
		logging.Errorf("failed to acquire tenant mapping lock for tenant %d: %s: %v", tenantID, dbErr.Message, res.Error)
		return nil, dbErr
	}

	return func() {
		res := dao_utils.DB().WithContext(ctx).
			Model(&dbmodel.MappingLocking{}).
			Where(fmt.Sprintf("%s = ?", dbtables.MappingLockingColumnTenantId), tenantID).
			Delete(&dbmodel.MappingLocking{})
		if res.Error != nil {
			logging.Errorf("failed to release tenant mapping lock for tenant %d: %v", tenantID, res.Error)
		}
	}, nil
}

func isUniqueViolation(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}

	if pqErr, ok := errors.AsType[*pq.Error](err); ok && pqErr.Code == postgresDuplicateKey {
		return true
	}

	// pgx / wrapped drivers may not surface *pq.Error; match the stable Postgres text.
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}
