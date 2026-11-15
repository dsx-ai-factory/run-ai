/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

// Retired templates remain in upgraded databases, but must not outlive their scope.
func deleteLegacyTemplatesByScope(ctx context.Context, db orm.DB, scopeType api.ScopeType, scopeId string) error {
	scopeColumn, err := dbtables.ScopeIdColumnName(scopeType)
	if err != nil {
		return err
	}

	return db.Transaction(func(tx orm.DB) error {
		for {
			query := tx.Model(&dbmodel.AssetMetaRec{}).Unscoped().
				Where(fmt.Sprintf("%s = ?", scopeColumn), scopeId).
				Where("kind = ?", "workload-template")
			// Tenant deletion identifies its target in the event, not the caller's context.
			if scopeType != api.SCOPETYPE_TENANT {
				query = query.Where("tenant_id = ?", auth.TenantIdOf(ctx))
			}

			var ids []string
			if err := query.Order("id ASC").Limit(deleteBatchSize).Pluck("id", &ids); err != nil {
				return fmt.Errorf("fetch legacy templates for scope deletion: %w", err)
			}
			if len(ids) == 0 {
				return nil
			}

			for _, table := range []struct{ name, idColumn string }{
				{"assets_template_content", "template_asset_id"},
				{"assets_template", "asset_id"},
			} {
				if _, err := tx.Table(table.name).Where(fmt.Sprintf("%s IN (?)", table.idColumn), ids).Delete(nil); err != nil {
					return fmt.Errorf("delete legacy template records from %s: %w", table.name, err)
				}
			}
			if err := deleteAssetsFromBookkeepingTable(tx, ids); err != nil {
				return err
			}
			if _, err := tx.Unscoped().Where("id IN (?)", ids).Delete(&dbmodel.AssetMetaRec{}); err != nil {
				return fmt.Errorf("delete legacy template metadata: %w", err)
			}
		}
	})
}
