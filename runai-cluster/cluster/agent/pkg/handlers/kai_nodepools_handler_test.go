/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"time"

	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	kainodepool "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"k8s.io/utils/ptr"

	_ "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/tests/helper"
)

func gpuPodsPlacementStrategy(node, device *openapi.PlacementStrategy) openapi.NullableGpuPodsPlacementStrategy {
	return *openapi.NewNullableGpuPodsPlacementStrategy(&openapi.GpuPodsPlacementStrategy{
		Node:   *openapi.NewNullablePlacementStrategy(node),
		Device: *openapi.NewNullablePlacementStrategy(device),
	})
}

func cpuPodsPlacementStrategy(node *openapi.PlacementStrategy) openapi.NullableCpuPodsPlacementStrategy {
	return *openapi.NewNullableCpuPodsPlacementStrategy(&openapi.CpuPodsPlacementStrategy{
		Node: *openapi.NewNullablePlacementStrategy(node),
	})
}

var _ = Describe("convertKaiNodePoolFields", func() {
	baseNodepool := func() *openapi.Nodepool {
		return &openapi.Nodepool{
			Name:       "test-nodepool",
			LabelKey:   "run.ai/node-pool",
			LabelValue: "test-nodepool",
		}
	}

	It("copies labelKey, labelValue and topology directly", func() {
		nodepool := baseNodepool()
		nodepool.NetworkTopologyName = *openapi.NewNullableString(ptr.To("topology-a"))

		crd, err := convertKaiNodePoolFields(nodepool)

		Expect(err).NotTo(HaveOccurred())
		Expect(crd.Name).To(Equal("test-nodepool"))
		Expect(crd.Spec.LabelKey).To(Equal("run.ai/node-pool"))
		Expect(crd.Spec.LabelValue).To(Equal("test-nodepool"))
		Expect(crd.Spec.PreferredNetworkTopologyName).To(Equal("topology-a"))
	})

	It("copies the node-level GPU/CPU placement strategy", func() {
		nodepool := baseNodepool()
		nodepool.SchedulingConfiguration = *openapi.NewNullableSchedulingConfiguration(&openapi.SchedulingConfiguration{
			GpuPodsPlacementStrategy: gpuPodsPlacementStrategy(ptr.To(openapi.PLACEMENTSTRATEGY_SPREAD), nil),
			CpuPodsPlacementStrategy: cpuPodsPlacementStrategy(ptr.To(openapi.PLACEMENTSTRATEGY_SPREAD)),
		})

		crd, err := convertKaiNodePoolFields(nodepool)

		Expect(err).NotTo(HaveOccurred())
		Expect(crd.Spec.SchedulingShardConfig).NotTo(BeNil())
		Expect(*crd.Spec.SchedulingShardConfig.PlacementStrategy.GPU).To(Equal("spread"))
		Expect(*crd.Spec.SchedulingShardConfig.PlacementStrategy.CPU).To(Equal("spread"))
	})

	It("splits minGuaranteedRuntime into preempt and reclaim min runtime", func() {
		nodepool := baseNodepool()
		nodepool.SchedulingConfiguration = *openapi.NewNullableSchedulingConfiguration(&openapi.SchedulingConfiguration{
			MinGuaranteedRuntime: *openapi.NewNullableString(ptr.To("5d8h40m")),
		})

		crd, err := convertKaiNodePoolFields(nodepool)

		Expect(err).NotTo(HaveOccurred())
		Expect(crd.Spec.SchedulingShardConfig.MinRuntime).To(Equal(&kaiv1.MinRuntime{
			PreemptMinRuntime: ptr.To("5d8h40m"),
			ReclaimMinRuntime: ptr.To("5d8h40m"),
		}))
	})

	It("leaves minRuntime unset when minGuaranteedRuntime is empty", func() {
		nodepool := baseNodepool()
		nodepool.SchedulingConfiguration = *openapi.NewNullableSchedulingConfiguration(&openapi.SchedulingConfiguration{})

		crd, err := convertKaiNodePoolFields(nodepool)

		Expect(err).NotTo(HaveOccurred())
		Expect(crd.Spec.SchedulingShardConfig.MinRuntime).To(BeNil())
	})

	It("maps timeBasedFairShare fields when enabled", func() {
		startTime := time.Date(2026, 1, 6, 8, 0, 0, 0, time.UTC)
		nodepool := baseNodepool()
		nodepool.SchedulingConfiguration = *openapi.NewNullableSchedulingConfiguration(&openapi.SchedulingConfiguration{
			TimeBasedFairShare: *openapi.NewNullableTimeBasedFairShare(&openapi.TimeBasedFairShare{
				Enabled:                 *openapi.NewNullableBool(ptr.To(true)),
				HistoricalUsageWeight:   *openapi.NewNullableFloat32(ptr.To(float32(0.5))),
				DecayHalfLife:           *openapi.NewNullableString(ptr.To("24h")),
				WindowType:              *openapi.NewNullableTimeBasedFairShareWindowType(ptr.To(openapi.TIMEBASEDFAIRSHAREWINDOWTYPE_TUMBLING)),
				WindowDuration:          *openapi.NewNullableString(ptr.To("7d")),
				TumblingWindowStartTime: *openapi.NewNullableTime(&startTime),
			}),
		})

		crd, err := convertKaiNodePoolFields(nodepool)

		Expect(err).NotTo(HaveOccurred())
		tbfs := crd.Spec.SchedulingShardConfig.TimeBasedFairShare
		Expect(tbfs).NotTo(BeNil())
		Expect(*tbfs.Enabled).To(BeTrue())
		Expect(*tbfs.HistoricalUsageWeight).To(Equal(float32(0.5)))
		Expect(*tbfs.HalfLifePeriod).To(Equal("24h"))
		Expect(*tbfs.Window.Type).To(Equal("tumbling"))
		Expect(*tbfs.Window.Size).To(Equal("7d"))
		Expect(tbfs.Window.TumblingStartTime.Time).To(Equal(startTime))
	})

	It("leaves timeBasedFairShare unset when disabled", func() {
		nodepool := baseNodepool()
		nodepool.SchedulingConfiguration = *openapi.NewNullableSchedulingConfiguration(&openapi.SchedulingConfiguration{
			TimeBasedFairShare: *openapi.NewNullableTimeBasedFairShare(&openapi.TimeBasedFairShare{
				Enabled: *openapi.NewNullableBool(ptr.To(false)),
			}),
		})

		crd, err := convertKaiNodePoolFields(nodepool)

		Expect(err).NotTo(HaveOccurred())
		Expect(crd.Spec.SchedulingShardConfig.TimeBasedFairShare).To(BeNil())
	})

	It("derives gpupack/gpuspread/gpusharingorder plugin overrides when the device strategy differs from the node strategy", func() {
		nodepool := baseNodepool()
		nodepool.SchedulingConfiguration = *openapi.NewNullableSchedulingConfiguration(&openapi.SchedulingConfiguration{
			GpuPodsPlacementStrategy: gpuPodsPlacementStrategy(
				ptr.To(openapi.PLACEMENTSTRATEGY_BINPACK), ptr.To(openapi.PLACEMENTSTRATEGY_SPREAD)),
		})

		crd, err := convertKaiNodePoolFields(nodepool)

		Expect(err).NotTo(HaveOccurred())
		plugins := crd.Spec.SchedulingShardConfig.Plugins
		Expect(*plugins["gpupack"].Enabled).To(BeFalse())
		Expect(*plugins["gpuspread"].Enabled).To(BeTrue())
		Expect(*plugins["gpusharingorder"].Enabled).To(BeFalse())
	})

	It("still sets plugins matching the node strategy when the device strategy equals it", func() {
		nodepool := baseNodepool()
		nodepool.SchedulingConfiguration = *openapi.NewNullableSchedulingConfiguration(&openapi.SchedulingConfiguration{
			GpuPodsPlacementStrategy: gpuPodsPlacementStrategy(
				ptr.To(openapi.PLACEMENTSTRATEGY_BINPACK), ptr.To(openapi.PLACEMENTSTRATEGY_BINPACK)),
		})

		crd, err := convertKaiNodePoolFields(nodepool)

		Expect(err).NotTo(HaveOccurred())
		plugins := crd.Spec.SchedulingShardConfig.Plugins
		Expect(*plugins["gpupack"].Enabled).To(BeTrue())
		Expect(*plugins["gpuspread"].Enabled).To(BeFalse())
		Expect(*plugins["gpusharingorder"].Enabled).To(BeTrue())
	})

	It("leaves plugins unset when no device strategy was requested", func() {
		nodepool := baseNodepool()
		nodepool.SchedulingConfiguration = *openapi.NewNullableSchedulingConfiguration(&openapi.SchedulingConfiguration{
			GpuPodsPlacementStrategy: gpuPodsPlacementStrategy(ptr.To(openapi.PLACEMENTSTRATEGY_BINPACK), nil),
		})

		crd, err := convertKaiNodePoolFields(nodepool)

		Expect(err).NotTo(HaveOccurred())
		Expect(crd.Spec.SchedulingShardConfig.Plugins).To(BeNil())
	})

	It("enables the numa plugin when NUMA-aware scheduling is on", func() {
		nodepool := baseNodepool()
		nodepool.SchedulingConfiguration = *openapi.NewNullableSchedulingConfiguration(&openapi.SchedulingConfiguration{
			NumaAwareScheduling: *openapi.NewNullableBool(ptr.To(true)),
		})

		crd, err := convertKaiNodePoolFields(nodepool)

		Expect(err).NotTo(HaveOccurred())
		Expect(*crd.Spec.SchedulingShardConfig.Plugins["numa"].Enabled).To(BeTrue())
	})

	It("keeps the device placement plugins when NUMA-aware scheduling is also on", func() {
		nodepool := baseNodepool()
		nodepool.SchedulingConfiguration = *openapi.NewNullableSchedulingConfiguration(&openapi.SchedulingConfiguration{
			GpuPodsPlacementStrategy: gpuPodsPlacementStrategy(
				ptr.To(openapi.PLACEMENTSTRATEGY_BINPACK), ptr.To(openapi.PLACEMENTSTRATEGY_SPREAD)),
			NumaAwareScheduling: *openapi.NewNullableBool(ptr.To(true)),
		})

		crd, err := convertKaiNodePoolFields(nodepool)

		Expect(err).NotTo(HaveOccurred())
		plugins := crd.Spec.SchedulingShardConfig.Plugins
		Expect(*plugins["numa"].Enabled).To(BeTrue())
		Expect(*plugins["gpupack"].Enabled).To(BeFalse())
		Expect(*plugins["gpuspread"].Enabled).To(BeTrue())
		Expect(*plugins["gpusharingorder"].Enabled).To(BeFalse())
	})

	It("leaves the numa plugin unset when NUMA-aware scheduling is off", func() {
		nodepool := baseNodepool()
		nodepool.SchedulingConfiguration = *openapi.NewNullableSchedulingConfiguration(&openapi.SchedulingConfiguration{
			GpuPodsPlacementStrategy: gpuPodsPlacementStrategy(
				ptr.To(openapi.PLACEMENTSTRATEGY_BINPACK), ptr.To(openapi.PLACEMENTSTRATEGY_SPREAD)),
			NumaAwareScheduling: *openapi.NewNullableBool(ptr.To(false)),
		})

		crd, err := convertKaiNodePoolFields(nodepool)

		Expect(err).NotTo(HaveOccurred())
		Expect(crd.Spec.SchedulingShardConfig.Plugins).NotTo(HaveKey("numa"))
	})

	It("moves GPU network acceleration detection and label key to annotations", func() {
		nodepool := baseNodepool()
		nodepool.GpuNetworkAccelerationDetection = *openapi.NewNullableNodepoolGPUNetworkAccelerationDetection(
			ptr.To(openapi.NODEPOOLGPUNETWORKACCELERATIONDETECTION_USE))
		nodepool.GpuNetworkAccelerationLabelKey = ptr.To("nvidia.com/gpu.clique")

		crd, err := convertKaiNodePoolFields(nodepool)

		Expect(err).NotTo(HaveOccurred())
		Expect(crd.Annotations[kainodepool.AnnotationGPUNetworkAccelerationDetection]).To(Equal("Use"))
		Expect(crd.Annotations[kainodepool.AnnotationGPUNetworkAccelerationLabelKey]).To(Equal("nvidia.com/gpu.clique"))
	})
})

var _ = Describe("KaiNodePoolObjectInspector", func() {
	var inspector KaiNodePoolObjectInspector

	Describe("DeepCopySpec", func() {
		It("copies spec and annotations", func() {
			from := &kainodepool.NodePool{
				Spec: kainodepool.NodePoolSpec{LabelKey: "k", LabelValue: "v"},
			}
			from.Annotations = map[string]string{"a": "b"}
			to := &kainodepool.NodePool{}

			inspector.DeepCopySpec(from, to)

			Expect(to.Spec.LabelKey).To(Equal("k"))
			Expect(to.Spec.LabelValue).To(Equal("v"))
			Expect(to.Annotations).To(Equal(map[string]string{"a": "b"}))
		})
	})

	Describe("IsUpToDate", func() {
		It("returns true for identical spec and annotations", func() {
			left := &kainodepool.NodePool{Spec: kainodepool.NodePoolSpec{LabelKey: "k"}}
			left.Annotations = map[string]string{"a": "b"}
			right := &kainodepool.NodePool{Spec: kainodepool.NodePoolSpec{LabelKey: "k"}}
			right.Annotations = map[string]string{"a": "b"}

			Expect(inspector.IsUpToDate(left, right)).To(BeTrue())
		})

		It("returns false when annotations differ", func() {
			left := &kainodepool.NodePool{Spec: kainodepool.NodePoolSpec{LabelKey: "k"}}
			left.Annotations = map[string]string{"a": "b"}
			right := &kainodepool.NodePool{Spec: kainodepool.NodePoolSpec{LabelKey: "k"}}
			right.Annotations = map[string]string{"a": "c"}

			Expect(inspector.IsUpToDate(left, right)).To(BeFalse())
		})
	})
})
