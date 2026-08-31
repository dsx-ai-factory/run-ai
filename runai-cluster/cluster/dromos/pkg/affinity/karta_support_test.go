/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package affinity

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/utils/ptr"
)

func componentWith(spec *kartav1alpha1.SpecDefinition) kartav1alpha1.ComponentDefinition {
	return kartav1alpha1.ComponentDefinition{Name: "component", SpecDefinition: spec}
}

func kartaWith(root kartav1alpha1.ComponentDefinition, children ...kartav1alpha1.ComponentDefinition) *kartav1alpha1.Karta {
	return &kartav1alpha1.Karta{Spec: kartav1alpha1.KartaSpec{
		StructureDefinition: kartav1alpha1.StructureDefinition{RootComponent: root, ChildComponents: children},
	}}
}

var _ = Describe("SupportsNodeAffinity", func() {
	It("accepts a component addressed by pod template spec", func() {
		karta := kartaWith(componentWith(&kartav1alpha1.SpecDefinition{PodTemplateSpecPath: ptr.To("spec.template")}))

		Expect(SupportsNodeAffinity(karta)).To(Succeed())
	})

	It("accepts a component addressed by pod spec", func() {
		karta := kartaWith(componentWith(&kartav1alpha1.SpecDefinition{PodSpecPath: ptr.To("spec")}))

		Expect(SupportsNodeAffinity(karta)).To(Succeed())
	})

	It("accepts a fragmented component that names the node-affinity path", func() {
		karta := kartaWith(componentWith(&kartav1alpha1.SpecDefinition{
			FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{NodeAffinityPath: ptr.To("spec.affinity.nodeAffinity")},
		}))

		Expect(SupportsNodeAffinity(karta)).To(Succeed())
	})

	It("rejects a fragmented component with no node-affinity path", func() {
		karta := kartaWith(componentWith(&kartav1alpha1.SpecDefinition{
			FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{},
		}))

		Expect(SupportsNodeAffinity(karta)).To(MatchError(ErrNoNodeAffinityPath))
	})

	It("rejects a karta with no pod-bearing component", func() {
		Expect(SupportsNodeAffinity(kartaWith(componentWith(nil)))).To(MatchError(ErrNoSpecDefinition))
	})

	It("rejects when a child component cannot carry node affinity", func() {
		karta := kartaWith(
			componentWith(&kartav1alpha1.SpecDefinition{PodTemplateSpecPath: ptr.To("spec.template")}),
			componentWith(&kartav1alpha1.SpecDefinition{FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{}}),
		)

		Expect(SupportsNodeAffinity(karta)).To(MatchError(ErrNoNodeAffinityPath))
	})
})
