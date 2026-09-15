/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"
	"sync"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/metrics"
	"k8s.io/utils/ptr"

	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"go.uber.org/zap"
	_ "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NodePoolHandler struct {
	resourceName  constants.ClusterResource
	cachedReader  client.Reader
	backendClient *runaiBackendClient.BackendClient
}

func NewNodePoolHandler(runaiBackendClient *runaiBackendClient.BackendClient, cachedReader client.Reader) *NodePoolHandler {
	return &NodePoolHandler{
		resourceName:  constants.NodePoolResourceName,
		cachedReader:  cachedReader,
		backendClient: runaiBackendClient,
	}
}

func (nodePoolHandler *NodePoolHandler) HandleUpdateItem(key client.ObjectKey) error {
	nodePool := &v1alpha1.NodePool{}
	err := nodePoolHandler.cachedReader.Get(context.Background(), key, nodePool)
	if err != nil {
		return err
	}

	return nodePoolHandler.HandleUpdate(nodePool)
}

func (nodePoolHandler *NodePoolHandler) HandleAdd(nodePool *v1alpha1.NodePool) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(nodePoolHandler.resourceName, constants.AddOperation)()
	return nodePoolHandler.updateNodePoolStatus(nodePool)
}

func (nodePoolHandler *NodePoolHandler) HandleUpdate(nodePool *v1alpha1.NodePool) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(nodePoolHandler.resourceName, constants.UpdateOperation)()
	return nodePoolHandler.updateNodePoolStatus(nodePool)
}

func (nodePoolHandler *NodePoolHandler) HandleDelete(nodePool *v1alpha1.NodePool) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(nodePoolHandler.resourceName, constants.DeleteOperation)()
	toDelete := parsers.NewNodepoolAdapter(nodePool).ToDeleteRequest()

	nodepoolApi := nodePoolHandler.backendClient.ClusterServiceClient.NodePoolsAPI
	request := nodepoolApi.SyncNodepoolsStatus(nodePoolHandler.backendClient.Context(), config.Config().ClusterUuid).
		SyncNodepoolsStatusRequest(openapi.SyncNodepoolsStatusRequest{
			Nodepools:     []openapi.NodepoolSyncFields{*toDelete},
			DeleteMissing: ptr.To(false),
		})

	res, err := request.Execute()
	if res != nil {
		_ = res.Body.Close()

	}
	return nodePoolHandler.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Deleted node pool %s", nodePool.Name))
}

func (nodePoolHandler *NodePoolHandler) updateNodePoolStatus(nodePool *v1alpha1.NodePool) error {
	toUpdate, err := parsers.NewNodepoolAdapter(nodePool).ToServerScheme()
	if err != nil {
		zap.S().Infof("No need to update node pool <%s> status - status is %s", nodePool.Name, nodePool.Status.Phase)
		return nil
	}
	nodepoolApi := nodePoolHandler.backendClient.ClusterServiceClient.NodePoolsAPI
	request := nodepoolApi.SyncNodepoolsStatus(nodePoolHandler.backendClient.Context(), config.Config().ClusterUuid).
		SyncNodepoolsStatusRequest(openapi.SyncNodepoolsStatusRequest{
			Nodepools:     []openapi.NodepoolSyncFields{*toUpdate},
			DeleteMissing: ptr.To(false),
		})

	res, err := request.Execute()
	if res != nil {
		_ = res.Body.Close()

	}

	return nodePoolHandler.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Updated node pool %s", nodePool.Name))
}

func (nodePoolHandler *NodePoolHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	zap.S().Infof("Start node pool Sync")

	// Don't convert to var since this sends an empty array when no data is in cluster vs null (which will result in bad request (400))
	nodePools := &v1alpha1.NodePoolList{}
	kubeErr := nodePoolHandler.cachedReader.List(context.Background(), nodePools)
	if kubeErr != nil {
		zap.S().Errorf("%+v", kubeErr)
		return
	}

	nodePoolsInCluster, parseErr := parsers.NewNodepoolsAdapter(nodePools.Items).ToServerScheme()
	if parseErr != nil {
		zap.S().Errorf("Had an error while parsing node pools: %+v", parseErr)
		return
	}

	nodepoolApi := nodePoolHandler.backendClient.ClusterServiceClient.NodePoolsAPI

	request := nodepoolApi.SyncNodepoolsStatus(nodePoolHandler.backendClient.Context(), config.Config().ClusterUuid).
		SyncNodepoolsStatusRequest(openapi.SyncNodepoolsStatusRequest{
			Nodepools:     nodePoolsInCluster,
			DeleteMissing: ptr.To(true),
		})
	res, err := request.Execute()
	if res != nil {
		_ = res.Body.Close()

	}
	err = nodePoolHandler.backendClient.HandleHttpResponse(res, err, "Sync node pools")
	if err != nil {
		zap.S().Errorf("Had an error while syncing node pools: %+v", err)
	}
}
