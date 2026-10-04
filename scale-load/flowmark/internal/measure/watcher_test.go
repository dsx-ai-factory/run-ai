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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

const testNamespace = "runai-flowmark"

func runner(name string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(WorkloadRunnerKind)
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

func waitForArrival(t *testing.T, w *Watcher, key string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		w.mu.Lock()
		_, seen := w.arrived[key]
		w.mu.Unlock()
		if seen {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("arrival for %q never recorded", key)
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
	waitForArrival(t, w, "w1")
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
	if got.Outcome != Pending || !errors.Is(got.Err, ErrWatchClosed) || !errors.Is(got.Err, context.Canceled) {
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

func externalWorkload(name, sourceUID, ownerUID string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(ExternalWorkloadKind)
	obj.SetNamespace(testNamespace)
	obj.SetName(name)
	if sourceUID != "" {
		_ = unstructured.SetNestedField(obj.Object, sourceUID, "spec", "originalWorkloadUID")
	}
	if ownerUID != "" {
		obj.SetOwnerReferences([]metav1.OwnerReference{{APIVersion: "apps/v1", Kind: "Deployment", Name: "dep", UID: types.UID(ownerUID)}})
	}
	return obj
}

func TestExternalWorkloadSourcePrefersUIDThenOwnerThenName(t *testing.T) {
	cases := []struct {
		name string
		obj  *unstructured.Unstructured
		want string
	}{
		{"spec uid wins", externalWorkload("ew-deployment-dep", "uid-spec", "uid-owner"), "uid-spec"},
		{"owner uid fallback", externalWorkload("ew-deployment-dep", "", "uid-owner"), "uid-owner"},
		{"name fallback", externalWorkload("ew-deployment-flowmark-ew-1a2b3c4d-r100-7", "", ""), "flowmark-ew-1a2b3c4d-r100-7"},
		{"name fallback other kind", externalWorkload("ew-tfjob-train", "", ""), "train"},
		{"name fallback no prefix", externalWorkload("plain-name", "", ""), "plain-name"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := externalWorkloadSource(tc.obj); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestExternalWorkloadArrivalMatchesByUIDInBothOrders(t *testing.T) {
	for _, eventFirst := range []bool{false, true} {
		name := "mark first"
		if eventFirst {
			name = "event first"
		}
		t.Run(name, func(t *testing.T) {
			kube := fake.NewClientBuilder().Build()
			w, err := WatchExternalWorkloads(context.Background(), kube, testNamespace)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(w.Stop)

			t0 := time.Now()
			create := func() {
				if err := kube.Create(context.Background(), externalWorkload("ew-deployment-dep-1", "uid-1", "uid-1")); err != nil {
					t.Fatal(err)
				}
			}
			if eventFirst {
				create()
				waitForArrival(t, w, "uid-1")
				w.MarkCreated("uid-1", t0)
			} else {
				w.MarkCreated("uid-1", t0)
				create()
			}

			got := byName(w.WaitAll(2 * time.Second))
			if got["uid-1"].Outcome != Arrived || got["uid-1"].Duration <= 0 {
				t.Fatalf("event must be matched to the deployment UID, got %+v", got)
			}
			if _, byEWName := got["dep-1"]; byEWName {
				t.Fatal("the item must be keyed by UID, not by the ExternalWorkload name")
			}
		})
	}
}

func TestUnmatchedCountsArrivalsNobodyMarked(t *testing.T) {
	kube := fake.NewClientBuilder().Build()
	w, err := WatchExternalWorkloads(context.Background(), kube, testNamespace)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(w.Stop)

	w.MarkCreated("uid-marked", time.Now())
	if err := kube.Create(context.Background(), externalWorkload("ew-deployment-marked", "uid-marked", "")); err != nil {
		t.Fatal(err)
	}
	if err := kube.Create(context.Background(), externalWorkload("ew-deployment-orphan", "", "")); err != nil {
		t.Fatal(err)
	}
	waitForArrival(t, w, "uid-marked")
	waitForArrival(t, w, "orphan")

	if got := w.Unmatched(); got != 1 {
		t.Fatalf("the object without a source UID must count as unmatched, got %d", got)
	}
	if got := byName(w.WaitAll(time.Second))["uid-marked"]; got.Outcome != Arrived {
		t.Fatalf("the marked item must still arrive: %+v", got)
	}
}

func TestServerClosedWatchIsReopenedAndLaterArrivalsStillCount(t *testing.T) {
	reconnectDelay = 10 * time.Millisecond
	t.Cleanup(func() { reconnectDelay = time.Second })

	first := watch.NewFake()
	opened := 0
	kube := fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
		Watch: func(ctx context.Context, c k8sclient.WithWatch, list k8sclient.ObjectList, opts ...k8sclient.ListOption) (watch.Interface, error) {
			opened++
			if opened == 1 {
				return first, nil
			}
			return c.Watch(ctx, list, opts...)
		},
	}).Build()
	w, err := WatchWorkloads(context.Background(), kube, testNamespace)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(w.Stop)

	w.MarkCreated("w1", time.Now())
	first.Stop()

	deadline := time.Now().Add(2 * time.Second)
	for w.Reconnects() == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if w.Reconnects() != 1 {
		t.Fatalf("the watch must be reopened once after the server closed it, got %d", w.Reconnects())
	}
	if err := kube.Create(context.Background(), runner("w1")); err != nil {
		t.Fatal(err)
	}
	got := byName(w.WaitAll(2 * time.Second))["w1"]
	if got.Outcome != Arrived {
		t.Fatalf("an arrival after the reconnect must count, got %+v", got)
	}
}

func TestWatchThatCannotBeReopenedEndsWaitWithClosedError(t *testing.T) {
	reconnectDelay = time.Millisecond
	t.Cleanup(func() { reconnectDelay = time.Second })

	first := watch.NewFake()
	opened := 0
	kube := fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
		Watch: func(context.Context, k8sclient.WithWatch, k8sclient.ObjectList, ...k8sclient.ListOption) (watch.Interface, error) {
			opened++
			if opened == 1 {
				return first, nil
			}
			return nil, errors.New("apiserver unreachable")
		},
	}).Build()
	w, err := WatchWorkloads(context.Background(), kube, testNamespace)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(w.Stop)
	w.MarkCreated("w1", time.Now())
	first.Stop()

	got := byName(w.WaitAll(5 * time.Second))["w1"]
	if got.Outcome != Pending || !errors.Is(got.Err, ErrWatchClosed) {
		t.Fatalf("a watch that cannot be reopened must surface on pending items, got %+v", got)
	}
	if opened != 1+maxReconnects {
		t.Fatalf("expected %d reopen attempts, got %d", maxReconnects, opened-1)
	}
}

func TestUnmatchedIgnoresArrivalsOfFailedCreates(t *testing.T) {
	kube, w := startWatch(t)
	w.MarkFailed("w1", errors.New("504 from the gateway, but the object was stored"))
	if err := kube.Create(context.Background(), runner("w1")); err != nil {
		t.Fatal(err)
	}
	waitForArrival(t, w, "w1")

	if got := w.Unmatched(); got != 0 {
		t.Fatalf("an arrival for a failed create is not a correlation problem, got %d unmatched", got)
	}
}

func TestAWatchTheServerKeepsEndingIsGivenUpWithTheClosedError(t *testing.T) {
	reconnectDelay = time.Millisecond
	t.Cleanup(func() { reconnectDelay = time.Second })

	opened := 0
	kube := fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
		Watch: func(context.Context, k8sclient.WithWatch, k8sclient.ObjectList, ...k8sclient.ListOption) (watch.Interface, error) {
			opened++
			ended := watch.NewFake()
			ended.Stop()
			return ended, nil
		},
	}).Build()
	w, err := WatchWorkloads(context.Background(), kube, testNamespace)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(w.Stop)
	w.MarkCreated("w1", time.Now())

	got := byName(w.WaitAll(5 * time.Second))["w1"]
	if got.Outcome != Pending || !errors.Is(got.Err, ErrWatchClosed) {
		t.Fatalf("a watch that keeps closing must end as a closed watch, not a slow cluster: %+v", got)
	}
	if opened != 1+maxReconnectCycles || w.Reconnects() != maxReconnectCycles {
		t.Fatalf("expected %d reopen cycles, got opened=%d reconnects=%d", maxReconnectCycles, opened, w.Reconnects())
	}
}

func TestAWatchTheServerKeepsFailingWithAnErrorEventIsGivenUp(t *testing.T) {
	reconnectDelay = time.Millisecond
	t.Cleanup(func() { reconnectDelay = time.Second })

	opened := 0
	kube := fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
		Watch: func(context.Context, k8sclient.WithWatch, k8sclient.ObjectList, ...k8sclient.ListOption) (watch.Interface, error) {
			opened++
			failing := watch.NewFake()
			go failing.Error(&metav1.Status{Code: 504, Message: "Too large resource version"})
			return failing, nil
		},
	}).Build()
	w, err := WatchWorkloads(context.Background(), kube, testNamespace)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(w.Stop)
	w.MarkCreated("w1", time.Now())

	got := byName(w.WaitAll(5 * time.Second))["w1"]
	if got.Outcome != Pending || !errors.Is(got.Err, ErrWatchClosed) {
		t.Fatalf("a watch that keeps failing must end as a closed watch: %+v", got)
	}
	if opened != 1+maxReconnectCycles {
		t.Fatalf("expected %d reopen cycles, got %d opens", maxReconnectCycles, opened-1)
	}
}
