/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm/clause"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateNodeType(ctx context.Context, nodeType *models.NodeType) (*models.NodeType, error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	if _, err := tx.Create(nodeType); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed node type creation: %v", err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed commit node type creation: %v", err)
		return nil, err
	}
	committed = true
	return nodeType, nil
}

func CreateNodeTypes(ctx context.Context, nodeTypes []*models.NodeType) error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	if _, err := tx.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(nodeTypes, 20); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed node types creation: %v", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed commit node types creation: %v", err)
		return err
	}
	committed = true
	return nil
}

func UpdateNodeType(ctx context.Context, nodeType *models.NodeType) (*models.NodeType, error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	if _, err := tx.Where(&models.NodeType{IntBaseModel: models.IntBaseModel{ID: nodeType.ID}, ClusterId: nodeType.ClusterId, TenantId: nodeType.TenantId}).Updates(nodeType); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed node type update: %v", err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed commit node type update: %v", err)
		return nil, err
	}
	committed = true
	InvalidateAllCache(ctx, nodeType.TenantId, nodeType.ClusterId.String())

	return nodeType, nil
}

func getOrgUnitIdsOfNodeType(tx orm.DB, nodeTypeId int64) ([]string, error) {
	orgUnitIds := []string{}

	ids := []string{}
	if err := tx.Table(models.OrgUnitNodeTypeTrainTable).Where("node_type_id = ?", nodeTypeId).Select("org_unit_data_id").Find(&ids); err != nil {
		return orgUnitIds, err
	}
	idsSet := map[string]bool{}
	for _, item := range ids {
		idsSet[item] = true
	}
	if err := tx.Table(models.OrgUnitNodeTypeWorkspaceTable).Where("node_type_id = ?", nodeTypeId).Select("org_unit_data_id").Find(&ids); err != nil {
		return orgUnitIds, err
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
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	orgUnitIds, err := getOrgUnitIdsOfNodeType(tx, nodeTypeId)
	if err != nil {
		return err
	}

	if _, err := tx.Table(models.OrgUnitNodeTypeTrainTable).Delete(&models.OrgUnitNodeType{}, "node_type_id = ?", nodeTypeId); err != nil {
		return err
	}
	if _, err := tx.Table(models.OrgUnitNodeTypeWorkspaceTable).Delete(&models.OrgUnitNodeType{}, "node_type_id = ?", nodeTypeId); err != nil {
		return err
	}

	err = updateUpdatedAt(ctx, orgUnitIds, tx, subjectId)
	if err != nil {
		return err
	}

	if _, err := tx.Where(nodeType).Delete(&nodeType); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete node type: %v", err)
		return err
	}
	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed commit node type deletion: %v", err)
		return err
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
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	if _, err := tx.Table(models.OrgUnitNodeTypeTrainTable).Delete(&models.OrgUnitNodeType{}, "node_type_id = ?", nodeTypeId); err != nil {
		return err
	}
	if _, err := tx.Table(models.OrgUnitNodeTypeWorkspaceTable).Delete(&models.OrgUnitNodeType{}, "node_type_id = ?", nodeTypeId); err != nil {
		return err
	}
	if _, err := tx.Where(nodeType).Delete(&nodeType); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete node type: %v", err)
		return err
	}
	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed commit node type deletion: %v", err)
		return err
	}
	committed = true
	InvalidateAllCache(ctx, nodeType.TenantId, nodeType.ClusterId.String())

	return nil
}

func GetNodeTypes(ctx context.Context, tenantId string, clusterId uuid.UUID, permittedScopes *openapi.PermittedScopes) (*[]models.NodeType, error) {
	nodeTypes := &[]models.NodeType{}
	db := runaidb.DB().WithContext(ctx).Table(models.NodeTypeTable)
	if permittedScopes != nil {
		if !IsTenantPermitted(tenantId, permittedScopes) && !IsClusterPermitted(clusterId.String(), permittedScopes) {
			return nodeTypes, nil
		}
	}
	if err := db.Where(&models.NodeType{TenantId: tenantId, ClusterId: clusterId}).Find(&nodeTypes); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed get node type: %v", err)
		return nil, err
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
	db := runaidb.DB().WithContext(ctx).Table(models.NodeTypeTable)
	if err := db.Where(nodeType).First(&nodeType); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed get node type: %v", err)
		}
		return nil, err
	}
	return nodeType, nil
}

func DeleteClusterNodeTypes(ctx context.Context, tenantId string, clusterId uuid.UUID) error {
	nodeTypes := &[]models.NodeType{}

	nodeType := &models.NodeType{
		ClusterId: clusterId,
		TenantId:  tenantId,
	}
	if _, err := runaidb.DB().WithContext(ctx).Table(models.NodeTypeTable).Where(nodeType).Delete(nodeTypes); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get node type: %v", err)
		return err
	}
	InvalidateAllCache(ctx, nodeType.TenantId, nodeType.ClusterId.String())

	return nil
}
