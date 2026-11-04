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
	"k8s.io/apimachinery/pkg/api/resource"
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

// newDRATestParser builds a parser whose template requests gpuCount GPUs; the first *failures reads fail.
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
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = resourcev1.AddToScheme(scheme)
	_ = resourcev1beta1.AddToScheme(scheme)
	_ = resourcev1beta2.AddToScheme(scheme)
	cachedClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(template).
		WithInterceptorFuncs(interceptor.Funcs{
			Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
				if failures != nil && *failures > 0 {
					*failures--
					return errors.New("temporary read failure")
				}
				return c.Get(ctx, key, obj, opts...)
			},
		}).Build()
	return dra.NewDRAParser(cachedClient, kubefake.NewClientset())
}

// newDRATestPod references the template through the claim "gpu" and has no GPU in its spec.
func newDRATestPod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "p", Namespace: draTestNamespace},
		Spec: corev1.PodSpec{
			ResourceClaims: []corev1.PodResourceClaim{{Name: "gpu", ResourceClaimTemplateName: ptr.To("gpu-template")}},
			Containers: []corev1.Container{{
				Name:      "c",
				Resources: corev1.ResourceRequirements{Claims: []corev1.ResourceClaim{{Name: "gpu"}}},
			}},
		},
	}
}

// withOrdinaryGpu adds one GPU request to the pod spec, next to any DRA claim.
func withOrdinaryGpu(pod *corev1.Pod) *corev1.Pod {
	gpuName := corev1.ResourceName(clustercommonconstants.NvidiaGpuResourceName)
	pod.Spec.Containers[0].Resources.Requests = corev1.ResourceList{gpuName: resource.MustParse("1")}
	pod.Spec.Containers[0].Resources.Limits = corev1.ResourceList{gpuName: resource.MustParse("1")}
	return pod
}

func TestAddDRAGpus(t *testing.T) {
	ctx := context.Background()

	t.Run("Two-GPU template: sets requested GPUs and GPU limit to two, keeping existing memory values", func(t *testing.T) {
		requested := &ResourcesInfo{GpuMemoryMegaBytes: 500}
		limits := &ResourcesInfo{GpuMemoryMegaBytes: 700}
		if err := AddDRAGpus(ctx, newDRATestPod(), newDRATestParser(nil, 2), requested, limits); err != nil {
			t.Fatal(err)
		}
		if requested.Gpus != 2 || limits.Gpus != 2 || requested.GpuMemoryMegaBytes != 500 || limits.GpuMemoryMegaBytes != 700 {
			t.Fatalf("got requested=%+v limits=%+v", *requested, *limits)
		}
	})

	t.Run("Pod requests one ordinary GPU and two independent DRA GPUs: reports three requested and limit GPUs", func(t *testing.T) {
		pod := withOrdinaryGpu(newDRATestPod())
		requested, limits := GetPodResources(pod)
		if requested.Gpus != 1 || limits.Gpus != 1 {
			t.Fatalf("the pod spec alone should report 1/1 GPUs, got requested=%v limits=%v", requested.Gpus, limits.Gpus)
		}
		if err := AddDRAGpus(ctx, pod, newDRATestParser(nil, 2), requested, limits); err != nil {
			t.Fatal(err)
		}
		if requested.Gpus != 3 || limits.Gpus != 3 {
			t.Fatalf("got requested=%v limits=%v, want 3/3", requested.Gpus, limits.Gpus)
		}
	})

	t.Run("Traditional GPU Pod without claims: leaves requested and limit GPUs unchanged", func(t *testing.T) {
		pod := withOrdinaryGpu(newDRATestPod())
		pod.Spec.ResourceClaims = nil
		pod.Spec.Containers[0].Resources.Claims = nil
		requested, limits := GetPodResources(pod)
		if err := AddDRAGpus(ctx, pod, newDRATestParser(nil, 2), requested, limits); err != nil {
			t.Fatal(err)
		}
		if requested.Gpus != 1 || limits.Gpus != 1 {
			t.Fatalf("got requested=%v limits=%v, want 1/1", requested.Gpus, limits.Gpus)
		}
	})

	t.Run("DRA parser is nil: preserves existing resource values", func(t *testing.T) {
		requested := &ResourcesInfo{Gpus: 1, GpuMemoryMegaBytes: 500}
		limits := &ResourcesInfo{Gpus: 1, GpuMemoryMegaBytes: 700}
		if err := AddDRAGpus(ctx, newDRATestPod(), nil, requested, limits); err != nil {
			t.Fatal(err)
		}
		if requested.Gpus != 1 || limits.Gpus != 1 || requested.GpuMemoryMegaBytes != 500 || limits.GpuMemoryMegaBytes != 700 {
			t.Fatalf("got requested=%+v limits=%+v", *requested, *limits)
		}
	})

	t.Run("Template read fails temporarily: preserves existing resources and adds requested GPUs after recovery", func(t *testing.T) {
		failures := 1
		parser := newDRATestParser(&failures, 2)
		requested := &ResourcesInfo{Gpus: 1, GpuMemoryMegaBytes: 500}
		limits := &ResourcesInfo{Gpus: 1, GpuMemoryMegaBytes: 700}

		if err := AddDRAGpus(ctx, newDRATestPod(), parser, requested, limits); err == nil {
			t.Fatal("expected an error on the first attempt")
		}
		if requested.Gpus != 1 || limits.Gpus != 1 || requested.GpuMemoryMegaBytes != 500 || limits.GpuMemoryMegaBytes != 700 {
			t.Fatalf("a failed attempt must not change the resources, got requested=%+v limits=%+v", *requested, *limits)
		}

		if err := AddDRAGpus(ctx, newDRATestPod(), parser, requested, limits); err != nil {
			t.Fatal(err)
		}
		if requested.Gpus != 3 || limits.Gpus != 3 || requested.GpuMemoryMegaBytes != 500 || limits.GpuMemoryMegaBytes != 700 {
			t.Fatalf("after recovery the two DRA GPUs must be added once, got requested=%+v limits=%+v", *requested, *limits)
		}
	})
}

func TestDRAPodAllocation(t *testing.T) {
	newRunningDRAPod := func() *corev1.Pod {
		pod := newDRATestPod()
		pod.Status.Phase = corev1.PodRunning
		pod.Spec.NodeName = "node-1"
		return pod
	}

	t.Run("Running DRA Pod without allocation annotations: reports zero allocated GPUs and GPU memory", func(t *testing.T) {
		allocated := GetPodAllocatedResources(newRunningDRAPod(), running)
		if allocated.Gpus != 0 || allocated.GpuMemoryMegaBytes != 0 {
			t.Fatalf("got allocated=%+v, want zero GPUs and memory", *allocated)
		}
	})

	t.Run("Running DRA Pod with allocation annotations: reads allocated GPU count and memory", func(t *testing.T) {
		pod := newRunningDRAPod()
		pod.Annotations = map[string]string{
			clustercommonconstants.PodAllocatedGpusAnnotation:      "2",
			clustercommonconstants.PodAllocatedGpuMemoryAnnotation: "10000",
		}
		allocated := GetPodAllocatedResources(pod, running)
		if allocated.Gpus != 2 || allocated.GpuMemoryMegaBytes != 10000 {
			t.Fatalf("got allocated=%+v, want 2 GPUs and 10000 MB", *allocated)
		}
	})

	t.Run("Allocation annotations change: reads the updated values", func(t *testing.T) {
		pod := newRunningDRAPod()
		pod.Annotations = map[string]string{
			clustercommonconstants.PodAllocatedGpusAnnotation:      "2",
			clustercommonconstants.PodAllocatedGpuMemoryAnnotation: "10000",
		}
		pod.Annotations[clustercommonconstants.PodAllocatedGpusAnnotation] = "3"
		pod.Annotations[clustercommonconstants.PodAllocatedGpuMemoryAnnotation] = "20000"
		allocated := GetPodAllocatedResources(pod, running)
		if allocated.Gpus != 3 || allocated.GpuMemoryMegaBytes != 20000 {
			t.Fatalf("got allocated=%+v, want 3 GPUs and 20000 MB", *allocated)
		}
	})
}
