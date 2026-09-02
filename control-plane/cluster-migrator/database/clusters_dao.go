/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-migrator/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

func GetTenantClusterNames(tenantId int32) (map[string]string, error) {
	clusters := []models.Cluster{}

	err := targetClient.
		Where("tenant_id = ?", tenantId).
		Find(&clusters)
	if err != nil {
		logging.Errorf("Failed to query tenant clusters for tenant %d. err: %v", tenantId, err)
		return nil, err
	}

	mapClusterIdToName := map[string]string{}

	for _, cluster := range clusters {
		mapClusterIdToName[cluster.Id] = cluster.Name
	}

	return mapClusterIdToName, nil
}
