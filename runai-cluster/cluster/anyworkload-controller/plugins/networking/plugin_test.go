/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Plugin", func() {
	Describe("networking annotations watch predicate", func() {
		var predicate = networkingAnnotationsChangedPredicate()

		DescribeTable("accepts updates when a networking annotation value changes", func(annotation string) {
			Expect(predicate.Update(event.UpdateEvent{
				ObjectOld: workloadRunnerWithAnnotations(map[string]string{annotation: "old"}),
				ObjectNew: workloadRunnerWithAnnotations(map[string]string{annotation: "new"}),
			})).To(BeTrue())
		},
			Entry("endpoints", endpoints.Annotation),
			Entry("submission profile", constants.NetworkingSubmissionProfileAnnotation),
		)

		DescribeTable("accepts updates when a networking annotation is added or removed", func(annotation string) {
			Expect(predicate.Update(event.UpdateEvent{
				ObjectOld: workloadRunnerWithAnnotations(nil),
				ObjectNew: workloadRunnerWithAnnotations(map[string]string{annotation: "new"}),
			})).To(BeTrue())

			Expect(predicate.Update(event.UpdateEvent{
				ObjectOld: workloadRunnerWithAnnotations(map[string]string{annotation: "old"}),
				ObjectNew: workloadRunnerWithAnnotations(nil),
			})).To(BeTrue())
		},
			Entry("endpoints", endpoints.Annotation),
			Entry("submission profile", constants.NetworkingSubmissionProfileAnnotation),
		)

		It("ignores unrelated annotation updates", func() {
			Expect(predicate.Update(event.UpdateEvent{
				ObjectOld: workloadRunnerWithAnnotations(map[string]string{endpoints.Annotation: "same", "other": "old"}),
				ObjectNew: workloadRunnerWithAnnotations(map[string]string{endpoints.Annotation: "same", "other": "new"}),
			})).To(BeFalse())
		})

		It("ignores non-update events", func() {
			runner := workloadRunnerWithAnnotations(map[string]string{endpoints.Annotation: "value"})

			Expect(predicate.Create(event.CreateEvent{Object: runner})).To(BeFalse())
			Expect(predicate.Delete(event.DeleteEvent{Object: runner})).To(BeFalse())
			Expect(predicate.Generic(event.GenericEvent{Object: runner})).To(BeFalse())
		})
	})

	Describe("mapWorkloadRunnerToSelf", func() {
		It("maps the WorkloadRunner object to its own reconcile request", func() {
			requests := mapWorkloadRunnerToSelf(nil, workloadRunnerWithAnnotations(nil))

			Expect(requests).To(HaveLen(1))
			Expect(requests[0].NamespacedName.Name).To(Equal("test-runner"))
			Expect(requests[0].NamespacedName.Namespace).To(Equal("test-namespace"))
		})
	})

	Describe("dependency watch predicates", func() {
		It("maps only owner-reference changes for intermediate owners", func() {
			predicate := ownerReferencesChangedPredicate()
			oldObject := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "owner"}}
			statusOnly := oldObject.DeepCopy()
			statusOnly.Labels = map[string]string{"unrelated": "change"}
			Expect(predicate.Update(event.UpdateEvent{ObjectOld: oldObject, ObjectNew: statusOnly})).To(BeFalse())

			ownerChanged := oldObject.DeepCopy()
			ownerChanged.OwnerReferences = []metav1.OwnerReference{{Kind: "WorkloadRunner", Name: "runner"}}
			Expect(predicate.Update(event.UpdateEvent{ObjectOld: oldObject, ObjectNew: ownerChanged})).To(BeTrue())
		})

		It("maps Pod ownership, selector, and IP changes but ignores unrelated status", func() {
			predicate := podDependencyChangedPredicate()
			oldPod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod", Labels: map[string]string{"app": "one"}}}
			statusOnly := oldPod.DeepCopy()
			statusOnly.Status.Phase = corev1.PodRunning
			Expect(predicate.Update(event.UpdateEvent{ObjectOld: oldPod, ObjectNew: statusOnly})).To(BeFalse())

			labelChanged := oldPod.DeepCopy()
			labelChanged.Labels["app"] = "two"
			Expect(predicate.Update(event.UpdateEvent{ObjectOld: oldPod, ObjectNew: labelChanged})).To(BeTrue())

			ipChanged := oldPod.DeepCopy()
			ipChanged.Status.PodIP = "10.0.0.1"
			Expect(predicate.Update(event.UpdateEvent{ObjectOld: oldPod, ObjectNew: ipChanged})).To(BeTrue())
		})
	})

	Describe("Service watch mapping", func() {
		It("maps a directly owned Service to its WorkloadRunner", func() {
			plugin := NewPlugin(nil)
			svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{
				Name:      "metrics",
				Namespace: "test-namespace",
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: dromosv1alpha1.GroupVersion.String(),
					Kind:       "WorkloadRunner",
					Name:       "test-runner",
				}},
			}}

			requests := plugin.mapToWorkloadRunner(context.Background(), svc)
			Expect(requests).To(HaveLen(1))
			Expect(requests[0].NamespacedName.String()).To(Equal("test-namespace/test-runner"))
		})

		It("maps a Service through its primary resource owner", func() {
			primary := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name:      "primary",
				Namespace: "test-namespace",
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: dromosv1alpha1.GroupVersion.String(),
					Kind:       "WorkloadRunner",
					Name:       "test-runner",
				}},
			}}
			reader := fake.NewClientBuilder().WithScheme(newTestScheme()).WithObjects(primary).Build()
			plugin := NewPlugin(reader)
			svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{
				Name:      "metrics",
				Namespace: "test-namespace",
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: "v1",
					Kind:       "ConfigMap",
					Name:       "primary",
				}},
			}}

			requests := plugin.mapToWorkloadRunner(context.Background(), svc)
			Expect(requests).To(HaveLen(1))
			Expect(requests[0].NamespacedName.String()).To(Equal("test-namespace/test-runner"))
		})

		It("maps through a metadata-only watched Job owner", func() {
			primary := &batchv1.Job{ObjectMeta: metav1.ObjectMeta{
				Name:      "primary-job",
				Namespace: "test-namespace",
				UID:       "primary-job-uid",
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: dromosv1alpha1.GroupVersion.String(),
					Kind:       dromosv1alpha1.WorkloadRunnerKind,
					Name:       "test-runner",
				}},
			}}
			reader := fake.NewClientBuilder().WithScheme(newTestScheme()).WithObjects(primary).Build()
			plugin := NewPlugin(reader)
			service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{
				Name:      "workload-service",
				Namespace: "test-namespace",
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: batchv1.SchemeGroupVersion.String(),
					Kind:       "Job",
					Name:       primary.Name,
					UID:        primary.UID,
				}},
			}}

			requests := plugin.mapToWorkloadRunner(context.Background(), service)
			Expect(requests).To(HaveLen(1))
			Expect(requests[0].NamespacedName.String()).To(Equal("test-namespace/test-runner"))
		})
	})
})

func workloadRunnerWithAnnotations(annotations map[string]string) *dromosv1alpha1.WorkloadRunner {
	return &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-runner",
			Namespace:   "test-namespace",
			Annotations: annotations,
		},
	}
}
