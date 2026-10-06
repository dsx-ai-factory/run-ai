/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster_delete

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	authorization_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	extapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	org_unit_client "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/versions"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
)

type ClusterDeleter interface {
	AttemptToGracefullyDeleteCluster(ctx context.Context, cluster *models.Cluster) *api.Error
	ForceDeleteCluster(ctx context.Context, cluster *models.Cluster) *api.Error
	NonAsyncDelete(ctx context.Context, cluster *models.Cluster) *api.Error
	DeleteTenantClusters(ctx context.Context, id int) error
	DeleteClusterFromDB(ctx context.Context, clusterId string, tenantId int32) error
	PublishClusterDeleteNotification(ctx context.Context, clusterId string, tenantId int32) error
	PublishClusterPendingDeleteNotification(ctx context.Context, clusterId string, tenantId int32) error
}

const (
	MinVersionForGracefulDeletion = "2.20.0"
)

type ClusterDeleterImpl struct {
	orgUnitClient       org_unit_client.IOrgUnitClient
	authorizationClient authorization_client.IAuthorizationClient
	nodepoolsService    api.NodePoolsAPIServicer
}

func NewClusterDeleter(authClient authorization_client.IAuthorizationClient, orgUnitClient org_unit_client.IOrgUnitClient, nodepoolsService api.NodePoolsAPIServicer,
) ClusterDeleter {
	return &ClusterDeleterImpl{
		orgUnitClient:       orgUnitClient,
		authorizationClient: authClient,
		nodepoolsService:    nodepoolsService,
	}
}

// isPubsubEnabled checks the current config value instead of using a cached field.
// This ensures we always check the latest pubsub initialization status,
// which is important since pubsub initializes asynchronously.
func (s *ClusterDeleterImpl) isPubsubEnabled() bool {
	serviceConfig := runai_config.Get[*config.Service]()
	return serviceConfig.PubsubInitialized
}

func (s *ClusterDeleterImpl) DeleteTenantClusters(ctx context.Context, tenantId int) error {
	var errs error
	if err := dao.HardDeleteAllTenantClusters(ctx, int32(tenantId)); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("error on tenant deletion. tenant id: {%d}, service {%s}, asset {clusters}, error: {%v}", tenantId, "cluster-service", err)
		errs = errors.Join(errs, err)
	}
	if err := dao.HardDeleteAllTenantNodes(ctx, int32(tenantId)); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("error on tenant deletion. tenant id: {%d}, service {%s}, asset {nodes}, error: {%v}", tenantId, "cluster-service", err)
		errs = errors.Join(errs, err)
	}
	if errs == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("All clusters of Tenant:[%d] successfully deleted", tenantId)
		return nil
	}
	return errs
}

func (s *ClusterDeleterImpl) AttemptToGracefullyDeleteCluster(ctx context.Context, cluster *models.Cluster) *api.Error {
	if err := dao.UpdateClusterDeleteRequestTimestamp(ctx, cluster.ID.String(), cluster.TenantId, time.Now()); err != nil {
		return err
	}
	pubsubEnabled := s.isPubsubEnabled()
	if pubsubEnabled {
		if err := s.PublishClusterPendingDeleteNotification(ctx, cluster.ID.String(), cluster.TenantId); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to publish cluster pending delete notification for cluster %s: %v", cluster.ID.String(), err)
		}
	}
	clusterSupportVersionSupportGracefulShutdown := s.clusterSupportVersionSupportGracefulShutdown(ctx, cluster.Version)
	if !clusterSupportVersionSupportGracefulShutdown || !pubsubEnabled {
		err := s.ForceDeleteCluster(ctx, cluster)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ClusterDeleterImpl) ForceDeleteCluster(ctx context.Context, cluster *models.Cluster) *api.Error {
	if cluster.DeleteRequestedAt == nil {
		if err := dao.UpdateClusterDeleteRequestTimestamp(ctx, cluster.ID.String(), cluster.TenantId, time.Now()); err != nil {
			return err
		}
	}
	if !s.isPubsubEnabled() {
		if err := s.NonAsyncDelete(ctx, cluster); err != nil {
			return err
		}
	} else {
		if err := s.PublishClusterDeleteNotification(ctx, cluster.ID.String(), cluster.TenantId); err != nil {
			return api.NewError(http.StatusInternalServerError, "Failed to publish cluster delete notification", err)
		}
	}
	return nil
}

func (s *ClusterDeleterImpl) NonAsyncDelete(ctx context.Context, cluster *models.Cluster) *api.Error {
	clusterId := cluster.ID.String()
	err := s.DeleteClusterFromDB(ctx, clusterId, cluster.TenantId)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, "Failed to delete cluster", err)
	}
	// This part is handled by the services when the cluster delete event is published
	err = s.orgUnitClient.ClusterDeleted(ctx, clusterId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to notify org-unit of cluster: %s deletion, [%v]", clusterId, err)
	}
	_, apiErr := s.authorizationClient.
		DeleteScopeAccessRules(ctx).
		ScopeType("cluster").
		ScopeId(clusterId).
		Execute()
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete cluster %s scoped access rules. Err %v", cluster.ID.String(), apiErr)
	}
	return nil
}

func (s *ClusterDeleterImpl) DeleteClusterFromDB(ctx context.Context, clusterId string, tenantId int32) error {
	_, apiErr := dao.GetClusterByIdAndTenantId(ctx, clusterId, tenantId, extapi.VerbosityMetadata)
	if apiErr != nil {
		if apiErr.Code == http.StatusNotFound {
			return nil
		} else {
			return fmt.Errorf("failed to check cluster existence  %s: %v", clusterId, apiErr)
		}
	}
	if err := dao.DeleteCluster(ctx, clusterId, tenantId); err != nil {
		return fmt.Errorf("failed to delete cluster %s: %v", clusterId, err)
	}
	if daoErr := dao.DeleteClusterNodepools(ctx, tenantId, "", clusterId); daoErr != nil {
		return fmt.Errorf("failed to delete cluster %s nodepools: %v", clusterId, daoErr)
	}
	if daoErr := dao.DeleteClusterConfig(ctx, tenantId, clusterId); daoErr != nil {
		return fmt.Errorf("failed to delete cluster config %s: %v", clusterId, daoErr)
	}
	if daoErr := dao.DeleteClusterNodes(ctx, tenantId, clusterId); daoErr != nil {
		return fmt.Errorf("failed to delete cluster %s nodes: %v", clusterId, daoErr)
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Cluster: [%s] and its Nodepools and Nodes of Tenant:[%d] successfully deleted", clusterId, tenantId)
	return nil
}

func (s *ClusterDeleterImpl) PublishClusterDeleteNotification(ctx context.Context, clusterId string, tenantId int32) error {
	ctxForPubSub := context.WithValue(ctx, runaictx.CfTenantId, strconv.Itoa(int(tenantId)))
	event := events.NewClusterDelete(clusterId)
	return runaipubsub.Publish[events.ClusterDelete](ctxForPubSub, event)
}

func (s *ClusterDeleterImpl) PublishClusterPendingDeleteNotification(ctx context.Context, clusterId string, tenantId int32) error {
	ctxForPubSub := context.WithValue(ctx, runaictx.CfTenantId, strconv.Itoa(int(tenantId)))
	event := events.NewClusterPendingDelete(clusterId)
	return runaipubsub.Publish[events.ClusterPendingDelete](ctxForPubSub, event)
}

func (s *ClusterDeleterImpl) clusterSupportVersionSupportGracefulShutdown(ctx context.Context, existingClusterVersion string) bool {
	if existingClusterVersion == "" {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("cluster version is empty, cluster graceful deletion is not supported")
		return false
	}
	result, innerErr := versions.AreAllVersionsSupported([]string{existingClusterVersion}, MinVersionForGracefulDeletion)
	if innerErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to check if cluster async delete is supported for cluster version, error: %v", innerErr)
		return false
	}
	return result
}
