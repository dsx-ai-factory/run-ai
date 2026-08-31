/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"time"

	"github.com/run-ai/runai/runai-application-packages/convutils"

	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"gorm.io/gorm"
	"k8s.io/utils/ptr"
)

func DeleteProjectInDb(ctx context.Context, project *models.Project, subjectId *string, deletionType api.DeletionTypes, force bool) (string, int64, error) {
	if project.DeletionRequest == nil {
		project.DeletionRequest = ptr.To(time.Now())
		project.ForceDelete = &force
		project.DeletionType = &deletionType
	}
	if subjectId != nil {
		project.DeletedBy = *subjectId
	}

	tx := dao_utils.DB().WithContext(ctx).Begin()
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)

	err := DeleteProjectInDbInTx(ctx, tx, project)
	if err != nil {
		return "", -1, err
	}

	result := tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit delete project: %v", result.Error)
		return "", -1, result.Error
	}

	InvalidateAllCache(ctx, project.TenantId, project.ClusterId.String())

	return project.Name, convutils.Deref(project.ParentID, -1), nil
}

func DeleteProjectInDbInTx(ctx context.Context, tx *gorm.DB, project *models.Project) error {
	_, err := deleteOrgUnitDataInTx(tx, project.OrgUnitDataID)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete project orgUnitData: %v", err)
		return err
	}
	_, err = deleteQuotaStatusInTx(tx, project.QuotaStatusID)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete project quotaStatus: %v", err)
		return err
	}
	result := tx.Delete(&project, project.ID)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete project: %v", result.Error)
		return result.Error
	}

	return nil
}

func DeleteProjectsInDb(ctx context.Context, projects []*models.Project) (deletedProjects []*models.Project, err error) {
	tx := dao_utils.DB().WithContext(ctx).Begin()
	committed := false
	defer dao_utils.FinalizeTransaction(tx, &committed)
	const savingPoint = "sp"
	tx.SavePoint(savingPoint)

	for _, project := range projects {
		err := DeleteProjectInDbInTx(ctx, tx, project)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete project: %v", project.ID)
			tx.RollbackTo(savingPoint) // Rollback to last saved point
		} else {
			tx.SavePoint(savingPoint)
			deletedProjects = append(deletedProjects, project)
		}
	}

	result := tx.Commit()
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit delete project: %v", result.Error)
		return []*models.Project{}, result.Error
	}

	return deletedProjects, nil
}

func SetProjectAsDeleting(ctx context.Context, project *models.Project, subjectId *string, deletionType api.DeletionTypes, force bool) (string, error) {
	project.DeletionRequest = ptr.To(time.Now())
	project.ForceDelete = &force
	project.DeletionType = &deletionType
	project.Phase = ptr.To(string(api.PROJECTPHASE_DELETING))

	if subjectId != nil {
		project.DeletedBy = *subjectId
	}

	tx := dao_utils.DB().WithContext(ctx)

	result := tx.Where("tenant_id = ? AND id = ?", project.TenantId, project.ID).Updates(project)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to set project as deleting: %v", result.Error)
		return "", result.Error
	}

	InvalidateAllCache(ctx, project.TenantId, project.ClusterId.String())

	return project.Name, nil
}
