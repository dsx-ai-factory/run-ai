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
	"k8s.io/apimachinery/pkg/types"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

const (
	podTestNamespace = "dra-pod-ns"
	podTestName      = "p"
	podTestUID       = types.UID("pod-uid")
)

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

// testGeneratedClaim is a claim created from a template for the Pod with the given UID.
func testGeneratedClaim(name string, count int64, ownerPodUID types.UID) *resourcev1.ResourceClaim {
	claim := testClaim(name, count)
	claim.OwnerReferences = []metav1.OwnerReference{{
		APIVersion: "v1", Kind: "Pod", Name: podTestName, UID: ownerPodUID, Controller: ptr.To(true),
	}}
	return claim
}

func testTemplate(name string, count int64) *resourcev1.ResourceClaimTemplate {
	return &resourcev1.ResourceClaimTemplate{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: podTestNamespace},
		Spec:       resourcev1.ResourceClaimTemplateSpec{Spec: gpuClaimSpec(nvidiaGpuDeviceClassName, count)},
	}
}

// newPodTestParser builds a parser over a fake cache; the first *failures Get calls fail with a temporary error.
func newPodTestParser(failures *int, objs ...client.Object) *DRAParser {
	cachedClient := fake.NewClientBuilder().WithScheme(testScheme()).WithObjects(objs...).
		WithInterceptorFuncs(interceptor.Funcs{
			Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
				if failures != nil && *failures > 0 {
					*failures--
					return errors.New("temporary read failure")
				}
				return c.Get(ctx, key, obj, opts...)
			},
		}).Build()
	return NewDRAParser(cachedClient, kubefake.NewClientset())
}

func newPodWithClaims(claimStatuses []corev1.PodResourceClaimStatus, claimRefs ...corev1.PodResourceClaim) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: podTestName, Namespace: podTestNamespace, UID: podTestUID},
		Spec:       corev1.PodSpec{ResourceClaims: claimRefs},
		Status:     corev1.PodStatus{ResourceClaimStatuses: claimStatuses},
	}
}

func TestResolvePodRequestedGpuCount(t *testing.T) {
	ctx := context.Background()
	// The generated claim deliberately differs from its template so the tests show which one was read.
	parser := newPodTestParser(nil,
		testClaim("existing", 4),
		testTemplate("tmpl", 2),
		testGeneratedClaim("p-gpu-abcde", 3, podTestUID),
		testGeneratedClaim("p-gpu-foreign", 3, types.UID("another-pod-uid")),
		&resourcev1.ResourceClaim{
			ObjectMeta: metav1.ObjectMeta{Name: "other-class", Namespace: podTestNamespace},
			Spec:       gpuClaimSpec("other.vendor.com", 8),
		},
	)

	existingClaimRef := corev1.PodResourceClaim{Name: "existing", ResourceClaimName: ptr.To("existing")}
	templateRef := corev1.PodResourceClaim{Name: "gpu", ResourceClaimTemplateName: ptr.To("tmpl")}
	generatedClaimStatuses := []corev1.PodResourceClaimStatus{{Name: "gpu", ResourceClaimName: ptr.To("p-gpu-abcde")}}
	missingGeneratedClaimStatuses := []corev1.PodResourceClaimStatus{{Name: "gpu", ResourceClaimName: ptr.To("not-created-yet")}}
	foreignGeneratedClaimStatuses := []corev1.PodResourceClaimStatus{{Name: "gpu", ResourceClaimName: ptr.To("p-gpu-foreign")}}
	skippedClaimStatuses := []corev1.PodResourceClaimStatus{{Name: "gpu", ResourceClaimName: nil}}

	tests := []struct {
		name    string
		pod     *corev1.Pod
		want    int64
		wantErr bool
	}{
		{"Pod references an existing four-GPU claim: returns four requested GPUs",
			newPodWithClaims(nil, existingClaimRef), 4, false},
		{"Pod references a two-GPU template without generated-claim status: returns two requested GPUs",
			newPodWithClaims(nil, templateRef), 2, false},
		{"Pod identifies a generated three-GPU claim: returns three GPUs instead of the template's two",
			newPodWithClaims(generatedClaimStatuses, templateRef), 3, false},
		{"Pod identifies a generated claim that cannot be found: returns an error",
			newPodWithClaims(missingGeneratedClaimStatuses, templateRef), 0, true},
		{"Pod identifies a generated claim owned by another Pod: returns an error",
			newPodWithClaims(foreignGeneratedClaimStatuses, templateRef), 0, true},
		{"Generated-claim status has a nil claim name: returns zero GPUs without an error",
			newPodWithClaims(skippedClaimStatuses, templateRef), 0, false},
		{"Pod references an independent four-GPU claim and two-GPU template: returns six requested GPUs",
			newPodWithClaims(nil, existingClaimRef, templateRef), 6, false},
		{"Two Pod references name the same four-GPU claim: returns four GPUs, counting the claim once",
			newPodWithClaims(nil, existingClaimRef, corev1.PodResourceClaim{Name: "again", ResourceClaimName: ptr.To("existing")}), 4, false},
		{"Claim requests devices from an unsupported DeviceClass: returns zero GPUs without an error",
			newPodWithClaims(nil, corev1.PodResourceClaim{Name: "x", ResourceClaimName: ptr.To("other-class")}), 0, false},
		{"Pod references an existing claim that cannot be found: returns an error",
			newPodWithClaims(nil, corev1.PodResourceClaim{Name: "x", ResourceClaimName: ptr.To("missing")}), 0, true},
		{"Pod references a template that cannot be found: returns an error",
			newPodWithClaims(nil, corev1.PodResourceClaim{Name: "x", ResourceClaimTemplateName: ptr.To("missing")}), 0, true},
		{"Pod reference sets neither a claim nor a template: returns an error",
			newPodWithClaims(nil, corev1.PodResourceClaim{Name: "x"}), 0, true},
		{"Pod has no claim references: returns zero GPUs without an error",
			newPodWithClaims(nil), 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ResolvePodRequestedGpuCount(ctx, tt.pod)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ResolvePodRequestedGpuCount() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ResolvePodRequestedGpuCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestResolvePodRequestedGpuCountRecoversFromTemporaryReadFailure(t *testing.T) {
	ctx := context.Background()
	failures := 1
	parser := newPodTestParser(&failures, testTemplate("tmpl", 2))
	pod := newPodWithClaims(nil, corev1.PodResourceClaim{Name: "gpu", ResourceClaimTemplateName: ptr.To("tmpl")})

	if _, err := parser.ResolvePodRequestedGpuCount(ctx, pod); err == nil {
		t.Fatal("expected the temporary read failure to be returned, got nil")
	}
	got, err := parser.ResolvePodRequestedGpuCount(ctx, pod)
	if err != nil {
		t.Fatalf("unexpected error after recovery: %v", err)
	}
	if got != 2 {
		t.Errorf("ResolvePodRequestedGpuCount() = %d, want 2", got)
	}
}

func TestResolvePodRequestedGpuCountRecoversWhenGeneratedClaimReachesCache(t *testing.T) {
	ctx := context.Background()
	cachedClient := fake.NewClientBuilder().WithScheme(testScheme()).WithObjects(testTemplate("tmpl", 2)).Build()
	parser := NewDRAParser(cachedClient, kubefake.NewClientset())
	// The Pod already names its generated claim, but the claim has not reached the cache yet.
	pod := newPodWithClaims(
		[]corev1.PodResourceClaimStatus{{Name: "gpu", ResourceClaimName: ptr.To("p-gpu-abcde")}},
		corev1.PodResourceClaim{Name: "gpu", ResourceClaimTemplateName: ptr.To("tmpl")})

	if _, err := parser.ResolvePodRequestedGpuCount(ctx, pod); err == nil {
		t.Fatal("expected an error while the generated claim is missing from the cache, got nil")
	}

	if err := cachedClient.Create(ctx, testGeneratedClaim("p-gpu-abcde", 3, podTestUID)); err != nil {
		t.Fatal(err)
	}
	got, err := parser.ResolvePodRequestedGpuCount(ctx, pod)
	if err != nil {
		t.Fatalf("unexpected error after the claim reached the cache: %v", err)
	}
	if got != 3 {
		t.Errorf("ResolvePodRequestedGpuCount() = %d, want 3", got)
	}
}
