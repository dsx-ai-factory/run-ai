/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepools

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNodePools(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NodePools suite")
}

// failingReader returns an error from List, to cover the error path.
type failingReader struct {
	client.Reader
}

func (failingReader) List(_ context.Context, _ client.ObjectList, _ ...client.ListOption) error {
	return errors.New("boom")
}

func newScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	Expect(kaiv1alpha1.AddToScheme(scheme)).To(Succeed())
	Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
	return scheme
}

func newReader(objs ...runtime.Object) client.Reader {
	return fake.NewClientBuilder().WithScheme(newScheme()).WithRuntimeObjects(objs...).Build()
}

func kaiNodePool(name string, phase kaiv1alpha1.NodePoolPhase) *kaiv1alpha1.NodePool {
	return &kaiv1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status:     kaiv1alpha1.NodePoolStatus{Phase: phase},
	}
}

func legacyNodePool(name string, phase v1alpha1.NodePoolPhase) *v1alpha1.NodePool {
	return &v1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status:     v1alpha1.NodePoolStatus{Phase: phase},
	}
}

var _ = Describe("NameLister", func() {
	ctx := context.Background()

	// Both read paths are exercised directly. UseKaiNodePools is a compile-time
	// constant, so going through ListSchedulableNodePoolNames would only ever reach
	// whichever branch the flag currently selects.
	Describe("kai.resources node pools", func() {
		It("returns the names of all node pools", func() {
			lister := NewNameLister(newReader(
				kaiNodePool("pool-a", ""),
				kaiNodePool("pool-b", ""),
			))

			Expect(lister.listKaiNames(ctx)).To(ConsistOf("pool-a", "pool-b"))
		})

		It("skips node pools that are being deleted", func() {
			lister := NewNameLister(newReader(
				kaiNodePool("alive", ""),
				kaiNodePool("dying", kaiv1alpha1.NodePoolDeleting),
			))

			Expect(lister.listKaiNames(ctx)).To(ConsistOf("alive"))
		})

		It("returns an empty list when there are no node pools", func() {
			Expect(NewNameLister(newReader()).listKaiNames(ctx)).To(BeEmpty())
		})

		It("propagates a list error", func() {
			_, err := NewNameLister(failingReader{}).listKaiNames(ctx)
			Expect(err).To(HaveOccurred())
		})
	})

	// TODO: delete once clusterconstants.UseKaiNodePools=true
	Describe("legacy run.ai node pools", func() {
		It("returns the names of all node pools", func() {
			lister := NewNameLister(newReader(
				legacyNodePool("pool-a", ""),
				legacyNodePool("pool-b", ""),
			))

			Expect(lister.listLegacyNames(ctx)).To(ConsistOf("pool-a", "pool-b"))
		})

		It("skips node pools that are being deleted", func() {
			lister := NewNameLister(newReader(
				legacyNodePool("alive", ""),
				legacyNodePool("dying", v1alpha1.NodePoolDeleting),
			))

			Expect(lister.listLegacyNames(ctx)).To(ConsistOf("alive"))
		})

		It("returns an empty list when there are no node pools", func() {
			Expect(NewNameLister(newReader()).listLegacyNames(ctx)).To(BeEmpty())
		})

		It("propagates a list error", func() {
			_, err := NewNameLister(failingReader{}).listLegacyNames(ctx)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ListSchedulableNodePoolNames", func() {
		It("reads the CRD selected by the UseKaiNodePools flag", func() {
			lister := NewNameLister(newReader(
				kaiNodePool("kai-pool", ""),
				legacyNodePool("legacy-pool", ""),
			))

			names, err := lister.ListSchedulableNodePoolNames(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(names).To(HaveLen(1))
		})
	})
})
