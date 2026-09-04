/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/hashicorp/go-multierror"
)

// HandleProjectDeleteRequest handles an incoming project deletion request
func HandleProjectDeleteRequest(ctx context.Context, req *events.ProjectDeleteRequest) error {
	logging.Infof("Handling project delete request: %v", req)
	if req.LeaveOrphans {
		return nil
	}

	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to parse tenant id. err: %s", apiErr.Message)
		return fmt.Errorf("failed to parse tenant id. err: %s", apiErr.Message)
	}
	// Why don't we store tenantId as uint in the database?
	// To align it with other places where the tenantId is stored as string, maybe one day it will be UUID
	tenantIdStr := fmt.Sprintf("%d", tenantId)

	var resultErr error
	if err := dao_crud.DeleteWorkloadsOfProjectId(ctx, tenantIdStr, req.ClusterId, req.ProjectId); err != nil {
		resultErr = multierror.Append(resultErr, err)
	}
	if err := dao_crud.DeleteTemplatesOfProjectId(ctx, tenantIdStr, req.ProjectId); err != nil {
		resultErr = multierror.Append(resultErr, err)
	}
	return resultErr
}

func HandleClusterDeleteRequest(ctx context.Context, req *events.ClusterDelete) error {
	var resultErr error
	logging.Infof("Handling cluster deleted request: %v", req)
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to parse tenant id. err: %s", apiErr.Message)
		return fmt.Errorf("failed to parse tenant id. err: %s", apiErr.Message)
	}
	tenantIdStr := fmt.Sprintf("%d", tenantId)

	// When a cluster is deleted, there is no more communication between the control plane and the cluster.
	// Therefor entities in workload-crud are clean out right away from the db. and not mark as deleted and handled
	// by the "main workload deletion flow"
	if err := dao_crud.DeleteWorkloadsOfClusterId(ctx, tenantIdStr, req.ClusterId); err != nil {
		resultErr = multierror.Append(resultErr, err)
	}

	if err := dao.DeleteWorkloadsOfClusterId(ctx, tenantIdStr, req.ClusterId); err != nil {
		resultErr = multierror.Append(resultErr, err)
	}

	if err := dao.DbDeleteWorkloadEventsOfClusterId(ctx, tenantIdStr, req.ClusterId); err != nil {
		resultErr = multierror.Append(resultErr, err)
	}
	if err := dao.DeletePrioritiesOfClusterId(ctx, tenantIdStr, req.ClusterId); err != nil {
		resultErr = multierror.Append(resultErr, err)
	}
	if err := dao_crud.DeleteTemplatesOfClusterId(ctx, tenantIdStr, req.ClusterId); err != nil {
		resultErr = multierror.Append(resultErr, err)
	}
	return resultErr
}

func HandleTenantDeleteRequest(ctx context.Context, req *events.TenantDelete) error {
	logging.Infof("Handling tenant deleted request: %v", req)
	var resultErr error
	if err := dao_crud.DeleteWorkloadsOfTenantId(ctx, req.TenantId); err != nil {
		resultErr = multierror.Append(resultErr, err)
	}

	if err := dao.DeleteWorkloadsOfTenantId(ctx, req.TenantId); err != nil {
		resultErr = multierror.Append(resultErr, err)
	}

	if err := dao.DbDeleteWorkloadEventsOfTenantId(ctx, req.TenantId); err != nil {
		resultErr = multierror.Append(resultErr, err)
	}

	if err := dao.DeleteExternalWorkloadTypeResourceInterfacesOfTenantId(ctx, req.TenantId); err != nil {
		resultErr = multierror.Append(resultErr, err)
	}

	if err := dao.DeleteExternalWorkloadTypeConfigsOfTenantId(ctx, req.TenantId); err != nil {
		resultErr = multierror.Append(resultErr, err)
	}

	if err := dao.DeleteWorkloadTypeStatusesOfTenantId(ctx, req.TenantId); err != nil {
		resultErr = multierror.Append(resultErr, err)
	}

	if err := dao.DeletePrioritiesOfTenantId(ctx, req.TenantId); err != nil {
		resultErr = multierror.Append(resultErr, err)
	}
	if err := dao_crud.DeleteTemplatesOfTenantId(ctx, req.TenantId); err != nil {
		resultErr = multierror.Append(resultErr, err)
	}
	return resultErr
}
