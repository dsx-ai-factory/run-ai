/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Plugin", func() {
	Describe("endpoint annotation watch predicate", func() {
		var predicate = endpointAnnotationChangedPredicate()

		It("accepts updates when the endpoints annotation value changes", func() {
			Expect(predicate.Update(event.UpdateEvent{
				ObjectOld: workloadRunnerWithAnnotations(map[string]string{endpoints.Annotation: "old"}),
				ObjectNew: workloadRunnerWithAnnotations(map[string]string{endpoints.Annotation: "new"}),
			})).To(BeTrue())
		})

		It("accepts updates when the endpoints annotation is added or removed", func() {
			Expect(predicate.Update(event.UpdateEvent{
				ObjectOld: workloadRunnerWithAnnotations(nil),
				ObjectNew: workloadRunnerWithAnnotations(map[string]string{endpoints.Annotation: "new"}),
			})).To(BeTrue())

			Expect(predicate.Update(event.UpdateEvent{
				ObjectOld: workloadRunnerWithAnnotations(map[string]string{endpoints.Annotation: "old"}),
				ObjectNew: workloadRunnerWithAnnotations(nil),
			})).To(BeTrue())
		})

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
