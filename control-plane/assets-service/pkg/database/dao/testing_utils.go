/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"

	"github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
)

// TestingAddHoursToUpdatedAt utility functions for changing the updated at column for testing purposes.
// we do it using exec and not save to prevent gorm to change it to current time.
func TestingAddHoursToUpdatedAt(ctx context.Context, id string, hours int) {
	query := fmt.Sprintf("UPDATE %s SET updated_at = updated_at + interval '%d hour' where id = ?", dbtables.TableNameAssetsMeta, hours)
	_, err := runaidb.DB().WithContext(ctx).Exec(query, id)
	gomega.Expect(err).To(gomega.BeNil())
}

func ForceTenantCloneForTesting(tenantId int32) {
	presetsCloneAttemptedForTenant.RLock()
	delete(presetsCloneAttemptedForTenant.tenants, tenantId)
	presetsCloneAttemptedForTenant.RUnlock()
}
