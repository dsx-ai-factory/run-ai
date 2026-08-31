/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/run-ai/runai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/database/dao"
	events "github.com/run-ai/runai/runai-application-packages/async-events"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

func apiErrToError(err *api.Error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %s", err.Message, err.Details)
}

func HandleTenantDeleteEvent(ctx context.Context, event *events.TenantDelete) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Handling tenant delete event for tenant: %s", event.TenantId)

	tenantId, err := strconv.ParseInt(event.TenantId, 10, 32)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to parse tenant ID %s: %v", event.TenantId, err)
		return err
	}

	// Clean workloads queue records for the deleted tenant
	dbErr := dao.CleanWorkloadsOfScope(ctx, int32(tenantId), nil, nil, nil)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to clean workloads queue for tenant %s: %v", event.TenantId, dbErr)
		return apiErrToError(dbErr)
	}

	// Clean sync status records for the deleted tenant
	dbErr = dao.CleanSyncStatusOfScope(ctx, int32(tenantId), nil, nil, nil)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to clean sync status for tenant %s: %v", event.TenantId, dbErr)
		return apiErrToError(dbErr)
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully cleaned workloads-manager data for deleted tenant: %s", event.TenantId)
	return nil
}

// HandleClusterDeleteEvent handles event that a cluster was deleted from the control plane
func HandleClusterDeleteEvent(ctx context.Context, event *events.ClusterDelete) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Handling cluster delete event for cluster: %s", event.ClusterId)

	// Get tenant ID from context
	tenantIdStr := ctx.Value(runaictx.CfTenantId)
	if tenantIdStr == nil {
		err := fmt.Errorf("tenant ID not found in context")
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get tenant ID from context: %v", err)
		return err
	}

	tenantId, err := strconv.ParseInt(tenantIdStr.(string), 10, 32)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to parse tenant ID %s: %v", tenantIdStr, err)
		return err
	}

	// Clean workloads queue records for the deleted cluster
	dbErr := dao.CleanWorkloadsOfScope(ctx, int32(tenantId), &event.ClusterId, nil, nil)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to clean workloads queue for cluster %s: %v", event.ClusterId, dbErr)
		return apiErrToError(dbErr)
	}

	// Clean sync status records for the deleted cluster
	dbErr = dao.CleanSyncStatusOfScope(ctx, int32(tenantId), &event.ClusterId, nil, nil)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to clean sync status for cluster %s: %v", event.ClusterId, dbErr)
		return apiErrToError(dbErr)
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully cleaned workloads-manager data for deleted cluster: %s", event.ClusterId)
	return nil
}

// HandleProjectDeletedEvent handles event that a project was deleted
func HandleProjectDeletedEvent(ctx context.Context, event *events.ProjectDeleted) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Handling project delete event for project: %s", event.ProjectId)

	if event.LeaveOrphans {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Project deletion event for %s has LeaveOrphans=true, skipping cleanup", event.ProjectId)
		return nil
	}

	// Get tenant ID from context
	tenantIdStr := ctx.Value(runaictx.CfTenantId)
	if tenantIdStr == nil {
		err := fmt.Errorf("tenant ID not found in context")
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get tenant ID from context: %v", err)
		return err
	}

	tenantId, err := strconv.ParseInt(tenantIdStr.(string), 10, 32)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to parse tenant ID %s: %v", tenantIdStr, err)
		return err
	}

	// Clean workloads queue records for the deleted project
	dbErr := dao.CleanWorkloadsOfScope(ctx, int32(tenantId), nil, nil, &event.ProjectId)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to clean workloads queue for project %s: %v", event.ProjectId, dbErr)
		return apiErrToError(dbErr)
	}

	// Clean sync status records for the deleted project
	dbErr = dao.CleanSyncStatusOfScope(ctx, int32(tenantId), nil, nil, &event.ProjectId)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to clean sync status for project %s: %v", event.ProjectId, dbErr)
		return apiErrToError(dbErr)
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully cleaned workloads-manager data for deleted project: %s", event.ProjectId)
	return nil
}
