/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Knative requested resources", func() {
	It("uses the current Revision minimum for an in-place decrease, then the template minimum during a rollout", func() {
		scheme := runtime.NewScheme()
		Expect(servingv1.AddToScheme(scheme)).To(Succeed())
		revision := &servingv1.Revision{ObjectMeta: metav1.ObjectMeta{
			Name: "model-00001", Namespace: "project", Annotations: map[string]string{MinScaleAnnotation: "0"},
		}}
		handler := NewKNativeTopOwnerHandler(fake.NewClientBuilder().WithScheme(scheme).WithObjects(revision).Build(), nil)
		service := &servingv1.Service{ObjectMeta: metav1.ObjectMeta{
			Name: "model", Namespace: "project", Labels: map[string]string{
				constants.ConvertedFromNativeLabel: "true", constants.ControlPlaneWorkloadKindLabelForWorkload: constants.V2InferenceWorkloadKind,
			},
		}}
		service.Spec.Template.Annotations = map[string]string{MinScaleAnnotation: "1"}
		service.Status.LatestCreatedRevisionName = revision.Name
		service.Status.LatestReadyRevisionName = revision.Name
		Expect(handler.minimumReplicas(context.Background(), service)).To(HaveValue(Equal(int32(0))))
		service.Status.LatestCreatedRevisionName = "model-00002"
		Expect(handler.minimumReplicas(context.Background(), service)).To(HaveValue(Equal(int32(1))))
	})
})
