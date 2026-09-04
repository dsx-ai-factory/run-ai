/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-migrator/database/models"
	clusterServiceModels "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm/clause"
)

func GetTenantNodePools(tenantId int32) ([]models.NodepoolV1, error) {
	nodepools := []models.NodepoolV1{}

	err := sourceClient.
		Where("tenant_id = ?", tenantId).
		Find(&nodepools)
	if err != nil {
		logging.Errorf("Failed to query tenant nodepools for tenant %d. err: %v", tenantId, err)
		return nil, err
	}

	return nodepools, nil
}

func CreateNodepools(ctx context.Context, nodepools []clusterServiceModels.Nodepool) error {
	if len(nodepools) == 0 {
		return nil
	}

	return targetClient.WithContext(ctx).Transaction(func(tx orm.DB) error {
		_, err := tx.
			Clauses(clause.OnConflict{UpdateAll: true}).
			CreateInBatches(nodepools, 20)
		return err
	})
}
