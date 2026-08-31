/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	grovev1alpha1 "github.com/ai-dynamo/grove/operator/api/core/v1alpha1"

	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"k8s.io/utils/ptr"
)

type TopologyAdapter kaiv1alpha1.Topology
type TopologiesAdapter []kaiv1alpha1.Topology
type TopologyNodeLevelAdapter kaiv1alpha1.TopologyLevel
type TopologyNodeLevelsAdapter []kaiv1alpha1.TopologyLevel
type ClusterTopologyAdapter grovev1alpha1.ClusterTopologyBinding

const (
	NetworkTopologiesClusterSource = "Cluster"
	NetworkTopologiesRunAISource   = "Run.ai"

	NetworkTopologySourceLabel = "run.ai/network-topology-source"
)

// NewTopologyAdapter creates a new TopologyAdapter from a kaiv1alpha1.Topology
func NewTopologyAdapter(topology *kaiv1alpha1.Topology) *TopologyAdapter {
	adapter := TopologyAdapter(*topology)
	return &adapter
}

// NewTopologyNodeLevelAdapter creates a new TopologyNodeLevelAdapter from a kaiv1alpha1.TopologyLevel
func NewTopologyNodeLevelAdapter(topologyLevel *kaiv1alpha1.TopologyLevel) *TopologyNodeLevelAdapter {
	adapter := TopologyNodeLevelAdapter(*topologyLevel)
	return &adapter
}

// NewTopologyNodeLevelsAdapter creates a new TopologyNodeLevelsAdapter from a slice of kaiv1alpha1.TopologyLevel
func NewTopologyNodeLevelsAdapter(topologyLevels []kaiv1alpha1.TopologyLevel) *TopologyNodeLevelsAdapter {
	adapter := TopologyNodeLevelsAdapter(topologyLevels)
	return &adapter
}

// NewTopologiesAdapter creates a new TopologiesAdapter from a slice of kaiv1alpha1.Topology
func NewTopologiesAdapter(topologies []kaiv1alpha1.Topology) *TopologiesAdapter {
	adapter := TopologiesAdapter(topologies)
	return &adapter
}

// NewClusterTopologyAdapter creates a new ClusterTopologyAdapter from a grovev1alpha1.ClusterTopologyBinding
func NewClusterTopologyAdapter(clusterTopology *grovev1alpha1.ClusterTopologyBinding) *ClusterTopologyAdapter {
	adapter := ClusterTopologyAdapter(*clusterTopology)
	return &adapter
}

// ToNetworkTopologySyncGrove converts a ClusterTopology to NetworkTopologySyncGrove
func (cta *ClusterTopologyAdapter) ToNetworkTopologySyncGrove() openapi.NetworkTopologySyncGrove {
	clusterTopology := grovev1alpha1.ClusterTopologyBinding(*cta)

	domains := make([]string, 0, len(clusterTopology.Spec.Levels))
	for _, level := range clusterTopology.Spec.Levels {
		domains = append(domains, string(level.Domain))
	}

	conditions := make([]openapi.NetworkTopologySyncGroveCondition, 0, len(clusterTopology.Status.Conditions))
	for _, c := range clusterTopology.Status.Conditions {
		conditions = append(conditions, openapi.NetworkTopologySyncGroveCondition{
			Type:               ptr.To(c.Type),
			Status:             ptr.To(string(c.Status)),
			Reason:             ptr.To(c.Reason),
			Message:            ptr.To(c.Message),
			ObservedGeneration: ptr.To(c.ObservedGeneration),
		})
	}

	return openapi.NetworkTopologySyncGrove{
		Name:       clusterTopology.Name,
		Domains:    domains,
		Conditions: conditions,
	}
}

// ToNetworkTopologySyncFields converts a Topology to NetworkTopologySyncFields
// We set Cluster Label for Source - when no Source label found
func (nta *TopologyAdapter) ToNetworkTopologySyncFields(groveTopology *grovev1alpha1.ClusterTopologyBinding) *openapi.NetworkTopologySyncFields {
	// Cast back to the original type to access fields
	topology := kaiv1alpha1.Topology(*nta)
	levelsAdapter := NewTopologyNodeLevelsAdapter(topology.Spec.Levels)
	source, ok := topology.Labels[NetworkTopologySourceLabel]
	if !ok {
		source = NetworkTopologiesClusterSource
	}
	syncFields := &openapi.NetworkTopologySyncFields{
		Name:   topology.Name,
		Source: source,
		Levels: levelsAdapter.ToTopologyLevelSyncFields(),
	}

	if groveTopology != nil {
		groveAdapter := NewClusterTopologyAdapter(groveTopology)
		syncFields.SetGrove(groveAdapter.ToNetworkTopologySyncGrove())
	}

	return syncFields
}

// ToNetworkTopologiesSyncFields converts a slice of Topology to a slice of NetworkTopologySyncFields
func (ntas *TopologiesAdapter) ToNetworkTopologiesSyncFields(groveTopologies map[string]*grovev1alpha1.ClusterTopologyBinding) []openapi.NetworkTopologySyncFields {
	topologies := make([]openapi.NetworkTopologySyncFields, 0, len(*ntas))

	for _, topology := range *ntas {
		var groveTopology *grovev1alpha1.ClusterTopologyBinding
		if groveTopologies != nil {
			groveTopology = groveTopologies[topology.Name]
		}
		topologySync := NewTopologyAdapter(&topology).
			ToNetworkTopologySyncFields(groveTopology)
		topologies = append(topologies, *topologySync)
	}

	return topologies
}

// ToTopologyLevelSyncFields converts TopologyNodeLevelsAdapter to TopologyLevelSyncField slice
func (ntla *TopologyNodeLevelsAdapter) ToTopologyLevelSyncFields() []string {
	levels := make([]string, 0, len(*ntla))

	for _, level := range *ntla {
		levelAdapter := NewTopologyNodeLevelAdapter(&level)
		levelSync := levelAdapter.ToTopologyLevelSyncField()
		levels = append(levels, *levelSync)
	}

	return levels
}

// ToTopologyLevelSyncField converts a TopologyNodeLevelAdapter to TopologyLevelSyncField
func (ntla *TopologyNodeLevelAdapter) ToTopologyLevelSyncField() *string {
	return ptr.To(ntla.ToLevel())
}

// ToLevel converts a TopologyLevel to its node label representation
func (ntla *TopologyNodeLevelAdapter) ToLevel() string {
	level := kaiv1alpha1.TopologyLevel(*ntla)
	return level.NodeLabel
}
