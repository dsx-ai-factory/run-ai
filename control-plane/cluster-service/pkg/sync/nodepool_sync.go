/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sync

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

type NodePoolsSync struct {
	SyncConfig      *config.Sync
	nodepoolsClient api.NodePoolsAPIServicer
}

func NewNodePoolsSync(nodepoolsClient api.NodePoolsAPIServicer) *NodePoolsSync {
	return &NodePoolsSync{
		SyncConfig:      runai_config.Get[*config.Sync](),
		nodepoolsClient: nodepoolsClient,
	}
}

func (nps *NodePoolsSync) Start() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "internal-NodePoolsSync")

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Starting NodePoolsSync within cluster service")

	intervalMinutes := nps.SyncConfig.IntervalMinutes
	if intervalMinutes <= 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("NodePoolsSync config is miss configured, missing IntervalMinutes")
		return
	}
	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)

	go func() {
		for range ticker.C {
			nps.syncClustersDefaultNodepool(ctx)
		}
	}()
}

func (nps *NodePoolsSync) syncClustersDefaultNodepool(ctx context.Context) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("sync cluster default nodepools per tenant")
	db := runaidb.DB()
	tenantIds, err := getTenantsWithConnectedCluster(ctx, db)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get tenants, err %v", err)
		return err
	}

	tenantsNum := len(tenantIds)
	if tenantsNum == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("there are no tenants with connected clusters")
		return nil
	}

	tenantsChunk := nps.SyncConfig.TenantsInChunk
	tenantInterval := nps.SyncConfig.TenantsIntervalSeconds
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("sync cluster default nodepools per tenant in chunks of %v every %v seconds", tenantsChunk, tenantInterval)

	ticker := time.NewTicker(time.Duration(tenantInterval) * time.Second)
	defer ticker.Stop()

	for i := 0; i < tenantsNum; i += tenantsChunk {
		end := i + tenantsChunk

		if end > tenantsNum {
			end = tenantsNum
		}
		tenantIdsChunk := tenantIds[i:end]
		<-ticker.C
		err = errors.Join(err, nps.syncDefaultNodepoolPerCluster(ctx, db, tenantIdsChunk))
	}

	return err
}

func (nps *NodePoolsSync) syncDefaultNodepoolPerCluster(ctx context.Context, db orm.DB, tenantIds []string) error {
	toFixClusters, err := getClustersWithoutDefaultNodepool(ctx, db, tenantIds)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get connected clusters, err %v", err)
		return err
	}
	if len(toFixClusters) == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("All clusters in tenants %v have default nodepool", tenantIds)
		return nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Found %d clusters that need a fix", len(toFixClusters))

	failedClusters := nps.fixClustersDefaultNodepool(ctx, toFixClusters)
	if len(failedClusters) > 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to create default nodepool for clusters, failed ids: %v", failedClusters)
		return fmt.Errorf("failed to create default nodepool for %d clusters", len(failedClusters))
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Fixed %d clusters in tenants %v to have default nodepool", len(toFixClusters), tenantIds)
	return nil
}

func (nps *NodePoolsSync) fixClustersDefaultNodepool(ctx context.Context, clusters []models.Cluster) []string {
	failedClusters := []string{}
	for _, cluster := range clusters {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("cluster %s missing a default nodepool, fixing that", cluster.ID)

		ctx = context.WithValue(ctx, runaictx.CfTenantId, fmt.Sprintf("%d", cluster.TenantId))
		response, err := nps.nodepoolsClient.ClusterNotification(ctx, api.ClusterNotification{
			Operation: api.NOTIFICATIONOPERATION_CREATE,
			Id:        cluster.ID.String(),
			Name:      cluster.Name,
		})
		if err != nil || response.Code >= http.StatusBadRequest {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to create default nodepool for cluster %s, response: %v err: %v", cluster.Name, response.Body, err)
			failedClusters = append(failedClusters, cluster.ID.String())
		}
	}

	return failedClusters
}

func getTenantsWithConnectedCluster(ctx context.Context, db orm.DB) ([]string, error) {
	tx := db.WithContext(ctx)

	tenantIds := []string{}

	// Pluck is not part of the orm.DB interface - drop to the raw handle.
	result := tx.Table(models.ClusterTable).
		Where("deleted_at IS NULL").
		Distinct("tenant_id").
		Actual().
		Pluck("tenant_id", &tenantIds)

	return tenantIds, result.Error
}

func getClustersWithoutDefaultNodepool(ctx context.Context, db orm.DB, tenantIds []string) ([]models.Cluster, error) {
	tx := db.WithContext(ctx)

	var clusters []models.Cluster
	err := tx.
		Table(models.ClusterTable).
		Joins(fmt.Sprintf("left join %s ON %s.id = %s.cluster_id AND %s.is_default AND %s.deleted_at IS NULL", models.NodepoolTable, models.ClusterTable, models.NodepoolTable, models.NodepoolTable, models.NodepoolTable)).
		Where(fmt.Sprintf("%s.tenant_id IN ?", models.ClusterTable), tenantIds).
		Where(fmt.Sprintf("%s.deleted_at IS NULL", models.ClusterTable)).
		Where(fmt.Sprintf("%s.id IS NULL", models.NodepoolTable)). // Default nodepool of this cluster does not exist
		Find(&clusters)

	return clusters, err
}
