/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dra

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	resourcev1beta1 "k8s.io/api/resource/v1beta1"
	resourcev1beta2 "k8s.io/api/resource/v1beta2"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func testScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = corev1.AddToScheme(s)
	_ = resourcev1.AddToScheme(s)
	_ = resourcev1beta1.AddToScheme(s)
	_ = resourcev1beta2.AddToScheme(s)
	return s
}

func TestExtractGpuCountFromClaimSpec(t *testing.T) {
	tests := []struct {
		name     string
		spec     resourcev1.ResourceClaimSpec
		expected int64
	}{
		{
			name: "ExactCount with 4 NVIDIA GPUs",
			spec: resourcev1.ResourceClaimSpec{
				Devices: resourcev1.DeviceClaim{
					Requests: []resourcev1.DeviceRequest{
						{
							Name: "gpu",
							Exactly: &resourcev1.ExactDeviceRequest{
								DeviceClassName: nvidiaGpuDeviceClassName,
								AllocationMode:  resourcev1.DeviceAllocationModeExactCount,
								Count:           4,
							},
						},
					},
				},
			},
			expected: 4,
		},
		{
			name: "ExactCount with count 0 is skipped",
			spec: resourcev1.ResourceClaimSpec{
				Devices: resourcev1.DeviceClaim{
					Requests: []resourcev1.DeviceRequest{
						{
							Name: "gpu",
							Exactly: &resourcev1.ExactDeviceRequest{
								DeviceClassName: nvidiaGpuDeviceClassName,
								AllocationMode:  resourcev1.DeviceAllocationModeExactCount,
								Count:           0,
							},
						},
					},
				},
			},
			expected: 0,
		},
		{
			name: "Default AllocationMode (empty) treated as ExactCount",
			spec: resourcev1.ResourceClaimSpec{
				Devices: resourcev1.DeviceClaim{
					Requests: []resourcev1.DeviceRequest{
						{
							Name: "gpu",
							Exactly: &resourcev1.ExactDeviceRequest{
								DeviceClassName: nvidiaGpuDeviceClassName,
								Count:           2,
							},
						},
					},
				},
			},
			expected: 2,
		},
		{
			name: "AllocationMode All is unsupported",
			spec: resourcev1.ResourceClaimSpec{
				Devices: resourcev1.DeviceClaim{
					Requests: []resourcev1.DeviceRequest{
						{
							Name: "gpu",
							Exactly: &resourcev1.ExactDeviceRequest{
								DeviceClassName: nvidiaGpuDeviceClassName,
								AllocationMode:  resourcev1.DeviceAllocationModeAll,
							},
						},
					},
				},
			},
			expected: 0,
		},
		{
			name: "FirstAvailable (no Exactly) returns 0",
			spec: resourcev1.ResourceClaimSpec{
				Devices: resourcev1.DeviceClaim{
					Requests: []resourcev1.DeviceRequest{
						{
							Name:           "gpu",
							FirstAvailable: []resourcev1.DeviceSubRequest{},
						},
					},
				},
			},
			expected: 0,
		},
		{
			name: "Non-NVIDIA device class returns 0",
			spec: resourcev1.ResourceClaimSpec{
				Devices: resourcev1.DeviceClaim{
					Requests: []resourcev1.DeviceRequest{
						{
							Name: "gpu",
							Exactly: &resourcev1.ExactDeviceRequest{
								DeviceClassName: "gpu.amd.com",
								AllocationMode:  resourcev1.DeviceAllocationModeExactCount,
								Count:           2,
							},
						},
					},
				},
			},
			expected: 0,
		},
		{
			name: "Multiple requests sums NVIDIA ones only",
			spec: resourcev1.ResourceClaimSpec{
				Devices: resourcev1.DeviceClaim{
					Requests: []resourcev1.DeviceRequest{
						{
							Name: "nvidia-gpu",
							Exactly: &resourcev1.ExactDeviceRequest{
								DeviceClassName: nvidiaGpuDeviceClassName,
								AllocationMode:  resourcev1.DeviceAllocationModeExactCount,
								Count:           2,
							},
						},
						{
							Name: "amd-gpu",
							Exactly: &resourcev1.ExactDeviceRequest{
								DeviceClassName: "gpu.amd.com",
								AllocationMode:  resourcev1.DeviceAllocationModeExactCount,
								Count:           3,
							},
						},
						{
							Name: "more-nvidia",
							Exactly: &resourcev1.ExactDeviceRequest{
								DeviceClassName: nvidiaGpuDeviceClassName,
								AllocationMode:  resourcev1.DeviceAllocationModeExactCount,
								Count:           1,
							},
						},
					},
				},
			},
			expected: 3,
		},
		{
			name:     "Empty spec returns 0",
			spec:     resourcev1.ResourceClaimSpec{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractGpuCountFromClaimSpec(tt.spec)
			if result != tt.expected {
				t.Errorf("extractGpuCountFromClaimSpec() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestResolveDRAGpuCount(t *testing.T) {
	namespace := "test-ns"

	claim := &resourcev1.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-gpu-claim",
			Namespace: namespace,
		},
		Spec: resourcev1.ResourceClaimSpec{
			Devices: resourcev1.DeviceClaim{
				Requests: []resourcev1.DeviceRequest{
					{
						Name: "gpu",
						Exactly: &resourcev1.ExactDeviceRequest{
							DeviceClassName: nvidiaGpuDeviceClassName,
							AllocationMode:  resourcev1.DeviceAllocationModeExactCount,
							Count:           4,
						},
					},
				},
			},
		},
	}

	template := &resourcev1.ResourceClaimTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-gpu-template",
			Namespace: namespace,
		},
		Spec: resourcev1.ResourceClaimTemplateSpec{
			Spec: resourcev1.ResourceClaimSpec{
				Devices: resourcev1.DeviceClaim{
					Requests: []resourcev1.DeviceRequest{
						{
							Name: "gpu",
							Exactly: &resourcev1.ExactDeviceRequest{
								DeviceClassName: nvidiaGpuDeviceClassName,
								AllocationMode:  resourcev1.DeviceAllocationModeExactCount,
								Count:           2,
							},
						},
					},
				},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(testScheme()).WithObjects(claim, template).Build()
	fakeKubeClient := kubefake.NewClientset(claim, template)
	parser := NewDRAParser(fakeClient, fakeKubeClient)

	ctx := context.Background()

	t.Run("ResourceClaimName reference", func(t *testing.T) {
		podClaims := []corev1.PodResourceClaim{
			{
				Name:              "gpu-claim",
				ResourceClaimName: ptr.To("my-gpu-claim"),
			},
		}
		result := parser.ResolveDRAGpuCount(ctx, namespace, podClaims)
		if result != 4 {
			t.Errorf("ResolveDRAGpuCount() = %d, want 4", result)
		}
	})

	t.Run("ResourceClaimTemplateName reference", func(t *testing.T) {
		podClaims := []corev1.PodResourceClaim{
			{
				Name:                      "gpu-template",
				ResourceClaimTemplateName: ptr.To("my-gpu-template"),
			},
		}
		result := parser.ResolveDRAGpuCount(ctx, namespace, podClaims)
		if result != 2 {
			t.Errorf("ResolveDRAGpuCount() = %d, want 2", result)
		}
	})

	t.Run("Claim not found returns 0", func(t *testing.T) {
		podClaims := []corev1.PodResourceClaim{
			{
				Name:              "missing",
				ResourceClaimName: ptr.To("nonexistent-claim"),
			},
		}
		result := parser.ResolveDRAGpuCount(ctx, namespace, podClaims)
		if result != 0 {
			t.Errorf("ResolveDRAGpuCount() = %d, want 0", result)
		}
	})

	t.Run("Multiple claims summed", func(t *testing.T) {
		podClaims := []corev1.PodResourceClaim{
			{
				Name:              "claim1",
				ResourceClaimName: ptr.To("my-gpu-claim"),
			},
			{
				Name:                      "claim2",
				ResourceClaimTemplateName: ptr.To("my-gpu-template"),
			},
		}
		result := parser.ResolveDRAGpuCount(ctx, namespace, podClaims)
		if result != 6 {
			t.Errorf("ResolveDRAGpuCount() = %d, want 6", result)
		}
	})

	t.Run("Empty claims returns 0", func(t *testing.T) {
		result := parser.ResolveDRAGpuCount(ctx, namespace, nil)
		if result != 0 {
			t.Errorf("ResolveDRAGpuCount() = %d, want 0", result)
		}
	})
}

func TestAddDRAGpuToResources(t *testing.T) {
	t.Run("Adds GPU to empty resources", func(t *testing.T) {
		resources := &corev1.ResourceRequirements{}
		AddDRAGpuToResources(4, resources)

		expectedQty := resource.NewQuantity(4, resource.DecimalSI)
		if qty, ok := resources.Requests[NvidiaGpuResourceName]; !ok || !qty.Equal(*expectedQty) {
			t.Errorf("Requests GPU = %v, want %v", qty, expectedQty)
		}
		if qty, ok := resources.Limits[NvidiaGpuResourceName]; !ok || !qty.Equal(*expectedQty) {
			t.Errorf("Limits GPU = %v, want %v", qty, expectedQty)
		}
	})

	t.Run("Adds GPU to existing resources", func(t *testing.T) {
		resources := &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				NvidiaGpuResourceName: *resource.NewQuantity(2, resource.DecimalSI),
			},
			Limits: corev1.ResourceList{
				NvidiaGpuResourceName: *resource.NewQuantity(2, resource.DecimalSI),
			},
		}
		AddDRAGpuToResources(3, resources)

		expectedQty := resource.NewQuantity(5, resource.DecimalSI)
		if qty := resources.Requests[NvidiaGpuResourceName]; !qty.Equal(*expectedQty) {
			t.Errorf("Requests GPU = %v, want %v", qty, expectedQty)
		}
		if qty := resources.Limits[NvidiaGpuResourceName]; !qty.Equal(*expectedQty) {
			t.Errorf("Limits GPU = %v, want %v", qty, expectedQty)
		}
	})

	t.Run("Zero GPU count does nothing", func(t *testing.T) {
		resources := &corev1.ResourceRequirements{}
		AddDRAGpuToResources(0, resources)
		if len(resources.Requests) != 0 || len(resources.Limits) != 0 {
			t.Errorf("Expected empty resources for 0 GPU count")
		}
	})

	t.Run("Negative GPU count does nothing", func(t *testing.T) {
		resources := &corev1.ResourceRequirements{}
		AddDRAGpuToResources(-1, resources)
		if len(resources.Requests) != 0 || len(resources.Limits) != 0 {
			t.Errorf("Expected empty resources for negative GPU count")
		}
	})
}
