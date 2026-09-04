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
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
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

	db := runaidb.DB().WithContext(ctx)

	// delete obsolete entries from a possibly panicked service (5 seconds grace time)
	delRows, delErr := db.Model(&dbmodel.MappingLocking{}).
		Where(fmt.Sprintf("%s = ? AND %s <= ?", dbtables.MappingLockingColumnTenantId, dbtables.MappingLockingColumnLockedAt), tenantID, staleCutoff).
		Delete(&dbmodel.MappingLocking{})
	if delErr != nil {
		dbErr := database.Error(database.DELETE, database.MappingLocking).WithDetailsFromDbResult(delRows, delErr)
		logging.Errorf("failed to delete stale tenant mapping locks: %s: %v", dbErr.Message, delErr)
		return nil, dbErr
	}

	// Attempt to add a record for the given tenant. if another action just added one, we will be rejected
	// for duplicate record.
	row := dbmodel.MappingLocking{
		TenantId: tenantID,
		LockedAt: now,
	}

	if _, createErr := db.Model(&row).Create(&row); createErr != nil {
		if isUniqueViolation(createErr) {
			return nil, api.NewError(http.StatusConflict, mappingLockBusyMessage, nil)
		}
		dbErr := database.Error(database.CREATE, database.MappingLocking).WithDetails(createErr)
		logging.Errorf("failed to acquire tenant mapping lock for tenant %d: %s: %v", tenantID, dbErr.Message, createErr)
		return nil, dbErr
	}

	return func() {
		if _, err := runaidb.DB().WithContext(ctx).
			Model(&dbmodel.MappingLocking{}).
			Where(fmt.Sprintf("%s = ?", dbtables.MappingLockingColumnTenantId), tenantID).
			Delete(&dbmodel.MappingLocking{}); err != nil {
			logging.Errorf("failed to release tenant mapping lock for tenant %d: %v", tenantID, err)
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
