/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package status_post_update_handlers

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	clusterdelete "github.com/run-ai/runai/control-plane/cluster-service/pkg/services/cluster-delete/mocks"
	"github.com/stretchr/testify/mock"
)

func TestHandleClusterConfigStatusChangeClusterDeactivated(t *testing.T) {
	clusterDeleter := clusterdelete.NewClusterDeleter(t)
	clusterDeleter.On("ForceDeleteCluster", mock.Anything, mock.Anything).Return(nil).Maybe()
	ctx := context.Background()
	uuid7, _ := uuid.NewV7()
	cluster := &models.Cluster{ID: uuid7, TenantId: 1}
	clusterConfig := &models.ClusterConfig{
		Config: models.ClusterConfigData{
			UserConfig: api.ClusterUserConfigData{},
			SystemConfig: api.ClusterSystemConfigData{
				Deactivated: true,
			},
		},
		Status: models.ClusterConfigStatus{
			ActualConfig: api.ClusterFullConfigData{
				SystemConfig: api.ClusterSystemConfigData{
					Deactivated: true,
				},
				UserConfig: api.ClusterUserConfigData{},
			},
			Conditions: []api.Condition{
				{
					Type:               "Applied",
					Status:             "True",
					LastTransitionTime: time.Now().Add(-2 * time.Minute),
				},
				{
					Type:               "Applying",
					Status:             "True",
					LastTransitionTime: time.Now().Add(-3 * time.Minute),
				},
			},
		},
	}
	handler := newConfigStatusHandlerClusterDeactivationHandler(clusterDeleter)
	handler.HandleClusterConfigStatusChange(ctx, cluster, clusterConfig)
	clusterDeleter.AssertCalled(t, "ForceDeleteCluster", ctx, cluster)
}

func TestHandleClusterConfigStatusChangeClusterNotDeactivated(t *testing.T) {
	clusterDeleter := clusterdelete.NewClusterDeleter(t)
	clusterDeleter.On("ForceDeleteCluster", mock.Anything, mock.Anything).Return(nil).Maybe()
	ctx := context.Background()
	uuid7, _ := uuid.NewV7()
	cluster := &models.Cluster{ID: uuid7, TenantId: 1}
	clusterConfig := &models.ClusterConfig{
		Config: models.ClusterConfigData{
			UserConfig: api.ClusterUserConfigData{},
			SystemConfig: api.ClusterSystemConfigData{
				Deactivated: false,
			},
		},
		Status: models.ClusterConfigStatus{
			ActualConfig: api.ClusterFullConfigData{
				SystemConfig: api.ClusterSystemConfigData{
					Deactivated: false,
				},
				UserConfig: api.ClusterUserConfigData{},
			},
			Conditions: []api.Condition{
				{
					Type:               "Applied",
					Status:             "True",
					LastTransitionTime: time.Now().Add(-2 * time.Minute),
				},
				{
					Type:               "Applying",
					Status:             "True",
					LastTransitionTime: time.Now().Add(-3 * time.Minute),
				},
			},
		},
	}
	handler := newConfigStatusHandlerClusterDeactivationHandler(clusterDeleter)
	handler.HandleClusterConfigStatusChange(ctx, cluster, clusterConfig)
	clusterDeleter.AssertNotCalled(t, "ForceDeleteCluster", ctx, cluster)
}
