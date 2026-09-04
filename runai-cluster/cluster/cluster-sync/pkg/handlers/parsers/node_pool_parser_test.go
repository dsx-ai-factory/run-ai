/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Direct coverage for the kai.resources adapters. The legacy adapters keep their own specs
// in node_pool_legacy_parser_test.go; asserting on kai node pools directly is how the
// accessrule-controller covers its Kai* twins, rather than flipping the compile-time const.

func testNodePool(name string, annotations map[string]string) *kaiv1alpha1.NodePool {
	return &kaiv1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{Name: name, Annotations: annotations},
		Status: kaiv1alpha1.NodePoolStatus{
			Phase:   kaiv1alpha1.NodePoolReady,
			Message: "all good",
			Nodes: []kaiv1alpha1.NodeInNodePool{
				{Name: "node1"},
				{Name: "node2"},
			},
		},
	}
}

var _ = Describe("NodePoolParsers", func() {
	Describe("NodepoolAdapter", func() {
		Describe("ToServerScheme", func() {
			It("should convert a node pool to NodepoolSyncFields", func() {
				nodePool := testNodePool("np-1", map[string]string{
					kaiv1alpha1.AnnotationGPUNetworkAccelerationDetected: "true",
				})

				synced, err := NewNodepoolAdapter(nodePool).ToServerScheme()

				Expect(err).ToNot(HaveOccurred())
				Expect(synced.Name).To(Equal("np-1"))
				Expect(*synced.Phase).To(Equal(openapi.NODEPOOLPHASE_READY))
				Expect(*synced.PhaseMessage).To(Equal("all good"))
				Expect(synced.Status.Nodes).To(Equal([]string{"node1", "node2"}))
				Expect(*synced.Status.GpuNetworkAccelerationDetected).To(BeTrue())
			})

			// The detection result moved from status to a controller-written annotation. An
			// absent or unparseable value must read as not detected rather than defaulting
			// to true, which would place workloads on unsupported nodes.
			DescribeTable("should read the acceleration detection annotation",
				func(annotations map[string]string, expected bool) {
					synced, err := NewNodepoolAdapter(testNodePool("np-1", annotations)).ToServerScheme()

					Expect(err).ToNot(HaveOccurred())
					Expect(*synced.Status.GpuNetworkAccelerationDetected).To(Equal(expected))
				},
				Entry("detected",
					map[string]string{kaiv1alpha1.AnnotationGPUNetworkAccelerationDetected: "true"}, true),
				Entry("not detected",
					map[string]string{kaiv1alpha1.AnnotationGPUNetworkAccelerationDetected: "false"}, false),
				Entry("empty value",
					map[string]string{kaiv1alpha1.AnnotationGPUNetworkAccelerationDetected: ""}, false),
				Entry("unparseable value",
					map[string]string{kaiv1alpha1.AnnotationGPUNetworkAccelerationDetected: "yes-please"}, false),
				Entry("no annotations", map[string]string(nil), false),
			)
		})

		// minGuaranteedRuntime moved under schedulingShardConfig.minRuntime, which splits it
		// into a preempt and a reclaim value; the migration writes the legacy value to both,
		// and the legacy field is defined in terms of preemption, so preempt is read back.
		Describe("scheduling configuration", func() {
			It("should report the preempt min runtime", func() {
				nodePool := testNodePool("np-1", nil)
				nodePool.Spec.SchedulingShardConfig = &kaiv1alpha1.SchedulingShardConfig{
					MinRuntime: &kaiv1.MinRuntime{
						PreemptMinRuntime: ptr.To("5d8h40m"),
						ReclaimMinRuntime: ptr.To("5d8h40m"),
					},
				}

				synced, err := NewNodepoolAdapter(nodePool).ToServerScheme()

				Expect(err).ToNot(HaveOccurred())
				Expect(synced.Status.SchedulingConfiguration.IsSet()).To(BeTrue())
				Expect(*synced.Status.SchedulingConfiguration.Get().MinGuaranteedRuntime.Get()).
					To(Equal("5d8h40m"))
			})

			DescribeTable("should leave it unset",
				func(mutate func(*kaiv1alpha1.NodePool)) {
					nodePool := testNodePool("np-1", nil)
					mutate(nodePool)

					synced, err := NewNodepoolAdapter(nodePool).ToServerScheme()

					Expect(err).ToNot(HaveOccurred())
					Expect(synced.Status.SchedulingConfiguration.IsSet()).To(BeFalse())
				},
				Entry("when there is no shard config", func(np *kaiv1alpha1.NodePool) {
					np.Spec.SchedulingShardConfig = nil
				}),
				Entry("when there is no min runtime", func(np *kaiv1alpha1.NodePool) {
					np.Spec.SchedulingShardConfig = &kaiv1alpha1.SchedulingShardConfig{}
				}),
				Entry("when there is no preempt min runtime", func(np *kaiv1alpha1.NodePool) {
					np.Spec.SchedulingShardConfig = &kaiv1alpha1.SchedulingShardConfig{
						MinRuntime: &kaiv1.MinRuntime{},
					}
				}),
			)
		})

		// The topology-mismatch condition reports the nodes that carry the mismatch, so the
		// control plane can name them.
		Describe("conditions", func() {
			It("should report only the nodes carrying the topology mismatch", func() {
				nodePool := testNodePool("np-1", nil)
				nodePool.Status.Nodes = []kaiv1alpha1.NodeInNodePool{
					{Name: "node1", TopologyMismatch: true},
					{Name: "node2", TopologyMismatch: false},
				}
				nodePool.Status.Conditions = []kaiv1alpha1.NodePoolCondition{{
					Type:    kaiv1alpha1.NodeTopologyMismatch,
					Status:  corev1.ConditionTrue,
					Reason:  "NodeTopologyMismatchReason",
					Message: "mismatch",
				}}

				synced, err := NewNodepoolAdapter(nodePool).ToServerScheme()

				Expect(err).ToNot(HaveOccurred())
				Expect(synced.Status.Conditions).To(HaveLen(1))
				Expect(synced.Status.Conditions[0].Type).
					To(Equal(string(kaiv1alpha1.NodeTopologyMismatch)))
				Expect(synced.Status.Conditions[0].Nodes).To(Equal([]string{"node1"}))
			})
		})

		// The nodepool-controller is the only writer of the NRT prerequisite phase and
		// condition, so this path matters more on the kai CRD than on the legacy one. It
		// also covers the non-topology-mismatch branch of the condition adapter, where the
		// related-nodes list stays empty.
		Describe("missing NRT prerequisites", func() {
			const phaseMessage = "The following prerequisites are not met on one or more " +
				"nodes in the node pool: nrt-healthy"

			It("should convert the phase and the condition", func() {
				nodePool := testNodePool("np-1", nil)
				nodePool.Status.Phase = kaiv1alpha1.NodePoolMissingPrerequisites
				nodePool.Status.Message = phaseMessage
				nodePool.Status.Conditions = []kaiv1alpha1.NodePoolCondition{{
					Type:    kaiv1alpha1.NodePoolMissingNrtHealthyPrerequisite,
					Status:  corev1.ConditionTrue,
					Reason:  kaiv1alpha1.NodePoolMissingNrtHealthyPrerequisiteReason,
					Message: phaseMessage,
				}}

				synced, err := NewNodepoolAdapter(nodePool).ToServerScheme()

				Expect(err).ToNot(HaveOccurred())
				Expect(*synced.Phase).To(Equal(openapi.NODEPOOLPHASE_MISSING_PREREQUISITES))
				Expect(*synced.PhaseMessage).To(Equal(phaseMessage))
				Expect(synced.Status.Conditions).To(HaveLen(1))

				condition := synced.Status.Conditions[0]
				Expect(condition.Type).
					To(Equal(string(kaiv1alpha1.NodePoolMissingNrtHealthyPrerequisite)))
				Expect(condition.Reason).
					To(Equal(kaiv1alpha1.NodePoolMissingNrtHealthyPrerequisiteReason))
				Expect(*condition.Message).To(Equal(phaseMessage))
				Expect(*condition.Status).To(Equal("True"))
				Expect(condition.Nodes).To(BeEmpty())
			})
		})

		Describe("ToDeleteRequest", func() {
			It("should create a delete request with the correct phase", func() {
				deleted := NewNodepoolAdapter(testNodePool("np-1", nil)).ToDeleteRequest()

				Expect(deleted.Name).To(Equal("np-1"))
				Expect(*deleted.Phase).To(Equal(openapi.NODEPOOLPHASE_DELETED))
			})
		})
	})

	// The list adapter is what a full resync goes through, so it gets its own coverage
	// rather than relying on the single-node-pool path.
	Describe("NodepoolsAdapter", func() {
		Describe("ToServerScheme", func() {
			It("should convert every node pool, preserving order", func() {
				first := *testNodePool("np-1", map[string]string{
					kaiv1alpha1.AnnotationGPUNetworkAccelerationDetected: "true",
				})
				second := *testNodePool("np-2", nil)

				synced, err := NewNodepoolsAdapter([]kaiv1alpha1.NodePool{first, second}).ToServerScheme()

				Expect(err).ToNot(HaveOccurred())
				Expect(synced).To(HaveLen(2))
				Expect(synced[0].Name).To(Equal("np-1"))
				Expect(*synced[0].Status.GpuNetworkAccelerationDetected).To(BeTrue())
				Expect(synced[1].Name).To(Equal("np-2"))
				Expect(*synced[1].Status.GpuNetworkAccelerationDetected).To(BeFalse())
			})

			It("should handle an empty slice", func() {
				synced, err := NewNodepoolsAdapter([]kaiv1alpha1.NodePool{}).ToServerScheme()

				Expect(err).ToNot(HaveOccurred())
				Expect(synced).To(BeEmpty())
			})
		})
	})
})
