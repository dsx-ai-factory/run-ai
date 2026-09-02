/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"

	"github.com/google/uuid"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CreateNodeType(ctx context.Context, nodeType *models.NodeType) (*models.NodeType, error) {
	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to db session start: %v", tx.Error)
		return nil, tx.Error
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

	result := tx.Create(nodeType)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed node type creation: %v", result.Error)
		return nil, result.Error
	}

	result = tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed commit node type creation: %v", result.Error)
		return nil, result.Error
	}
	committed = true
	return nodeType, nil
}

func CreateNodeTypes(ctx context.Context, nodeTypes []*models.NodeType) error {
	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to db session start: %v", tx.Error)
		return tx.Error
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

	result := tx.
		Clauses(clause.OnConflict{
			UpdateAll: true,
		}).
		CreateInBatches(nodeTypes, 20)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed node types creation: %v", result.Error)
		return result.Error
	}

	result = tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed commit node types creation: %v", result.Error)
		return result.Error
	}
	committed = true
	return nil
}

func UpdateNodeType(ctx context.Context, nodeType *models.NodeType) (*models.NodeType, error) {
	tx := dao_utils.DB().WithContext(ctx).Begin()
	if tx.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to db session start: %v", tx.Error)
		return nil, tx.Error
	}
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

	result := tx.Where(&models.NodeType{
		IntBaseModel: models.IntBaseModel{
			ID: nodeType.ID,
		},
		ClusterId: nodeType.ClusterId,
		TenantId:  nodeType.TenantId,
	}).Updates(nodeType)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed node type creation: %v", result.Error)
		return nil, result.Error
	}

	result = tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed commit node type creation: %v", result.Error)
		return nil, result.Error
	}
	committed = true
	InvalidateAllCache(ctx, nodeType.TenantId, nodeType.ClusterId.String())

	return nodeType, nil
}

func getOrgUnitIdsOfNodeType(tx *gorm.DB, nodeTypeId int64) ([]string, error) {
	orgUnitIds := []string{}

	ids := []string{}
	result := tx.Table(models.OrgUnitNodeTypeTrainTable).
		Where("node_type_id = ?", nodeTypeId).
		Select("org_unit_data_id").
		Find(&ids)
	if result.Error != nil {
		return orgUnitIds, result.Error
	}
	idsSet := map[string]bool{}
	for _, item := range ids {
		idsSet[item] = true
	}
	result = tx.Table(models.OrgUnitNodeTypeWorkspaceTable).
		Where("node_type_id = ?", nodeTypeId).
		Select("org_unit_data_id").
		Find(&ids)
	if result.Error != nil {
		return orgUnitIds, result.Error
	}
	for _, item := range ids {
		idsSet[item] = true
	}

	for id := range idsSet {
		orgUnitIds = append(orgUnitIds, id)
	}
	return orgUnitIds, nil
}

func DeleteNodeType(ctx context.Context, nodeTypeId int64, tenantId string, subjectId string) error {
	nodeType := &models.NodeType{
		IntBaseModel: models.IntBaseModel{ID: nodeTypeId},
		TenantId:     tenantId,
	}
	tx := dao_utils.DB().WithContext(ctx).Begin()
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

	orgUnitIds, err := getOrgUnitIdsOfNodeType(tx, nodeTypeId)
	if err != nil {
		return err
	}

	result := tx.Table(models.OrgUnitNodeTypeTrainTable).Delete(&models.OrgUnitNodeType{}, "node_type_id = ?", nodeTypeId)
	if result.Error != nil {
		return result.Error
	}
	result = tx.Table(models.OrgUnitNodeTypeWorkspaceTable).Delete(&models.OrgUnitNodeType{}, "node_type_id = ?", nodeTypeId)
	if result.Error != nil {
		return result.Error
	}

	err = updateUpdatedAt(ctx, orgUnitIds, tx, subjectId)
	if err != nil {
		return err
	}

	result = tx.Where(nodeType).Delete(&nodeType)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete node type: %v", result.Error)
		return result.Error
	}
	result = tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed commit node type deletion: %v", result.Error)
		return result.Error
	}
	committed = true
	InvalidateAllCache(ctx, nodeType.TenantId, nodeType.ClusterId.String())

	return nil
}

func DeleteNodeTypeFromCluster(ctx context.Context, nodeTypeId int64, tenantId string, clusterId uuid.UUID) error {
	nodeType := &models.NodeType{
		IntBaseModel: models.IntBaseModel{ID: nodeTypeId},
		ClusterId:    clusterId,
		TenantId:     tenantId,
	}
	tx := dao_utils.DB().WithContext(ctx).Begin()
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

	result := tx.Table(models.OrgUnitNodeTypeTrainTable).Delete(&models.OrgUnitNodeType{}, "node_type_id = ?", nodeTypeId)
	if result.Error != nil {
		return result.Error
	}
	result = tx.Table(models.OrgUnitNodeTypeWorkspaceTable).Delete(&models.OrgUnitNodeType{}, "node_type_id = ?", nodeTypeId)
	if result.Error != nil {
		return result.Error
	}
	result = tx.Where(nodeType).Delete(&nodeType)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete node type: %v", result.Error)
		return result.Error
	}
	result = tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed commit node type deletion: %v", result.Error)
		return result.Error
	}
	committed = true
	InvalidateAllCache(ctx, nodeType.TenantId, nodeType.ClusterId.String())

	return nil
}

func GetNodeTypes(ctx context.Context, tenantId string, clusterId uuid.UUID, permittedScopes *openapi.PermittedScopes) (*[]models.NodeType, error) {
	nodeTypes := &[]models.NodeType{}
	db := dao_utils.DB().WithContext(ctx).Table(models.NodeTypeTable)
	if permittedScopes != nil {
		if !IsTenantPermitted(tenantId, permittedScopes) && !IsClusterPermitted(clusterId.String(), permittedScopes) {
			return nodeTypes, nil
		}
	}
	result := db.Where(&models.NodeType{
		TenantId:  tenantId,
		ClusterId: clusterId,
	}).Find(&nodeTypes)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed get node type: %v", result.Error)
		return nil, result.Error
	}
	return nodeTypes, nil
}

func GetNodeType(ctx context.Context, tenantId string, clusterId uuid.UUID, nodeTypeId int64) (*models.NodeType, error) {
	nodeType := &models.NodeType{
		IntBaseModel: models.IntBaseModel{
			ID: nodeTypeId,
		},
		ClusterId: clusterId,
		TenantId:  tenantId,
	}
	db := dao_utils.DB().WithContext(ctx).Table(models.NodeTypeTable)
	result := db.Where(nodeType).First(&nodeType)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed get node type: %v", result.Error)
		}
		return nil, result.Error
	}
	return nodeType, nil
}

func DeleteClusterNodeTypes(ctx context.Context, tenantId string, clusterId uuid.UUID) error {
	nodeTypes := &[]models.NodeType{}

	nodeType := &models.NodeType{
		ClusterId: clusterId,
		TenantId:  tenantId,
	}
	result := dao_utils.DB().WithContext(ctx).Table(models.NodeTypeTable).Where(nodeType).Delete(nodeTypes)

	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get node type: %v", result.Error)
		return result.Error
	}
	InvalidateAllCache(ctx, nodeType.TenantId, nodeType.ClusterId.String())

	return nil
}
