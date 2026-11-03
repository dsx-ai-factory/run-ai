/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package data

import (
	"context"
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	resourcev1beta1 "k8s.io/api/resource/v1beta1"
	resourcev1beta2 "k8s.io/api/resource/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra"
)

const draTestNamespace = "dra-ns"

func newDRATestParser(failures *int, gpuCount int64) *dra.DRAParser {
	template := &resourcev1.ResourceClaimTemplate{
		ObjectMeta: metav1.ObjectMeta{Name: "gpu-template", Namespace: draTestNamespace},
		Spec: resourcev1.ResourceClaimTemplateSpec{Spec: resourcev1.ResourceClaimSpec{
			Devices: resourcev1.DeviceClaim{Requests: []resourcev1.DeviceRequest{{
				Name: "gpu",
				Exactly: &resourcev1.ExactDeviceRequest{
					DeviceClassName: "gpu.nvidia.com",
					AllocationMode:  resourcev1.DeviceAllocationModeExactCount,
					Count:           gpuCount,
				},
			}}},
		}},
	}
	s := runtime.NewScheme()
	_ = corev1.AddToScheme(s)
	_ = resourcev1.AddToScheme(s)
	_ = resourcev1beta1.AddToScheme(s)
	_ = resourcev1beta2.AddToScheme(s)
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(template).
		WithInterceptorFuncs(interceptor.Funcs{
			Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
				if failures != nil && *failures > 0 {
					*failures--
					return errors.New("temporary read failure")
				}
				return c.Get(ctx, key, obj, opts...)
			},
		}).Build()
	return dra.NewDRAParser(c, kubefake.NewClientset())
}

func newDRATestPod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "p", Namespace: draTestNamespace},
		Spec: corev1.PodSpec{ResourceClaims: []corev1.PodResourceClaim{{
			Name: "gpu", ResourceClaimTemplateName: ptr.To("gpu-template"),
		}}},
	}
}

func TestAddDRAGpus(t *testing.T) {
	ctx := context.Background()

	t.Run("DRA pod: requested and limit report the claim's GPUs, GPU memory is untouched", func(t *testing.T) {
		req := &ResourcesInfo{GpuMemoryMegaBytes: 500}
		lim := &ResourcesInfo{GpuMemoryMegaBytes: 700}
		if err := AddDRAGpus(ctx, newDRATestPod(), newDRATestParser(nil, 2), req, lim); err != nil {
			t.Fatal(err)
		}
		if req.Gpus != 2 || lim.Gpus != 2 || req.GpuMemoryMegaBytes != 500 || lim.GpuMemoryMegaBytes != 700 {
			t.Fatalf("got req=%+v lim=%+v", *req, *lim)
		}
	})

	t.Run("pod with ordinary GPUs and a claim: each independent request counted once", func(t *testing.T) {
		req, lim := &ResourcesInfo{Gpus: 1}, &ResourcesInfo{Gpus: 1}
		if err := AddDRAGpus(ctx, newDRATestPod(), newDRATestParser(nil, 2), req, lim); err != nil {
			t.Fatal(err)
		}
		if req.Gpus != 3 || lim.Gpus != 3 {
			t.Fatalf("got req=%v lim=%v, want 3/3", req.Gpus, lim.Gpus)
		}
	})

	t.Run("traditional GPU pod (no claims) is unchanged", func(t *testing.T) {
		req, lim := &ResourcesInfo{Gpus: 1}, &ResourcesInfo{Gpus: 1}
		if err := AddDRAGpus(ctx, &corev1.Pod{}, newDRATestParser(nil, 2), req, lim); err != nil {
			t.Fatal(err)
		}
		if req.Gpus != 1 || lim.Gpus != 1 {
			t.Fatalf("got req=%v lim=%v, want 1/1", req.Gpus, lim.Gpus)
		}
	})

	t.Run("nil parser is a no-op", func(t *testing.T) {
		req, lim := &ResourcesInfo{}, &ResourcesInfo{}
		if err := AddDRAGpus(ctx, newDRATestPod(), nil, req, lim); err != nil {
			t.Fatal(err)
		}
		if req.Gpus != 0 || lim.Gpus != 0 {
			t.Fatalf("got req=%v lim=%v, want 0/0", req.Gpus, lim.Gpus)
		}
	})

	t.Run("temporary read failure is returned, then the retry reports the GPUs", func(t *testing.T) {
		failures := 1
		parser := newDRATestParser(&failures, 2)
		req, lim := &ResourcesInfo{}, &ResourcesInfo{}
		if err := AddDRAGpus(ctx, newDRATestPod(), parser, req, lim); err == nil {
			t.Fatal("expected an error on the first attempt")
		}
		if req.Gpus != 0 || lim.Gpus != 0 {
			t.Fatalf("failed attempt must not change the resources, got req=%v lim=%v", req.Gpus, lim.Gpus)
		}
		if err := AddDRAGpus(ctx, newDRATestPod(), parser, req, lim); err != nil {
			t.Fatal(err)
		}
		if req.Gpus != 2 || lim.Gpus != 2 {
			t.Fatalf("got req=%v lim=%v, want 2/2", req.Gpus, lim.Gpus)
		}
	})
}

func TestDRAPodAllocationComesFromAnnotations(t *testing.T) {
	pod := newDRATestPod()
	pod.Status.Phase = corev1.PodRunning
	pod.Spec.NodeName = "node-1"

	if got := GetPodAllocatedResources(pod, GetPodDetailedPhase(pod)).Gpus; got != 0 {
		t.Fatalf("allocated GPUs without the status-updater annotation = %v, want 0", got)
	}

	pod.Annotations = map[string]string{clustercommonconstants.PodAllocatedGpusAnnotation: "2"}
	if got := GetPodAllocatedResources(pod, GetPodDetailedPhase(pod)).Gpus; got != 2 {
		t.Fatalf("allocated GPUs with annotation = %v, want 2", got)
	}

	pod.Annotations[clustercommonconstants.PodAllocatedGpusAnnotation] = "3"
	if got := GetPodAllocatedResources(pod, GetPodDetailedPhase(pod)).Gpus; got != 3 {
		t.Fatalf("allocated GPUs after annotation update = %v, want 3", got)
	}
}
