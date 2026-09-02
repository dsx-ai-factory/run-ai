/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	monorepocommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

var _ = Describe("JobTopOwnerHandler.setRunaiWorkloadObservedStatus", func() {
	// On a read failure we keep the Job phase and don't flag ewi-status (cluster-sync uses the wlr phase).
	It("keeps the Job phase and skips ewi-status on a pod-list error", func() {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())

		c := fakeclient.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
			List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
				return apierrors.NewServiceUnavailable("pod list unavailable")
			},
		}).Build()

		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{Name: "my-job", Namespace: "team-a"},
			Status:     batchv1.JobStatus{Active: 1, Ready: ptr.To[int32](1)}, // setJobObservedStatus -> Running
		}
		// A stale authoritative flag from a previous reconcile must be cleared so cluster-sync isn't blocked.
		ew := &v1alpha1.ExternalWorkload{ObjectMeta: metav1.ObjectMeta{
			Name: "ew-1", Namespace: "team-a",
			Annotations: map[string]string{monorepocommonconstants.EwiStatusAnnotation: "true"},
		}}

		handler := &JobTopOwnerHandler{cachedClient: c}
		handler.setRunaiWorkloadObservedStatus(context.Background(), job, "wlr-1", ew)

		Expect(ew.Annotations).NotTo(HaveKey(monorepocommonconstants.EwiStatusAnnotation))
		Expect(ew.Status.ObservedStatus.Phase).To(Equal(v1alpha1.ExWRunningPhase))
	})

	It("keeps the Job phase and skips ewi-status on a WorkloadRunner fetch error", func() {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())

		// Pod list succeeds (empty); only the WorkloadRunner Get fails.
		c := fakeclient.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
			Get: func(context.Context, client.WithWatch, client.ObjectKey, client.Object, ...client.GetOption) error {
				return apierrors.NewServiceUnavailable("workloadrunner get unavailable")
			},
		}).Build()

		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{Name: "my-job", Namespace: "team-a"},
			Status:     batchv1.JobStatus{Active: 1, Ready: ptr.To[int32](0)}, // setJobObservedStatus -> Pending
		}
		// A stale authoritative flag from a previous reconcile must be cleared so cluster-sync isn't blocked.
		ew := &v1alpha1.ExternalWorkload{ObjectMeta: metav1.ObjectMeta{
			Name: "ew-1", Namespace: "team-a",
			Annotations: map[string]string{monorepocommonconstants.EwiStatusAnnotation: "true"},
		}}

		handler := &JobTopOwnerHandler{cachedClient: c}
		handler.setRunaiWorkloadObservedStatus(context.Background(), job, "wlr-1", ew)

		Expect(ew.Annotations).NotTo(HaveKey(monorepocommonconstants.EwiStatusAnnotation))
		Expect(ew.Status.ObservedStatus.Phase).To(Equal(v1alpha1.ExWPendingPhase))
	})
})
