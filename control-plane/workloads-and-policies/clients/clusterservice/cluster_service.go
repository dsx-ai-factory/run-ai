package clusterservice

import (
	"context"
	"fmt"
	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"net/http"
)

var (
	errTopologyNotFound = workloadsApi.NewError(http.StatusNotFound, "topology not found")
	errClientIsNil      = workloadsApi.NewError(http.StatusInternalServerError, "no client provided")
)

func GetNodePoolsByNames(ctx context.Context, client ClientInterface, clusterId string, nodePoolsNames []string) ([]openapi.Nodepool, *workloadsApi.Error) {
	if client == nil {
		return nil, errClientIsNil
	}

	nodePools := make([]openapi.Nodepool, 0)
	apiResponse, err := client.GetNodePoolsInternalByCluster(ctx, clusterId)

	if err != nil {
		return nil, workloadsApi.NewError(http.StatusServiceUnavailable, err.Error())
	}

	for _, nodePool := range apiResponse.GetNodepools() {
		for _, nodePoolName := range nodePoolsNames {
			if nodePoolName == nodePool.GetName() {
				nodePools = append(nodePools, nodePool)
			}
		}
	}

	return nodePools, nil
}

func GetNodePoolsMinimalByNames(ctx context.Context, client ClientInterface, clusterId string, nodePoolsNames []string) ([]openapi.NodepoolMinimal, *workloadsApi.Error) {
	if client == nil {
		return nil, errClientIsNil
	}

	nodePools := make([]openapi.NodepoolMinimal, 0)
	apiResponse, err := client.GetNodePoolsMinimalByCluster(ctx, clusterId)
	if err != nil {
		return nil, workloadsApi.NewError(http.StatusServiceUnavailable, err.Error())
	}

	for _, nodePool := range apiResponse.GetNodepools() {
		for _, nodePoolName := range nodePoolsNames {
			if nodePoolName == nodePool.GetName() {
				nodePools = append(nodePools, nodePool)
			}
		}
	}

	return nodePools, nil
}

func GetTopologyByName(ctx context.Context, client ClientInterface, name string) (*openapi.NetworkTopology, *workloadsApi.Error) {
	if client == nil {
		return nil, errClientIsNil
	}

	topology, err := client.GetNetworkTopologies(ctx, []string{fmt.Sprintf("name==%s", name)})
	if err != nil {
		return nil, workloadsApi.NewError(http.StatusServiceUnavailable, err.Error())
	}

	if topology == nil || len(topology.GetNetworkTopologies()) == 0 {
		return nil, errTopologyNotFound
	}

	return &topology.NetworkTopologies[0], nil
}
