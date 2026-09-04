/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package externalworkloads

import (
	"context"

	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgocache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

var _ = Describe("Run:ai status pod watcher", func() {
	DescribeTable("enqueues the workload whose status depends on the pod",
		func(labels map[string]string, expectedName string, expectedGVK schema.GroupVersionKind) {
			q := workqueue.NewTypedRateLimitingQueue[utils.ExternalWorkloadTask](
				workqueue.DefaultTypedControllerRateLimiter[utils.ExternalWorkloadTask]())
			handler := &ExternalWorkloadHandler{tasksQueue: q}
			pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
				Name: "pod-1", Namespace: "team-a", Labels: labels,
			}}

			enqueueRunaiStatusForPod(context.Background(), handler, pod)

			Expect(q.Len()).To(Equal(1))
			task, shutdown := q.Get()
			Expect(shutdown).To(BeFalse())
			Expect(task.Key.Name).To(Equal(expectedName))
			Expect(task.Key.Namespace).To(Equal("team-a"))
			Expect(task.GVK).To(Equal(expectedGVK))
			q.Done(task)
		},
		Entry("Knative revision pod", map[string]string{knativeServicePodLabel: "model"}, "model",
			workloadTypeGVKs[inferenceWorkloadType]),
		Entry("LeaderWorkerSet pod", map[string]string{leaderWorkerSetPodLabel: "model"}, "model",
			workloadTypeGVKs[distributedInferenceWorkloadType]),
	)

	It("ignores pods that do not feed a Run:ai phase calculator", func() {
		q := workqueue.NewTypedRateLimitingQueue[utils.ExternalWorkloadTask](
			workqueue.DefaultTypedControllerRateLimiter[utils.ExternalWorkloadTask]())
		handler := &ExternalWorkloadHandler{tasksQueue: q}

		enqueueRunaiStatusForPod(context.Background(), handler, &corev1.Pod{})

		Expect(q.Len()).To(BeZero())
	})

	It("filters informer events for pods that do not feed a Run:ai phase calculator", func() {
		q := workqueue.NewTypedRateLimitingQueue[utils.ExternalWorkloadTask](
			workqueue.DefaultTypedControllerRateLimiter[utils.ExternalWorkloadTask]())
		handler := &ExternalWorkloadHandler{tasksQueue: q}
		eventHandler := runaiStatusPodEventHandler(context.Background(), handler)

		eventHandler.OnAdd(&corev1.Pod{}, false)
		eventHandler.OnUpdate(&corev1.Pod{}, &corev1.Pod{})
		eventHandler.OnDelete(clientgocache.DeletedFinalStateUnknown{Obj: &corev1.Pod{}})

		Expect(q.Len()).To(BeZero())
	})

	It("keeps relevant tombstone delete events", func() {
		q := workqueue.NewTypedRateLimitingQueue[utils.ExternalWorkloadTask](
			workqueue.DefaultTypedControllerRateLimiter[utils.ExternalWorkloadTask]())
		handler := &ExternalWorkloadHandler{tasksQueue: q}
		eventHandler := runaiStatusPodEventHandler(context.Background(), handler)
		pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
			Name: "pod-1", Namespace: "team-a",
			Labels: map[string]string{knativeServicePodLabel: "model"},
		}}

		eventHandler.OnDelete(clientgocache.DeletedFinalStateUnknown{Obj: pod})

		Expect(q.Len()).To(Equal(1))
	})

	It("detects every pod field consumed by the phase calculators", func() {
		oldPod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
			Labels:      map[string]string{knativeServicePodLabel: "model"},
			Annotations: map[string]string{clustercommonconstants.PodStatusAnnotation: "ContainerCreating"},
		}}
		Expect(runaiStatusPodInputsChanged(oldPod, oldPod.DeepCopy())).To(BeFalse())

		readyPod := oldPod.DeepCopy()
		readyPod.Status.Conditions = []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}
		Expect(runaiStatusPodInputsChanged(oldPod, readyPod)).To(BeTrue())

		annotatedPod := oldPod.DeepCopy()
		annotatedPod.Annotations[clustercommonconstants.PodStatusAnnotation] = "Pending"
		Expect(runaiStatusPodInputsChanged(oldPod, annotatedPod)).To(BeTrue())
	})

	It("enqueues a Kubeflow training CR by matching the pod workload name to its ExternalWorkload", func() {
		const ns, jobName, wlrName = "team-a", "train-1", "wlr-1"
		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())

		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ew-train-1",
				Namespace: ns,
				Labels: map[string]string{
					commonconstants.WorkloadRunnerNameLabel:  wlrName,
					commonconstants.ConvertedFromNativeLabel: "true",
				},
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: "kubeflow.org/v1",
					Kind:       "PyTorchJob",
					Name:       jobName,
				}},
			},
			Spec: v1alpha1.ExternalWorkloadSpec{
				OriginalWorkloadName: jobName,
				Group:                kubeflowGroup,
				Version:              "v1",
			},
		}
		otherEW := ew.DeepCopy()
		otherEW.Name = "ew-other"
		otherEW.Labels = map[string]string{commonconstants.WorkloadRunnerNameLabel: "other-wlr"}
		notConvertedEW := ew.DeepCopy()
		notConvertedEW.Name = "ew-not-converted"
		notConvertedEW.Labels = map[string]string{commonconstants.WorkloadRunnerNameLabel: wlrName}
		c := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(ew, otherEW, notConvertedEW).Build()
		q := workqueue.NewTypedRateLimitingQueue[utils.ExternalWorkloadTask](
			workqueue.DefaultTypedControllerRateLimiter[utils.ExternalWorkloadTask]())
		handler := &ExternalWorkloadHandler{cachedClient: c, tasksQueue: q}
		pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-1",
			Namespace: ns,
			Labels: map[string]string{
				clustercommonconstants.LabelWorkloadName: wlrName,
				kubeflowv1.JobNameLabel:                  jobName,
			},
		}}

		enqueueRunaiStatusForPod(context.Background(), handler, pod)

		Expect(q.Len()).To(Equal(1))
		task, shutdown := q.Get()
		Expect(shutdown).To(BeFalse())
		Expect(task.Key.Name).To(Equal(jobName))
		Expect(task.Key.Namespace).To(Equal(ns))
		Expect(task.GVK).To(Equal(schema.GroupVersionKind{Group: kubeflowGroup, Version: "v1", Kind: "PyTorchJob"}))
		q.Done(task)
	})

	It("does not list ExternalWorkloads for a Kubeflow pod without a job-name label", func() {
		c := fakeclient.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
			List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
				Fail("ExternalWorkloads should not be listed without a job-name label")
				return nil
			},
		}).Build()
		q := workqueue.NewTypedRateLimitingQueue[utils.ExternalWorkloadTask](
			workqueue.DefaultTypedControllerRateLimiter[utils.ExternalWorkloadTask]())
		handler := &ExternalWorkloadHandler{cachedClient: c, tasksQueue: q}

		enqueueKubeflowRunaiStatusForPod(context.Background(), handler, &corev1.Pod{})

		Expect(q.Len()).To(BeZero())
	})

	It("does not list ExternalWorkloads for a Kubeflow pod without a workload-runner label", func() {
		c := fakeclient.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
			List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
				Fail("ExternalWorkloads should not be listed without a workload-runner label")
				return nil
			},
		}).Build()
		q := workqueue.NewTypedRateLimitingQueue[utils.ExternalWorkloadTask](
			workqueue.DefaultTypedControllerRateLimiter[utils.ExternalWorkloadTask]())
		handler := &ExternalWorkloadHandler{cachedClient: c, tasksQueue: q}
		pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{kubeflowv1.JobNameLabel: "train-1"},
		}}

		enqueueKubeflowRunaiStatusForPod(context.Background(), handler, pod)

		Expect(q.Len()).To(BeZero())
	})

	It("does not enqueue a Kubeflow task when no matching Kubeflow ExternalWorkload exists", func() {
		const ns, jobName, wlrName = "team-a", "train-1", "wlr-1"
		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ew-batch-1",
				Namespace: ns,
				Labels: map[string]string{
					commonconstants.WorkloadRunnerNameLabel:  wlrName,
					commonconstants.ConvertedFromNativeLabel: "true",
				},
			},
			Spec: v1alpha1.ExternalWorkloadSpec{
				OriginalWorkloadName: jobName,
				Group:                "batch",
				Version:              "v1",
			},
		}
		c := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(ew).Build()
		q := workqueue.NewTypedRateLimitingQueue[utils.ExternalWorkloadTask](
			workqueue.DefaultTypedControllerRateLimiter[utils.ExternalWorkloadTask]())
		handler := &ExternalWorkloadHandler{cachedClient: c, tasksQueue: q}
		pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Labels: map[string]string{
				clustercommonconstants.LabelWorkloadName: wlrName,
				kubeflowv1.JobNameLabel:                  jobName,
			},
		}}

		enqueueKubeflowRunaiStatusForPod(context.Background(), handler, pod)

		Expect(q.Len()).To(BeZero())
	})

	It("does not enqueue a Kubeflow task when listing ExternalWorkloads fails", func() {
		const ns, jobName, wlrName = "team-a", "train-1", "wlr-1"
		c := fakeclient.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
			List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
				return apierrors.NewServiceUnavailable("list failed")
			},
		}).Build()
		q := workqueue.NewTypedRateLimitingQueue[utils.ExternalWorkloadTask](
			workqueue.DefaultTypedControllerRateLimiter[utils.ExternalWorkloadTask]())
		handler := &ExternalWorkloadHandler{cachedClient: c, tasksQueue: q}
		pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Labels: map[string]string{
				clustercommonconstants.LabelWorkloadName: wlrName,
				kubeflowv1.JobNameLabel:                  jobName,
			},
		}}

		enqueueKubeflowRunaiStatusForPod(context.Background(), handler, pod)

		Expect(q.Len()).To(BeZero())
	})

	It("uses only owner references whose APIVersion matches the ExternalWorkload group and version", func() {
		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{
					{APIVersion: "batch/v1", Kind: "Job"},
					{APIVersion: "kubeflow.org/v1", Kind: "PyTorchJob"},
				},
			},
			Spec: v1alpha1.ExternalWorkloadSpec{
				Group:   kubeflowGroup,
				Version: "v1",
			},
		}

		gvk, ok := gvkFromExternalWorkloadOwner(ew)

		Expect(ok).To(BeTrue())
		Expect(gvk).To(Equal(schema.GroupVersionKind{Group: kubeflowGroup, Version: "v1", Kind: "PyTorchJob"}))
	})
})
