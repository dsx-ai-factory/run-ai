/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package measure

import (
	"context"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func workloadRunner(name string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(WorkloadRunnerKind)
	obj.SetNamespace(testNamespace)
	obj.SetName(name)
	return obj
}

func workloadPod(workload, suffix string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(PodKind)
	obj.SetNamespace(testNamespace)
	obj.SetName(workload + suffix)
	obj.SetOwnerReferences([]metav1.OwnerReference{{APIVersion: "run.ai/v1", Kind: "RunaiJob", Name: workload, UID: types.UID("uid-" + workload)}})
	return obj
}

func TestArrivalsCorrelateRunnerAndPodInEitherOrder(t *testing.T) {
	for name, order := range map[string][]k8sclient.Object{
		"runner then pod": {workloadRunner("w1"), workloadPod("w1", "-0-0")},
		"pod then runner": {workloadPod("w1", "-0-0"), workloadRunner("w1")},
	} {
		t.Run(name, func(t *testing.T) {
			kube := fake.NewClientBuilder().Build()
			watchers, err := OpenArrivals(context.Background(), kube, testNamespace)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(watchers.Stop)

			accepted := time.Now()
			watchers.MarkAccepted("w1", accepted)
			for _, obj := range order {
				if err := kube.Create(context.Background(), obj); err != nil {
					t.Fatal(err)
				}
			}

			filled := watchers.Fill([]Submission{{Name: "w1", RequestedAt: accepted.Add(-time.Second), ApiAcceptedAt: accepted}}, 5*time.Second)
			got := filled[0]
			if got.ClusterSyncedAt.IsZero() || got.PodCreatedAt.IsZero() {
				t.Fatalf("both stamps must be filled, got %+v", got)
			}
			if got.ClusterSyncedAt.Before(accepted) || got.PodCreatedAt.Before(accepted) {
				t.Fatalf("stamps come from the watch after the accept, got %+v", got)
			}
			if spans := CollectSpans(filled); spans.TTJ[0].Outcome != Arrived {
				t.Fatalf("a fully arrived workspace has a ttj, got %+v", spans.TTJ[0])
			}
		})
	}
}

func TestArrivalsLeaveStampsEmptyWhenNothingArrives(t *testing.T) {
	kube := fake.NewClientBuilder().Build()
	watchers, err := OpenArrivals(context.Background(), kube, testNamespace)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(watchers.Stop)
	accepted := time.Now()
	watchers.MarkAccepted("w1", accepted)

	filled := watchers.Fill([]Submission{{Name: "w1", RequestedAt: accepted, ApiAcceptedAt: accepted}}, 300*time.Millisecond)
	if !filled[0].ClusterSyncedAt.IsZero() || !filled[0].PodCreatedAt.IsZero() {
		t.Fatalf("nothing arrived, stamps must stay empty: %+v", filled[0])
	}
	if spans := CollectSpans(filled); spans.TTJ[0].Outcome != NeverArrived {
		t.Fatalf("a workspace that never reached the cluster is marked never arrived, got %+v", spans.TTJ[0])
	}
}

func TestPodWithoutOwnerIsMatchedByItsNamePrefix(t *testing.T) {
	kube := fake.NewClientBuilder().Build()
	watchers, err := OpenArrivals(context.Background(), kube, testNamespace)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(watchers.Stop)
	accepted := time.Now()
	watchers.MarkAccepted("w1", accepted)

	pod := workloadPod("w1", "-0-0")
	pod.SetOwnerReferences(nil)
	if err := kube.Create(context.Background(), pod); err != nil {
		t.Fatal(err)
	}
	filled := watchers.Fill([]Submission{{Name: "w1", RequestedAt: accepted, ApiAcceptedAt: accepted}}, 2*time.Second)
	if filled[0].PodCreatedAt.IsZero() {
		t.Fatalf("pod w1-0-0 belongs to workspace w1 by its name, got %+v", filled[0])
	}
}
