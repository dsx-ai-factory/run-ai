/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"

	_ "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/tests/helper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func statusUpdateKarta() *kartav1alpha1.Karta {
	return &kartav1alpha1.Karta{
		TypeMeta:   metav1.TypeMeta{APIVersion: kartav1alpha1.GroupVersion.String(), Kind: "Karta"},
		ObjectMeta: metav1.ObjectMeta{Name: "kserve", UID: "existing-karta", Labels: map[string]string{"updated-at": "old"}, Finalizers: []string{"karta.run.ai/finalize"}},
		Spec: kartav1alpha1.KartaSpec{StructureDefinition: kartav1alpha1.StructureDefinition{
			ChildComponents: []kartav1alpha1.ComponentDefinition{{Name: "predictor", StatusDefinition: &kartav1alpha1.StatusDefinition{}}},
			RootComponent: kartav1alpha1.ComponentDefinition{
				Kind: &kartav1alpha1.GroupVersionKind{Group: "serving.kserve.io", Version: "v1beta1", Kind: "InferenceService"},
				StatusDefinition: &kartav1alpha1.StatusDefinition{
					ConditionsDefinition: &kartav1alpha1.ConditionsDefinition{Path: ".status.conditions"},
					StatusMappings:       kartav1alpha1.StatusMappings{Running: []kartav1alpha1.StatusMatcher{{ByPhase: "old-running"}}},
				},
			},
		}},
	}
}

type statusUpdateExpectedGetter struct {
	WorkloadTypeKartasHandler
	expected *kartav1alpha1.Karta
}

func (g statusUpdateExpectedGetter) GetExpectedObject(_, _ string) (client.Object, error) {
	return g.expected.DeepCopy(), nil
}

var _ = Describe("Karta status mapping updates", func() {
	DescribeTable("only permits root mapping changes without recreation",
		func(change func(*kartav1alpha1.Karta), recreate bool) {
			existing := statusUpdateKarta()
			expected := existing.DeepCopy()
			expected.Spec.StructureDefinition.RootComponent.StatusDefinition.StatusMappings.Running = []kartav1alpha1.StatusMatcher{{ByPhase: "new-running"}}
			change(expected)
			beforeExisting, beforeExpected := existing.DeepCopy(), expected.DeepCopy()
			h := &WorkloadTypeKartasHandler{}
			Expect(h.ShouldReCreate(expected, existing)).To(Equal(recreate))
			Expect(existing).To(Equal(beforeExisting))
			Expect(expected).To(Equal(beforeExpected))
		},
		Entry("running mapping", func(*kartav1alpha1.Karta) {}, false),
		Entry("failed mapping", func(k *kartav1alpha1.Karta) {
			k.Spec.StructureDefinition.RootComponent.StatusDefinition.StatusMappings.Failed = []kartav1alpha1.StatusMatcher{{ByPhase: "failed"}}
		}, false),
		Entry("remove mappings", func(k *kartav1alpha1.Karta) {
			k.Spec.StructureDefinition.RootComponent.StatusDefinition.StatusMappings = kartav1alpha1.StatusMappings{}
		}, false),
		Entry("conditions path", func(k *kartav1alpha1.Karta) {
			k.Spec.StructureDefinition.RootComponent.StatusDefinition.ConditionsDefinition.Path = ".status.otherConditions"
		}, true),
		Entry("phase definition", func(k *kartav1alpha1.Karta) {
			k.Spec.StructureDefinition.RootComponent.StatusDefinition.PhaseDefinition = &kartav1alpha1.PhaseDefinition{Path: ".status.phase"}
		}, true),
		Entry("resource identity", func(k *kartav1alpha1.Karta) { k.Spec.StructureDefinition.RootComponent.Kind.Kind = "OtherService" }, true),
		Entry("root structure", func(k *kartav1alpha1.Karta) { k.Spec.StructureDefinition.RootComponent.Name = "other-root" }, true),
		Entry("additional watched kinds", func(k *kartav1alpha1.Karta) {
			k.Spec.StructureDefinition.AdditionalChildKinds = []kartav1alpha1.GroupVersionKind{{Group: "apps", Version: "v1", Kind: "Deployment"}}
		}, true),
		Entry("child definitions", func(k *kartav1alpha1.Karta) {
			k.Spec.StructureDefinition.ChildComponents = []kartav1alpha1.ComponentDefinition{{Name: "predictor"}}
		}, true),
		Entry("child status mappings", func(k *kartav1alpha1.Karta) {
			k.Spec.StructureDefinition.ChildComponents[0].StatusDefinition.StatusMappings.Running = []kartav1alpha1.StatusMatcher{{ByPhase: "ready"}}
		}, true),
		Entry("optimization instructions", func(k *kartav1alpha1.Karta) {
			k.Spec.Instructions.GangScheduling = &kartav1alpha1.GangSchedulingInstruction{}
		}, true),
		Entry("remove status definition", func(k *kartav1alpha1.Karta) { k.Spec.StructureDefinition.RootComponent.StatusDefinition = nil }, true),
	)

	It("requires recreation when adding a previously absent status definition", func() {
		expected := statusUpdateKarta()
		existing := expected.DeepCopy()
		existing.Spec.StructureDefinition.RootComponent.StatusDefinition = nil
		h := &WorkloadTypeKartasHandler{}
		Expect(h.ShouldReCreate(expected, existing)).To(BeTrue())
	})

	It("preserves defaulted condition fields without mutating the desired update", func() {
		expected := statusUpdateKarta()
		existing := expected.DeepCopy()
		existing.Spec.StructureDefinition.RootComponent.StatusDefinition.ConditionsDefinition.MessageFieldName = ptr.To("message")
		h := &WorkloadTypeKartasHandler{}
		Expect(h.ShouldReCreate(expected, existing)).To(BeFalse())
		Expect(existing.Spec.StructureDefinition.RootComponent.StatusDefinition.ConditionsDefinition.MessageFieldName).To(Equal(ptr.To("message")))
	})

	It("applies the mapping and update label together without deleting a finalized Karta", func() {
		ctx := context.Background()
		scheme := runtime.NewScheme()
		Expect(kartav1alpha1.AddToScheme(scheme)).To(Succeed())
		existing := statusUpdateKarta()
		expected := existing.DeepCopy()
		expected.Labels["updated-at"] = "new"
		expected.Spec.StructureDefinition.RootComponent.StatusDefinition.StatusMappings.Running = []kartav1alpha1.StatusMatcher{{ByPhase: "new-running"}}
		deletes := 0
		c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existing).WithInterceptorFuncs(interceptor.Funcs{
			Delete: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
				deletes++
				return c.Delete(ctx, obj, opts...)
			},
		}).Build()
		h := &CommonHandler{runaiClient: c, cachedClient: c, gvk: kartav1alpha1.GroupVersion.WithKind("Karta"),
			objectInspector: WtcKartaObjectInspector{}, expectedObjectGetter: &statusUpdateExpectedGetter{expected: expected}}
		Expect(h.handleExisting(expected, existing, "")).To(Succeed())
		stored := &kartav1alpha1.Karta{}
		Expect(c.Get(ctx, client.ObjectKeyFromObject(existing), stored)).To(Succeed())
		Expect(stored.Spec).To(Equal(expected.Spec))
		Expect(stored.Labels).To(Equal(expected.Labels))
		Expect(stored.UID).To(Equal(existing.UID))
		Expect(stored.Finalizers).To(Equal(existing.Finalizers))
		Expect(stored.DeletionTimestamp.IsZero()).To(BeTrue())
		Expect(h.handleExisting(expected, stored, "")).To(Succeed())
		Expect(deletes).To(BeZero())
	})
})
