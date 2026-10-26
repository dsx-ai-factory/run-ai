/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	nrtv1alpha2 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2"
	kainodepool "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const numaPluginName = "numa"

type NumaNodeHandler struct {
	cachedReader            client.Reader
	backendClient           *runaiBackendClient.BackendClient
	maxNumOfNodesPerRequest int

	syncedObjects      map[string]bool
	syncedObjectsMutex sync.RWMutex
	syncMutex          sync.Mutex
	syncWaiting        atomic.Bool
}

func NewNumaNodeHandler(
	backendClient *runaiBackendClient.BackendClient,
	cachedReader client.Reader,
) *NumaNodeHandler {
	return &NumaNodeHandler{
		cachedReader:            cachedReader,
		backendClient:           backendClient,
		maxNumOfNodesPerRequest: config.Config().MaxNumOfNodesPerRequest,
		syncedObjects:           map[string]bool{},
	}
}

func (h *NumaNodeHandler) HandleUpdateItem(key client.ObjectKey) error {
	nrt := &nrtv1alpha2.NodeResourceTopology{}
	if err := h.cachedReader.Get(context.Background(), key, nrt); err != nil {
		return err
	}
	return h.syncNode(parsers.ParseNodeResourceTopology(nrt), "Updated", constants.UpdateOperation)
}

func (h *NumaNodeHandler) HandleAdd(nrt *nrtv1alpha2.NodeResourceTopology) error {
	h.syncedObjectsMutex.RLock()
	wasSynced := h.syncedObjects[nrt.Name]
	h.syncedObjectsMutex.RUnlock()
	// An informer can deliver its initial Add event after the full snapshot has
	// already included this NRT. Avoid sending the same snapshot a second time.
	if wasSynced {
		zap.S().Infof("NUMA nodes for physical node <%s> were already synced - not adding", nrt.Name)
		return nil
	}

	return h.syncNode(parsers.ParseNodeResourceTopology(nrt), "Added", constants.AddOperation)
}

func (h *NumaNodeHandler) HandleDelete(nrt *nrtv1alpha2.NodeResourceTopology) error {
	return h.syncNode(openapi.NodeNumaForSync{
		Name:      nrt.Name,
		NumaNodes: []openapi.NumaNodeDetails{},
	}, "Deleted", constants.DeleteOperation)
}

func (h *NumaNodeHandler) ShouldSyncNumaNode(nodeName string) (bool, error) {
	// NRT objects may exist for every cluster node, but the control plane should
	// receive them only for nodes assigned to NUMA-enabled KAI nodepools.
	nodePoolName, err := h.getNodePoolName(nodeName)
	if err != nil || nodePoolName == "" {
		return false, err
	}

	nodePool := &kainodepool.NodePool{}
	if err = h.cachedReader.Get(
		context.Background(),
		client.ObjectKey{Name: nodePoolName},
		nodePool,
	); err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return IsNumaEnabled(nodePool), nil
}

func (h *NumaNodeHandler) getNodePoolName(nodeName string) (string, error) {
	node := &corev1.Node{}
	if err := h.cachedReader.Get(
		context.Background(),
		client.ObjectKey{Name: nodeName},
		node,
	); err != nil {
		if apierrors.IsNotFound(err) {
			return "", nil
		}
		return "", err
	}

	return nodePoolNameForNode(node), nil
}

func (h *NumaNodeHandler) syncNode(
	node openapi.NodeNumaForSync,
	action string,
	operation constants.OperationName,
) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(
		constants.NumaNodeResourceName, operation,
	)()
	return h.sendSyncRequest(
		[]openapi.NodeNumaForSync{node},
		false,
		"",
		"",
		fmt.Sprintf("%s NUMA details for node %s", action, node.Name),
	)
}

func (h *NumaNodeHandler) SyncAll() {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(
		constants.NumaNodeResourceName, constants.SyncOperation,
	)()
	if !h.syncWaiting.CompareAndSwap(false, true) {
		zap.S().Debug("Skipping NUMA sync - another goroutine is already waiting")
		return
	}

	h.syncMutex.Lock()
	h.syncWaiting.Store(false)
	defer h.syncMutex.Unlock()

	zap.S().Info("Start NUMA nodes sync")
	nrtList := &nrtv1alpha2.NodeResourceTopologyList{}
	if err := h.cachedReader.List(context.Background(), nrtList); err != nil {
		zap.S().Errorf("Failed listing NodeResourceTopology objects: %+v", err)
		h.replaceSyncedObjects(map[string]bool{})
		return
	}

	nodeNameToNodepoolName, err := h.getNodeNameToNodepoolName()
	if err != nil {
		zap.S().Errorf("Failed listing Node objects: %+v", err)
		h.replaceSyncedObjects(map[string]bool{})
		return
	}

	nodepoolNameToNumaEnabled, err := h.getNodepoolNameToNumaEnabled()
	if err != nil {
		zap.S().Errorf("Failed listing KAI NodePool objects: %+v", err)
		h.replaceSyncedObjects(map[string]bool{})
		return
	}

	numaNodes := make([]openapi.NodeNumaForSync, 0, len(nrtList.Items))
	syncedObjects := make(map[string]bool, len(nrtList.Items))
	for i := range nrtList.Items {
		if !shouldSyncNumaNode(
			nrtList.Items[i].Name,
			nodeNameToNodepoolName,
			nodepoolNameToNumaEnabled,
		) {
			zap.S().Debugf(
				"Skipping NUMA nodes for physical node %s because NUMA-aware scheduling is not enabled",
				nrtList.Items[i].Name,
			)
			continue
		}
		numaNodes = append(numaNodes, parsers.ParseNodeResourceTopology(&nrtList.Items[i]))
		syncedObjects[nrtList.Items[i].Name] = true
	}
	if err = h.syncAllWithPagination(numaNodes); err != nil {
		zap.S().Errorf("Had an error while syncing NUMA nodes: %+v", err)
		h.replaceSyncedObjects(map[string]bool{})
		return
	}
	h.replaceSyncedObjects(syncedObjects)
}

func (h *NumaNodeHandler) getNodeNameToNodepoolName() (map[string]string, error) {
	nodeList := &corev1.NodeList{}
	if err := h.cachedReader.List(context.Background(), nodeList); err != nil {
		return nil, err
	}
	nodeNameToNodepoolName := make(map[string]string, len(nodeList.Items))
	for i := range nodeList.Items {
		nodeNameToNodepoolName[nodeList.Items[i].Name] = nodePoolNameForNode(&nodeList.Items[i])
	}
	return nodeNameToNodepoolName, nil
}

func (h *NumaNodeHandler) getNodepoolNameToNumaEnabled() (map[string]bool, error) {
	nodePoolList := &kainodepool.NodePoolList{}
	if err := h.cachedReader.List(context.Background(), nodePoolList); err != nil {
		return nil, err
	}
	nodepoolNameToNumaEnabled := make(map[string]bool, len(nodePoolList.Items))
	for i := range nodePoolList.Items {
		nodepoolNameToNumaEnabled[nodePoolList.Items[i].Name] = IsNumaEnabled(&nodePoolList.Items[i])
	}
	return nodepoolNameToNumaEnabled, nil
}

func (h *NumaNodeHandler) replaceSyncedObjects(syncedObjects map[string]bool) {
	h.syncedObjectsMutex.Lock()
	h.syncedObjects = syncedObjects
	h.syncedObjectsMutex.Unlock()
}

func (h *NumaNodeHandler) syncAllWithPagination(nodes []openapi.NodeNumaForSync) error {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})

	if len(nodes) <= h.maxNumOfNodesPerRequest {
		return h.sendSyncRequest(nodes, true, "", "", "Sync NUMA nodes")
	}

	for i := 0; i < len(nodes); i += h.maxNumOfNodesPerRequest - 1 {
		chunkEndIndex := i + h.maxNumOfNodesPerRequest
		if chunkEndIndex > len(nodes) {
			chunkEndIndex = len(nodes)
		}
		chunk := nodes[i:chunkEndIndex]

		from := ""
		to := ""
		if i > 0 {
			from = chunk[0].Name
		}
		if chunkEndIndex < len(nodes) {
			to = chunk[len(chunk)-1].Name
		}
		if err := h.sendSyncRequest(chunk, true, from, to, "Sync NUMA nodes"); err != nil {
			return err
		}
	}
	return nil
}

func (h *NumaNodeHandler) sendSyncRequest(
	nodes []openapi.NodeNumaForSync,
	deleteMissing bool,
	from string,
	to string,
	message string,
) error {
	request := openapi.NumaNodesSyncRequest{
		Nodes: nodes,
	}
	if deleteMissing {
		request.DeleteMissing = &deleteMissing
	}
	if from != "" {
		request.From = &from
		message = fmt.Sprintf("%s; from <%s>", message, from)
	}
	if to != "" {
		request.To = &to
		message = fmt.Sprintf("%s; to <%s>", message, to)
	}

	response, httpErr := h.backendClient.ClusterServiceClient.NodesAPI.SyncNumaNodes(
		h.backendClient.Context(),
		config.Config().ClusterUuid,
	).NumaNodesSyncRequest(request).Execute()
	if response != nil && response.Body != nil {
		defer func() { _ = response.Body.Close() }()
	}
	statusCode, err := h.backendClient.HandleHttpResponseGetStatus(response, httpErr, message)
	if statusCode == http.StatusRequestEntityTooLarge && h.maxNumOfNodesPerRequest > 2 {
		h.maxNumOfNodesPerRequest = h.maxNumOfNodesPerRequest * 4 / 5
		zap.S().Warnf("Changing max NUMA nodes per request to: %d", h.maxNumOfNodesPerRequest)
	}
	return err
}

func shouldSyncNumaNode(
	nodeName string,
	nodeNameToNodepoolName map[string]string,
	nodepoolNameToNumaEnabled map[string]bool,
) bool {
	nodePoolName, found := nodeNameToNodepoolName[nodeName]
	if !found {
		return false
	}
	return nodepoolNameToNumaEnabled[nodePoolName]
}

func nodePoolNameForNode(node *corev1.Node) string {
	nodePoolName := node.Labels[v1alpha1.NodePoolNameLabel]
	if nodePoolName == "" {
		nodePoolName = v1alpha1.DefaultNodePoolName
	}
	return nodePoolName
}

// IsNumaEnabled mirrors nodepool-controller's NUMA plugin check.
func IsNumaEnabled(nodePool *kainodepool.NodePool) bool {
	config := nodePool.Spec.SchedulingShardConfig
	if config == nil {
		return false
	}
	plugin, found := config.Plugins[numaPluginName]
	return found && plugin.Enabled != nil && *plugin.Enabled
}
