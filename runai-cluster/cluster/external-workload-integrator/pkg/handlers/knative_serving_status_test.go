/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"

	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	knativeapis "knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

var _ = Describe("Knative serving workload update counter", func() {
	var (
		service *servingv1.Service
		runner  *dromosv1alpha1.WorkloadRunner
		pod     *corev1.Pod
		ew      *v1alpha1.ExternalWorkload
	)

	BeforeEach(func() {
		runner = emptyWLR("runner", "team-a")
		runner.UID = "runner-uid"
		runner.Annotations = map[string]string{clusterconstants.AnnotationsBackendWorkloadCounter: "7"}
		service = &servingv1.Service{ObjectMeta: metav1.ObjectMeta{
			Name: "inference", Namespace: runner.Namespace, Generation: 2,
			Annotations: map[string]string{clusterconstants.AnnotationsBackendWorkloadCounter: "7"},
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(runner,
				dromosv1alpha1.SchemeGroupVersion.WithKind(dromosv1alpha1.WorkloadRunnerKind))},
		}}
		service.Spec.Template.Annotations = map[string]string{MinScaleAnnotation: "1"}
		service.Spec.Template.Spec.Containers = []corev1.Container{{Name: "model", Image: "test-image"}}
		service.Status.ObservedGeneration = service.Generation
		service.Status.Conditions = duckv1.Conditions{{Type: knativeapis.ConditionReady, Status: corev1.ConditionTrue}}
		service.Status.LatestCreatedRevisionName = "inference-00002"
		service.Status.LatestReadyRevisionName = "inference-00002"
		service.Status.Traffic = []servingv1.TrafficTarget{{RevisionName: "inference-00002", Percent: ptr.To[int64](100)}}
		pod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "model-pod", Namespace: runner.Namespace, Labels: map[string]string{
				serving.ServiceLabelKey: service.Name, serving.RevisionLabelKey: service.Status.LatestReadyRevisionName,
			}},
			Status: corev1.PodStatus{Phase: corev1.PodRunning, Conditions: []corev1.PodCondition{{
				Type: corev1.PodReady, Status: corev1.ConditionTrue,
			}}},
		}
		ew = &v1alpha1.ExternalWorkload{}
		ew.Spec.AdditionalFields = ptr.To(`{"servingWorkloadUpdateCounter":"6","otherKey":"preserved"}`)
	})

	newHandler := func(funcs interceptor.Funcs) *KNativeTopOwnerHandler {
		scheme := newTestScheme()
		Expect(servingv1.AddToScheme(scheme)).To(Succeed())
		c := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(runner, pod).WithInterceptorFuncs(funcs).Build()
		return NewKNativeTopOwnerHandler(c, nil)
	}

	It("reports the applied counter through Parse and keeps unrelated fields", func() {
		Expect(newHandler(interceptor.Funcs{}).Parse(context.Background(), service, ew)).To(Succeed())
		Expect(decodedAdditionalFields(ew)).To(Equal(map[string]any{
			commonconstants.ServingWorkloadUpdateCounterField: "7", "otherKey": "preserved",
		}))
	})

	It("accepts the initial update counter zero", func() {
		service.Annotations[clusterconstants.AnnotationsBackendWorkloadCounter] = "0"
		runner.Annotations[clusterconstants.AnnotationsBackendWorkloadCounter] = "0"
		newHandler(interceptor.Funcs{}).setServingWorkloadUpdateCounter(context.Background(), service, ew)
		Expect(decodedAdditionalFields(ew)).To(HaveKeyWithValue(commonconstants.ServingWorkloadUpdateCounterField, "0"))
	})

	It("allows an already idle ready revision with minimum zero", func() {
		service.Spec.Template.Annotations[MinScaleAnnotation] = "0"
		pod.Status.Conditions = nil
		newHandler(interceptor.Funcs{}).setServingWorkloadUpdateCounter(context.Background(), service, ew)
		Expect(decodedAdditionalFields(ew)).To(HaveKeyWithValue(commonconstants.ServingWorkloadUpdateCounterField, "7"))
	})

	DescribeTable("clears stale proof when", func(mutate func()) {
		mutate()
		newHandler(interceptor.Funcs{}).setServingWorkloadUpdateCounter(context.Background(), service, ew)
		Expect(decodedAdditionalFields(ew)).To(Equal(map[string]any{"otherKey": "preserved"}))
	},
		Entry("the root annotation is absent", func() { service.Annotations = nil }),
		Entry("the counter only exists on the pod template", func() {
			service.Annotations = nil
			service.Spec.Template.Annotations[clusterconstants.AnnotationsBackendWorkloadCounter] = "7"
		}),
		Entry("the counter is negative", func() { service.Annotations[clusterconstants.AnnotationsBackendWorkloadCounter] = "-1" }),
		Entry("the counter is not an integer", func() { service.Annotations[clusterconstants.AnnotationsBackendWorkloadCounter] = "invalid" }),
		Entry("the counter overflows", func() {
			service.Annotations[clusterconstants.AnnotationsBackendWorkloadCounter] = "9223372036854775808"
		}),
		Entry("the counter has a leading zero", func() { service.Annotations[clusterconstants.AnnotationsBackendWorkloadCounter] = "07" }),
		Entry("the counter has a plus sign", func() { service.Annotations[clusterconstants.AnnotationsBackendWorkloadCounter] = "+7" }),
		Entry("the service has no runner controller", func() { service.OwnerReferences = nil }),
		Entry("the runner reference is not controlling", func() { service.OwnerReferences[0].Controller = ptr.To(false) }),
		Entry("the owner has another kind", func() { service.OwnerReferences[0].Kind = "OtherKind" }),
		Entry("the owner has another API version", func() { service.OwnerReferences[0].APIVersion = "other/v1" }),
		Entry("the owner UID is absent", func() { service.OwnerReferences[0].UID = "" }),
		Entry("the owner UID does not match", func() { runner.UID = "replacement-runner" }),
		Entry("a newer runner update is pending", func() { runner.Annotations[clusterconstants.AnnotationsBackendWorkloadCounter] = "8" }),
		Entry("the runner counter is absent", func() { runner.Annotations = nil }),
		Entry("the runner is deleting", func() {
			runner.DeletionTimestamp = ptr.To(metav1.Now())
			runner.Finalizers = []string{"test"}
		}),
		Entry("the service is deleting", func() { service.DeletionTimestamp = ptr.To(metav1.Now()) }),
		Entry("the service generation has not been observed", func() { service.Status.ObservedGeneration-- }),
		Entry("the service is not ready", func() { service.Status.Conditions[0].Status = corev1.ConditionFalse }),
		Entry("service readiness is unknown", func() { service.Status.Conditions = nil }),
		Entry("a new image revision is still pending behind the serving revision", func() {
			service.Status.LatestCreatedRevisionName = "inference-00003"
		}),
		Entry("the ready revision is absent", func() { service.Status.LatestReadyRevisionName = "" }),
		Entry("traffic is unresolved", func() { service.Status.Traffic = nil }),
		Entry("traffic points to a different revision", func() { service.Status.Traffic[0].RevisionName = "inference-00001" }),
		Entry("the traffic percentage is absent", func() { service.Status.Traffic[0].Percent = nil }),
		Entry("the revision does not receive all traffic", func() { service.Status.Traffic[0].Percent = ptr.To[int64](50) }),
		Entry("traffic is split across revisions", func() {
			service.Status.Traffic = append(service.Status.Traffic, servingv1.TrafficTarget{RevisionName: "inference-00001", Percent: ptr.To[int64](0)})
		}),
		Entry("a positive minimum has no ready pod", func() { pod.Status.Conditions[0].Status = corev1.ConditionFalse }),
		Entry("the only ready pod belongs to an older revision", func() { pod.Labels[serving.RevisionLabelKey] = "inference-00001" }),
		Entry("the only ready pod belongs to another service", func() { pod.Labels[serving.ServiceLabelKey] = "other-service" }),
		Entry("the only ready pod is in another namespace", func() { pod.Namespace = "other-project" }),
		Entry("the pod is terminating", func() {
			pod.DeletionTimestamp = ptr.To(metav1.Now())
			pod.Finalizers = []string{"test"}
		}),
		Entry("the pod is not running", func() { pod.Status.Phase = corev1.PodPending }),
		Entry("the minimum is invalid", func() { service.Spec.Template.Annotations[MinScaleAnnotation] = "invalid" }),
		Entry("the minimum is negative", func() { service.Spec.Template.Annotations[MinScaleAnnotation] = "-1" }),
		Entry("the implicit positive minimum has no ready pod", func() {
			service.Spec.Template.Annotations = nil
			pod.Status.Conditions = nil
		}),
	)

	DescribeTable("clears stale proof on read failure", func(funcs interceptor.Funcs) {
		newHandler(funcs).setServingWorkloadUpdateCounter(context.Background(), service, ew)
		Expect(decodedAdditionalFields(ew)).To(Equal(map[string]any{"otherKey": "preserved"}))
	},
		Entry("runner read", interceptor.Funcs{Get: func(context.Context, client.WithWatch, client.ObjectKey, client.Object, ...client.GetOption) error {
			return apierrors.NewServiceUnavailable("runner unavailable")
		}}),
		Entry("pod list", interceptor.Funcs{List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
			return apierrors.NewServiceUnavailable("pods unavailable")
		}}),
	)

	It("clears stale proof even when Parse rejects the object type", func() {
		Expect(newHandler(interceptor.Funcs{}).Parse(context.Background(), pod, ew)).To(HaveOccurred())
		Expect(decodedAdditionalFields(ew)).To(Equal(map[string]any{"otherKey": "preserved"}))
	})

	It("clears stale proof on a later reconciliation after a rollout starts", func() {
		h := newHandler(interceptor.Funcs{})
		h.setServingWorkloadUpdateCounter(context.Background(), service, ew)
		Expect(decodedAdditionalFields(ew)).To(HaveKey(commonconstants.ServingWorkloadUpdateCounterField))
		service.Status.LatestCreatedRevisionName = "inference-00003"
		h.setServingWorkloadUpdateCounter(context.Background(), service, ew)
		Expect(decodedAdditionalFields(ew)).To(Equal(map[string]any{"otherKey": "preserved"}))
	})
})
