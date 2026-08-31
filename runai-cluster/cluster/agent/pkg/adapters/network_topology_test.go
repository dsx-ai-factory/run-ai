/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package adapters

import (
	"testing"

	"k8s.io/utils/ptr"

	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/stretchr/testify/assert"
)

func TestNetworkTopologyAdapterConstructors(t *testing.T) {
	t.Run("NewNetworkTopologyAdapter", func(t *testing.T) {
		// Create a sample openapi.NetworkTopology
		openapiTopology := openapi.NetworkTopology{
			Name:   "test-topology",
			Levels: []string{"topology-hostname", "topology-rack"},
			Source: ptr.To(NetworkTopologyRunAiSource),
		}

		// Use the constructor
		adapter := NewNetworkTopologyAdapter(&openapiTopology)

		// Verify the adapter was created correctly
		assert.NotNil(t, adapter)
		assert.Equal(t, "test-topology", adapter.Name)
		assert.Equal(t, []string{"topology-hostname", "topology-rack"}, adapter.Levels)

		// Test conversion to cluster topology
		clusterTopology := adapter.ToClusterTopology()
		assert.NotNil(t, clusterTopology)
		assert.Equal(t, "test-topology", clusterTopology.Name)
		assert.Len(t, clusterTopology.Spec.Levels, 2)
		assert.Equal(t, "topology-hostname", clusterTopology.Spec.Levels[0].NodeLabel)
		assert.Equal(t, "topology-rack", clusterTopology.Spec.Levels[1].NodeLabel)
	})

	t.Run("NewNetworkTopologiesAdapter", func(t *testing.T) {
		// Create sample openapi.NetworkTopology objects
		topology1 := openapi.NetworkTopology{
			Name:   "topology-1",
			Levels: []string{"topology-hostname", "topology-rack"},
			Source: ptr.To(NetworkTopologyRunAiSource),
		}
		topology2 := openapi.NetworkTopology{
			Name:   "topology-2",
			Levels: []string{"topology-zone", "topology-block"},
			Source: ptr.To(NetworkTopologyRunAiSource),
		}
		topology3 := openapi.NetworkTopology{
			Name:   "topology-3",
			Levels: []string{"topology-datacenter", "topology-rack", "topology-hostname"},
			Source: ptr.To(NetworkTopologyRunAiSource),
		}

		networkTopologies := []openapi.NetworkTopology{topology1, topology2, topology3}

		// Use the constructor
		adapter := NewNetworkTopologiesAdapter(networkTopologies)

		// Verify the adapter was created correctly
		assert.NotNil(t, adapter)
		assert.Len(t, *adapter, 3)
		assert.Equal(t, "topology-1", (*adapter)[0].Name)
		assert.Equal(t, "topology-2", (*adapter)[1].Name)
		assert.Equal(t, "topology-3", (*adapter)[2].Name)

		// Test conversion to cluster topologies
		clusterTopologies := adapter.ToClusterTopologiesBySource()

		assert.Len(t, clusterTopologies, 3)

		// Verify first topology
		clusterTopology1 := clusterTopologies[0].(*kaiv1alpha1.Topology)
		assert.Equal(t, "topology-1", clusterTopology1.Name)
		assert.Len(t, clusterTopology1.Spec.Levels, 2)
		assert.Equal(t, "topology-hostname", clusterTopology1.Spec.Levels[0].NodeLabel)
		assert.Equal(t, "topology-rack", clusterTopology1.Spec.Levels[1].NodeLabel)

		// Verify second topology
		clusterTopology2 := clusterTopologies[1].(*kaiv1alpha1.Topology)
		assert.Equal(t, "topology-2", clusterTopology2.Name)
		assert.Len(t, clusterTopology2.Spec.Levels, 2)
		assert.Equal(t, "topology-zone", clusterTopology2.Spec.Levels[0].NodeLabel)
		assert.Equal(t, "topology-block", clusterTopology2.Spec.Levels[1].NodeLabel)

		// Verify third topology
		clusterTopology3 := clusterTopologies[2].(*kaiv1alpha1.Topology)
		assert.Equal(t, "topology-3", clusterTopology3.Name)
		assert.Len(t, clusterTopology3.Spec.Levels, 3)
		assert.Equal(t, "topology-datacenter", clusterTopology3.Spec.Levels[0].NodeLabel)
		assert.Equal(t, "topology-rack", clusterTopology3.Spec.Levels[1].NodeLabel)
		assert.Equal(t, "topology-hostname", clusterTopology3.Spec.Levels[2].NodeLabel)
	})

	t.Run("NewNetworkTopologyLevelAdapter", func(t *testing.T) {
		level := "topology-hostname"
		adapter := NewNetworkTopologyLevelAdapter(level)

		assert.NotNil(t, adapter)
		assert.Equal(t, NetworkTopologyLevelAdapter("topology-hostname"), *adapter)

		// Test conversion to cluster topology level
		clusterLevel := adapter.ToClusterTopologyLevel()
		assert.NotNil(t, clusterLevel)
		assert.Equal(t, "topology-hostname", clusterLevel.NodeLabel)
	})

	t.Run("NewNetworkTopologyLevelsAdapter", func(t *testing.T) {
		levels := []string{"topology-hostname", "topology-rack", "topology-block"}
		adapter := NewNetworkTopologyLevelsAdapter(levels)

		assert.NotNil(t, adapter)
		assert.Equal(t, NetworkTopologyLevelsAdapter(levels), *adapter)

		// Test conversion to cluster topology levels
		clusterLevels := adapter.ToClusterTopologyLevels()
		assert.Len(t, clusterLevels, 3)
		assert.Equal(t, "topology-hostname", clusterLevels[0].NodeLabel)
		assert.Equal(t, "topology-rack", clusterLevels[1].NodeLabel)
		assert.Equal(t, "topology-block", clusterLevels[2].NodeLabel)
	})
}

func TestNetworkTopologiesAdapterEdgeCases(t *testing.T) {
	t.Run("Empty slice", func(t *testing.T) {
		networkTopologies := []openapi.NetworkTopology{}
		adapter := NewNetworkTopologiesAdapter(networkTopologies)

		assert.NotNil(t, adapter)
		assert.Len(t, *adapter, 0)

		clusterTopologies := adapter.ToClusterTopologiesBySource()

		assert.Len(t, clusterTopologies, 0)
	})

	t.Run("Nil slice", func(t *testing.T) {
		var networkTopologies []openapi.NetworkTopology
		adapter := NewNetworkTopologiesAdapter(networkTopologies)

		assert.NotNil(t, adapter)
		assert.Len(t, *adapter, 0)

		clusterTopologies := adapter.ToClusterTopologiesBySource()

		assert.Len(t, clusterTopologies, 0)
	})

	t.Run("Single topology with empty levels", func(t *testing.T) {
		topology := openapi.NetworkTopology{
			Name:   "empty-levels-topology",
			Levels: []string{},
			Source: ptr.To(NetworkTopologyRunAiSource),
		}
		networkTopologies := []openapi.NetworkTopology{topology}
		adapter := NewNetworkTopologiesAdapter(networkTopologies)
		clusterTopologies := adapter.ToClusterTopologiesBySource()

		assert.Len(t, clusterTopologies, 1)

		clusterTopology := clusterTopologies[0].(*kaiv1alpha1.Topology)
		assert.Equal(t, "empty-levels-topology", clusterTopology.Name)
		assert.Len(t, clusterTopology.Spec.Levels, 0)
	})

	t.Run("Single topology with nil levels", func(t *testing.T) {
		topology := openapi.NetworkTopology{
			Name:   "nil-levels-topology",
			Levels: nil,
			Source: ptr.To(NetworkTopologyRunAiSource),
		}
		networkTopologies := []openapi.NetworkTopology{topology}
		adapter := NewNetworkTopologiesAdapter(networkTopologies)

		clusterTopologies := adapter.ToClusterTopologiesBySource()

		assert.Len(t, clusterTopologies, 1)

		clusterTopology := clusterTopologies[0].(*kaiv1alpha1.Topology)
		assert.Equal(t, "nil-levels-topology", clusterTopology.Name)
		assert.Len(t, clusterTopology.Spec.Levels, 0)
	})

	t.Run("Multiple topologies with mixed level configurations", func(t *testing.T) {
		topology1 := openapi.NetworkTopology{
			Name:   "simple-topology",
			Levels: []string{"topology-hostname"},
			Source: ptr.To(NetworkTopologyRunAiSource),
		}
		topology2 := openapi.NetworkTopology{
			Name:   "complex-topology",
			Levels: []string{"topology-zone", "topology-rack", "topology-hostname", "topology-core"},
			Source: ptr.To(NetworkTopologyRunAiSource),
		}
		topology3 := openapi.NetworkTopology{
			Name:   "empty-topology",
			Levels: []string{},
			Source: ptr.To(NetworkTopologyRunAiSource),
		}

		networkTopologies := []openapi.NetworkTopology{topology1, topology2, topology3}
		adapter := NewNetworkTopologiesAdapter(networkTopologies)

		clusterTopologies := adapter.ToClusterTopologiesBySource()

		assert.Len(t, clusterTopologies, 3)

		// Verify simple topology
		clusterTopology1 := clusterTopologies[0].(*kaiv1alpha1.Topology)
		assert.Equal(t, "simple-topology", clusterTopology1.Name)
		assert.Len(t, clusterTopology1.Spec.Levels, 1)
		assert.Equal(t, "topology-hostname", clusterTopology1.Spec.Levels[0].NodeLabel)

		// Verify complex topology
		clusterTopology2 := clusterTopologies[1].(*kaiv1alpha1.Topology)
		assert.Equal(t, "complex-topology", clusterTopology2.Name)
		assert.Len(t, clusterTopology2.Spec.Levels, 4)
		assert.Equal(t, "topology-zone", clusterTopology2.Spec.Levels[0].NodeLabel)
		assert.Equal(t, "topology-rack", clusterTopology2.Spec.Levels[1].NodeLabel)
		assert.Equal(t, "topology-hostname", clusterTopology2.Spec.Levels[2].NodeLabel)
		assert.Equal(t, "topology-core", clusterTopology2.Spec.Levels[3].NodeLabel)

		// Verify empty topology
		clusterTopology3 := clusterTopologies[2].(*kaiv1alpha1.Topology)
		assert.Equal(t, "empty-topology", clusterTopology3.Name)
		assert.Len(t, clusterTopology3.Spec.Levels, 0)
	})
}

func TestNetworkTopologyAdapterEdgeCases(t *testing.T) {
	t.Run("Empty levels", func(t *testing.T) {
		openapiTopology := openapi.NetworkTopology{
			Name:   "empty-topology",
			Levels: []string{},
		}

		adapter := NewNetworkTopologyAdapter(&openapiTopology)
		clusterTopology := adapter.ToClusterTopology()

		assert.NotNil(t, clusterTopology)
		assert.Equal(t, "empty-topology", clusterTopology.Name)
		assert.Len(t, clusterTopology.Spec.Levels, 0)
	})

	t.Run("Nil levels", func(t *testing.T) {
		openapiTopology := openapi.NetworkTopology{
			Name:   "nil-levels-topology",
			Levels: nil,
		}

		adapter := NewNetworkTopologyAdapter(&openapiTopology)
		clusterTopology := adapter.ToClusterTopology()

		assert.NotNil(t, clusterTopology)
		assert.Equal(t, "nil-levels-topology", clusterTopology.Name)
		assert.Len(t, clusterTopology.Spec.Levels, 0)
	})

	t.Run("Empty level string", func(t *testing.T) {
		level := ""
		adapter := NewNetworkTopologyLevelAdapter(level)
		clusterLevel := adapter.ToClusterTopologyLevel()

		assert.NotNil(t, clusterLevel)
		assert.Equal(t, "", clusterLevel.NodeLabel)
	})

	t.Run("Special characters in level names", func(t *testing.T) {
		levels := []string{"level-with-dashes", "level_with_underscores", "level.with.dots", "level/sublevel"}
		adapter := NewNetworkTopologyLevelsAdapter(levels)
		clusterLevels := adapter.ToClusterTopologyLevels()

		assert.Len(t, clusterLevels, 4)
		assert.Equal(t, "level-with-dashes", clusterLevels[0].NodeLabel)
		assert.Equal(t, "level_with_underscores", clusterLevels[1].NodeLabel)
		assert.Equal(t, "level.with.dots", clusterLevels[2].NodeLabel)
		assert.Equal(t, "level/sublevel", clusterLevels[3].NodeLabel)
	})
}

func TestNetworkTopologiesAdapterUsageExample(t *testing.T) {
	// Example of how to use the NetworkTopologiesAdapter in practice
	topology1 := openapi.NetworkTopology{
		Name:   "production-topology",
		Levels: []string{"topology-zone", "topology-rack", "topology-hostname"},
		Source: ptr.To(NetworkTopologyRunAiSource),
	}
	topology2 := openapi.NetworkTopology{
		Name:   "development-topology",
		Levels: []string{"topology-rack", "topology-hostname"},
		Source: ptr.To(NetworkTopologyRunAiSource),
	}
	topology3 := openapi.NetworkTopology{
		Name:   "testing-topology",
		Levels: []string{"topology-hostname"},
		Source: ptr.To(NetworkTopologyRunAiSource),
	}

	networkTopologies := []openapi.NetworkTopology{topology1, topology2, topology3}

	// Create adapter using constructor
	adapter := NewNetworkTopologiesAdapter(networkTopologies)

	// Convert to cluster topologies
	clusterTopologies := adapter.ToClusterTopologiesBySource()
	assert.Len(t, clusterTopologies, 3)

	// Verify the conversions
	productionTopology := clusterTopologies[0].(*kaiv1alpha1.Topology)
	assert.Equal(t, "production-topology", productionTopology.Name)
	assert.Len(t, productionTopology.Spec.Levels, 3)
	assert.Equal(t, "topology-zone", productionTopology.Spec.Levels[0].NodeLabel)
	assert.Equal(t, "topology-rack", productionTopology.Spec.Levels[1].NodeLabel)
	assert.Equal(t, "topology-hostname", productionTopology.Spec.Levels[2].NodeLabel)

	developmentTopology := clusterTopologies[1].(*kaiv1alpha1.Topology)
	assert.Equal(t, "development-topology", developmentTopology.Name)
	assert.Len(t, developmentTopology.Spec.Levels, 2)
	assert.Equal(t, "topology-rack", developmentTopology.Spec.Levels[0].NodeLabel)
	assert.Equal(t, "topology-hostname", developmentTopology.Spec.Levels[1].NodeLabel)

	testingTopology := clusterTopologies[2].(*kaiv1alpha1.Topology)
	assert.Equal(t, "testing-topology", testingTopology.Name)
	assert.Len(t, testingTopology.Spec.Levels, 1)
	assert.Equal(t, "topology-hostname", testingTopology.Spec.Levels[0].NodeLabel)
}

func TestNetworkTopologyAdapterUsageExample(t *testing.T) {
	// Example of how to use the adapters in practice
	openapiTopology := openapi.NetworkTopology{
		Name:   "production-topology",
		Levels: []string{"topology-zone", "topology-rack", "topology-hostname"},
	}

	// Create adapter using constructor
	adapter := NewNetworkTopologyAdapter(&openapiTopology)

	// Convert to cluster topology
	clusterTopology := adapter.ToClusterTopology()

	// Verify the conversion
	assert.Equal(t, "production-topology", clusterTopology.Name)
	assert.Len(t, clusterTopology.Spec.Levels, 3)
	assert.Equal(t, "topology-zone", clusterTopology.Spec.Levels[0].NodeLabel)
	assert.Equal(t, "topology-rack", clusterTopology.Spec.Levels[1].NodeLabel)
	assert.Equal(t, "topology-hostname", clusterTopology.Spec.Levels[2].NodeLabel)
}

func TestNetworkTopologyAdapterSourceLabel(t *testing.T) {
	t.Run("Source label is set when Source is not nil", func(t *testing.T) {
		source := "Cluster"
		openapiTopology := openapi.NetworkTopology{
			Name:   "test-topology-with-source",
			Levels: []string{"topology-hostname", "topology-rack"},
			Source: &source,
		}

		adapter := NewNetworkTopologyAdapter(&openapiTopology)
		clusterTopology := adapter.ToClusterTopology()

		assert.NotNil(t, clusterTopology)
		assert.Equal(t, "test-topology-with-source", clusterTopology.Name)
		assert.Len(t, clusterTopology.Spec.Levels, 2)

		// Verify the source label is set correctly
		assert.NotNil(t, clusterTopology.Labels)
		assert.Equal(t, source, clusterTopology.Labels[NetworkTopologySourceLabel])
		assert.Len(t, clusterTopology.Labels, 1)
	})

	t.Run("Source label is not set when Source is nil", func(t *testing.T) {
		openapiTopology := openapi.NetworkTopology{
			Name:   "test-topology-without-source",
			Levels: []string{"topology-hostname", "topology-rack"},
			Source: nil,
		}

		adapter := NewNetworkTopologyAdapter(&openapiTopology)
		clusterTopology := adapter.ToClusterTopology()

		assert.NotNil(t, clusterTopology)
		assert.Equal(t, "test-topology-without-source", clusterTopology.Name)
		assert.Len(t, clusterTopology.Spec.Levels, 2)

		// Verify no labels are set when Source is nil
		assert.Nil(t, clusterTopology.Labels)
	})

	t.Run("Source label with empty string source", func(t *testing.T) {
		source := ""
		openapiTopology := openapi.NetworkTopology{
			Name:   "test-topology-empty-source",
			Levels: []string{"topology-hostname"},
			Source: &source,
		}

		adapter := NewNetworkTopologyAdapter(&openapiTopology)
		clusterTopology := adapter.ToClusterTopology()

		assert.NotNil(t, clusterTopology)
		assert.Equal(t, "test-topology-empty-source", clusterTopology.Name)

		// Verify the source label is set even with empty string
		assert.NotNil(t, clusterTopology.Labels)
		assert.Equal(t, "", clusterTopology.Labels[NetworkTopologySourceLabel])
		assert.Len(t, clusterTopology.Labels, 1)
	})

	t.Run("Source label with special characters in source", func(t *testing.T) {
		source := "external-system.with-dashes_and_underscores"
		openapiTopology := openapi.NetworkTopology{
			Name:   "test-topology-special-chars",
			Levels: []string{"topology-zone"},
			Source: &source,
		}

		adapter := NewNetworkTopologyAdapter(&openapiTopology)
		clusterTopology := adapter.ToClusterTopology()

		assert.NotNil(t, clusterTopology)
		assert.Equal(t, "test-topology-special-chars", clusterTopology.Name)

		// Verify the source label preserves special characters
		assert.NotNil(t, clusterTopology.Labels)
		assert.Equal(t, source, clusterTopology.Labels[NetworkTopologySourceLabel])
		assert.Len(t, clusterTopology.Labels, 1)
	})

	t.Run("Source label through NetworkTopologiesAdapter", func(t *testing.T) {
		source1 := NetworkTopologyRunAiSource
		source2 := "source2"
		topology1 := openapi.NetworkTopology{
			Name:   "topology-1",
			Levels: []string{"topology-hostname"},
			Source: &source1,
		}
		topology2 := openapi.NetworkTopology{
			Name:   "topology-2",
			Levels: []string{"topology-rack"},
			Source: &source1,
		}
		topology3 := openapi.NetworkTopology{
			Name:   "topology-3",
			Levels: []string{"topology-zone"},
			Source: &source2,
		}

		networkTopologies := []openapi.NetworkTopology{topology1, topology2, topology3}
		adapter := NewNetworkTopologiesAdapter(networkTopologies)
		clusterTopologies := adapter.ToClusterTopologiesBySource()

		assert.Len(t, clusterTopologies, 2)

		// Verify first topology has source label
		clusterTopology1 := clusterTopologies[0].(*kaiv1alpha1.Topology)
		assert.NotNil(t, clusterTopology1.Labels)
		assert.Equal(t, source1, clusterTopology1.Labels[NetworkTopologySourceLabel])

		// Verify second topology has source label
		clusterTopology2 := clusterTopologies[1].(*kaiv1alpha1.Topology)
		assert.NotNil(t, clusterTopology2.Labels)
		assert.Equal(t, source1, clusterTopology2.Labels[NetworkTopologySourceLabel])

	})
}
