/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	"k8s.io/utils/ptr"

	grovev1alpha1 "github.com/ai-dynamo/grove/operator/api/core/v1alpha1"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/metrics"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NetworkTopologyHandler struct {
	resourceName  constants.ClusterResource
	cachedReader  client.Reader
	backendClient *runaiBackendClient.BackendClient
}

func NewNetworkTopologyHandler(runaiBackendClient *runaiBackendClient.BackendClient, cachedReader client.Reader) *NetworkTopologyHandler {
	return &NetworkTopologyHandler{
		resourceName:  constants.NetworkTopologyResourceName,
		cachedReader:  cachedReader,
		backendClient: runaiBackendClient,
	}
}

func (networkTopologyHandler *NetworkTopologyHandler) HandleUpdateItem(_ client.ObjectKey) error {
	return nil
}

func (networkTopologyHandler *NetworkTopologyHandler) HandleAdd(topology *kaiv1alpha1.Topology) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(networkTopologyHandler.resourceName, constants.AddOperation)()
	return networkTopologyHandler.addNetworkTopology(topology)
}

func (networkTopologyHandler *NetworkTopologyHandler) HandleUpdate(topology *kaiv1alpha1.Topology) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(networkTopologyHandler.resourceName, constants.UpdateOperation)()
	return networkTopologyHandler.updateNetworkTopology(topology)
}

func (networkTopologyHandler *NetworkTopologyHandler) HandleDelete(topology *kaiv1alpha1.Topology) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(networkTopologyHandler.resourceName, constants.DeleteOperation)()
	return networkTopologyHandler.deleteNetworkTopology(topology)
}

func (networkTopologyHandler *NetworkTopologyHandler) addNetworkTopology(topology *kaiv1alpha1.Topology) error {
	return networkTopologyHandler.syncNetworkTopology(topology, openapi.NETWORKTOPOLOGYSYNCACTION_ADD)
}

func (networkTopologyHandler *NetworkTopologyHandler) updateNetworkTopology(topology *kaiv1alpha1.Topology) error {
	return networkTopologyHandler.syncNetworkTopology(topology, openapi.NETWORKTOPOLOGYSYNCACTION_UPDATE)
}

func (networkTopologyHandler *NetworkTopologyHandler) syncNetworkTopology(topology *kaiv1alpha1.Topology, action openapi.NetworkTopologySyncAction) error {
	groveTopology := networkTopologyHandler.fetchGroveTopology(topology)
	adapter := parsers.NewTopologyAdapter(topology)
	syncFields := adapter.ToNetworkTopologySyncFields(groveTopology)

	syncFields.Action = ptr.To(action)
	request := openapi.NetworkTopologySyncRequest{
		NetworkTopologies: []openapi.NetworkTopologySyncFields{*syncFields},
	}
	res, err := networkTopologyHandler.sendSyncRequest(request)
	if res != nil && res.Body != nil {
		defer func() { _ = res.Body.Close() }()
	}
	return networkTopologyHandler.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("%s network topology %s", networkTopologySyncActionVerb(action), topology.Name))
}

func (networkTopologyHandler *NetworkTopologyHandler) deleteNetworkTopology(topology *kaiv1alpha1.Topology) error {
	adapter := parsers.NewTopologyAdapter(topology)
	toDelete := adapter.ToNetworkTopologySyncFields(nil)
	if toDelete.Source != parsers.NetworkTopologiesClusterSource {
		return nil
	}
	toDelete.Action = ptr.To(openapi.NETWORKTOPOLOGYSYNCACTION_DELETE)
	request := openapi.NetworkTopologySyncRequest{
		NetworkTopologies: []openapi.NetworkTopologySyncFields{*toDelete},
	}
	res, err := networkTopologyHandler.sendSyncRequest(request)
	if res != nil {
		_ = res.Body.Close()
	}

	return networkTopologyHandler.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("%s network topology %s", networkTopologySyncActionVerb(openapi.NETWORKTOPOLOGYSYNCACTION_DELETE), topology.Name))
}

func networkTopologySyncActionVerb(action openapi.NetworkTopologySyncAction) string {
	switch action {
	case openapi.NETWORKTOPOLOGYSYNCACTION_ADD:
		return "Added"
	case openapi.NETWORKTOPOLOGYSYNCACTION_UPDATE:
		return "Updated"
	case openapi.NETWORKTOPOLOGYSYNCACTION_DELETE:
		return "Deleted"
	}
	return ""
}

func (networkTopologyHandler *NetworkTopologyHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	zap.S().Infof("Start network topology Sync")

	// Don't convert to var since this sends an empty array when no data is in cluster vs null (which will result in bad request (400))
	topologiesList := &kaiv1alpha1.TopologyList{}
	kubeErr := networkTopologyHandler.cachedReader.List(context.Background(), topologiesList)
	if kubeErr != nil {
		zap.S().Errorf("%+v", kubeErr)
		return
	}

	groveTopologies := networkTopologyHandler.fetchGroveTopologies(topologiesList.Items)
	networkTopologies := parsers.NewTopologiesAdapter(topologiesList.Items).
		ToNetworkTopologiesSyncFields(groveTopologies)
	for i := range networkTopologies {
		networkTopologies[i].Action = ptr.To(openapi.NETWORKTOPOLOGYSYNCACTION_ADD)
	}

	request := openapi.NetworkTopologySyncRequest{
		FullSync:          ptr.To(true),
		NetworkTopologies: networkTopologies,
	}
	res, err := networkTopologyHandler.sendSyncRequest(request)
	if res != nil {
		_ = res.Body.Close()
	}

	err = networkTopologyHandler.backendClient.HandleHttpResponse(res, err, "Sync network topologies")
	if err != nil {
		zap.S().Errorf("Had an error while syncing network topologies: %+v", err)
	}
}

func (networkTopologyHandler *NetworkTopologyHandler) fetchGroveTopology(topology *kaiv1alpha1.Topology) *grovev1alpha1.ClusterTopologyBinding {
	groveTopologyName := topology.Annotations[commonconstants.GroveTopologyAnnotation]
	if groveTopologyName == "" {
		return nil
	}
	clusterTopology := &grovev1alpha1.ClusterTopologyBinding{}
	err := networkTopologyHandler.cachedReader.Get(context.Background(), types.NamespacedName{Name: groveTopologyName}, clusterTopology)
	if err != nil {
		zap.S().Errorf("Failed to fetch ClusterTopology %s: %+v", groveTopologyName, err)
		return nil
	}
	return clusterTopology
}

func (networkTopologyHandler *NetworkTopologyHandler) fetchGroveTopologies(topologies []kaiv1alpha1.Topology) map[string]*grovev1alpha1.ClusterTopologyBinding {
	groveTopologies := make(map[string]*grovev1alpha1.ClusterTopologyBinding)
	for _, topology := range topologies {
		groveTopology := networkTopologyHandler.fetchGroveTopology(&topology)
		if groveTopology != nil {
			groveTopologies[topology.Name] = groveTopology
		}
	}
	return groveTopologies
}

func (networkTopologyHandler *NetworkTopologyHandler) sendSyncRequest(syncRequest openapi.NetworkTopologySyncRequest) (*http.Response, error) {
	networkTopologyApi := networkTopologyHandler.backendClient.ClusterServiceClient.NetworkTopologiesAPI
	request := networkTopologyApi.SyncNetworkTopologies(networkTopologyHandler.backendClient.Context(), config.Config().ClusterUuid).
		NetworkTopologySyncRequest(syncRequest)

	return request.Execute()
}
