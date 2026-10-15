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
	"sync"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	maxReconnects          = 5
	maxConsecutiveClosures = 10
	maxBackoffSteps        = 5
)

var reconnectDelay = time.Second

type Outcome string

const (
	Arrived      Outcome = "arrived"
	NeverArrived Outcome = "never_arrived"
	Failed       Outcome = "failed"
)

var WorkloadRunnerKind = schema.GroupVersionKind{Group: "run.ai", Version: "v1alpha1", Kind: "WorkloadRunner"}

var ErrWatchClosed = errors.New("watch closed")

type Timing struct {
	Name     string
	Outcome  Outcome
	Created  time.Time
	Arrived  time.Time
	Duration time.Duration
	Err      error
}

type Watcher struct {
	cancel    context.CancelFunc
	nameOf    func(*unstructured.Unstructured) string
	kube      k8sclient.WithWatch
	kind      schema.GroupVersionKind
	namespace string

	resourceVersion string

	mu            sync.Mutex
	created       map[string]time.Time
	arrived       map[string]time.Time
	failed        map[string]error
	lastMark      time.Time
	closed        error
	reconnects    int
	interruptions int
	changed       chan struct{}
	done          chan struct{}
}

func Watch(ctx context.Context, kube k8sclient.WithWatch, kind schema.GroupVersionKind, namespace string, nameOf func(*unstructured.Unstructured) string) (*Watcher, error) {
	ctx, cancel := context.WithCancel(ctx)
	w := &Watcher{
		cancel:    cancel,
		nameOf:    nameOf,
		kube:      kube,
		kind:      kind,
		namespace: namespace,
		created:   map[string]time.Time{},
		arrived:   map[string]time.Time{},
		failed:    map[string]error{},
		changed:   make(chan struct{}, 1),
		done:      make(chan struct{}),
	}
	events, err := w.open(ctx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("watching %s in namespace %q: %w", kind.Kind, namespace, err)
	}
	go w.consume(ctx, events)
	return w, nil
}

func (w *Watcher) open(ctx context.Context) (watch.Interface, error) {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(schema.GroupVersionKind{Group: w.kind.Group, Version: w.kind.Version, Kind: w.kind.Kind + "List"})
	options := &k8sclient.ListOptions{
		Namespace: w.namespace,
		Raw:       &metav1.ListOptions{AllowWatchBookmarks: true, ResourceVersion: w.resourceVersion},
	}
	return w.kube.Watch(ctx, list, options)
}

func (w *Watcher) reopen(ctx context.Context, closures int) (watch.Interface, error) {
	var last error
	for attempt := 1; attempt <= maxReconnects; attempt++ {
		if steps := min(closures-1+attempt-1, maxBackoffSteps); steps > 0 {
			timer := time.NewTimer(time.Duration(steps) * reconnectDelay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, fmt.Errorf("%w: %w", ErrWatchClosed, ctx.Err())
			case <-timer.C:
			}
		}
		events, err := w.open(ctx)
		if err == nil {
			w.mu.Lock()
			w.reconnects++
			if w.resourceVersion == "" {
				w.interruptions++
			}
			w.mu.Unlock()
			return events, nil
		}
		last = err
		if apierrors.IsGone(err) || apierrors.IsResourceExpired(err) {
			w.resourceVersion = ""
		}
	}
	return nil, fmt.Errorf("%w: reopening the %s watch failed %d times (last: %v)", ErrWatchClosed, w.kind.Kind, maxReconnects, last)
}

func (w *Watcher) Reconnects() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.reconnects
}

func (w *Watcher) Interruptions() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.interruptions
}

func WatchWorkloads(ctx context.Context, kube k8sclient.WithWatch, namespace string) (*Watcher, error) {
	return Watch(ctx, kube, WorkloadRunnerKind, namespace, (*unstructured.Unstructured).GetName)
}

func (w *Watcher) MarkCreated(name string, t0 time.Time) {
	w.mu.Lock()
	w.created[name] = t0
	w.lastMark = time.Now()
	w.mu.Unlock()
	w.signal()
}

func (w *Watcher) MarkFailed(name string, err error) {
	w.mu.Lock()
	w.failed[name] = err
	w.mu.Unlock()
	w.signal()
}

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

func (w *Watcher) Unmatched() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	unmatched := 0
	for key := range w.arrived {
		_, marked := w.created[key]
		_, failed := w.failed[key]
		if !marked && !failed {
			unmatched++
		}
	}
	return unmatched
}

func (w *Watcher) Stop() {
	w.cancel()
	<-w.done
}

func (w *Watcher) consume(ctx context.Context, events watch.Interface) {
	defer close(w.done)
	closures := 0
	for {
		select {
		case <-ctx.Done():
			events.Stop()
			w.close(fmt.Errorf("%w: %w", ErrWatchClosed, ctx.Err()))
			return
		case event, open := <-events.ResultChan():
			if !open {
				events.Stop()
				closures++
				reopened, err := w.reopenAfterClosure(ctx, closures, nil)
				if err != nil {
					w.close(err)
					return
				}
				events = reopened
				continue
			}
			switch event.Type {
			case watch.Added, watch.Modified:
				closures = 0
				if obj := asUnstructured(event.Object); obj != nil {
					w.resourceVersion = obj.GetResourceVersion()
					w.recordArrival(w.nameOf(obj), time.Now())
				}
			case watch.Bookmark, watch.Deleted:
				closures = 0
				if obj := asUnstructured(event.Object); obj != nil {
					w.resourceVersion = obj.GetResourceVersion()
				}
			case watch.Error:
				events.Stop()
				closures++
				if status := apierrors.FromObject(event.Object); apierrors.IsGone(status) || apierrors.IsResourceExpired(status) {
					w.resourceVersion = ""
				}
				reopened, err := w.reopenAfterClosure(ctx, closures, event.Object)
				if err != nil {
					w.close(err)
					return
				}
				events = reopened
			}
		}
	}
}

func (w *Watcher) reopenAfterClosure(ctx context.Context, closures int, cause runtime.Object) (watch.Interface, error) {
	switch {
	case closures > maxConsecutiveClosures:
		return nil, fmt.Errorf("%w: the server ended the %s watch %d times in a row without delivering an event", ErrWatchClosed, w.kind.Kind, closures-1)
	case w.Interruptions() >= maxConsecutiveClosures:
		return nil, fmt.Errorf("%w: the %s watch lost its resume point %d times", ErrWatchClosed, w.kind.Kind, maxConsecutiveClosures)
	}
	reopened, err := w.reopen(ctx, closures)
	if err != nil && cause != nil {
		return nil, fmt.Errorf("%w after a watch error (%v)", err, cause)
	}
	return reopened, err
}

func asUnstructured(object runtime.Object) *unstructured.Unstructured {
	if obj, ok := object.(*unstructured.Unstructured); ok {
		return obj
	}
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	if err != nil {
		return nil
	}
	return &unstructured.Unstructured{Object: content}
}

func (w *Watcher) recordArrival(name string, at time.Time) {
	w.mu.Lock()
	if _, seen := w.arrived[name]; !seen {
		w.arrived[name] = at
	}
	w.mu.Unlock()
	w.signal()
}

func (w *Watcher) Err() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.closed
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
			timing.Outcome = NeverArrived
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
