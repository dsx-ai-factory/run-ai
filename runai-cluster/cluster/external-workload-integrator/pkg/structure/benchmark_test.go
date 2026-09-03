/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package structure

import (
	"context"
	"fmt"
	"testing"

	v1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
)

func makeBenchKarta() *v1alpha1.Karta {
	workerValue := "worker"
	return makeKarta("root", []v1alpha1.ComponentDefinition{
		{
			Name: "group",
			Kind: &v1alpha1.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "StatefulSet",
			},
		},
		{
			Name:     "worker",
			OwnerRef: ptr.To("group"),
			PodSelector: &v1alpha1.PodSelector{
				ComponentTypeSelector: &v1alpha1.ComponentTypeSelector{
					KeyPath: `.metadata.labels["role"]`,
					Value:   &workerValue,
				},
			},
		},
	})
}

func makeBenchPods(n int) []corev1.Pod {
	pods := make([]corev1.Pod, n)
	for i := range n {
		pods[i] = corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				UID:  types.UID(fmt.Sprintf("pod-uid-%d", i)),
				Name: fmt.Sprintf("worker-%d", i),
				Labels: map[string]string{
					"role": "worker",
				},
			},
		}
	}
	return pods
}

func makeBenchOwnerChain() OwnerChain {
	return OwnerChain{
		{UID: "rs-uid", Name: "worker-rs", Kind: "ReplicaSet", Group: "apps", Version: "v1"},
		{UID: "sts-uid", Name: "worker-sts", Kind: "StatefulSet", Group: "apps", Version: "v1"},
		{UID: "lws-uid", Name: "my-lws", Kind: "LeaderWorkerSet", Group: "leaderworkerset.x-k8s.io", Version: "v1"},
	}
}

func benchmarkGetPodElementsPath(b *testing.B, numPods int) {
	karta := makeBenchKarta()
	pods := makeBenchPods(numPods)
	ownerChain := makeBenchOwnerChain()

	b.ResetTimer()
	for range b.N {
		for i := range pods {
			_ = GetPodElementsPath(context.Background(), &pods[i], karta, ownerChain, nil)
		}
	}
}

func BenchmarkGetPodElementsPath_10Pods(b *testing.B) {
	benchmarkGetPodElementsPath(b, 10)
}

func BenchmarkGetPodElementsPath_100Pods(b *testing.B) {
	benchmarkGetPodElementsPath(b, 100)
}

func BenchmarkGetPodElementsPath_1000Pods(b *testing.B) {
	benchmarkGetPodElementsPath(b, 1000)
}

// BenchmarkBuildElementStructure_FirstTime benchmarks the first Build call where all
// pods are unlabeled and need full element path calculation.
func BenchmarkBuildElementStructure_FirstTime_1000Pods(b *testing.B) {
	benchmarkBuildFirstTime(b, 1000)
}

// BenchmarkBuildElementStructure_AllLabeled benchmarks the fast path where all pods
// are already labeled and the structure is unchanged - should be near-zero cost.
// Uses a two-phase approach: Build once from unlabeled pods, label them, then benchmark
// the second Build call which should hit the fast path.
func BenchmarkBuildElementStructure_AllLabeled_1000Pods(b *testing.B) {
	benchmarkBuildAllLabeled(b, 1000)
}

func benchmarkBuildFirstTime(b *testing.B, numPods int) {
	karta := makeBenchKarta()
	pods := makeBenchPods(numPods)
	ownerChain := makeBenchOwnerChain()

	ownerChainByPodName := make(map[string]*OwnerChain, numPods)
	for i := range pods {
		chain := ownerChain
		ownerChainByPodName[pods[i].Name] = &chain
	}

	b.ResetTimer()
	for range b.N {
		_, _, err := BuildElementStructure(context.Background(), "wl-uid", karta, pods, ownerChainByPodName, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkBuildAllLabeled(b *testing.B, numPods int) {
	karta := makeBenchKarta()
	pods := makeBenchPods(numPods)
	ownerChain := makeBenchOwnerChain()

	// Phase 1: Build structure from unlabeled pods (one-time setup)
	ownerChainByPodName := make(map[string]*OwnerChain, numPods)
	for i := range pods {
		chain := ownerChain
		ownerChainByPodName[pods[i].Name] = &chain
	}

	existing, addresses, err := BuildElementStructure(context.Background(), "wl-uid", karta, pods, ownerChainByPodName, nil)
	if err != nil {
		b.Fatal(err)
	}

	// Phase 2: Label pods with computed element paths
	labeledPods := make([]corev1.Pod, len(pods))
	copy(labeledPods, pods)
	for _, addr := range addresses {
		labels := GeneratePodElementLabels(addr.ElementsPath)
		for i := range labeledPods {
			if string(labeledPods[i].UID) == addr.PodID {
				for k, v := range labels {
					labeledPods[i].Labels[k] = v
				}
			}
		}
	}

	// Phase 3: Benchmark the second Build call — all pods labeled, same structure
	b.ResetTimer()
	for range b.N {
		result, _, err := BuildElementStructure(context.Background(), "wl-uid", karta, labeledPods, nil, existing)
		if err != nil {
			b.Fatal(err)
		}
		if len(result) != len(existing) {
			b.Fatalf("expected %d elements, got %d", len(existing), len(result))
		}
	}
}
