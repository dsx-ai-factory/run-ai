/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func counterService() *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "serving.knative.dev/v1",
		"kind":       "Service",
		"metadata": map[string]interface{}{
			"name": "inference",
			"annotations": map[string]interface{}{
				backendWorkloadCounterAnnotation: "999",
				"example.com/preserve":           "true",
			},
		},
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						backendWorkloadCounterAnnotation: "user-supplied",
					},
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{map[string]interface{}{"image": "example/image:old"}},
				},
			},
		},
	}}
}

var _ = Describe("Knative applied workload counter", func() {
	It("reconciles a counter-only update and ignores unrelated metadata changes", func() {
		before := &dromosv1alpha1.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{
			Generation: 1, Annotations: map[string]string{backendWorkloadCounterAnnotation: "5"},
		}}
		after := before.DeepCopy()
		after.Annotations["example.com/status"] = "updated"
		update := event.UpdateEvent{ObjectOld: before, ObjectNew: after}
		Expect(workloadRunnerUpdateChanged(update)).To(BeFalse())
		after.Annotations[backendWorkloadCounterAnnotation] = "6"
		Expect(workloadRunnerUpdateChanged(update)).To(BeTrue())
		after.Annotations[backendWorkloadCounterAnnotation] = "5"
		after.Generation++
		Expect(workloadRunnerUpdateChanged(update)).To(BeTrue())
	})

	DescribeTable("trusts only canonical nonnegative runner counters",
		func(value, expected string) {
			runner := &dromosv1alpha1.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{backendWorkloadCounterAnnotation: value},
			}}
			service := counterService()
			originalSpec := service.DeepCopy().Object["spec"]

			setKnativeWorkloadCounter(runner, service)

			Expect(service.GetAnnotations()).To(HaveKeyWithValue("example.com/preserve", "true"))
			if expected == "" {
				Expect(service.GetAnnotations()).NotTo(HaveKey(backendWorkloadCounterAnnotation))
			} else {
				Expect(service.GetAnnotations()).To(HaveKeyWithValue(backendWorkloadCounterAnnotation, expected))
			}
			Expect(service.Object["spec"]).To(Equal(originalSpec))
		},
		Entry("initial version zero", "0", "0"),
		Entry("updated version", "12", "12"),
		Entry("largest supported counter", "9223372036854775807", "9223372036854775807"),
		Entry("missing counter", "", ""),
		Entry("negative counter", "-1", ""),
		Entry("leading zero", "01", ""),
		Entry("signed positive", "+1", ""),
		Entry("whitespace", " 1", ""),
		Entry("nondecimal", "1.0", ""),
		Entry("overflow", "9223372036854775808", ""),
	)

	DescribeTable("does not change other attached resource kinds or groups",
		func(apiVersion, kind string) {
			resource := counterService()
			resource.SetAPIVersion(apiVersion)
			resource.SetKind(kind)
			before := resource.DeepCopy()
			setKnativeWorkloadCounter(&dromosv1alpha1.WorkloadRunner{}, resource)
			Expect(resource).To(Equal(before))
		},
		Entry("core Service", "v1", "Service"),
		Entry("Knative Revision", "serving.knative.dev/v1", "Revision"),
		Entry("Deployment", "apps/v1", "Deployment"),
	)

	It("adds only a root annotation when the rendered Service has no annotations", func() {
		service := counterService()
		service.SetAnnotations(nil)
		unstructured.RemoveNestedField(service.Object, "spec", "template", "metadata", "annotations")
		originalSpec := service.DeepCopy().Object["spec"]
		runner := &dromosv1alpha1.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{backendWorkloadCounterAnnotation: "1"},
		}}

		setKnativeWorkloadCounter(runner, service)

		Expect(service.GetAnnotations()).To(Equal(map[string]string{backendWorkloadCounterAnnotation: "1"}))
		Expect(service.Object["spec"]).To(Equal(originalSpec))
	})

	It("publishes the runner counter with the post-plugin spec in the same apply", func() {
		scheme := runtime.NewScheme()
		Expect(dromosv1alpha1.AddToScheme(scheme)).To(Succeed())
		var applied []*unstructured.Unstructured
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
			Patch: func(_ context.Context, _ client.WithWatch, object client.Object, patch client.Patch, options ...client.PatchOption) error {
				Expect(patch.Type()).To(Equal(types.ApplyPatchType))
				settings := (&client.PatchOptions{}).ApplyOptions(options)
				Expect(settings.FieldManager).To(Equal(controllerName))
				Expect(settings.Force).To(HaveValue(BeTrue()))
				applied = append(applied, object.(*unstructured.Unstructured).DeepCopy())
				return nil
			},
		}).Build()
		controller := NewWorkloadRunnerController(k8sClient, framework.NewFramework(framework.NewRegistry(&counterMutationPlugin{}), nil), nil)
		controller.manager = &counterTestManager{scheme: scheme}
		runner := &dromosv1alpha1.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{
			Name: "inference", Namespace: "project", UID: "runner-uid",
			Annotations: map[string]string{backendWorkloadCounterAnnotation: "0"},
		}}
		service := counterService()
		_, err := controller.handleAttachedResource(context.Background(), runner, service, &kartav1alpha1.Karta{}, reconcile.New())
		Expect(err).NotTo(HaveOccurred())
		Expect(applied).To(HaveLen(1))
		Expect(applied[0].GetAnnotations()).To(HaveKeyWithValue(backendWorkloadCounterAnnotation, "0"))
		containers, _, err := unstructured.NestedSlice(applied[0].Object, "spec", "template", "spec", "containers")
		Expect(err).NotTo(HaveOccurred())
		Expect(containers).To(Equal([]interface{}{map[string]interface{}{"image": "example/image:mutated"}}))
		podAnnotations, _, err := unstructured.NestedStringMap(applied[0].Object, "spec", "template", "metadata", "annotations")
		Expect(err).NotTo(HaveOccurred())
		Expect(podAnnotations).To(HaveKeyWithValue(backendWorkloadCounterAnnotation, "user-supplied"))

		// A previously existing runner acquires or refreshes the marker without a template edit.
		runner.Annotations[backendWorkloadCounterAnnotation] = "1"
		_, err = controller.handleAttachedResource(context.Background(), runner, service.DeepCopy(), &kartav1alpha1.Karta{}, reconcile.New())
		Expect(err).NotTo(HaveOccurred())
		Expect(applied).To(HaveLen(2))
		Expect(applied[1].GetAnnotations()).To(HaveKeyWithValue(backendWorkloadCounterAnnotation, "1"))
		Expect(applied[1].Object["spec"]).To(Equal(applied[0].Object["spec"]))
	})
})

type counterTestManager struct {
	ctrl.Manager
	scheme *runtime.Scheme
}

func (m *counterTestManager) GetScheme() *runtime.Scheme { return m.scheme }

type counterMutationPlugin struct{}

func (*counterMutationPlugin) Name() string { return "counter-test-mutation" }

func (*counterMutationPlugin) MutateResource(_ context.Context, _ *dromosv1alpha1.WorkloadRunner, resource *unstructured.Unstructured, _ *kartav1alpha1.Karta, _ *reconcile.State) error {
	annotations := resource.GetAnnotations()
	annotations[backendWorkloadCounterAnnotation] = "spoofed-by-plugin"
	resource.SetAnnotations(annotations)
	return unstructured.SetNestedSlice(resource.Object, []interface{}{map[string]interface{}{"image": "example/image:mutated"}}, "spec", "template", "spec", "containers")
}
