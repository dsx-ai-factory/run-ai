/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/common"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/dao"
	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"strconv"
)

func apiErrToError(err *api.Error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %s", err.Message, err.Details)
}

func HandleProjectDeletedEvent(ctx context.Context, event *events.ProjectDeleted) error {
	logging.Infof("Handling Project Deleted event: %v", event)
	if event.LeaveOrphans {
		return nil
	}

	tenantId, err := common.TenantIdFromContext(ctx)
	if err != nil {
		return apiErrToError(err)
	}

	// First delete issues that are related to other entries (e.g. issues on pvc) in other tables
	if err = dao.DeleteProjectResourceIssues(ctx, tenantId, event.ProjectId); err != nil {
		return apiErrToError(err)
	}

	// Delete all the objects that are related to the project, no order is needed
	if err = dao.DeleteProjectConfigMaps(ctx, tenantId, event.ProjectId); err != nil {
		return apiErrToError(err)
	}

	if err = dao.DeleteProjectObjects(ctx, tenantId, event.ProjectId); err != nil {
		return apiErrToError(err)
	}
	if err = dao.DeleteProjectPvcs(ctx, tenantId, event.ProjectId); err != nil {
		return apiErrToError(err)
	}

	if err = dao.DeleteProjectSecrets(ctx, tenantId, event.ProjectId); err != nil {
		return apiErrToError(err)
	}

	if err = dao.DeleteProjectStorageClasses(ctx, tenantId, event.ProjectId); err != nil {
		return apiErrToError(err)
	}

	return nil
}

// HandleClusterDeleteEvent handles event that a cluster was deleted from the control plane
func HandleClusterDeleteEvent(ctx context.Context, event *events.ClusterDelete) error {
	logging.Infof("Handling Clusted Deleted event: %v", event)

	tenantId, err := common.TenantIdFromContext(ctx)
	if err != nil {
		return apiErrToError(err)
	}

	// First delete issues that are related to other entries (e.g. issues on pvc) in other tables
	if err = dao.DeleteClusterResourceIssues(ctx, tenantId, event.ClusterId); err != nil {
		return apiErrToError(err)
	}

	// Delete all the objects that are related to the cluster, no order is needed

	// ConfigMap isn't supported at cluster level, but still need to delete ConfigMap in a cluster's project
	if err = dao.DeleteClusterConfigMaps(ctx, tenantId, event.ClusterId); err != nil {
		return apiErrToError(err)
	}

	if err = dao.DeleteClusterObjects(ctx, tenantId, event.ClusterId); err != nil {
		return apiErrToError(err)
	}
	if err = dao.DeleteClusterPvcs(ctx, tenantId, event.ClusterId); err != nil {
		return apiErrToError(err)
	}

	if err = dao.DeleteClusterSecrets(ctx, tenantId, event.ClusterId); err != nil {
		return apiErrToError(err)
	}

	if err = dao.DeleteClusterStorageClasses(ctx, tenantId, event.ClusterId); err != nil {
		return apiErrToError(err)
	}

	return nil
}

// HandleTenantDeleteEvent handles event that a tenant was deleted from the control plane
func HandleTenantDeleteEvent(ctx context.Context, event *events.TenantDelete) error {
	logging.Infof("Handling Tenant Deleted event: %v", event)

	// The tenant ID comes from the event, not the context,
	// since a tenant deletion request wouldn't originate from within the tenant itself.
	parsedTenantId, parseErr := strconv.ParseInt(event.TenantId, 10, 32)
	if parseErr != nil {
		logging.Errorf("Failed to parse tenant ID %s: %v", event.TenantId, parseErr)
		return parseErr
	}
	tenantId := int32(parsedTenantId)

	var err *api.Error

	// First delete issues that are related to other entries (e.g. issues on pvc) in other tables
	if err = dao.DeleteTenantResourceIssues(ctx, tenantId); err != nil {
		return apiErrToError(err)
	}

	// ConfigMap isn't supported at tenant level, but still need to delete ConfigMap in a tenant's project
	if err = dao.DeleteTenantConfigMaps(ctx, tenantId); err != nil {
		return apiErrToError(err)
	}

	if err = dao.DeleteTenantObjects(ctx, tenantId); err != nil {
		return apiErrToError(err)
	}

	if err = dao.DeleteTenantPvcs(ctx, tenantId); err != nil {
		return apiErrToError(err)
	}

	if err = dao.DeleteTenantSecrets(ctx, tenantId); err != nil {
		return apiErrToError(err)
	}

	if err = dao.DeleteTenantStorageClasses(ctx, tenantId); err != nil {
		return apiErrToError(err)
	}
	return nil
}
