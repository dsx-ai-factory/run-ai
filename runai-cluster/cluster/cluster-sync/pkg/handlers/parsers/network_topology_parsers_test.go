/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	grovev1alpha1 "github.com/ai-dynamo/grove/operator/api/core/v1alpha1"
	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	RegionNodeLabelWest = "region=us-west"
	ZoneNodeLabelWest1  = "zone=us-west-1"
	Rack1NodeLabel      = "rack=rack-1"
	Node1NodeLabel      = "node=node-1"

	// Test data constants
	TestTopologyName = "test-topology"
	Topology1Name    = "topology-1"
	Topology2Name    = "topology-2"

	// Additional test topology labels
	ZoneNodeLabelEast1 = "zone=us-east-1"
	Rack2NodeLabel     = "rack=rack-2"

	// Grove topology test constants
	GroveTopologyName = "grove-topology"
)

var _ = Describe("NetworkTopologyParsers", func() {
	Describe("TopologyAdapter", func() {
		var (
			topology *kaiv1alpha1.Topology
			adapter  *TopologyAdapter
		)

		BeforeEach(func() {
			topology = &kaiv1alpha1.Topology{
				ObjectMeta: metav1.ObjectMeta{
					Name: TestTopologyName,
					Labels: map[string]string{
						NetworkTopologySourceLabel: NetworkTopologiesRunAISource,
					},
				},
				Spec: kaiv1alpha1.TopologySpec{
					Levels: []kaiv1alpha1.TopologyLevel{
						{NodeLabel: ZoneNodeLabelWest1},
						{NodeLabel: Rack1NodeLabel},
					},
				},
			}
			adapter = NewTopologyAdapter(topology)
		})

		Describe("NewTopologyAdapter", func() {
			It("should create a new TopologyAdapter", func() {
				Expect(adapter).ToNot(BeNil())
			})

			It("should properly wrap the topology", func() {
				Expect(adapter).ToNot(BeNil())
			})
		})

		Describe("ToNetworkTopologySyncFields", func() {
			It("should convert topology to NetworkTopologySyncFields", func() {
				result := adapter.ToNetworkTopologySyncFields(nil)
				Expect(result).ToNot(BeNil())
				Expect(result.Name).To(Equal(TestTopologyName))
				Expect(result.Source).To(Equal(NetworkTopologiesRunAISource))
				Expect(result.Levels).To(HaveLen(2))
				Expect(result.Levels).To(ContainElements(ZoneNodeLabelWest1, Rack1NodeLabel))
			})

			It("should use default source when label is missing", func() {
				delete(topology.Labels, NetworkTopologySourceLabel)
				adapter = NewTopologyAdapter(topology)

				result := adapter.ToNetworkTopologySyncFields(nil)

				Expect(result.Source).To(Equal(NetworkTopologiesClusterSource))
			})

			It("should handle empty levels", func() {
				topology.Spec.Levels = []kaiv1alpha1.TopologyLevel{}
				adapter = NewTopologyAdapter(topology)

				result := adapter.ToNetworkTopologySyncFields(nil)

				Expect(result.Levels).To(HaveLen(0))
			})

			Context("with grove topology", func() {
				var groveTopology *grovev1alpha1.ClusterTopologyBinding

				BeforeEach(func() {
					groveTopology = &grovev1alpha1.ClusterTopologyBinding{
						ObjectMeta: metav1.ObjectMeta{
							Name: GroveTopologyName,
						},
						Spec: grovev1alpha1.ClusterTopologyBindingSpec{
							Levels: []grovev1alpha1.TopologyLevel{
								{Domain: grovev1alpha1.TopologyDomainRegion, Key: "topology.kubernetes.io/region"},
								{Domain: grovev1alpha1.TopologyDomainZone, Key: "topology.kubernetes.io/zone"},
								{Domain: grovev1alpha1.TopologyDomainRack, Key: "rack-label"},
							},
						},
					}
				})

				It("should populate grove field when groveTopology is provided", func() {
					result := adapter.ToNetworkTopologySyncFields(groveTopology)

					Expect(result).ToNot(BeNil())
					Expect(result.Name).To(Equal(TestTopologyName))
					Expect(result.Source).To(Equal(NetworkTopologiesRunAISource))
					Expect(result.Levels).To(HaveLen(2))

					grove, ok := result.GetGroveOk()
					Expect(ok).To(BeTrue())
					Expect(grove).ToNot(BeNil())
					Expect(grove.Name).To(Equal(GroveTopologyName))
				})

				It("should not populate grove field when groveTopology is nil", func() {
					result := adapter.ToNetworkTopologySyncFields(nil)

					Expect(result).ToNot(BeNil())
					grove := result.Grove
					Expect(grove.IsSet()).To(BeFalse())
				})

				It("should populate grove field with empty spec", func() {
					groveTopology.Spec = grovev1alpha1.ClusterTopologyBindingSpec{}

					result := adapter.ToNetworkTopologySyncFields(groveTopology)

					grove, ok := result.GetGroveOk()
					Expect(ok).To(BeTrue())
					Expect(grove).ToNot(BeNil())
					Expect(grove.Name).To(Equal(GroveTopologyName))
				})

				It("should populate grove domains from spec levels", func() {
					result := adapter.ToNetworkTopologySyncFields(groveTopology)

					grove, ok := result.GetGroveOk()
					Expect(ok).To(BeTrue())
					Expect(grove.Domains).To(HaveLen(3))
					Expect(grove.Domains).To(ContainElements(
						string(grovev1alpha1.TopologyDomainRegion),
						string(grovev1alpha1.TopologyDomainZone),
						string(grovev1alpha1.TopologyDomainRack),
					))
				})

				It("should populate grove conditions from status", func() {
					groveTopology.Status = grovev1alpha1.ClusterTopologyBindingStatus{
						ObservedGeneration: 7,
						Conditions: []metav1.Condition{
							{
								Type:               "Ready",
								Status:             metav1.ConditionTrue,
								Reason:             "Synced",
								Message:            "topology is in sync",
								ObservedGeneration: 7,
							},
							{
								Type:               "SchedulerTopologyDrift",
								Status:             metav1.ConditionFalse,
								Reason:             "NoDrift",
								Message:            "scheduler backend topology matches",
								ObservedGeneration: 7,
							},
						},
					}

					result := adapter.ToNetworkTopologySyncFields(groveTopology)

					grove, ok := result.GetGroveOk()
					Expect(ok).To(BeTrue())
					Expect(grove.Conditions).To(HaveLen(2))

					Expect(*grove.Conditions[0].Type).To(Equal("Ready"))
					Expect(*grove.Conditions[0].Status).To(Equal(string(metav1.ConditionTrue)))
					Expect(*grove.Conditions[0].Reason).To(Equal("Synced"))
					Expect(*grove.Conditions[0].Message).To(Equal("topology is in sync"))
					Expect(*grove.Conditions[0].ObservedGeneration).To(Equal(int64(7)))

					Expect(*grove.Conditions[1].Type).To(Equal("SchedulerTopologyDrift"))
					Expect(*grove.Conditions[1].Status).To(Equal(string(metav1.ConditionFalse)))
				})

				It("should produce empty conditions when status has none", func() {
					result := adapter.ToNetworkTopologySyncFields(groveTopology)

					grove, ok := result.GetGroveOk()
					Expect(ok).To(BeTrue())
					Expect(grove.Conditions).To(BeEmpty())
				})
			})
		})
	})

	Describe("TopologiesAdapter", func() {
		var (
			topologies []kaiv1alpha1.Topology
			adapter    *TopologiesAdapter
		)

		BeforeEach(func() {
			topologies = []kaiv1alpha1.Topology{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: Topology1Name,
						Labels: map[string]string{
							NetworkTopologySourceLabel: NetworkTopologiesClusterSource,
						},
					},
					Spec: kaiv1alpha1.TopologySpec{
						Levels: []kaiv1alpha1.TopologyLevel{
							{NodeLabel: ZoneNodeLabelWest1},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: Topology2Name,
						Labels: map[string]string{
							NetworkTopologySourceLabel: NetworkTopologiesClusterSource,
						},
					},
					Spec: kaiv1alpha1.TopologySpec{
						Levels: []kaiv1alpha1.TopologyLevel{
							{NodeLabel: ZoneNodeLabelEast1},
							{NodeLabel: Rack2NodeLabel},
						},
					},
				},
			}
			adapter = NewTopologiesAdapter(topologies)
		})

		Describe("NewTopologiesAdapter", func() {
			It("should create a new TopologiesAdapter", func() {
				Expect(adapter).ToNot(BeNil())
			})
		})

		Describe("ToNetworkTopologiesSyncFields", func() {
			It("should convert multiple topologies to NetworkTopologySyncFields", func() {
				result := adapter.ToNetworkTopologiesSyncFields(nil)

				Expect(result).To(HaveLen(2))

				// Check first topology
				Expect(result[0].Name).To(Equal(Topology1Name))
				Expect(result[0].Source).To(Equal(NetworkTopologiesClusterSource))
				Expect(result[0].Levels).To(HaveLen(1))
				Expect(result[0].Levels).To(ContainElement(ZoneNodeLabelWest1))

				// Check second topology
				Expect(result[1].Name).To(Equal(Topology2Name))
				Expect(result[1].Source).To(Equal(NetworkTopologiesClusterSource))
				Expect(result[1].Levels).To(HaveLen(2))
				Expect(result[1].Levels).To(ContainElements(ZoneNodeLabelEast1, Rack2NodeLabel))
			})

			It("should handle empty topologies slice", func() {
				emptyAdapter := NewTopologiesAdapter([]kaiv1alpha1.Topology{})
				result := emptyAdapter.ToNetworkTopologiesSyncFields(nil)

				Expect(result).To(HaveLen(0))
			})

			It("should match grove topologies by name", func() {
				groveTopologies := map[string]*grovev1alpha1.ClusterTopologyBinding{
					Topology1Name: {
						ObjectMeta: metav1.ObjectMeta{Name: Topology1Name},
						Spec: grovev1alpha1.ClusterTopologyBindingSpec{
							Levels: []grovev1alpha1.TopologyLevel{
								{Domain: grovev1alpha1.TopologyDomainZone, Key: "topology.kubernetes.io/zone"},
							},
						},
					},
				}

				result := adapter.ToNetworkTopologiesSyncFields(groveTopologies)

				Expect(result).To(HaveLen(2))

				// First topology should have grove set
				grove1, ok := result[0].GetGroveOk()
				Expect(ok).To(BeTrue())
				Expect(grove1).ToNot(BeNil())
				Expect(grove1.Name).To(Equal(Topology1Name))

				// Second topology should not have grove set (no matching grove topology)
				Expect(result[1].Grove.IsSet()).To(BeFalse())
			})

			It("should handle grove topologies map with no matches", func() {
				groveTopologies := map[string]*grovev1alpha1.ClusterTopologyBinding{
					"non-existent": {
						ObjectMeta: metav1.ObjectMeta{Name: "non-existent"},
					},
				}

				result := adapter.ToNetworkTopologiesSyncFields(groveTopologies)

				Expect(result).To(HaveLen(2))
				Expect(result[0].Grove.IsSet()).To(BeFalse())
				Expect(result[1].Grove.IsSet()).To(BeFalse())
			})
		})
	})

	Describe("TopologyNodeLevelAdapter", func() {
		var (
			level   *kaiv1alpha1.TopologyLevel
			adapter *TopologyNodeLevelAdapter
		)

		BeforeEach(func() {
			level = &kaiv1alpha1.TopologyLevel{
				NodeLabel: ZoneNodeLabelWest1,
			}
			adapter = NewTopologyNodeLevelAdapter(level)
		})

		Describe("NewTopologyNodeLevelAdapter", func() {
			It("should create a new TopologyNodeLevelAdapter", func() {
				Expect(adapter).ToNot(BeNil())
			})
		})

		Describe("ToLevel", func() {
			It("should return the node label", func() {
				result := adapter.ToLevel()
				Expect(result).To(Equal(ZoneNodeLabelWest1))
			})
		})

		Describe("ToTopologyLevelSyncField", func() {
			It("should convert to string pointer", func() {
				result := adapter.ToTopologyLevelSyncField()

				Expect(result).ToNot(BeNil())
				Expect(*result).To(Equal(ZoneNodeLabelWest1))
			})
		})
	})

	Describe("TopologyNodeLevelsAdapter", func() {
		var (
			levels  []kaiv1alpha1.TopologyLevel
			adapter *TopologyNodeLevelsAdapter
		)

		BeforeEach(func() {
			levels = []kaiv1alpha1.TopologyLevel{
				{NodeLabel: ZoneNodeLabelWest1},
				{NodeLabel: Rack1NodeLabel},
				{NodeLabel: Node1NodeLabel},
			}
			adapter = NewTopologyNodeLevelsAdapter(levels)
		})

		Describe("NewTopologyNodeLevelsAdapter", func() {
			It("should create a new TopologyNodeLevelsAdapter", func() {
				Expect(adapter).ToNot(BeNil())
			})
		})

		Describe("ToTopologyLevelSyncFields", func() {
			It("should convert levels to string slice", func() {
				result := adapter.ToTopologyLevelSyncFields()

				Expect(result).To(HaveLen(3))
				Expect(result).To(ContainElements(ZoneNodeLabelWest1, Rack1NodeLabel, Node1NodeLabel))
			})

			It("should handle empty levels", func() {
				emptyAdapter := NewTopologyNodeLevelsAdapter([]kaiv1alpha1.TopologyLevel{})
				result := emptyAdapter.ToTopologyLevelSyncFields()

				Expect(result).To(HaveLen(0))
			})

			It("should handle single level", func() {
				singleLevel := []kaiv1alpha1.TopologyLevel{
					{NodeLabel: ZoneNodeLabelWest1},
				}
				singleAdapter := NewTopologyNodeLevelsAdapter(singleLevel)
				result := singleAdapter.ToTopologyLevelSyncFields()

				Expect(result).To(HaveLen(1))
				Expect(result[0]).To(Equal(ZoneNodeLabelWest1))
			})
		})
	})

	Describe("Integration Tests", func() {
		It("should handle complete topology conversion flow", func() {
			// Create a complex topology
			topology := &kaiv1alpha1.Topology{
				ObjectMeta: metav1.ObjectMeta{
					Name: "complex-topology",
					Labels: map[string]string{
						NetworkTopologySourceLabel: NetworkTopologiesRunAISource,
					},
				},
				Spec: kaiv1alpha1.TopologySpec{
					Levels: []kaiv1alpha1.TopologyLevel{
						{NodeLabel: RegionNodeLabelWest},
						{NodeLabel: ZoneNodeLabelWest1},
						{NodeLabel: Rack1NodeLabel},
						{NodeLabel: Node1NodeLabel},
					},
				},
			}

			// Convert through the entire flow
			topologyAdapter := NewTopologyAdapter(topology)
			syncFields := topologyAdapter.ToNetworkTopologySyncFields(nil)

			// Verify the result
			Expect(syncFields.Name).To(Equal("complex-topology"))
			Expect(syncFields.Source).To(Equal(NetworkTopologiesRunAISource))
			Expect(syncFields.Levels).To(HaveLen(4))
			Expect(syncFields.Levels).To(ContainElements(
				RegionNodeLabelWest,
				ZoneNodeLabelWest1,
				Rack1NodeLabel,
				Node1NodeLabel,
			))
		})

		It("should handle multiple topologies with different sources", func() {
			topologies := []kaiv1alpha1.Topology{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "topology-a",
						Labels: map[string]string{
							NetworkTopologySourceLabel: NetworkTopologiesRunAISource,
						},
					},
					Spec: kaiv1alpha1.TopologySpec{
						Levels: []kaiv1alpha1.TopologyLevel{
							{NodeLabel: "zone=zone-a"},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "topology-b",
						// No source label - should use Cluster
					},
					Spec: kaiv1alpha1.TopologySpec{
						Levels: []kaiv1alpha1.TopologyLevel{
							{NodeLabel: "zone=zone-b"},
						},
					},
				},
			}

			topologiesAdapter := NewTopologiesAdapter(topologies)
			syncFields := topologiesAdapter.ToNetworkTopologiesSyncFields(nil)

			Expect(syncFields).To(HaveLen(2))
			Expect(syncFields[0].Source).To(Equal(NetworkTopologiesRunAISource))
			Expect(syncFields[1].Source).To(Equal(NetworkTopologiesClusterSource))
		})
	})

	Describe("Edge Cases", func() {
		It("should handle topology with nil labels", func() {
			topology := &kaiv1alpha1.Topology{
				ObjectMeta: metav1.ObjectMeta{
					Name: "no-labels-topology",
					// No labels map
				},
				Spec: kaiv1alpha1.TopologySpec{
					Levels: []kaiv1alpha1.TopologyLevel{
						{NodeLabel: ZoneNodeLabelWest1},
					},
				},
			}

			adapter := NewTopologyAdapter(topology)
			syncFields := adapter.ToNetworkTopologySyncFields(nil)

			Expect(syncFields.Source).To(Equal(NetworkTopologiesClusterSource))
		})

		It("should handle topology with empty labels", func() {
			topology := &kaiv1alpha1.Topology{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "empty-labels-topology",
					Labels: map[string]string{},
				},
				Spec: kaiv1alpha1.TopologySpec{
					Levels: []kaiv1alpha1.TopologyLevel{
						{NodeLabel: ZoneNodeLabelWest1},
					},
				},
			}

			adapter := NewTopologyAdapter(topology)
			syncFields := adapter.ToNetworkTopologySyncFields(nil)

			Expect(syncFields.Source).To(Equal(NetworkTopologiesClusterSource))
		})
	})
})
