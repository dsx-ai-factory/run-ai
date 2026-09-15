/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"sync"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/metrics"
	"go.uber.org/zap"

	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
)

type ClusterInfoHandler struct {
	resourceName       constants.ClusterResource
	runaiBackendClient *backendclient.BackendClient
	clusterInfoParser  *parsers.ClusterInfoParser
	cachedClient       cachedclient.CachedClient
}

func NewClusterInfoHandler(runaiBackendClient *backendclient.BackendClient, cachedClient cachedclient.CachedClient) *ClusterInfoHandler {
	// we are using watchClient and not cachedClient because the cluster-info-controller is using the watchClient,
	// so we want them to use the same client
	return &ClusterInfoHandler{
		resourceName:       constants.ClusterInfoResourceName,
		runaiBackendClient: runaiBackendClient,
		clusterInfoParser:  parsers.NewClusterInfoParser(),
		cachedClient:       cachedClient,
	}
}

func (handler *ClusterInfoHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	zap.S().Infof("Start cluster info Sync")

	key := client.ObjectKey{
		Name:      constants.ClusterInfoConfigMapName,
		Namespace: runaiv1.DefaultNamespace,
	}
	err := handler.HandleUpdateItem(key)
	if err != nil {
		zap.S().Errorf("Failed to send cluster info: %v", err)
	}
}

func (handler *ClusterInfoHandler) HandleUpdateItem(key client.ObjectKey) error {
	clusterInfoCm := &corev1.ConfigMap{}
	err := handler.cachedClient.Get(context.Background(), key, clusterInfoCm)
	if err != nil {
		zap.S().Errorf("Failed to get runai public ConfigMap: %v", err)
		return err
	}

	return handler.HandleAddUpdate(clusterInfoCm, constants.UpdateOperation)
}

func (handler *ClusterInfoHandler) HandleAddUpdate(clusterInfoCm *corev1.ConfigMap, operation constants.OperationName) error {
	return handler.sendClusterInfo(clusterInfoCm, operation)
}

func (handler *ClusterInfoHandler) sendClusterInfo(clusterInfoCm *corev1.ConfigMap, operation constants.OperationName) error {
	clusterInfoToSend, err := handler.clusterInfoParser.ParseClusterInfoToServerScheme(clusterInfoCm)
	if err != nil {
		return err
	}

	{
		defer metrics.GetMetricsExporter().MeasureExecutionTime(handler.resourceName, operation)()
		res, err := handler.runaiBackendClient.ClusterServiceClient.ClustersAPI.UpdateClusterInfo(
			handler.runaiBackendClient.Context(),
			config.Config().ClusterUuid,
		).ClusterInfoSyncRequest(clusterInfoToSend).Execute()
		if res != nil {
			_ = res.Body.Close()
		}

		return handler.runaiBackendClient.HandleHttpResponse(
			res,
			err,
			"Updated ClusterInfo",
		)
	}
}
