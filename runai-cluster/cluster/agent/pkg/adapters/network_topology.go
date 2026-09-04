/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package adapters

import (
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	NetworkTopologyRunAiSource = "Run.ai"
	NetworkTopologySourceLabel = "run.ai/network-topology-source"
)

type NetworkTopologyAdapter openapi.NetworkTopology
type NetworkTopologiesAdapter []openapi.NetworkTopology
type NetworkTopologyLevelAdapter string
type NetworkTopologyLevelsAdapter []string

// NewNetworkTopologyAdapter creates a new NetworkTopologyAdapter from an openapi.NetworkTopology
func NewNetworkTopologyAdapter(networkTopology *openapi.NetworkTopology) *NetworkTopologyAdapter {
	adapter := NetworkTopologyAdapter(*networkTopology)
	return &adapter
}

// NewNetworkTopologiesAdapter creates a new NetworkTopologyLevelsAdapter from a slice of strings
func NewNetworkTopologiesAdapter(networkTopologies []openapi.NetworkTopology) *NetworkTopologiesAdapter {
	adapter := NetworkTopologiesAdapter(networkTopologies)
	return &adapter
}

// NewNetworkTopologyLevelAdapter creates a new NetworkTopologyLevelAdapter from a string
func NewNetworkTopologyLevelAdapter(level string) *NetworkTopologyLevelAdapter {
	adapter := NetworkTopologyLevelAdapter(level)
	return &adapter
}

// NewNetworkTopologyLevelsAdapter creates a new NetworkTopologyLevelsAdapter from a slice of strings
func NewNetworkTopologyLevelsAdapter(levels []string) *NetworkTopologyLevelsAdapter {
	adapter := NetworkTopologyLevelsAdapter(levels)
	return &adapter
}

func (nta *NetworkTopologyAdapter) ToClusterTopology() *kaiv1alpha1.Topology {
	topologyCr := &kaiv1alpha1.Topology{}
	levelsAdapter := NewNetworkTopologyLevelsAdapter(nta.Levels)
	topologyCr.Spec = kaiv1alpha1.TopologySpec{
		Levels: levelsAdapter.ToClusterTopologyLevels(),
	}
	if nta.Source != nil {
		topologyCr.Labels = map[string]string{NetworkTopologySourceLabel: *nta.Source}
	}
	topologyCr.Name = nta.Name
	return topologyCr
}

func (ntas *NetworkTopologiesAdapter) ToClusterTopologiesBySource() []client.Object {
	clusterTopologies := []client.Object{}

	for _, networkTopology := range *ntas {
		if networkTopology.Source != nil && *networkTopology.Source == NetworkTopologyRunAiSource {
			clusterTopology := NewNetworkTopologyAdapter(&networkTopology).ToClusterTopology()
			clusterTopologies = append(clusterTopologies, clusterTopology)
		}
	}

	return clusterTopologies
}

func (ntla *NetworkTopologyLevelAdapter) ToClusterTopologyLevel() *kaiv1alpha1.TopologyLevel {
	return &kaiv1alpha1.TopologyLevel{
		NodeLabel: string(*ntla),
	}
}

func (ntla *NetworkTopologyLevelsAdapter) ToClusterTopologyLevels() []kaiv1alpha1.TopologyLevel {
	topologyLevels := make([]kaiv1alpha1.TopologyLevel, 0, len(*ntla))
	for _, level := range *ntla {
		levelAdapter := NewNetworkTopologyLevelAdapter(level)
		topologyLevels = append(topologyLevels, *levelAdapter.ToClusterTopologyLevel())
	}
	return topologyLevels
}
