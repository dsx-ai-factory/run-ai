/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dra

import (
	"context"
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

const podTestNamespace = "dra-pod-ns"

func gpuClaimSpec(deviceClass string, count int64) resourcev1.ResourceClaimSpec {
	return resourcev1.ResourceClaimSpec{Devices: resourcev1.DeviceClaim{Requests: []resourcev1.DeviceRequest{{
		Name: "gpu",
		Exactly: &resourcev1.ExactDeviceRequest{
			DeviceClassName: deviceClass,
			AllocationMode:  resourcev1.DeviceAllocationModeExactCount,
			Count:           count,
		},
	}}}}
}

func testClaim(name string, count int64) *resourcev1.ResourceClaim {
	return &resourcev1.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: podTestNamespace},
		Spec:       gpuClaimSpec(nvidiaGpuDeviceClassName, count),
	}
}

func testTemplate(name string, count int64) *resourcev1.ResourceClaimTemplate {
	return &resourcev1.ResourceClaimTemplate{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: podTestNamespace},
		Spec:       resourcev1.ResourceClaimTemplateSpec{Spec: gpuClaimSpec(nvidiaGpuDeviceClassName, count)},
	}
}

// newPodTestParser builds a parser over a fake cache; the first *failures Get calls fail with a temporary error.
func newPodTestParser(failures *int, objs ...client.Object) *DRAParser {
	c := fake.NewClientBuilder().WithScheme(testScheme()).WithObjects(objs...).
		WithInterceptorFuncs(interceptor.Funcs{
			Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
				if failures != nil && *failures > 0 {
					*failures--
					return errors.New("temporary read failure")
				}
				return c.Get(ctx, key, obj, opts...)
			},
		}).Build()
	return NewDRAParser(c, kubefake.NewClientset())
}

func podWithClaims(statuses []corev1.PodResourceClaimStatus, claims ...corev1.PodResourceClaim) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "p", Namespace: podTestNamespace},
		Spec:       corev1.PodSpec{ResourceClaims: claims},
		Status:     corev1.PodStatus{ResourceClaimStatuses: statuses},
	}
}

func TestResolvePodGpuCount(t *testing.T) {
	ctx := context.Background()
	// The generated claim deliberately differs from its template so the tests show which one was read.
	parser := newPodTestParser(nil,
		testClaim("existing", 4),
		testTemplate("tmpl", 2),
		testClaim("p-gpu-abcde", 3),
		&resourcev1.ResourceClaim{
			ObjectMeta: metav1.ObjectMeta{Name: "other-class", Namespace: podTestNamespace},
			Spec:       gpuClaimSpec("other.vendor.com", 8),
		},
	)

	existing := corev1.PodResourceClaim{Name: "existing", ResourceClaimName: ptr.To("existing")}
	fromTemplate := corev1.PodResourceClaim{Name: "gpu", ResourceClaimTemplateName: ptr.To("tmpl")}
	generatedStatus := []corev1.PodResourceClaimStatus{{Name: "gpu", ResourceClaimName: ptr.To("p-gpu-abcde")}}
	missingGeneratedStatus := []corev1.PodResourceClaimStatus{{Name: "gpu", ResourceClaimName: ptr.To("not-created-yet")}}

	tests := []struct {
		name    string
		pod     *corev1.Pod
		want    int64
		wantErr bool
	}{
		{"existing claim", podWithClaims(nil, existing), 4, false},
		{"template of an unscheduled pod (no generated claim yet)", podWithClaims(nil, fromTemplate), 2, false},
		{"template with generated claim: claim is used and counted once", podWithClaims(generatedStatus, fromTemplate), 3, false},
		{"generated claim named in status but not readable: falls back to template", podWithClaims(missingGeneratedStatus, fromTemplate), 2, false},
		{"claim and template are independent requests", podWithClaims(nil, existing, fromTemplate), 6, false},
		{"same claim referenced twice is counted once", podWithClaims(nil, existing, corev1.PodResourceClaim{Name: "again", ResourceClaimName: ptr.To("existing")}), 4, false},
		{"non-nvidia device class is not counted", podWithClaims(nil, corev1.PodResourceClaim{Name: "x", ResourceClaimName: ptr.To("other-class")}), 0, false},
		{"missing explicit claim is an error", podWithClaims(nil, corev1.PodResourceClaim{Name: "x", ResourceClaimName: ptr.To("missing")}), 0, true},
		{"missing template is an error", podWithClaims(nil, corev1.PodResourceClaim{Name: "x", ResourceClaimTemplateName: ptr.To("missing")}), 0, true},
		{"no claims", podWithClaims(nil), 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ResolvePodGpuCount(ctx, tt.pod)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ResolvePodGpuCount() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ResolvePodGpuCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestResolvePodGpuCountRecoversFromTemporaryReadFailure(t *testing.T) {
	ctx := context.Background()
	failures := 1
	parser := newPodTestParser(&failures, testTemplate("tmpl", 2))
	pod := podWithClaims(nil, corev1.PodResourceClaim{Name: "gpu", ResourceClaimTemplateName: ptr.To("tmpl")})

	if _, err := parser.ResolvePodGpuCount(ctx, pod); err == nil {
		t.Fatal("expected the temporary read failure to be returned, got nil")
	}
	got, err := parser.ResolvePodGpuCount(ctx, pod)
	if err != nil {
		t.Fatalf("unexpected error after recovery: %v", err)
	}
	if got != 2 {
		t.Errorf("ResolvePodGpuCount() = %d, want 2", got)
	}
}
