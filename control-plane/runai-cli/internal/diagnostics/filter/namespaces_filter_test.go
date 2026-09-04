/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package filter

import (
	"context"
	"fmt"

	filtermocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/diagnostics/filter"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("NamespacesFilter", func() {
	Describe("NewNamespacesFilter", func() {
		It("should return a NamespacesFilter", func() {
			nsFilterClientMock := filtermocks.NewMockNamespacesFilterClienter(nil)

			nsFilter := NewNamespacesFilter(nsFilterClientMock, []string{"ns1"})
			Expect(nsFilter).NotTo(BeNil())
			Expect(nsFilter.namespaces).To(Equal([]string{"ns1"}))
			Expect(nsFilter.nsFilterClient).To(Equal(nsFilterClientMock))
		})

		It("should return a NamespacesFilter with a unique list of namespaces", func() {
			nsFilter := NewNamespacesFilter(nil, []string{"ns1", "ns2", "ns1"})
			Expect(nsFilter.namespaces).To(Equal([]string{"ns1", "ns2"}))
		})
	})

	Describe("Filter", func() {
		var mockCtrl *gomock.Controller
		var nsFilterClientMock *filtermocks.MockNamespacesFilterClienter
		var ctx context.Context

		BeforeEach(func() {
			ctx = context.Background()
			mockCtrl = gomock.NewController(GinkgoT())
			nsFilterClientMock = filtermocks.NewMockNamespacesFilterClienter(mockCtrl)
		})

		AfterEach(func() {
			mockCtrl.Finish()
		})

		It("should return the correct FilterResult when all namespaces are valid", func() {
			nsFilter := NewNamespacesFilter(nsFilterClientMock, []string{"ns1", "ns2"})

			nsFilterClientMock.EXPECT().ListNamespaces(ctx).Return(getNamespaceList([]string{"ns1", "ns2"}), nil)

			nsFilterClientMock.EXPECT().VerifyAccess(ctx, listPodsSAR("ns1")).Return(true, nil)
			nsFilterClientMock.EXPECT().VerifyAccess(ctx, podLogsSAR("ns1")).Return(true, nil)

			nsFilterClientMock.EXPECT().VerifyAccess(ctx, listPodsSAR("ns2")).Return(true, nil)
			nsFilterClientMock.EXPECT().VerifyAccess(ctx, podLogsSAR("ns2")).Return(true, nil)

			filterResult, err := nsFilter.Filter(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(filterResult.Valid).To(Equal([]string{"ns1", "ns2"}))
			Expect(filterResult.NonExisting).To(Equal([]string{}))
			Expect(filterResult.NonPermitted).To(Equal([]string{}))
		})

		It("should return the correct FilterResult when no namespaces are provided", func() {
			nsFilter := NewNamespacesFilter(nsFilterClientMock, []string{})

			filterResult, err := nsFilter.Filter(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(filterResult.Valid).To(Equal([]string{}))
			Expect(filterResult.NonExisting).To(Equal([]string{}))
			Expect(filterResult.NonPermitted).To(Equal([]string{}))
		})

		It("should return the correct FilterResult when there are valid, non-existing and non-permitted namespaces", func() {
			nsFilter := NewNamespacesFilter(nsFilterClientMock, []string{"ns1", "ns2", "ns3"})

			nsFilterClientMock.EXPECT().ListNamespaces(ctx).Return(getNamespaceList([]string{"ns1", "ns2"}), nil)

			nsFilterClientMock.EXPECT().VerifyAccess(ctx, listPodsSAR("ns1")).Return(true, nil)
			nsFilterClientMock.EXPECT().VerifyAccess(ctx, podLogsSAR("ns1")).Return(true, nil)

			nsFilterClientMock.EXPECT().VerifyAccess(ctx, listPodsSAR("ns2")).Return(false, nil)

			filterResult, err := nsFilter.Filter(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(filterResult.Valid).To(Equal([]string{"ns1"}))
			Expect(filterResult.NonExisting).To(Equal([]string{"ns3"}))
			Expect(filterResult.NonPermitted).To(Equal([]string{"ns2"}))
		})

		It("should return the correct FilterResult when client returns more namespaces than the provided namespaces", func() {
			nsFilter := NewNamespacesFilter(nsFilterClientMock, []string{"ns1"})

			nsFilterClientMock.EXPECT().ListNamespaces(ctx).Return(getNamespaceList([]string{"ns1", "ns2", "ns3"}), nil)

			nsFilterClientMock.EXPECT().VerifyAccess(ctx, listPodsSAR("ns1")).Return(true, nil)
			nsFilterClientMock.EXPECT().VerifyAccess(ctx, podLogsSAR("ns1")).Return(true, nil)

			filterResult, err := nsFilter.Filter(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(filterResult.Valid).To(Equal([]string{"ns1"}))
			Expect(filterResult.NonExisting).To(Equal([]string{}))
			Expect(filterResult.NonPermitted).To(Equal([]string{}))
		})

		It("should return the correct FilterResult when a namespace doesn't exist", func() {
			nsFilter := NewNamespacesFilter(nsFilterClientMock, []string{"ns1"})

			nsFilterClientMock.EXPECT().ListNamespaces(ctx).Return(getNamespaceList([]string{}), nil)

			filterResult, err := nsFilter.Filter(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(filterResult.Valid).To(Equal([]string{}))
			Expect(filterResult.NonExisting).To(Equal([]string{"ns1"}))
			Expect(filterResult.NonPermitted).To(Equal([]string{}))
		})

		It("should return an error when failed to list namespaces", func() {
			nsFilter := NewNamespacesFilter(nsFilterClientMock, []string{"ns1", "ns2"})

			nsFilterClientMock.EXPECT().ListNamespaces(ctx).Return(nil, fmt.Errorf("failed to list namespaces"))

			_, err := nsFilter.Filter(ctx)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to list namespaces"))
		})

		It("should return the correct FilterResult when a namespace is not permitted to list pods", func() {
			nsFilter := NewNamespacesFilter(nsFilterClientMock, []string{"ns1"})

			nsFilterClientMock.EXPECT().ListNamespaces(ctx).Return(getNamespaceList([]string{"ns1"}), nil)

			nsFilterClientMock.EXPECT().VerifyAccess(ctx, listPodsSAR("ns1")).Return(false, nil)

			filterResult, err := nsFilter.Filter(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(filterResult.Valid).To(Equal([]string{}))
			Expect(filterResult.NonExisting).To(Equal([]string{}))
			Expect(filterResult.NonPermitted).To(Equal([]string{"ns1"}))
		})

		It("should return the correct FilterResult when a namespace is not permitted to get pods logs", func() {
			nsFilter := NewNamespacesFilter(nsFilterClientMock, []string{"ns1"})

			nsFilterClientMock.EXPECT().ListNamespaces(ctx).Return(getNamespaceList([]string{"ns1"}), nil)

			nsFilterClientMock.EXPECT().VerifyAccess(ctx, listPodsSAR("ns1")).Return(true, nil)
			nsFilterClientMock.EXPECT().VerifyAccess(ctx, podLogsSAR("ns1")).Return(false, nil)

			filterResult, err := nsFilter.Filter(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(filterResult.Valid).To(Equal([]string{}))
			Expect(filterResult.NonExisting).To(Equal([]string{}))
			Expect(filterResult.NonPermitted).To(Equal([]string{"ns1"}))
		})

		It("should return an error when failed to verify access to list pods", func() {
			nsFilter := NewNamespacesFilter(nsFilterClientMock, []string{"ns1"})

			nsFilterClientMock.EXPECT().ListNamespaces(ctx).Return(getNamespaceList([]string{"ns1"}), nil)

			nsFilterClientMock.EXPECT().VerifyAccess(ctx, listPodsSAR("ns1")).Return(false, fmt.Errorf("error"))

			filterResult, err := nsFilter.Filter(ctx)
			Expect(filterResult).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to verify list pods access permissions in ns1: error"))
		})

		It("should return an error when failed to verify access to get pod logs", func() {
			nsFilter := NewNamespacesFilter(nsFilterClientMock, []string{"ns1"})

			nsFilterClientMock.EXPECT().ListNamespaces(ctx).Return(getNamespaceList([]string{"ns1"}), nil)

			nsFilterClientMock.EXPECT().VerifyAccess(ctx, listPodsSAR("ns1")).Return(true, nil)
			nsFilterClientMock.EXPECT().VerifyAccess(ctx, podLogsSAR("ns1")).Return(false, fmt.Errorf("error"))

			filterResult, err := nsFilter.Filter(ctx)
			Expect(filterResult).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to verify get pod logs access permissions in ns1: error"))
		})
	})
})

var _ = Describe("Diff", func() {
	It("should return the correct diff", func() {
		diff := diff([]string{"ns1", "ns2"}, []string{"ns1", "ns3"})
		Expect(diff).To(Equal([]string{"ns2"}))
	})
})

var _ = Describe("Unique", func() {
	It("should return the correct unique list", func() {
		unique := unique([]string{"ns1", "ns2", "ns1", "ns1"})
		Expect(unique).To(Equal([]string{"ns1", "ns2"}))
	})
})

func getNamespaceList(namespaces []string) *corev1.NamespaceList {
	items := make([]corev1.Namespace, len(namespaces))
	for i, n := range namespaces {
		items[i] = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: n,
			},
		}
	}

	return &corev1.NamespaceList{
		Items: items,
	}
}
