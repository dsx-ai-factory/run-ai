/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package measure

import (
	"context"
	"errors"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const testNamespace = "runai-flowmark"

func runner(name string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(workloadRunnerKind)
	obj.SetNamespace(testNamespace)
	obj.SetName(name)
	return obj
}

func startWatch(t *testing.T) (k8sclient.WithWatch, *Watcher) {
	t.Helper()
	kube := fake.NewClientBuilder().Build()
	w, err := WatchWorkloads(context.Background(), kube, testNamespace)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(w.Stop)
	return kube, w
}

func byName(timings []Timing) map[string]Timing {
	result := map[string]Timing{}
	for _, timing := range timings {
		result[timing.Name] = timing
	}
	return result
}

func TestMarkBeforeEventStampsDuration(t *testing.T) {
	kube, w := startWatch(t)

	t0 := time.Now()
	w.MarkCreated("w1", t0)
	time.Sleep(20 * time.Millisecond)
	if err := kube.Create(context.Background(), runner("w1")); err != nil {
		t.Fatal(err)
	}

	got := byName(w.WaitAll(2 * time.Second))["w1"]
	if got.Outcome != Arrived || got.Duration < 20*time.Millisecond || got.Duration > time.Second {
		t.Fatalf("expected an arrival of at least 20ms, got %+v", got)
	}
	if !got.Created.Equal(t0) || got.Arrived.Before(t0) {
		t.Fatalf("timestamps must come from the marks and the event: %+v", got)
	}
}

func TestEventBeforeMarkStampsDuration(t *testing.T) {
	kube, w := startWatch(t)

	t0 := time.Now()
	if err := kube.Create(context.Background(), runner("w1")); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for {
		w.mu.Lock()
		_, seen := w.arrived["w1"]
		w.mu.Unlock()
		if seen || time.Now().After(deadline) {
			break
		}
		time.Sleep(time.Millisecond)
	}
	w.MarkCreated("w1", t0)

	got := byName(w.WaitAll(2 * time.Second))["w1"]
	if got.Outcome != Arrived || got.Duration <= 0 || got.Duration > time.Second {
		t.Fatalf("a fast cluster must still yield a positive duration, got %+v", got)
	}
}

func TestNeverArrivingItemIsPendingAtTimeout(t *testing.T) {
	kube, w := startWatch(t)

	start := time.Now()
	w.MarkCreated("slow", start)
	w.MarkCreated("fast", start)
	if err := kube.Create(context.Background(), runner("fast")); err != nil {
		t.Fatal(err)
	}

	got := byName(w.WaitAll(50 * time.Millisecond))
	if waited := time.Since(start); waited < 50*time.Millisecond || waited > time.Second {
		t.Fatalf("WaitAll must return at the timeout counted from the last mark, waited %s", waited)
	}
	if got["fast"].Outcome != Arrived {
		t.Fatalf("fast must arrive: %+v", got["fast"])
	}
	if got["slow"].Outcome != Pending || got["slow"].Duration != 0 || got["slow"].Err != nil {
		t.Fatalf("slow must be pending with no duration and no error, got %+v", got["slow"])
	}
}

func TestMarkFailedIsReportedNotTimed(t *testing.T) {
	kube, w := startWatch(t)

	w.MarkCreated("ok", time.Now())
	w.MarkFailed("broken", errors.New("HTTP 500 from the control plane"))
	if err := kube.Create(context.Background(), runner("ok")); err != nil {
		t.Fatal(err)
	}

	got := byName(w.WaitAll(time.Second))
	if len(got) != 2 {
		t.Fatalf("expected both items, got %v", got)
	}
	if got["broken"].Outcome != Failed || got["broken"].Err == nil || got["broken"].Duration != 0 {
		t.Fatalf("failed item must carry its error and no duration: %+v", got["broken"])
	}
	if got["ok"].Outcome != Arrived {
		t.Fatalf("ok must arrive: %+v", got["ok"])
	}
}

func TestFirstEventWins(t *testing.T) {
	kube, w := startWatch(t)

	w.MarkCreated("w1", time.Now())
	obj := runner("w1")
	if err := kube.Create(context.Background(), obj); err != nil {
		t.Fatal(err)
	}
	first := byName(w.WaitAll(time.Second))["w1"]

	obj.SetLabels(map[string]string{"touched": "yes"})
	if err := kube.Update(context.Background(), obj); err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	second := byName(w.WaitAll(time.Second))["w1"]

	if !first.Arrived.Equal(second.Arrived) {
		t.Fatalf("a later update must not move the arrival stamp: %s vs %s", first.Arrived, second.Arrived)
	}
}

func TestCancelledContextEndsWaitWithError(t *testing.T) {
	kube := fake.NewClientBuilder().Build()
	ctx, cancel := context.WithCancel(context.Background())
	w, err := WatchWorkloads(ctx, kube, testNamespace)
	if err != nil {
		t.Fatal(err)
	}
	w.MarkCreated("w1", time.Now())

	cancel()
	start := time.Now()
	got := byName(w.WaitAll(time.Minute))["w1"]

	if waited := time.Since(start); waited > 5*time.Second {
		t.Fatalf("WaitAll must return promptly after cancellation, waited %s", waited)
	}
	if got.Outcome != Pending || !errors.Is(got.Err, errWatchClosed) || !errors.Is(got.Err, context.Canceled) {
		t.Fatalf("cancellation must surface on pending items, got %+v", got)
	}
}

func TestArrivalBeforeMarkIsZeroNotNegative(t *testing.T) {
	_, w := startWatch(t)

	w.recordArrival("w1", time.Now())
	w.MarkCreated("w1", time.Now().Add(50*time.Millisecond))

	got := byName(w.WaitAll(time.Second))["w1"]
	if got.Outcome != Arrived || got.Duration != 0 {
		t.Fatalf("an arrival stamped before t0 must be a zero-duration arrival, got %+v", got)
	}
}

func TestExternalWorkloadSourceName(t *testing.T) {
	cases := map[string]string{
		"ew-deployment-flowmark-basic-1a2b3c4d-w1": "flowmark-basic-1a2b3c4d-w1",
		"ew-tfjob-train": "train",
		"plain-name":     "plain-name",
	}
	for in, want := range cases {
		obj := &unstructured.Unstructured{}
		obj.SetName(in)
		if got := externalWorkloadSource(obj); got != want {
			t.Errorf("%s: got %q want %q", in, got, want)
		}
	}
}
