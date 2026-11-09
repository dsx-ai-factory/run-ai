/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gpuinventory

import (
	"context"
	"errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
	"time"
)

func TestCollectorPublishesRefreshesAndClearsOnlyItsObservation(t *testing.T) {
	ctx := context.Background()
	s := runtime.NewScheme()
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	n := testNode()
	n.Annotations = map[string]string{"unrelated": "keep"}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(n).Build()
	now := time.Date(2026, 11, 8, 10, 0, 0, 0, time.UTC)
	collector := NewCollector(c, n.Name, "", "")
	collector.bootID = func() (string, error) { return "boot", nil }
	collector.now = func() time.Time { return now }
	collector.read = func() ([]Device, error) { return testDevices(), nil }
	get := func() *corev1.Node {
		t.Helper()
		node := &corev1.Node{}
		if err := c.Get(ctx, client.ObjectKey{Name: n.Name}, node); err != nil {
			t.Fatal(err)
		}
		return node
	}
	if err := collector.Collect(ctx); err != nil {
		t.Fatal(err)
	}
	first := get()
	if ReportedIdentity(first, now) != "2237:10de" || first.Annotations["unrelated"] != "keep" {
		t.Fatalf("bad publication: %+v", first.Annotations)
	}
	now = now.Add(RefreshInterval)
	if err := collector.Collect(ctx); err != nil {
		t.Fatal(err)
	}
	if get().Annotations[Annotation] == first.Annotations[Annotation] {
		t.Fatal("observation did not refresh")
	}
	collector.read = func() ([]Device, error) { return nil, errors.New("driver unavailable") }
	if err := collector.Collect(ctx); err == nil {
		t.Fatal("collection failure was hidden")
	}
	cleared := get()
	if _, ok := cleared.Annotations[Annotation]; ok {
		t.Fatal("stale identity retained")
	}
	if cleared.Annotations["unrelated"] != "keep" {
		t.Fatal("unrelated annotation removed")
	}
	collector.read = func() ([]Device, error) { return testDevices(), nil }
	if err := collector.Collect(ctx); err != nil {
		t.Fatal(err)
	}
	if ReportedIdentity(get(), now) != "2237:10de" {
		t.Fatal("recovery did not restore identity")
	}
	collector.bootID = func() (string, error) { return "different-kernel-boot", nil }
	if err := collector.Collect(ctx); err == nil {
		t.Fatal("kernel/node boot mismatch was accepted")
	}
	if get().Annotations[Annotation] != "" {
		t.Fatal("boot mismatch retained identity")
	}
}

func TestCollectorDoesNotOverwriteConcurrentNodeChanges(t *testing.T) {
	ctx := context.Background()
	s := runtime.NewScheme()
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	n := testNode()
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(n).Build()
	collector := NewCollector(c, n.Name, "", "")
	collector.bootID = func() (string, error) { return "boot", nil }
	collector.read = func() ([]Device, error) {
		current := &corev1.Node{}
		if err := c.Get(ctx, client.ObjectKey{Name: n.Name}, current); err != nil {
			t.Fatal(err)
		}
		current.Labels[MemoryLabel] = "40960"
		current.Annotations = map[string]string{Annotation: "concurrent-observation"}
		if err := c.Update(ctx, current); err != nil {
			t.Fatal(err)
		}
		return testDevices(), nil
	}
	if err := collector.Collect(ctx); err == nil {
		t.Fatal("expected resource-version conflict")
	}
	got := &corev1.Node{}
	if err := c.Get(ctx, client.ObjectKey{Name: n.Name}, got); err != nil {
		t.Fatal(err)
	}
	if got.Annotations[Annotation] != "concurrent-observation" {
		t.Fatal("overwrote concurrent evidence")
	}
}
