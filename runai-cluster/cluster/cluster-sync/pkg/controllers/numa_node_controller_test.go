/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache/informertest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type numaNodeFakeManager struct {
	ctrl.Manager
	cache cache.Cache
}

func (m *numaNodeFakeManager) GetCache() cache.Cache {
	return m.cache
}

var _ = Describe("NRT CRD lifecycle", func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		Expect(apiextensionsv1.AddToScheme(scheme)).To(Succeed())
	})

	Context("when checking whether the CRD is installed", func() {
		It("returns true only for an established NRT CRD", func() {
			crd := &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{Name: nrtCRDName},
				Status: apiextensionsv1.CustomResourceDefinitionStatus{
					Conditions: []apiextensionsv1.CustomResourceDefinitionCondition{{
						Type:   apiextensionsv1.Established,
						Status: apiextensionsv1.ConditionTrue,
					}},
				},
			}
			reader := fake.NewClientBuilder().WithScheme(scheme).WithObjects(crd).Build()

			Expect(IsNRTCRDInstalled(ctx, reader)).To(BeTrue())
		})

		It("returns false when the NRT CRD does not exist", func() {
			reader := fake.NewClientBuilder().WithScheme(scheme).Build()

			Expect(IsNRTCRDInstalled(ctx, reader)).To(BeFalse())
		})

		It("returns false while the NRT CRD is not established", func() {
			crd := &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{Name: nrtCRDName},
			}
			reader := fake.NewClientBuilder().WithScheme(scheme).WithObjects(crd).Build()

			Expect(IsNRTCRDInstalled(ctx, reader)).To(BeFalse())
		})
	})

	Context("when watching for a missing CRD", func() {
		It("restarts only after the NRT CRD is installed", func() {
			fakeCache := &informertest.FakeInformers{Scheme: scheme}
			manager := &numaNodeFakeManager{cache: fakeCache}

			exitCode := -1
			originalExitProcess := exitProcess
			DeferCleanup(func() {
				exitProcess = originalExitProcess
			})
			exitProcess = func(code int) {
				exitCode = code
			}

			Expect(WatchForNRTCRD(ctx, manager)).To(Succeed())

			crdGVK := apiextensionsv1.SchemeGroupVersion.WithKind("CustomResourceDefinition")
			crdInformer, err := fakeCache.FakeInformerForKind(ctx, crdGVK)
			Expect(err).NotTo(HaveOccurred())

			otherCRD := &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{Name: "other.example.com"},
			}
			crdInformer.Add(otherCRD)
			establishedOtherCRD := otherCRD.DeepCopy()
			establishedOtherCRD.Status.Conditions = []apiextensionsv1.CustomResourceDefinitionCondition{{
				Type:   apiextensionsv1.Established,
				Status: apiextensionsv1.ConditionTrue,
			}}
			crdInformer.Update(otherCRD, establishedOtherCRD)
			Expect(exitCode).To(Equal(-1))

			nrtCRD := &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{Name: nrtCRDName},
			}
			crdInformer.Add(nrtCRD)
			Expect(exitCode).To(Equal(-1))

			establishedNRTCRD := nrtCRD.DeepCopy()
			establishedNRTCRD.Status.Conditions = []apiextensionsv1.CustomResourceDefinitionCondition{{
				Type:   apiextensionsv1.Established,
				Status: apiextensionsv1.ConditionTrue,
			}}
			crdInformer.Update(nrtCRD, establishedNRTCRD)
			Expect(exitCode).To(Equal(0))
		})
	})
})
