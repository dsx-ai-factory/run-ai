/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvlconv

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/plugins/mnnvl"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

func newRunner(uid string, gpuNetworkAccelerationLabels ...string) *dromosv1alpha1.WorkloadRunner {
	return &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{UID: types.UID(uid)},
		Spec: dromosv1alpha1.WorkloadRunnerSpec{
			Instructions: &dromosv1alpha1.Instructions{
				GpuOptimizations: &dromosv1alpha1.GpuOptimizations{
					Mnnvl: &dromosv1alpha1.MnnvlConfig{
						Mode:                            dromosv1alpha1.MnnvlModeRequired,
						GpuNetworkAccelerationLabelKeys: gpuNetworkAccelerationLabels,
					},
				},
			},
		},
	}
}

func newResource(annotations map[string]string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	if annotations != nil {
		u.SetAnnotations(annotations)
	}
	return u
}

var _ = Describe("ProvidePodAffinity", func() {
	// With no workloadId annotation on the runner, the selector falls back to the
	// runner UID - matching the integrator, which also falls back to the top-owner
	// UID when the annotation is absent.
	It("adds one preferred term per GPU network acceleration label, keyed on the runner UID", func() {
		pa := ProvidePodAffinity(newRunner("uid-1", "label.a", "label.b"), newResource(nil))
		Expect(pa).NotTo(BeNil())

		terms := pa.PreferredDuringSchedulingIgnoredDuringExecution
		Expect(terms).To(HaveLen(2))
		for _, term := range terms {
			Expect(term.Weight).To(Equal(mnnvl.PodAffinityWeight))
			Expect(term.PodAffinityTerm.LabelSelector.MatchLabels).
				To(HaveKeyWithValue(commonconstants.TopOwnerUidLabel, "uid-1"))
		}
	})

	// When the control plane set the workloadId annotation on the runner, the
	// selector must use that value (the integrator stamps it on the pods as
	// run.ai/top-owner-uid), NOT the runner's k8s UID.
	It("uses the workloadId annotation over the runner UID when set", func() {
		runner := newRunner("k8s-uid", "label.a")
		runner.SetAnnotations(map[string]string{
			commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: "cp-workload-99",
		})

		pa := ProvidePodAffinity(runner, newResource(nil))
		Expect(pa).NotTo(BeNil())
		Expect(pa.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector.MatchLabels).
			To(HaveKeyWithValue(commonconstants.TopOwnerUidLabel, "cp-workload-99"))
	})

	It("returns nil when the runner has no GPU network acceleration labels", func() {
		Expect(ProvidePodAffinity(newRunner("uid-1"), newResource(nil))).To(BeNil())
	})

	It("returns nil when the only label is empty", func() {
		Expect(ProvidePodAffinity(newRunner("uid-1", ""), newResource(nil))).To(BeNil())
	})

	// The provider owns the kai-skip: any KAI topology annotation means KAI owns
	// placement, so the provider returns nil even though the runner has labels.
	It("returns nil when a KAI topology annotation is present", func() {
		for _, key := range kaiTopologyAnnotationKeys {
			pa := ProvidePodAffinity(newRunner("uid-1", "label.a"), newResource(map[string]string{key: "x"}))
			Expect(pa).To(BeNil(), "annotation %q should yield nil (KAI owns placement)", key)
		}
	})
})
