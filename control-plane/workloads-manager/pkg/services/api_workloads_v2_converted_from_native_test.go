/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
)

var _ = Describe("isImmutableV1CastWorkload", func() {
	var k8sUtil *K8sResourceUtil

	BeforeEach(func() {
		k8sUtil = NewK8sResourceUtil()
	})

	buildManifest := func(labels map[string]string) map[string]any {
		labelMap := map[string]any{}
		for k, v := range labels {
			labelMap[k] = v
		}
		return map[string]any{
			"apiVersion": "batch/v1",
			"kind":       "Job",
			"metadata": map[string]any{
				"name":   "test-workload",
				"labels": labelMap,
			},
		}
	}

	DescribeTable("workloadKind allowlist (translatedFromV1=true)",
		func(labels map[string]string, expectImmutable bool, expectWorkloadKind string) {
			manifest := buildManifest(labels)
			immutable, workloadKind := isImmutableV1CastWorkload(true, manifest, k8sUtil)
			Expect(immutable).To(Equal(expectImmutable))
			Expect(workloadKind).To(Equal(expectWorkloadKind))
		},
		Entry("Workspace -> immutable",
			map[string]string{constants.LabelWorkloadKind: "Workspace"},
			true, "Workspace"),
		Entry("Training -> immutable",
			map[string]string{constants.LabelWorkloadKind: "Training"},
			true, "Training"),
		Entry("Distributed -> immutable",
			map[string]string{constants.LabelWorkloadKind: "Distributed"},
			true, "Distributed"),
		Entry("Inference -> mutable (not in allowlist)",
			map[string]string{constants.LabelWorkloadKind: "Inference"},
			false, ""),
		Entry("DistributedInference -> mutable (not in allowlist)",
			map[string]string{constants.LabelWorkloadKind: "DistributedInference"},
			false, ""),
		Entry("missing workloadKind label -> mutable",
			map[string]string{},
			false, ""),
	)

	It("returns false when translatedFromV1 is false regardless of labels", func() {
		manifest := buildManifest(map[string]string{constants.LabelWorkloadKind: "Training"})
		immutable, workloadKind := isImmutableV1CastWorkload(false, manifest, k8sUtil)
		Expect(immutable).To(BeFalse())
		Expect(workloadKind).To(BeEmpty())
	})

	It("returns false for nil manifest", func() {
		immutable, workloadKind := isImmutableV1CastWorkload(true, nil, k8sUtil)
		Expect(immutable).To(BeFalse())
		Expect(workloadKind).To(BeEmpty())
	})

	It("returns false when manifest has no metadata", func() {
		manifest := map[string]any{
			"apiVersion": "batch/v1",
			"kind":       "Job",
		}
		immutable, _ := isImmutableV1CastWorkload(true, manifest, k8sUtil)
		Expect(immutable).To(BeFalse())
	})
})

var _ = Describe("isConvertedFromNativeWorkload", func() {
	DescribeTable("label detection",
		func(labels map[string]string, expected bool) {
			Expect(isConvertedFromNativeWorkload(labels)).To(Equal(expected))
		},
		Entry("converted-from-native=true -> true",
			map[string]string{ConvertedFromNativeLabel: "true"}, true),
		Entry("converted-from-native=true with workloadKind -> true",
			map[string]string{ConvertedFromNativeLabel: "true", constants.LabelWorkloadKind: "Inference"}, true),
		Entry("converted-from-native=false -> false",
			map[string]string{ConvertedFromNativeLabel: "false"}, false),
		Entry("label absent -> false",
			map[string]string{constants.LabelWorkloadKind: "Training"}, false),
		Entry("nil labels -> false", nil, false),
	)
})

var _ = Describe("nativeWorkloadTypeFromV1Cast", func() {
	buildManifest := func(labels map[string]any) map[string]any {
		return map[string]any{
			"apiVersion": "batch/v1",
			"kind":       "Job",
			"metadata": map[string]any{
				"name":   "test-workload",
				"labels": labels,
			},
		}
	}

	It("returns the original native workload type for converted manifests", func() {
		manifest := buildManifest(map[string]any{
			ConvertedFromNativeLabel:    "true",
			constants.LabelWorkloadKind: "Training",
		})

		workloadType, ok := nativeWorkloadTypeFromV1Cast(true, manifest)

		Expect(ok).To(BeTrue())
		Expect(workloadType).To(Equal("Training"))
	})

	It("returns Distributed for converted distributed manifests", func() {
		manifest := buildManifest(map[string]any{
			ConvertedFromNativeLabel:    "true",
			constants.LabelWorkloadKind: "Distributed",
		})

		workloadType, ok := nativeWorkloadTypeFromV1Cast(true, manifest)

		Expect(ok).To(BeTrue())
		Expect(workloadType).To(Equal("Distributed"))
	})

	It("returns false when the workload was not translated from v1", func() {
		manifest := buildManifest(map[string]any{
			ConvertedFromNativeLabel:    "true",
			constants.LabelWorkloadKind: "Training",
		})

		_, ok := nativeWorkloadTypeFromV1Cast(false, manifest)

		Expect(ok).To(BeFalse())
	})
})

var _ = Describe("distributedFrameworkFromKind", func() {
	DescribeTable("maps Kubeflow job kinds to distributed frameworks",
		func(kind string, expected string) {
			Expect(distributedFrameworkFromKind(kind)).To(Equal(expected))
		},
		Entry("PyTorchJob", "PyTorchJob", "PyTorch"),
		Entry("MPIJob", "MPIJob", "MPI"),
		Entry("TFJob", "TFJob", "TF"),
		Entry("XGBoostJob", "XGBoostJob", "XGBoost"),
		Entry("JAXJob", "JAXJob", "JAX"),
		Entry("unknown", "Job", ""),
	)
})
