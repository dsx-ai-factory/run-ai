package clusterservice

import (
	"context"
	"fmt"
	"net/http"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

type SimpleMock struct {
	clusterNodePoolMap                          map[string]*openapi.GetNodepoolsInternal200Response
	topologyMap                                 map[string]*openapi.NetworkTopology
	clusters                                    map[string]*openapi.DisplayedCluster
	nodeNetworkTopologyLabelValues              map[string]map[string][]string // clusterUuid -> nodeName -> labels
	getNodeNetworkTopologyLabelValuesCallCount  int
	getNodeNetworkTopologyLabelValuesCallParams []string // tracks "clusterUuid:nodeName" for each call
	// Simulates caching like the real ClusterClient
	nodeNetworkTopologyLabelValuesCache map[string][]string // "clusterUuid:nodeName" -> labels
	getNodeInternalCallCount            int                 // tracks actual API calls (cache misses)
	getNodeInternalCallParams           []string            // tracks "clusterUuid:nodeName" for each API call
}

func NewSimpleMock() *SimpleMock {
	return &SimpleMock{
		clusterNodePoolMap:                  make(map[string]*openapi.GetNodepoolsInternal200Response),
		topologyMap:                         make(map[string]*openapi.NetworkTopology),
		clusters:                            make(map[string]*openapi.DisplayedCluster),
		nodeNetworkTopologyLabelValues:      make(map[string]map[string][]string),
		nodeNetworkTopologyLabelValuesCache: make(map[string][]string),
	}
}

func (smc *SimpleMock) GetNodePoolsInternalByCluster(_ context.Context, clusterUUID string) (*openapi.GetNodepoolsInternal200Response, error) {
	response, ok := smc.clusterNodePoolMap[clusterUUID]
	if !ok {
		return nil, fmt.Errorf("no node pool found for cluster %s", clusterUUID)
	}
	return response, nil
}

func (smc *SimpleMock) GetNodePoolsMinimalByCluster(ctx context.Context, clusterUUID string) (*openapi.GetNodepoolsMinimal200Response, error) {
	response, ok := smc.clusterNodePoolMap[clusterUUID]
	if !ok {
		return nil, fmt.Errorf("no node pool found for cluster %s", clusterUUID)
	}
	nodePools := []openapi.NodepoolMinimal{}
	for _, np := range response.GetNodepools() {
		nodePools = append(nodePools, openapi.NodepoolMinimal{
			Id:                     np.Id,
			Name:                   np.Name,
			ClusterId:              np.ClusterId,
			Phase:                  np.Phase,
			SchedulingCapabilities: np.SchedulingCapabilities,
		})
	}
	return &openapi.GetNodepoolsMinimal200Response{Nodepools: nodePools}, nil
}

func (smc *SimpleMock) GetNetworkTopologies(_ context.Context, filterBy []string) (*openapi.GetNetworkTopologies200Response, error) {
	topology, exists := smc.topologyMap[filterBy[0]]
	if !exists {
		return &openapi.GetNetworkTopologies200Response{NetworkTopologies: []openapi.NetworkTopology{}}, nil
	}

	return &openapi.GetNetworkTopologies200Response{NetworkTopologies: []openapi.NetworkTopology{*topology}}, nil
}

func (sm *SimpleMock) GetClusterByUuid(_ context.Context, id string) (*openapi.DisplayedCluster, *openapi.Error) {
	info := sm.clusters[id]
	if info == nil {
		return nil, openapi.NewError(http.StatusNotFound, "cluster not found")
	}

	return info, nil
}

func (sm *SimpleMock) GetInternalCluster(_ context.Context, clusterId string) (*openapi.DisplayedCluster, error) {
	info := sm.clusters[clusterId]
	if info == nil {
		return nil, fmt.Errorf("cluster %s not found", clusterId)
	}

	return info, nil
}

func (smc *SimpleMock) WithDefaultNodePool(clusterId string) *SimpleMock {
	smc.clusterNodePoolMap[clusterId] = &openapi.GetNodepoolsInternal200Response{Nodepools: []openapi.Nodepool{{Name: "default", IsDefault: true}}}
	return smc
}

func (smc *SimpleMock) WithNodePools(clusterId string, nodePoolNames ...string) *SimpleMock {
	nodePools := make([]openapi.Nodepool, 0)
	for _, name := range nodePoolNames {
		nodePools = append(nodePools, openapi.Nodepool{Name: name})
	}
	nodePool, ok := smc.clusterNodePoolMap[clusterId]
	if !ok {
		smc.clusterNodePoolMap[clusterId] = &openapi.GetNodepoolsInternal200Response{Nodepools: nodePools}
		return smc
	}
	nodePool.Nodepools = append(nodePool.Nodepools, nodePools...)
	return smc
}

func (smc *SimpleMock) WithNodePoolsWithMNNVL(clusterId string, mnnvlLabelKey string, nodePoolNames ...string) *SimpleMock {
	nodePools := make([]openapi.Nodepool, 0)
	for _, name := range nodePoolNames {
		nodePools = append(nodePools, openapi.Nodepool{
			Name: name, GpuNetworkAccelerationLabelKey: &mnnvlLabelKey,
			Status: &openapi.NodepoolSyncFieldsStatus{
				GpuNetworkAccelerationDetected: convutils.Ref(true),
			},
		},
		)
	}
	nodePool, ok := smc.clusterNodePoolMap[clusterId]
	if !ok {
		smc.clusterNodePoolMap[clusterId] = &openapi.GetNodepoolsInternal200Response{Nodepools: nodePools}
		return smc
	}
	nodePool.Nodepools = append(nodePool.Nodepools, nodePools...)
	return smc
}

func (smc *SimpleMock) WithNodePoolsWithTopology(clusterId string, topologyName string, nodePoolNames ...string) *SimpleMock {
	nodePools := make([]openapi.Nodepool, 0)
	for _, name := range nodePoolNames {
		nodePools = append(nodePools, openapi.Nodepool{
			Name:                name,
			NetworkTopologyName: *openapi.NewNullableString(&topologyName),
		})
	}

	nodePool, ok := smc.clusterNodePoolMap[clusterId]
	if !ok {
		smc.clusterNodePoolMap[clusterId] = &openapi.GetNodepoolsInternal200Response{Nodepools: nodePools}
		return smc
	}
	nodePool.Nodepools = append(nodePool.Nodepools, nodePools...)
	return smc
}

func (smc *SimpleMock) WithTopology(name string, topology *openapi.NetworkTopology) *SimpleMock {
	smc.topologyMap[fmt.Sprintf("name==%s", name)] = topology
	return smc
}

func (sm *SimpleMock) WithCluster(id string, info *openapi.DisplayedCluster) *SimpleMock {
	sm.clusters[id] = info
	return sm
}

func (smc *SimpleMock) GetNodeNetworkTopologyLabelValues(ctx context.Context, clusterUuid string, nodeName string) ([]string, error) {
	smc.getNodeNetworkTopologyLabelValuesCallCount++
	smc.getNodeNetworkTopologyLabelValuesCallParams = append(smc.getNodeNetworkTopologyLabelValuesCallParams, clusterUuid+":"+nodeName)

	cacheKey := clusterUuid + ":" + nodeName

	// Check cache first (simulates real client behavior)
	if cachedLabels, ok := smc.nodeNetworkTopologyLabelValuesCache[cacheKey]; ok {
		return cachedLabels, nil
	}

	// Cache miss - call GetNodeInternal (which tracks the API call)
	node, _, err := smc.GetNodeInternal(ctx, clusterUuid, nodeName)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, nil
	}

	// Cache the result
	smc.nodeNetworkTopologyLabelValuesCache[cacheKey] = node.NetworkTopologyLabelValues

	return node.NetworkTopologyLabelValues, nil
}

func (smc *SimpleMock) GetNodeNetworkTopologyLabelValuesCallCount() int {
	return smc.getNodeNetworkTopologyLabelValuesCallCount
}

func (smc *SimpleMock) GetNodeNetworkTopologyLabelValuesCallParams() []string {
	return smc.getNodeNetworkTopologyLabelValuesCallParams
}

func (smc *SimpleMock) GetNodeInternalCallCount() int {
	return smc.getNodeInternalCallCount
}

func (smc *SimpleMock) GetNodeInternalCallParams() []string {
	return smc.getNodeInternalCallParams
}

func (smc *SimpleMock) ResetCallCounts() *SimpleMock {
	smc.getNodeNetworkTopologyLabelValuesCallCount = 0
	smc.getNodeNetworkTopologyLabelValuesCallParams = nil
	smc.getNodeInternalCallCount = 0
	smc.getNodeInternalCallParams = nil
	return smc
}

func (smc *SimpleMock) ClearCache() *SimpleMock {
	smc.nodeNetworkTopologyLabelValuesCache = make(map[string][]string)
	return smc
}
func (smc *SimpleMock) GetNodeInternal(_ context.Context, clusterUuid string, nodeName string) (*openapi.NodeInternal, *http.Response, error) {
	smc.getNodeInternalCallCount++
	smc.getNodeInternalCallParams = append(smc.getNodeInternalCallParams, clusterUuid+":"+nodeName)

	clusterNodes, ok := smc.nodeNetworkTopologyLabelValues[clusterUuid]
	if !ok {
		return nil, nil, nil
	}
	labels, ok := clusterNodes[nodeName]
	if !ok {
		return nil, nil, nil
	}

	return &openapi.NodeInternal{
		NetworkTopologyLabelValues: labels,
	}, nil, nil
}

func (smc *SimpleMock) WithNodeNetworkTopologyLabelValues(clusterUuid string, nodeName string, labels []string) *SimpleMock {
	if smc.nodeNetworkTopologyLabelValues[clusterUuid] == nil {
		smc.nodeNetworkTopologyLabelValues[clusterUuid] = make(map[string][]string)
	}
	smc.nodeNetworkTopologyLabelValues[clusterUuid][nodeName] = labels
	return smc
}
