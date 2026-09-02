/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"context"
	"fmt"
	"time"

	orgUnitsServiceDao "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/dao"
	orgUnitServiceModels "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

func CreateNodeTypes(nodeTypes []models.NodeType) error {
	nodeTypesToCreate := convutils.ConvertSliceTypes(nodeTypes, mapNodeTypeToV2)
	return orgUnitsServiceDao.CreateNodeTypes(context.Background(), nodeTypesToCreate)
}

func mapNodeTypeToV2(nodeType models.NodeType) *orgUnitServiceModels.NodeType {
	return &orgUnitServiceModels.NodeType{
		IntBaseModel: orgUnitServiceModels.IntBaseModel{
			ID:        int64(nodeType.ID),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: "",
			UpdatedBy: "",
		},
		Name:      nodeType.Name,
		ClusterId: nodeType.ClusterUUID,
		TenantId:  fmt.Sprintf("%d", nodeType.TenantID),
	}
}
