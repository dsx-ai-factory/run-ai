/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
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

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	err := DeleteProjectInDbInTx(ctx, tx, project)
	if err != nil {
		return "", -1, err
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit delete project: %v", err)
		return "", -1, err
	}

	InvalidateAllCache(ctx, project.TenantId, project.ClusterId.String())

	return project.Name, convutils.Deref(project.ParentID, -1), nil
}

func DeleteProjectInDbInTx(ctx context.Context, tx orm.DB, project *models.Project) error {
	err := deleteOrgUnitDataInTx(tx, project.OrgUnitDataID)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete project orgUnitData: %v", err)
		return err
	}
	err = deleteQuotaStatusInTx(tx, project.QuotaStatusID)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete project quotaStatus: %v", err)
		return err
	}
	if _, err := tx.Delete(&project, project.ID); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete project: %v", err)
		return err
	}

	return nil
}

func DeleteProjectsInDb(ctx context.Context, projects []*models.Project) (deletedProjects []*models.Project, err error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)
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

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit delete project: %v", err)
		return []*models.Project{}, err
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

	tx := runaidb.DB().WithContext(ctx)

	if _, err := tx.Where("tenant_id = ? AND id = ?", project.TenantId, project.ID).Updates(project); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to set project as deleting: %v", err)
		return "", err
	}

	InvalidateAllCache(ctx, project.TenantId, project.ClusterId.String())

	return project.Name, nil
}
