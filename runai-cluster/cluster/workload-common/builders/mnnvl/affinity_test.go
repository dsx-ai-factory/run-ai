/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvl

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("MNNVL Affinity Functions", func() {

	Describe("AddPodAffinity", func() {
		var (
			podAffinity *corev1.PodAffinity
			workloadID  string
			nodePools   []v1alpha1.NodePool
		)

		BeforeEach(func() {
			podAffinity = &corev1.PodAffinity{}
			workloadID = "test-workload-123"

			// Setup test node pools with different MNNVL labels
			nodePools = []v1alpha1.NodePool{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "nodepool-1",
					},
					Spec: v1alpha1.NodePoolSpec{
						GPUNetworkAccelerationLabelKey: "gpu-network-acc-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "nodepool-2",
					},
					Spec: v1alpha1.NodePoolSpec{
						GPUNetworkAccelerationLabelKey: "gpu-network-acc-2",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "nodepool-3-duplicate-label",
					},
					Spec: v1alpha1.NodePoolSpec{
						GPUNetworkAccelerationLabelKey: "gpu-network-acc-1", // Duplicate label
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "nodepool-4-empty-label",
					},
					Spec: v1alpha1.NodePoolSpec{
						GPUNetworkAccelerationLabelKey: "", // Empty label
					},
				},
			}
		})

		Context("when podAffinity is nil", func() {
			BeforeEach(func() {
				podAffinity = nil
			})

			It("should initialize podAffinity", func() {
				AddPodAffinity(podAffinity, getGPUNetworkAccelerationLabels(nodePools), workloadID)
				// Since we pass a nil podAffinity, it gets initialized inside the function
				// but this change is not reflected in our variable as Go is pass-by-value
				Expect(podAffinity).To(BeNil())
			})
		})

		It("should add preferred pod affinity for each unique label", func() {
			AddPodAffinity(podAffinity, getGPUNetworkAccelerationLabels(nodePools), workloadID)
			Expect(podAffinity.PreferredDuringSchedulingIgnoredDuringExecution).NotTo(BeNil())

			// Should only have 2 terms (for the 2 unique non-empty labels)
			Expect(podAffinity.PreferredDuringSchedulingIgnoredDuringExecution).To(HaveLen(2))

			// Verify terms have correct topology keys
			topologyKeys := []string{}
			for _, term := range podAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
				topologyKeys = append(topologyKeys, term.PodAffinityTerm.TopologyKey)

				// Check label selector
				Expect(term.PodAffinityTerm.LabelSelector).NotTo(BeNil())
				Expect(term.PodAffinityTerm.LabelSelector.MatchLabels).To(HaveKeyWithValue(commonconstants.TopOwnerUidLabel, workloadID))
			}

			// Check that we have the expected topology keys
			Expect(topologyKeys).To(ConsistOf("gpu-network-acc-1", "gpu-network-acc-2"))
		})

		It("should append to existing preferred pod affinity", func() {
			// Setup existing preferred affinity
			podAffinity.PreferredDuringSchedulingIgnoredDuringExecution = []corev1.WeightedPodAffinityTerm{
				{
					Weight: 100,
					PodAffinityTerm: corev1.PodAffinityTerm{
						TopologyKey: "existing-key",
					},
				},
			}

			AddPodAffinity(podAffinity, getGPUNetworkAccelerationLabels(nodePools), workloadID)
			Expect(podAffinity.PreferredDuringSchedulingIgnoredDuringExecution).To(HaveLen(3))
		})

		It("should use the correct default weight", func() {
			AddPodAffinity(podAffinity, getGPUNetworkAccelerationLabels(nodePools), workloadID)
			for _, term := range podAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
				Expect(term.Weight).To(Equal(DefaultAffinityWeight))
			}
		})

		It("should handle empty node pool list", func() {
			AddPodAffinity(podAffinity, getGPUNetworkAccelerationLabels([]v1alpha1.NodePool{}), workloadID)
			Expect(podAffinity.PreferredDuringSchedulingIgnoredDuringExecution).To(BeEmpty())
		})
	})

	Describe("NodeAffinityFromNodePools", func() {
		var (
			affinity  *corev1.Affinity
			nodePools []v1alpha1.NodePool
		)

		BeforeEach(func() {
			affinity = nil

			// Setup test node pools
			nodePools = []v1alpha1.NodePool{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "nodepool-1",
					},
					Spec: v1alpha1.NodePoolSpec{
						LabelKey:                       "nodepool-key-1",
						LabelValue:                     "nodepool-value-1",
						GPUNetworkAccelerationLabelKey: "gpu-network-acc-1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: v1alpha1.DefaultNodePoolName,
					},
					Spec: v1alpha1.NodePoolSpec{
						LabelKey:                       v1alpha1.NodePoolNameLabel,
						GPUNetworkAccelerationLabelKey: "gpu-network-acc-default",
					},
				},
			}
		})

		Context("with invalid inputs", func() {
			It("should not return error for empty node pool list", func() {
				err := NodeAffinityFromLegacyNodePools([]v1alpha1.NodePool{}, &affinity, mnnvl.ModeRequired)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return error for ModeNone", func() {
				err := NodeAffinityFromLegacyNodePools(nodePools, &affinity, mnnvl.ModeNone)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not supported in None mode"))
			})
		})

		Context("with ModeRequired", func() {
			It("should initialize affinity if nil", func() {
				err := NodeAffinityFromLegacyNodePools(nodePools, &affinity, mnnvl.ModeRequired)
				Expect(err).NotTo(HaveOccurred())
				Expect(affinity).NotTo(BeNil())
				Expect(affinity.NodeAffinity).NotTo(BeNil())
			})

			It("should create node selector terms with both nodepool and MNNVL requirements", func() {
				err := NodeAffinityFromLegacyNodePools(nodePools, &affinity, mnnvl.ModeRequired)
				Expect(err).NotTo(HaveOccurred())

				terms := affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
				Expect(terms).To(HaveLen(2)) // One per nodepool

				// Check first nodepool (regular)
				Expect(terms[0].MatchExpressions).To(HaveLen(2)) // One for nodepool, one for MNNVL

				// Check second nodepool (default)
				Expect(terms[1].MatchExpressions).To(HaveLen(2)) // One for nodepool, one for MNNVL

				// Verify the MNNVL requirement is added with Exists operator
				mnnvlFound := false
				for _, expr := range terms[0].MatchExpressions {
					if expr.Key == "gpu-network-acc-1" && expr.Operator == corev1.NodeSelectorOpExists {
						mnnvlFound = true
						break
					}
				}
				Expect(mnnvlFound).To(BeTrue())
			})

			It("should handle default nodepool correctly", func() {
				err := NodeAffinityFromLegacyNodePools(nodePools, &affinity, mnnvl.ModeRequired)
				Expect(err).NotTo(HaveOccurred())

				terms := affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
				Expect(terms).NotTo(BeEmpty())

				// Find the default nodepool term
				var defaultTerm *corev1.NodeSelectorTerm
				for i, term := range terms {
					for _, expr := range term.MatchExpressions {
						if expr.Key == v1alpha1.NodePoolNameLabel {
							defaultTerm = &terms[i]
							break
						}
					}
					if defaultTerm != nil {
						break
					}
				}

				Expect(defaultTerm).NotTo(BeNil())

				// Verify the MNNVL requirement for default nodepool
				mnnvlFound := false
				for _, expr := range defaultTerm.MatchExpressions {
					if expr.Key == "gpu-network-acc-default" && expr.Operator == corev1.NodeSelectorOpExists {
						mnnvlFound = true
						break
					}
				}
				Expect(mnnvlFound).To(BeTrue())
			})
		})

		Context("with ModePreferred", func() {
			It("should add regular nodepool affinity and preferred MNNVL affinity", func() {
				err := NodeAffinityFromLegacyNodePools(nodePools, &affinity, mnnvl.ModePreferred)
				Expect(err).NotTo(HaveOccurred())

				// Check required node affinity for nodepools
				Expect(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution).NotTo(BeNil())

				// Check preferred affinity for MNNVL
				preferredTerms := affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution
				Expect(preferredTerms).NotTo(BeEmpty())

				// Should have one preferred term per unique MNNVL label
				uniqueLabels := make(map[string]bool)
				for _, term := range preferredTerms {
					for _, expr := range term.Preference.MatchExpressions {
						if expr.Operator == corev1.NodeSelectorOpExists {
							uniqueLabels[expr.Key] = true
						}
					}
				}

				// We should have 2 unique labels from our test data
				Expect(len(uniqueLabels)).To(Equal(2))
				Expect(uniqueLabels).To(HaveKey("gpu-network-acc-1"))
				Expect(uniqueLabels).To(HaveKey("gpu-network-acc-default"))
			})
		})
	})
})

func getGPUNetworkAccelerationLabels(nodePools []v1alpha1.NodePool) []string {
	return lo.Map(nodePools, func(nodePool v1alpha1.NodePool, _ int) string {
		return getLegacyGPUNetworkAccelerationLabel(&nodePool)
	})
}
