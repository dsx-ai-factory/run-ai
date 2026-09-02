/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvl

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Direct coverage for the kai.resources twins of the MNNVL affinity helpers. The legacy
// twins keep their own specs in affinity_test.go; asserting on kai node pools directly is
// how the accessrule-controller covers its Kai* twins, rather than flipping the
// compile-time const.
// Delete alongside the legacy twins when UseKaiNodePools is removed.
var _ = Describe("MNNVL affinity on kai node pools", func() {
	const (
		poolLabelKey   = "pool/key"
		poolLabelValue = "pool-value"
		accelLabel     = "vendor.com/clique"
	)

	// On the kai CRD the acceleration label key is an annotation, not a spec field.
	kaiPool := func(name string, accelerationLabelKey string) kaiv1alpha1.NodePool {
		pool := kaiv1alpha1.NodePool{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Spec: kaiv1alpha1.NodePoolSpec{
				LabelKey:   poolLabelKey,
				LabelValue: poolLabelValue,
			},
		}
		if accelerationLabelKey != "" {
			pool.Annotations = map[string]string{
				kaiv1alpha1.AnnotationGPUNetworkAccelerationLabelKey: accelerationLabelKey,
			}
		}
		return pool
	}

	Describe("getGPUNetworkAccelerationLabel", func() {
		It("reads the acceleration label key from the annotation", func() {
			pool := kaiPool("np1", accelLabel)

			Expect(getGPUNetworkAccelerationLabel(&pool)).To(Equal(accelLabel))
		})

		// The legacy CRD schema defaulted its spec field, so an unset annotation must
		// behave the same rather than yielding an empty key.
		It("falls back to the default when the annotation is absent", func() {
			pool := kaiPool("np1", "")

			Expect(getGPUNetworkAccelerationLabel(&pool)).
				To(Equal(kaiv1alpha1.DefaultGPUNetworkAccelerationLabelKey))
		})
	})

	Describe("NodeAffinityFromNodePools", func() {
		It("errors in None mode", func() {
			var affinity *corev1.Affinity

			err := NodeAffinityFromNodePools(
				[]kaiv1alpha1.NodePool{kaiPool("np1", accelLabel)}, &affinity, mnnvl.ModeNone)

			Expect(err).To(HaveOccurred())
		})

		It("does nothing for an empty node pool list", func() {
			var affinity *corev1.Affinity

			err := NodeAffinityFromNodePools([]kaiv1alpha1.NodePool{}, &affinity, mnnvl.ModeRequired)

			Expect(err).ToNot(HaveOccurred())
			Expect(affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution).To(BeNil())
		})

		// Required mode puts the pool selector and the acceleration requirement in the
		// same term, so a node must satisfy both.
		It("combines the pool selector and the acceleration label in Required mode", func() {
			var affinity *corev1.Affinity

			err := NodeAffinityFromNodePools(
				[]kaiv1alpha1.NodePool{kaiPool("np1", accelLabel)}, &affinity, mnnvl.ModeRequired)

			Expect(err).ToNot(HaveOccurred())
			terms := affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
			Expect(terms).To(HaveLen(1))
			Expect(terms[0].MatchExpressions).To(ConsistOf(
				corev1.NodeSelectorRequirement{
					Key:      poolLabelKey,
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{poolLabelValue},
				},
				corev1.NodeSelectorRequirement{
					Key:      accelLabel,
					Operator: corev1.NodeSelectorOpExists,
				},
			))
		})

		It("adds the acceleration label as preferred in Preferred mode", func() {
			var affinity *corev1.Affinity

			err := NodeAffinityFromNodePools(
				[]kaiv1alpha1.NodePool{kaiPool("np1", accelLabel)}, &affinity, mnnvl.ModePreferred)

			Expect(err).ToNot(HaveOccurred())
			preferred := affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution
			Expect(preferred).To(HaveLen(1))
			Expect(preferred[0].Preference.MatchExpressions).To(ConsistOf(
				corev1.NodeSelectorRequirement{Key: accelLabel, Operator: corev1.NodeSelectorOpExists},
			))
		})

		It("does not repeat a preferred term for pools sharing an acceleration label", func() {
			var affinity *corev1.Affinity

			err := NodeAffinityFromNodePools([]kaiv1alpha1.NodePool{
				kaiPool("np1", accelLabel),
				kaiPool("np2", accelLabel),
			}, &affinity, mnnvl.ModePreferred)

			Expect(err).ToNot(HaveOccurred())
			Expect(affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution).To(HaveLen(1))
		})
	})
})
