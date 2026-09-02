/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

func GetTenantNodeTypes(tenantId int32) ([]models.NodeType, error) {
	nodeTypes := []models.NodeType{}

	result := sourceClient.
		Where("tenant_id = ?", tenantId).
		Find(&nodeTypes)
	if result.Error != nil {
		logging.Errorf("Failed to query tenant node types for tenant %d. err: %v", tenantId, result.Error)
		return nil, result.Error
	}

	return nodeTypes, nil
}
