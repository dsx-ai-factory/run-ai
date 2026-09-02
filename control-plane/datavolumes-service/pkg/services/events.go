/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/datavolumes-service/pkg/database/dbtables"
	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

func HandleTenantDeleteEvent(ctx context.Context, req *events.TenantDelete) error {
	logging.Infof("Handeling tenant delete request: %v", req)
	return dao.DeleteDataVolumesByScope(ctx, dbtables.ScopeTenant, req.TenantId)
}

func HandleClusterDeleteEvent(ctx context.Context, req *events.ClusterDelete) error {
	logging.Infof("Handeling cluster delete request: %v", req)
	return dao.DeleteDataVolumesByScope(ctx, api.SCOPETYPE_CLUSTER, req.ClusterId)
}

func HandleProjectDeleteEvent(ctx context.Context, req *events.ProjectDeleted) error {
	logging.Infof("Handeling project delete request: %v", req)
	return dao.DeleteDataVolumesByScope(ctx, api.SCOPETYPE_PROJECT, req.ProjectId)
}
