/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package measure

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Outcome is the honest fate of one measured item: it arrived, it is still
// pending when the caller stopped waiting, or its creation failed.
type Outcome string

const (
	Arrived Outcome = "arrived"
	Pending Outcome = "pending"
	Failed  Outcome = "failed"
)

var (
	workloadRunnerKind   = schema.GroupVersionKind{Group: "run.ai", Version: "v1alpha1", Kind: "WorkloadRunner"}
	externalWorkloadKind = schema.GroupVersionKind{Group: "run.ai", Version: "v1alpha1", Kind: "ExternalWorkload"}
	errWatchClosed       = errors.New("watch closed by the server")
)

// Timing is the measurement of one item. Duration is only meaningful when
// Outcome is Arrived; pending and failed items carry no duration so they can
// never be mistaken for a fast arrival.
type Timing struct {
	Name     string
	Outcome  Outcome
	Created  time.Time
	Arrived  time.Time
	Duration time.Duration
	Err      error
}

// Watcher stamps the arrival of Kubernetes objects on the local monotonic
// clock. It never reads server timestamps, so a duration is always the gap
// between two readings of the same clock.
type Watcher struct {
	cancel context.CancelFunc
	nameOf func(*unstructured.Unstructured) string

	mu       sync.Mutex
	created  map[string]time.Time
	arrived  map[string]time.Time
	failed   map[string]error
	lastMark time.Time
	closed   error
	changed  chan struct{}
	done     chan struct{}
}

// Watch opens a watch on kind in namespace before the caller creates anything,
// so no arrival can be missed. nameOf maps an event object to the item name the
// caller will pass to MarkCreated. Arrival is the first event seen for a name;
// later events for the same name never move the stamp.
func Watch(ctx context.Context, kube k8sclient.WithWatch, kind schema.GroupVersionKind, namespace string, nameOf func(*unstructured.Unstructured) string) (*Watcher, error) {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(schema.GroupVersionKind{Group: kind.Group, Version: kind.Version, Kind: kind.Kind + "List"})

	ctx, cancel := context.WithCancel(ctx)
	events, err := kube.Watch(ctx, list, k8sclient.InNamespace(namespace))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("watching %s in namespace %q: %w", kind.Kind, namespace, err)
	}

	w := &Watcher{
		cancel:  cancel,
		nameOf:  nameOf,
		created: map[string]time.Time{},
		arrived: map[string]time.Time{},
		failed:  map[string]error{},
		changed: make(chan struct{}, 1),
		done:    make(chan struct{}),
	}
	go w.consume(ctx, events)
	return w, nil
}

// WatchWorkloads watches the WorkloadRunner objects Run:ai creates for
// workloads submitted through the control plane. The item name is the
// workload name.
func WatchWorkloads(ctx context.Context, kube k8sclient.WithWatch, namespace string) (*Watcher, error) {
	return Watch(ctx, kube, workloadRunnerKind, namespace, (*unstructured.Unstructured).GetName)
}

// WatchExternalWorkloads watches the ExternalWorkload objects Run:ai creates
// for plain Kubernetes workloads. Run:ai names them ew-<kind>-<name>; the item
// name is the original object name.
func WatchExternalWorkloads(ctx context.Context, kube k8sclient.WithWatch, namespace string) (*Watcher, error) {
	return Watch(ctx, kube, externalWorkloadKind, namespace, externalWorkloadSource)
}

func externalWorkloadSource(obj *unstructured.Unstructured) string {
	rest, ok := strings.CutPrefix(obj.GetName(), "ew-")
	if !ok {
		return obj.GetName()
	}
	if _, name, found := strings.Cut(rest, "-"); found {
		return name
	}
	return rest
}

// MarkCreated records when the caller finished creating name, taken from the
// same local clock the arrival is stamped with. It may be called before or
// after the arrival event; the duration is the same either way. An arrival
// stamped before t0, possible when the cluster reacts while the create call is
// still returning, is reported as zero, never negative.
func (w *Watcher) MarkCreated(name string, t0 time.Time) {
	w.mu.Lock()
	w.created[name] = t0
	w.lastMark = time.Now()
	w.mu.Unlock()
	w.signal()
}

// MarkFailed records that creating name failed. The item is reported as
// Failed and is never part of a latency distribution.
func (w *Watcher) MarkFailed(name string, err error) {
	w.mu.Lock()
	w.failed[name] = err
	w.mu.Unlock()
	w.signal()
}

// WaitAll blocks until every marked item has arrived or failed, or until
// timeout has passed since the last MarkCreated. Items still unresolved are
// reported as Pending, never as zero-duration arrivals. If the server closed
// the watch or the context was cancelled, WaitAll returns at once and pending
// items carry that error so a broken watch cannot pass as a slow cluster.
// Results are sorted by name.
func (w *Watcher) WaitAll(timeout time.Duration) []Timing {
	for {
		w.mu.Lock()
		unresolved := w.unresolvedLocked()
		deadline := w.lastMark.Add(timeout)
		closed := w.closed
		w.mu.Unlock()

		remaining := time.Until(deadline)
		if len(unresolved) == 0 || remaining <= 0 || closed != nil {
			return w.timings()
		}

		timer := time.NewTimer(remaining)
		select {
		case <-w.changed:
			timer.Stop()
		case <-timer.C:
		}
	}
}

// Stop closes the watch and waits for the event loop to exit.
func (w *Watcher) Stop() {
	w.cancel()
	<-w.done
}

func (w *Watcher) consume(ctx context.Context, events watch.Interface) {
	defer close(w.done)
	defer events.Stop()
	for {
		select {
		case <-ctx.Done():
			w.close(fmt.Errorf("%w: %w", errWatchClosed, ctx.Err()))
			return
		case event, open := <-events.ResultChan():
			if !open {
				w.close(errWatchClosed)
				return
			}
			switch event.Type {
			case watch.Added, watch.Modified:
				if obj, ok := event.Object.(*unstructured.Unstructured); ok {
					w.recordArrival(w.nameOf(obj), time.Now())
				}
			case watch.Error:
				w.close(fmt.Errorf("%w: %v", errWatchClosed, event.Object))
				return
			}
		}
	}
}

func (w *Watcher) recordArrival(name string, at time.Time) {
	w.mu.Lock()
	if _, seen := w.arrived[name]; !seen {
		w.arrived[name] = at
	}
	w.mu.Unlock()
	w.signal()
}

func (w *Watcher) close(err error) {
	w.mu.Lock()
	if w.closed == nil {
		w.closed = err
	}
	w.mu.Unlock()
	w.signal()
}

func (w *Watcher) signal() {
	select {
	case w.changed <- struct{}{}:
	default:
	}
}

func (w *Watcher) unresolvedLocked() []string {
	var names []string
	for name := range w.created {
		_, arrived := w.arrived[name]
		_, failed := w.failed[name]
		if !arrived && !failed {
			names = append(names, name)
		}
	}
	return names
}

func (w *Watcher) timings() []Timing {
	w.mu.Lock()
	defer w.mu.Unlock()

	result := make([]Timing, 0, len(w.created)+len(w.failed))
	for name, t0 := range w.created {
		timing := Timing{Name: name, Created: t0}
		switch {
		case w.failed[name] != nil:
			timing.Outcome = Failed
			timing.Err = w.failed[name]
		case !w.arrived[name].IsZero():
			timing.Outcome = Arrived
			timing.Arrived = w.arrived[name]
			timing.Duration = max(timing.Arrived.Sub(t0), 0)
		default:
			timing.Outcome = Pending
			timing.Err = w.closed
		}
		result = append(result, timing)
	}
	for name, err := range w.failed {
		if _, created := w.created[name]; !created {
			result = append(result, Timing{Name: name, Outcome: Failed, Err: err})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}
