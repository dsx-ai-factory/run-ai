/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/test"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("TestNimServiceHandler DRA", Ordered, func() {
	var (
		cachedClient client.WithWatch
		nimHandler   *NimServiceTopOwnerHandler
		ctx          context.Context
		cancel       context.CancelFunc
	)

	BeforeAll(func() {
		scheme := scheme_init.Scheme()
		ns := test.BuildNamespace(test.TestNamespace, map[string]string{})
		cachedClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(ns).Build()

		draFakeClient := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).WithObjects(test.TestDRAGpuClaim.DeepCopy(), test.TestDRAGpuClaimTemplate.DeepCopy()).Build()
		fakeKubeClient := kubefake.NewClientset(test.TestDRAGpuClaim.DeepCopy(), test.TestDRAGpuClaimTemplate.DeepCopy())
		draParser := dra.NewDRAParser(draFakeClient, fakeKubeClient)
		nimHandler = NewNimServiceTopOwnerHandler(cachedClient, draParser)

		ctx, cancel = context.WithCancel(context.Background())
		DeferCleanup(func() {
			cancel()
		})
	})

	It("NIMService with DRA ResourceClaim (3 replicas * 2 GPUs = 6 GPUs)", func() {
		nimService := test.TestNimServiceWithDRAClaim.DeepCopy()
		expectedEw := test.TestNimServiceWithDRAClaimExpectedEw.DeepCopy()

		ew := &v1alpha1.ExternalWorkload{}
		err := nimHandler.Parse(ctx, nimService, ew)
		Expect(err).To(BeNil())
		test.ValidateExternalWorkloadInnerSpec(Default, 1, ew, expectedEw)
	})

	It("NIMService with DRA ResourceClaimTemplate (2 replicas * 4 GPUs = 8 GPUs)", func() {
		nimService := test.TestNimServiceWithDRATemplate.DeepCopy()
		expectedEw := test.TestNimServiceWithDRATemplateExpectedEw.DeepCopy()

		ew := &v1alpha1.ExternalWorkload{}
		err := nimHandler.Parse(ctx, nimService, ew)
		Expect(err).To(BeNil())
		test.ValidateExternalWorkloadInnerSpec(Default, 1, ew, expectedEw)
	})

	It("NIMService with fractional GPU annotation (gpu-fraction: 0.5, no resources)", func() {
		nimService := test.TestNimServiceWithFractionalGpu.DeepCopy()
		expectedEw := test.TestNimServiceWithFractionalGpuExpectedEw.DeepCopy()

		ew := &v1alpha1.ExternalWorkload{}
		err := nimHandler.Parse(ctx, nimService, ew)
		Expect(err).To(BeNil())
		test.ValidateExternalWorkloadInnerSpec(Default, 1, ew, expectedEw)
	})

	It("NIMService fractional GPU is idempotent across re-parses of the same ExternalWorkload", func() {
		// The integrator re-parses the fetched ExternalWorkload on every reconcile, so
		// parsing the same object repeatedly must not accumulate the annotation GPU.
		nimService := test.TestNimServiceWithFractionalGpu.DeepCopy()
		expectedEw := test.TestNimServiceWithFractionalGpuExpectedEw.DeepCopy()

		ew := &v1alpha1.ExternalWorkload{}
		for i := 0; i < 3; i++ {
			Expect(nimHandler.Parse(ctx, nimService, ew)).To(BeNil())
		}
		test.ValidateExternalWorkloadInnerSpec(Default, 1, ew, expectedEw)
	})

	It("NIMService with no resources and no GPU annotations leaves requested resources empty", func() {
		nimService := test.TestNimServiceWithFractionalGpu.DeepCopy()
		nimService.Spec.Annotations = map[string]string{"some-other-annotation": "value"}

		ew := &v1alpha1.ExternalWorkload{}
		err := nimHandler.Parse(ctx, nimService, ew)
		Expect(err).To(BeNil())
		Expect(ew.Spec.RequestedResources.Limits).To(BeEmpty())
		Expect(ew.Spec.RequestedResources.Requests).To(BeEmpty())
	})
})
