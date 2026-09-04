/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package topology

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	clusterApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

var (
	ErrTopologyNameDoesNotMatch = workloadsApi.NewError(http.StatusBadRequest, "Workload submission failed: All selected node pools must have the same topology.")
	ErrNoNodePoolsNames         = workloadsApi.NewError(http.StatusBadRequest, "no node pool names provided")
	ErrTopologyHasNoLevels      = workloadsApi.NewError(http.StatusBadRequest, "topology has no levels")
	ErrTopologyNotFound         = workloadsApi.NewError(http.StatusNotFound, "topology not found")
)

// CheckNodePoolsNames filters actualNodePools to only include those in nodePoolsNames
// and validates that all requested node pools exist
func CheckNodePoolsNames(ctx context.Context, actualNodePools []clusterApi.Nodepool, nodePoolsNames []string) ([]clusterApi.Nodepool, *workloadsApi.Error) {
	if len(nodePoolsNames) == 0 {
		return nil, ErrNoNodePoolsNames
	}

	requestedNamesSet := datastructs.NewSet(nodePoolsNames...)
	requestedPools := make([]clusterApi.Nodepool, 0, len(nodePoolsNames))

	for _, nodePool := range actualNodePools {
		if requestedNamesSet.Has(nodePool.Name) {
			requestedPools = append(requestedPools, nodePool)
			requestedNamesSet.Delete(nodePool.Name)
		}
	}

	if len(requestedNamesSet) > 0 {
		sortedKeys := requestedNamesSet.Keys()
		sort.Strings(sortedKeys)
		return nil, workloadsApi.NewError(http.StatusNotFound, fmt.Sprintf("the following node pools do not exist: %s", strings.Join(sortedKeys, ", ")))
	}

	return requestedPools, nil
}

func CheckNodePoolsTopology(ctx context.Context, actualNodePools []clusterApi.Nodepool, nodePoolsNames []string) (bool, string, *workloadsApi.Error) {
	requestedPools, err := CheckNodePoolsNames(ctx, actualNodePools, nodePoolsNames)
	if err != nil {
		return false, "", err
	}

	found := requestedPools[0].HasNetworkTopologyName()
	topologyName := requestedPools[0].GetNetworkTopologyName()

	for _, nodePool := range requestedPools {
		if topologyName != nodePool.GetNetworkTopologyName() {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("node pool name mismatch, expected %s, got %s", topologyName, nodePool.GetNetworkTopologyName())
			return false, "", ErrTopologyNameDoesNotMatch
		}
	}

	return found, topologyName, nil
}

func GetByName(ctx context.Context, client clusterservice.ClientInterface, name string) (*clusterApi.NetworkTopology, *workloadsApi.Error) {
	topologyNetwork, err := clusterservice.GetTopologyByName(ctx, client, name)
	if err != nil {
		return nil, err
	}

	if len(topologyNetwork.Levels) == 0 {
		return nil, ErrTopologyHasNoLevels
	}
	return topologyNetwork, nil
}
