/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller_utils

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

var _ = Describe("GPU compute mode availability", func() {
	var (
		config   *unstructured.Unstructured
		operator *appsv1.Deployment
		scheme   *runtime.Scheme
	)
	BeforeEach(func() {
		scheme = gpuFractioningTestScheme()
		Expect(appsv1.AddToScheme(scheme)).To(Succeed())
		config = newGpuFractioningConfig()
		config.SetLabels(map[string]string{
			"app.kubernetes.io/name": "gpu-fractioning", "app.kubernetes.io/instance": "runai",
		})
		operator = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: "runai-gpu-fractioning", Namespace: "runtime", Labels: config.GetLabels(),
			},
			Spec: appsv1.DeploymentSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "manager", Env: []corev1.EnvVar{{Name: "SUPPORT_SM_SHARING", Value: "true"}}}},
			}}},
		}
	})
	checkUnavailable := func(objects ...client.Object) {
		k8s := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
		status := CheckGpuComputeModes(context.Background(), k8s)
		Expect(*status.Available).To(BeFalse())
		Expect(status.Reason).NotTo(BeEmpty())
		for _, mode := range []string{"time-slicing", "sm-sharing"} {
			Expect(*status.Components[mode].Available).To(BeFalse())
			Expect(status.Components[mode].Reason).NotTo(BeEmpty())
		}
	}
	checkSupported := func(objects ...client.Object) {
		k8s := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
		status := CheckGpuComputeModes(context.Background(), k8s)
		Expect(*status.Available).To(BeTrue())
		Expect(status.Reason).To(BeEmpty())
		for _, mode := range []string{"time-slicing", "sm-sharing"} {
			Expect(*status.Components[mode].Available).To(BeTrue())
			Expect(status.Components[mode].Reason).To(BeEmpty())
		}
	}

	It("reports both configured modes before any daemon or readiness status exists", func() {
		checkSupported(config, operator)
	})

	DescribeTable("keeps time-slicing available when SM-sharing is disabled or unknown",
		func(env []corev1.EnvVar, reason string) {
			operator.Spec.Template.Spec.Containers[0].Env = env
			k8s := fake.NewClientBuilder().WithScheme(scheme).WithObjects(config, operator).Build()
			status := CheckGpuComputeModes(context.Background(), k8s)
			Expect(*status.Available).To(BeTrue())
			Expect(*status.Components["time-slicing"].Available).To(BeTrue())
			Expect(*status.Components["sm-sharing"].Available).To(BeFalse())
			Expect(status.Components["sm-sharing"].Reason).To(ContainSubstring(reason))
		},
		Entry("disabled", []corev1.EnvVar{{Name: "SUPPORT_SM_SHARING", Value: "false"}}, "disabled"),
		Entry("missing", nil, "cannot be determined"),
		Entry("empty", []corev1.EnvVar{{Name: "SUPPORT_SM_SHARING"}}, "invalid"),
		Entry("invalid", []corev1.EnvVar{{Name: "SUPPORT_SM_SHARING", Value: "yes"}}, "invalid"),
		Entry("conflicting settings", []corev1.EnvVar{{Name: "SUPPORT_SM_SHARING", Value: "true"}, {Name: "SUPPORT_SM_SHARING", Value: "false"}}, "ambiguous"),
		Entry("duplicate settings", []corev1.EnvVar{{Name: "SUPPORT_SM_SHARING", Value: "true"}, {Name: "SUPPORT_SM_SHARING", Value: "true"}}, "ambiguous"),
		Entry("indirect setting", []corev1.EnvVar{{Name: "SUPPORT_SM_SHARING", ValueFrom: &corev1.EnvVarSource{
			ConfigMapKeyRef: &corev1.ConfigMapKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "runtime-settings"}, Key: "enabled"},
		}}}, "cannot be determined"),
		Entry("literal with indirect setting", []corev1.EnvVar{{Name: "SUPPORT_SM_SHARING", Value: "true", ValueFrom: &corev1.EnvVarSource{}}}, "cannot be determined"),
	)

	It("reports configured support without runtime readiness or DaemonSet reads", func() {
		Expect(unstructured.SetNestedSlice(config.Object, []interface{}{
			map[string]interface{}{"type": "Ready", "status": "False", "observedGeneration": int64(1)},
		}, "status", "conditions")).To(Succeed())
		config.SetGeneration(2)
		operator.Spec.Replicas = ptr.To(int32(0))
		operator.Generation = 3
		operator.Status = appsv1.DeploymentStatus{ObservedGeneration: 1, UnavailableReplicas: 1}
		k8s := fake.NewClientBuilder().WithScheme(scheme).WithObjects(config, operator).WithInterceptorFuncs(interceptor.Funcs{
			List: func(ctx context.Context, reader client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
				if _, ok := list.(*appsv1.DaemonSetList); ok {
					Fail("mode discovery must not inspect DaemonSets")
				}
				return reader.List(ctx, list, opts...)
			},
		}).Build()
		status := CheckGpuComputeModes(context.Background(), k8s)
		Expect(*status.Components["sm-sharing"].Available).To(BeTrue())
	})
	It("rejects a missing runtime without changing legacy GPU-stack availability", func() {
		k8s := fake.NewClientBuilder().WithScheme(scheme).Build()
		Expect(*CheckGpuFractioningReadiness(context.Background(), k8s).Available).To(BeTrue())
		checkUnavailable()
	})
	It("rejects a runtime being deleted", func() {
		config.SetFinalizers([]string{"cleanup"})
		config.SetDeletionTimestamp(ptr.To(metav1.Now()))
		checkUnavailable(config, operator)
	})
	DescribeTable("requires a compatible annotation prefix", func(prefix interface{}) {
		Expect(unstructured.SetNestedField(config.Object, prefix, "spec", "fractioningAgent", "annotationPrefix")).To(Succeed())
		checkUnavailable(config, operator)
	}, Entry("custom", "example.com/container."), Entry("malformed", int64(7)))
	It("accepts an explicit standard annotation prefix", func() {
		Expect(unstructured.SetNestedField(config.Object, "nvidia.com/container.", "spec", "fractioningAgent", "annotationPrefix")).To(Succeed())
		checkSupported(config, operator)
	})
	for _, label := range []string{"app.kubernetes.io/name", "app.kubernetes.io/instance"} {
		It("rejects configuration without "+label, func() {
			labels := config.GetLabels()
			delete(labels, label)
			config.SetLabels(labels)
			checkUnavailable(config, operator)
		})
		It("does not discover an operator from a different "+label, func() {
			operator.Labels = map[string]string{
				"app.kubernetes.io/name": "gpu-fractioning", "app.kubernetes.io/instance": "runai",
			}
			operator.Labels[label] = "other"
			checkUnavailable(config, operator)
		})
	}
	It("discovers chart name and release overrides without assuming deployment name or namespace", func() {
		labels := map[string]string{"app.kubernetes.io/name": "custom-fractions", "app.kubernetes.io/instance": "custom-release"}
		config.SetLabels(labels)
		operator.Labels = labels
		operator.Name = "custom-fullname"
		operator.Namespace = "custom-namespace"
		checkSupported(config, operator)
	})
	It("ignores deployments from unrelated releases", func() {
		unrelated := operator.DeepCopy()
		unrelated.Name = "other-release"
		unrelated.Labels["app.kubernetes.io/instance"] = "other"
		checkSupported(config, operator, unrelated)
	})
	It("reads the manager setting rather than a sidecar's setting", func() {
		operator.Spec.Template.Spec.Containers = append([]corev1.Container{{
			Name: "sidecar", Env: []corev1.EnvVar{{Name: "SUPPORT_SM_SHARING", Value: "false"}},
		}}, operator.Spec.Template.Spec.Containers...)
		checkSupported(config, operator)
	})
	It("requires the operator Deployment", func() { checkUnavailable(config) })
	It("rejects an operator being deleted", func() {
		operator.Finalizers = []string{"cleanup"}
		operator.DeletionTimestamp = ptr.To(metav1.Now())
		checkUnavailable(config, operator)
	})
	It("requires the operator manager container", func() {
		operator.Spec.Template.Spec.Containers[0].Name = "other"
		checkUnavailable(config, operator)
	})
	It("rejects ambiguous installations across namespaces", func() {
		duplicate := operator.DeepCopy()
		duplicate.Namespace = "other-runtime"
		checkUnavailable(config, operator, duplicate)
	})
	DescribeTable("does not treat API read errors as support", func(operation string) {
		funcs := interceptor.Funcs{}
		if operation == "get" {
			funcs.Get = func(context.Context, client.WithWatch, client.ObjectKey, client.Object, ...client.GetOption) error {
				return errors.New("forbidden")
			}
		} else {
			funcs.List = func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
				return errors.New("forbidden")
			}
		}
		k8s := fake.NewClientBuilder().WithScheme(scheme).WithObjects(config, operator).WithInterceptorFuncs(funcs).Build()
		status := CheckGpuComputeModes(context.Background(), k8s)
		Expect(*status.Available).To(BeFalse())
		Expect(status.Reason).To(ContainSubstring("forbidden"))
	}, Entry("configuration", "get"), Entry("operator", "list"))
})
