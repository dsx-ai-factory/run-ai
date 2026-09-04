/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/common"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

type PvcCreatedOrUpdated int

const (
	NewlyCreatedPvc PvcCreatedOrUpdated = iota
	ExistingPvcUpdated
)

// upsertPvc upserts a PVC into the database.
// returns a boolean value indicating if this is the first time we create the PVC or if we only update it.
func upsertPvc(db orm.DB, pvc *models.Pvc) (bool, error) {
	record := models.Pvc{
		PvcFilters: pvc.PvcFilters,
	}

	isNewPvc := false
	err := db.Model(&record).Where(&record).First(&record)
	if err == nil {
		pvc.ID = record.ID
		_, err = db.Model(&pvc).Save(&pvc)
	} else {
		if !errors.Is(err, orm.ErrRecordNotFound) {
			return isNewPvc, err
		}
		_, err = db.Model(&pvc).Create(&pvc)
		isNewPvc = true
	}

	return isNewPvc, err
}

func ListPvcs(ctx context.Context, filter *models.PvcFilters, options *models.ListOptions) ([]models.Pvc, *api.Error) {
	if options == nil {
		options = &models.ListOptions{}
	}

	record := models.NewPvc(filter)
	dbErr := api.NewError(http.StatusServiceUnavailable, record.ErrorMessage("list"), nil)

	db := runaidb.DB().WithContext(ctx)

	query, err := filter.Apply(ctx, db.Model(record), options.IncludeAncestors, options.IncludeDescendants)
	if err != nil {
		return nil, err
	}

	pvcs := &[]models.Pvc{}
	if err := query.Order(filter.PvcName).Find(&pvcs); err != nil {
		dbErr = dbErr.WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)

		return nil, dbErr
	}

	return *pvcs, nil
}

// CreatePvc creates a new PVC in the database.
// returns the PVC and a boolean value indicating if this is the first time we create the PVC or if we only update it.
func CreatePvc(ctx context.Context, pvc *models.Pvc) (*models.Pvc, *bool, *api.Error) {
	dbErr := api.NewError(http.StatusServiceUnavailable, pvc.ErrorMessage("insert"), nil)

	db := runaidb.DB()

	isNewPvc, err := upsertPvc(db, pvc)

	if err != nil {
		dbErr = dbErr.WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)

		return nil, nil, dbErr
	}

	return pvc, &isNewPvc, nil
}

func DeletePvc(ctx context.Context, filters *models.PvcFilters) *api.Error {
	record := models.NewPvc(filters)
	dbErr := api.NewError(http.StatusServiceUnavailable, record.ErrorMessage("delete"), nil)

	db := runaidb.DB()

	if _, err := db.
		Where(filters).
		Delete(record); err != nil {
		dbErr = dbErr.WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)

		return dbErr
	}

	return nil
}

// ApplyPvcs applies a list of PVCs to the database.
// returns a map of PVC ids to PVCs and a map of PVC ids to boolean values indicating if this is the first time we create the PVC or if we only update it.
func ApplyPvcs(ctx context.Context, filters *models.PvcFilters, pvcs []api.Pvc) (map[string]api.Pvc, map[string]PvcCreatedOrUpdated, *api.Error) {
	record := models.NewPvc(filters)
	dbErr := api.NewError(http.StatusServiceUnavailable, record.ErrorMessage("apply"), nil)

	db := runaidb.DB()

	tx := db.Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	pvcNames := make([]string, 0)
	// Keep track on the ids of the PVCs that were applied to relate the issues with the PVCs
	pvcIdMap := make(map[string]api.Pvc, len(pvcs))
	isNewPvcMap := make(map[string]PvcCreatedOrUpdated, len(pvcs))

	for _, pvc := range pvcs {
		toAdd := record.ScopeParams()
		toAdd.PvcName = &pvc.Name
		toAdd.Origin = OriginPtrToStringPtr(pvc.Origin)
		toAdd.StorageClassName = pvc.StorageClassName
		if pvc.Phase != nil {
			toAdd.Phase = convutils.Ref(string(*pvc.Phase))
		}
		toAdd.ActualCapacity = pvc.ActualCapacity
		toAdd.RequestedCapacity = pvc.RequestedCapacity

		pvcNames = append(pvcNames, pvc.Name)

		isNewPvc, err := upsertPvc(db, toAdd)
		if err != nil {
			dbErr = dbErr.WithDetails(err)
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)
			return nil, nil, dbErr
		}
		pvcIdMap[toAdd.ID.String()] = pvc
		if isNewPvc {
			isNewPvcMap[toAdd.ID.String()] = NewlyCreatedPvc
		} else {
			isNewPvcMap[toAdd.ID.String()] = ExistingPvcUpdated
		}
	}

	// Delete all records that are not in the list. for example
	// if I have in the table pvcs aaa bbb ccc and I receive from the sync bbb ccc ddd
	// I need to upsert ccc and ddd, but I also not to remove aaa (because it is no longer in the cluster)
	if _, err := tx.
		Where(filters).
		Where(fmt.Sprintf("%s not in ?", models.PvcsNameColumn), pvcNames).
		Delete(record); err != nil {
		dbErr = dbErr.WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)
		return nil, nil, dbErr
	}

	if err := tx.Commit(); err != nil {
		dbErr = dbErr.WithDetails(err)
		if dbErr.Code != http.StatusConflict {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)
			return nil, nil, dbErr
		}
	}
	committed = true

	return pvcIdMap, isNewPvcMap, nil
}

func DeletePvcsNotInCluster(ctx context.Context, filters *models.PvcFilters, pvcIds []string) *api.Error {
	record := models.NewPvc(filters)
	dbErr := api.NewError(http.StatusServiceUnavailable, record.ErrorMessage("delete"), nil)

	db := runaidb.DB()

	tenantId, err := common.TenantIdFromContext(ctx)
	if err != nil {
		return err
	}

	query := db.Model(record).
		Where("tenant_id = ?", tenantId).
		Where(filters)

	if len(pvcIds) != 0 {
		query = query.Not("id in ?", pvcIds)
	}

	if _, err := query.Delete(record); err != nil {
		dbErr = dbErr.WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)
		return dbErr
	}

	return nil
}

// DeleteProjectPvcs deletes all pvcs of a project
func DeleteProjectPvcs(ctx context.Context, tenantId int32, projectId string) *api.Error {
	return deleteEntitiesByScope(ctx, openapi.SCOPETYPE_PROJECT, tenantId, projectId, &models.Pvc{})
}

// DeleteClusterPvcs deletes all pvcs within a cluster
func DeleteClusterPvcs(ctx context.Context, tenantId int32, clusterId string) *api.Error {
	return deleteClusterEntities(ctx, tenantId, clusterId, &models.Pvc{})
}

// DeleteTenantPvcs deletes all pvcs within a tenant
func DeleteTenantPvcs(ctx context.Context, tenantId int32) *api.Error {
	return deleteTenantEntities(ctx, tenantId, &models.Pvc{})
}
