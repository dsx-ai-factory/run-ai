/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package controllers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/gpuinventory"
	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGPUInventoryExpiresWithoutNodeEventsAndRearmsAfterRecovery(t *testing.T) {
	now := time.Date(2026, 11, 8, 10, 0, 0, 0, time.UTC)
	node := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "gpu", UID: "node", Labels: map[string]string{gpuinventory.CountLabel: "1", gpuinventory.ProductLabel: "NVIDIA-A10G", gpuinventory.MemoryLabel: "23028"}}}
	node.Status.NodeInfo.BootID = "boot"
	obs := gpuinventory.NewObservation(node, []gpuinventory.Device{{UUID: "GPU-test", BusID: "0000:00:1e.0", PCIDeviceID: "2237:10de", Product: "NVIDIA-A10G"}}, now)
	set := func() {
		t.Helper()
		data, err := json.Marshal(obs)
		if err != nil {
			t.Fatal(err)
		}
		node.Annotations = map[string]string{gpuinventory.Annotation: string(data)}
	}
	set()
	s := runtime.NewScheme()
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(node).Build()
	queued := []client.ObjectKey{}
	tracker := newGPUInventoryExpiration(c, func(k client.ObjectKey) { queued = append(queued, k) })
	scan := func(at time.Time) {
		t.Helper()
		if err := tracker.scan(context.Background(), at); err != nil {
			t.Fatal(err)
		}
	}
	scan(now)
	if len(queued) != 0 {
		t.Fatal("fresh observation queued")
	}
	scan(now.Add(gpuinventory.MaxAge))
	if len(queued) != 1 || queued[0].Name != "gpu" {
		t.Fatalf("expiry not queued: %+v", queued)
	}
	scan(now.Add(gpuinventory.MaxAge + time.Minute))
	if len(queued) != 1 {
		t.Fatal("unchanged expiry repeatedly queued")
	}
	if err := c.Get(context.Background(), client.ObjectKey{Name: "gpu"}, node); err != nil {
		t.Fatal(err)
	}
	obs.ObservedAt = now.Add(10 * time.Minute)
	set()
	if err := c.Update(context.Background(), node); err != nil {
		t.Fatal(err)
	}
	scan(obs.ObservedAt)
	scan(obs.ObservedAt.Add(gpuinventory.MaxAge))
	if len(queued) != 2 {
		t.Fatal("recovered collector expiry not rearmed")
	}
	if err := c.Get(context.Background(), client.ObjectKey{Name: "gpu"}, node); err != nil {
		t.Fatal(err)
	}
	node.Labels[v1alpha1.NodePoolNameLabel] = clusterconstants.ExcludedNodepoolName
	node.Annotations[gpuinventory.Annotation] = "bad"
	if err := c.Update(context.Background(), node); err != nil {
		t.Fatal(err)
	}
	scan(now.Add(time.Hour))
	if len(queued) != 2 {
		t.Fatal("excluded node queued")
	}
}
