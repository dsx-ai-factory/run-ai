/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller_utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	rclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func gpuFractioningTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	s.AddKnownTypeWithName(gpuFractioningConfigGVK, &unstructured.Unstructured{})
	s.AddKnownTypeWithName(
		gpuFractioningConfigGVK.GroupVersion().WithKind(gpuFractioningConfigGVK.Kind+"List"),
		&unstructured.UnstructuredList{},
	)
	return s
}

func newGpuFractioningConfig(conditions ...map[string]interface{}) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gpuFractioningConfigGVK)
	obj.SetName(gpuFractioningConfigName)
	if len(conditions) > 0 {
		raw := make([]interface{}, len(conditions))
		for i, c := range conditions {
			raw[i] = c
		}
		Expect(unstructured.SetNestedSlice(obj.Object, raw, "status", "conditions")).To(Succeed())
	}
	return obj
}

func readyCond(status, message string) map[string]interface{} {
	return map[string]interface{}{
		"type":    "Ready",
		"status":  status,
		"message": message,
	}
}

var _ = Describe("CheckGpuFractioningReadiness", func() {
	var k8s rclient.Client

	newClient := func(objs ...rclient.Object) rclient.Client {
		return fake.NewClientBuilder().WithScheme(gpuFractioningTestScheme()).WithObjects(objs...).Build()
	}

	It("reports available when the Ready condition is True", func() {
		obj := newGpuFractioningConfig(
			readyCond("True", "all components are healthy"),
		)
		k8s = newClient(obj)

		status := CheckGpuFractioningReadiness(apiAccessContext, k8s)
		Expect(*status.Available).To(BeTrue())
		Expect(status.Reason).To(BeEmpty())
	})

	It("reports unavailable with the condition's message when Ready is False", func() {
		obj := newGpuFractioningConfig(
			readyCond("False", "NVIDIA GPU Operator version v25.3.0 is below minimum supported version v26.7.1"),
		)
		k8s = newClient(obj)

		status := CheckGpuFractioningReadiness(apiAccessContext, k8s)
		Expect(*status.Available).To(BeFalse())
		Expect(status.Reason).To(Equal("NVIDIA GPU Operator version v25.3.0 is below minimum supported version v26.7.1"))
	})

	It("falls back to a non-empty reason when Ready is False with no message", func() {
		obj := newGpuFractioningConfig(
			readyCond("False", ""),
		)
		k8s = newClient(obj)

		status := CheckGpuFractioningReadiness(apiAccessContext, k8s)
		Expect(*status.Available).To(BeFalse())
		Expect(status.Reason).NotTo(BeEmpty())
	})

	It("reports available when Ready is Unknown (not itself blocking)", func() {
		obj := newGpuFractioningConfig(
			readyCond("Unknown", "one or more components have unknown status"),
		)
		k8s = newClient(obj)

		status := CheckGpuFractioningReadiness(apiAccessContext, k8s)
		Expect(*status.Available).To(BeTrue())
	})

	It("finds Ready among multiple conditions", func() {
		obj := newGpuFractioningConfig(
			map[string]interface{}{"type": "FractiondReady", "status": "True"},
			map[string]interface{}{"type": "MpsdReady", "status": "True"},
			map[string]interface{}{"type": "DriverUpgradeInProgress", "status": "False"},
			readyCond("False", "not ready: MpsdReady"),
		)
		k8s = newClient(obj)

		status := CheckGpuFractioningReadiness(apiAccessContext, k8s)
		Expect(*status.Available).To(BeFalse())
		Expect(status.Reason).To(Equal("not ready: MpsdReady"))
	})

	It("reports available when the CR has no conditions yet", func() {
		obj := newGpuFractioningConfig()
		k8s = newClient(obj)

		status := CheckGpuFractioningReadiness(apiAccessContext, k8s)
		Expect(*status.Available).To(BeTrue())
	})

	It("reports available when the CR has conditions but no Ready entry", func() {
		obj := newGpuFractioningConfig(
			map[string]interface{}{"type": "FractiondReady", "status": "False"},
		)
		k8s = newClient(obj)

		status := CheckGpuFractioningReadiness(apiAccessContext, k8s)
		Expect(*status.Available).To(BeTrue())
	})

	It("reports available when the Ready condition's status is an invalid value", func() {
		obj := newGpuFractioningConfig(
			readyCond("Degraded", "not a valid condition status"),
		)
		k8s = newClient(obj)

		status := CheckGpuFractioningReadiness(apiAccessContext, k8s)
		Expect(*status.Available).To(BeTrue())
	})

	It("reports available when the GpuFractioningConfig CR does not exist", func() {
		k8s = newClient()

		status := CheckGpuFractioningReadiness(apiAccessContext, k8s)
		Expect(*status.Available).To(BeTrue())
		Expect(status.Reason).To(BeEmpty())
	})
})

var _ = Describe("readyCondition", func() {
	It("skips malformed condition entries instead of failing", func() {
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(gpuFractioningConfigGVK)
		Expect(unstructured.SetNestedSlice(obj.Object, []interface{}{
			"not-a-map",
			readyCond("False", "boom"),
		}, "status", "conditions")).To(Succeed())

		status, message, found := readyCondition(obj)
		Expect(found).To(BeTrue())
		Expect(status).To(Equal("False"))
		Expect(message).To(Equal("boom"))
	})

	It("reports not found when status.conditions is absent", func() {
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(gpuFractioningConfigGVK)

		_, _, found := readyCondition(obj)
		Expect(found).To(BeFalse())
	})

	It("reports not found when status.conditions is empty", func() {
		obj := newGpuFractioningConfig()

		_, _, found := readyCondition(obj)
		Expect(found).To(BeFalse())
	})

	It("reports not found when the Ready condition's status is missing", func() {
		obj := newGpuFractioningConfig(
			map[string]interface{}{"type": "Ready", "message": "no status field"},
		)

		_, _, found := readyCondition(obj)
		Expect(found).To(BeFalse())
	})

	It("reports not found when the Ready condition's status is an invalid value", func() {
		obj := newGpuFractioningConfig(
			readyCond("Degraded", "not a valid condition status"),
		)

		_, _, found := readyCondition(obj)
		Expect(found).To(BeFalse())
	})
})
