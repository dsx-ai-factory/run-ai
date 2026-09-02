/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

const (
	DeletionErrorLogPrefix = "error on tenant deletion."
	ServiceName            = "authorization"
)

func logError(ctx context.Context, tenantId string, err error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s tenantId: <%s>, service: <%s>, asset: <%s>, error: %s", DeletionErrorLogPrefix, tenantId, ServiceName, "access-rules", err)
}

func HardDeleteTenantAccessRules(ctx context.Context, tenantId string) error {
	err := runaidb.DB().WithContext(ctx).Unscoped().Transaction(func(tx orm.DB) error {
		_, err := tx.
			Where("tenant_id = ?", tenantId).
			Delete(&[]models.AccessRule{})
		return err
	})
	if err != nil {
		logError(ctx, tenantId, err)
	}
	return err
}
