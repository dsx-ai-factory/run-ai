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

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/metrics"
	"go.uber.org/zap"

	swagger "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	runaiBackendClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/util"
	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NodeHandler struct {
	cachedReader            client.Reader
	backendClient           *runaiBackendClient.BackendClient
	maxNumOfNodesPerRequest int
	nodeParser              *parsers.NodeParser
	numaNodeHandler         *NumaNodeHandler

	// names of already synced nodes
	syncedObjects      map[string]bool
	syncedObjectsMutex sync.RWMutex
	syncMutex          sync.Mutex
	syncWaiting        atomic.Bool
}

func NewNodeHandler(runaiBackendClient *runaiBackendClient.BackendClient, cachedReader client.Reader) *NodeHandler {
	nodeParser := parsers.NewNodeParser(cachedReader)
	return &NodeHandler{
		cachedReader:            cachedReader,
		backendClient:           runaiBackendClient,
		maxNumOfNodesPerRequest: config.Config().MaxNumOfNodesPerRequest,
		nodeParser:              nodeParser,
		syncedObjects:           map[string]bool{},
	}
}

func (nodeHandler *NodeHandler) HandleUpdateItem(key client.ObjectKey) error {
	node := &corev1.Node{}
	err := nodeHandler.cachedReader.Get(context.Background(), key, node)
	if err != nil {
		return err
	}

	return nodeHandler.HandleUpdate(node)
}

func (nodeHandler *NodeHandler) HandleAdd(node *corev1.Node) error {
	nodeHandler.syncedObjectsMutex.RLock()
	_, found := nodeHandler.syncedObjects[node.Name]
	nodeHandler.syncedObjectsMutex.RUnlock()
	if found {
		zap.S().Infof("Node <%s> was already synced - not adding", node.Name)
		return nil
	}

	return nodeHandler.updateNodeInfo(node, "Added")
}

func (nodeHandler *NodeHandler) HandleUpdate(node *corev1.Node) error {
	return nodeHandler.updateNodeInfo(node, "Updated")
}

func (nodeHandler *NodeHandler) HandleDelete(node *corev1.Node) error {
	defer metrics.GetMetricsExporter().MeasureExecutionTime(constants.NodeResourceName, constants.DeleteOperation)()
	res, err := nodeHandler.backendClient.ClusterServiceClient.NodesAPI.DeleteNode(
		nodeHandler.backendClient.Context(),
		config.Config().ClusterUuid,
		node.Name,
	).Execute()
	if res != nil {
		_ = res.Body.Close()
	}
	return nodeHandler.backendClient.HandleHttpResponse(res, err, fmt.Sprintf("Deleted node %s", node.Name))
}

func (nodeHandler *NodeHandler) updateNodeInfo(node *corev1.Node, action string) error {
	nodeInfo := nodeHandler.nodeParser.ParseNodeToServerScheme(node)

	{
		defer metrics.GetMetricsExporter().MeasureExecutionTime(constants.NodeResourceName, constants.UpdateOperation)()
		res, err := nodeHandler.backendClient.ClusterServiceClient.NodesAPI.UpsertNode(
			nodeHandler.backendClient.Context(),
			config.Config().ClusterUuid,
			node.Name,
		).NodeInfo(nodeInfo).Execute()
		if res != nil {
			_ = res.Body.Close()
		}
		return nodeHandler.backendClient.HandleHttpResponse(res, err,
			fmt.Sprintf("%s node %s", action, node.Name))
	}
}

func (nodeHandler *NodeHandler) HandleSyncEvent(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	// NUMA rows reference the physical node rows in cluster-service. Sync them only
	// after the node snapshot succeeds so cluster-service can resolve every node ID.
	// Registering both full-sync handlers independently would run them concurrently,
	// and the one-shot NUMA full sync has no rate-limited retry if it wins that race.
	if nodeHandler.syncAllNodes() && nodeHandler.numaNodeHandler != nil {
		nodeHandler.numaNodeHandler.SyncAll()
	}
}

func (nodeHandler *NodeHandler) SetNumaNodeHandler(numaNodeHandler *NumaNodeHandler) {
	nodeHandler.numaNodeHandler = numaNodeHandler
}

// UpdateGpuFractioningSupport stores the cluster's fractioning verdict and resyncs every node
// when it flips. The per-node conditions are derived from it, but the Node objects do not change
// when it does, so the update-diff path sees nothing and would leave the reported conditions
// stale until the next periodic full sync.
func (nodeHandler *NodeHandler) UpdateGpuFractioningSupport(clusterInfoCm *corev1.ConfigMap, isInInitialList bool) {
	if !nodeHandler.nodeParser.UpdateGpuFractioningSupport(clusterInfoCm) {
		return
	}
	// The startup full sync runs after the informer seeds this, so it already reports the verdict.
	if isInInitialList {
		return
	}
	zap.S().Info("Cluster GPU fractioning support changed, resyncing all nodes")
	nodeHandler.syncAllNodes()
}

func (nodeHandler *NodeHandler) syncAllNodes() bool {
	// Allow at most 1 waiting goroutine; additional callers are dropped
	if !nodeHandler.syncWaiting.CompareAndSwap(false, true) {
		zap.S().Debugf("Skipping nodes sync - another goroutine is already waiting")
		return false
	}

	nodeHandler.syncMutex.Lock()
	nodeHandler.syncWaiting.Store(false)
	defer nodeHandler.syncMutex.Unlock()

	zap.S().Infof("Start nodes Sync")

	nodes := &corev1.NodeList{}
	err := nodeHandler.cachedReader.List(context.Background(), nodes)
	if err != nil {
		zap.S().Errorf("Failed listing nodes in cluster, error: %+v", err)
		return false
	}

	managedNodes := filterOutExcludedNodes(nodes.Items)

	nodeNamesMap, nodesToSync := nodeHandler.nodeParser.ParseNodesToServerScheme(managedNodes)

	err = nodeHandler.handleNodesSyncWithPagination(nodesToSync)

	nodeHandler.syncedObjectsMutex.Lock()
	defer nodeHandler.syncedObjectsMutex.Unlock()
	if err != nil {
		zap.S().Errorf("Had an error while syncing nodes: %+v", err)
		nodeHandler.syncedObjects = map[string]bool{}
		return false
	}
	nodeHandler.syncedObjects = nodeNamesMap
	return true
}

// filterOutExcludedNodes drops nodes labeled with the excluded-nodes sentinel
// so they are never synced to cluster-service. The event-driven path
// (node_controller.go) already skips these nodes; this guards the periodic
// full resync, which lists all nodes regardless of their labels.
func filterOutExcludedNodes(nodes []corev1.Node) []corev1.Node {
	managedNodes := make([]corev1.Node, 0, len(nodes))
	for idx := range nodes {
		node := nodes[idx]
		if node.Labels[v1alpha1.NodePoolNameLabel] == clusterconstants.ExcludedNodepoolName {
			zap.S().Debugf("Node %s is excluded from the cluster - filtering it out", node.Name)
			continue
		}
		managedNodes = append(managedNodes, node)
	}
	return managedNodes
}

func (nodeHandler *NodeHandler) handleNodesSyncWithPagination(nodesToSync []swagger.NodeForSync) error {
	if len(nodesToSync) <= nodeHandler.maxNumOfNodesPerRequest {
		return nodeHandler.sendSyncNodesRequest(nodesToSync, "", "")
	}

	// sort all nodes first
	sortNodesByName(nodesToSync)

	for i := 0; i < len(nodesToSync); i += nodeHandler.maxNumOfNodesPerRequest - 1 {
		chunkEndIndex := i + nodeHandler.maxNumOfNodesPerRequest
		if chunkEndIndex > len(nodesToSync) {
			chunkEndIndex = len(nodesToSync)
		}

		chunk := nodesToSync[i:chunkEndIndex]

		from := ""
		to := ""
		if i > 0 {
			from = chunk[0].Name
		}
		if chunkEndIndex < len(nodesToSync) {
			to = chunk[len(chunk)-1].Name
		}

		if err := nodeHandler.sendSyncNodesRequest(chunk, from, to); err != nil {
			return err
		}
	}
	return nil
}

func (nodeHandler *NodeHandler) sendSyncNodesRequest(nodesToSync []swagger.NodeForSync, from, to string) error {
	msg := "Sync Nodes"
	if from != "" {
		msg = fmt.Sprintf("%s; from <%s>", msg, from)
	}
	if to != "" {
		msg = fmt.Sprintf("%s; to <%s>", msg, to)
	} else {
		msg = fmt.Sprintf("Finished %s", msg)
	}

	defer metrics.GetMetricsExporter().MeasureExecutionTime(constants.NodeResourceName, constants.SyncOperation)()
	nodesSyncRequest := swagger.NodesSyncRequest{
		NodesForSync: nodesToSync,
		From:         &from,
		To:           &to,
	}

	res, httpErr := nodeHandler.backendClient.ClusterServiceClient.NodesAPI.SyncNodes(
		nodeHandler.backendClient.Context(),
		config.Config().ClusterUuid,
	).NodesSyncRequest(nodesSyncRequest).Execute()
	if res != nil {
		_ = res.Body.Close()
	}
	statusCode, err := nodeHandler.backendClient.HandleHttpResponseGetStatus(res, httpErr, msg)
	if statusCode == http.StatusRequestEntityTooLarge && nodeHandler.maxNumOfNodesPerRequest > 2 {
		// reduce by 20%
		nodeHandler.maxNumOfNodesPerRequest = int(nodeHandler.maxNumOfNodesPerRequest * 4 / 5)
		zap.S().Warnf("Changing maxNumOfNodesPerRequest to: %d", nodeHandler.maxNumOfNodesPerRequest)
	}
	return err
}

func sortNodesByName(nodes []swagger.NodeForSync) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})
}

func (nodeHandler *NodeHandler) NodeChanged(old *corev1.Node, new *corev1.Node) bool {
	if old == nil || new == nil {
		return true
	}
	oldNodeInfo := nodeHandler.nodeParser.ParseNodeToServerScheme(old)
	newNodeInfo := nodeHandler.nodeParser.ParseNodeToServerScheme(new)

	if oldNodeInfo.NodePool != newNodeInfo.NodePool ||
		oldNodeInfo.Status != newNodeInfo.Status ||
		oldNodeInfo.CreatedAt != newNodeInfo.CreatedAt ||
		nodeHandler.nullableStringChanged(oldNodeInfo.NvLinkDomainUid, newNodeInfo.NvLinkDomainUid) ||
		nodeHandler.nullableStringChanged(oldNodeInfo.NvLinkCliqueId, newNodeInfo.NvLinkCliqueId) ||
		nodeHandler.topologyManagerPolicyChanged(oldNodeInfo.Numa, newNodeInfo.Numa) ||
		nodeHandler.conditionsChanged(oldNodeInfo.Conditions, newNodeInfo.Conditions) ||
		nodeHandler.gpuInfoChanged(oldNodeInfo.GpuInfo, newNodeInfo.GpuInfo) ||
		nodeHandler.taintsChanged(oldNodeInfo.Taints, newNodeInfo.Taints) ||
		nodeHandler.networkTopologyLabelsChanged(oldNodeInfo.NetworkTopologyLabels, newNodeInfo.NetworkTopologyLabels) {
		return true
	}
	return false
}

func (nodeHandler *NodeHandler) topologyManagerPolicyChanged(old swagger.NullableNumaInfo, new swagger.NullableNumaInfo) bool {
	oldNuma := old.Get()
	newNuma := new.Get()
	if oldNuma == nil || newNuma == nil {
		return oldNuma != newNuma
	}
	return util.CompareNullables(newNuma.TopologyManagerPolicy, oldNuma.TopologyManagerPolicy)
}

func (nodeHandler *NodeHandler) conditionsChanged(old []swagger.NodeStatusConditionDetails, new []swagger.NodeStatusConditionDetails) bool {
	if len(old) != len(new) {
		return true
	}
	oldConditionsMap := map[string]swagger.NodeStatusConditionDetails{}
	for _, oldCondition := range old {
		oldConditionsMap[oldCondition.Type] = oldCondition
	}
	for _, newCondition := range new {
		condition, found := oldConditionsMap[newCondition.Type]
		if !found {
			return true
		}
		if newCondition.Reason != condition.Reason ||
			ptrToStrChanges(condition.Message, newCondition.Message) {
			return true
		}
	}

	return false
}

func ptrToStrChanges(old *string, new *string) bool {
	if old == nil {
		return new != nil
	}
	if new == nil {
		return true
	}
	return *old != *new
}

func (nodeHandler *NodeHandler) nullableStringChanged(old swagger.NullableString, new swagger.NullableString) bool {
	return util.CompareNullables(new, old)
}

func (nodeHandler *NodeHandler) gpuInfoChanged(old swagger.NullableGpuInfo, new swagger.NullableGpuInfo) bool {
	oldGpuInfo, newGpuInfo := old.Get(), new.Get()
	if oldGpuInfo == nil || newGpuInfo == nil {
		return oldGpuInfo != newGpuInfo
	}
	return oldGpuInfo.GpuType != newGpuInfo.GpuType ||
		oldGpuInfo.GpuCount != newGpuInfo.GpuCount ||
		gpuMemoryBytesChanged(oldGpuInfo.GpuMemoryBytes, newGpuInfo.GpuMemoryBytes)
}

func gpuMemoryBytesChanged(old *int64, new *int64) bool {
	if old == nil || new == nil {
		return old != new
	}
	return *old != *new
}

func (nodeHandler *NodeHandler) networkTopologyLabelsChanged(old []swagger.NetworkTopologyLabel, new []swagger.NetworkTopologyLabel) bool {
	if len(old) != len(new) {
		return true
	}
	for i := range old {
		if old[i].LabelKey != new[i].LabelKey || old[i].LabelValue != new[i].LabelValue {
			return true
		}
	}
	return false
}

func (nodeHandler *NodeHandler) SyncNodesForNodepool(nodePoolName string) {
	zap.S().Infof("Syncing all nodes (triggered by node pool %s change)", nodePoolName)
	nodeHandler.syncAllNodes()
}

// SyncNumaNodesForNodepool resyncs NUMA details when a node pool's NUMA-aware
// scheduling is toggled. The NRT objects do not change when it flips, so the
// NRT informer sees nothing and the reported NUMA nodes would stay stale until
// the next periodic full sync.
func (nodeHandler *NodeHandler) SyncNumaNodesForNodepool(nodePoolName string) {
	// nil when the NodeResourceTopology CRD is absent and NUMA sync is disabled.
	if nodeHandler.numaNodeHandler == nil {
		return
	}
	zap.S().Infof("Syncing NUMA node details (triggered by node pool %s change)", nodePoolName)
	nodeHandler.numaNodeHandler.SyncAll()
}

func (nodeHandler *NodeHandler) taintsChanged(old []swagger.NodeTaint, new []swagger.NodeTaint) bool {
	if len(old) != len(new) {
		return true
	}
	oldTaintsMap := map[string]swagger.NodeTaint{}
	for _, oldTaint := range old {
		oldTaintsMap[oldTaint.Key] = oldTaint
	}
	for _, newTaint := range new {
		oldTaint, found := oldTaintsMap[newTaint.Key]
		if !found {
			return true
		}
		if newTaint.Effect != oldTaint.Effect ||
			ptrToStrChanges(oldTaint.Value, newTaint.Value) {
			return true
		}
	}
	return false
}
