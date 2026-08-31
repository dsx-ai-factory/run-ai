/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload_submitter

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// newTestSubmitter creates a minimal Submitter for unit tests with the given
// max concurrent slots and injected submitFn. Real clients are not needed.
func newTestSubmitter(maxConcurrent int, fn submitFunc) *Submitter {
	cfg := &Config{
		ClusterID:       "test-cluster",
		MaxConcurrent:   maxConcurrent,
		PodsPerWorkload: 1,
	}
	s := NewSubmitter(cfg, nil, Params{DSecs: 60}, []string{"proj-0"}, nil)
	s.submitFn = fn
	return s
}

// captureLog redirects the default logger to a buffer for the duration of the test.
func captureLog(t *testing.T) func() string {
	t.Helper()
	var buf bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(prev) })
	return func() string { return buf.String() }
}

func resp(status int) *http.Response {
	return &http.Response{StatusCode: status, Body: http.NoBody}
}

// TestRecordFailure_Accumulation verifies that repeated failures accumulate per status code.
func TestRecordFailure_Accumulation(t *testing.T) {
	s := newTestSubmitter(10, nil)

	s.recordFailure(resp(http.StatusInternalServerError), nil)
	s.recordFailure(resp(http.StatusTooManyRequests), nil)
	s.recordFailure(resp(http.StatusTooManyRequests), nil)

	s.failMu.Lock()
	defer s.failMu.Unlock()

	if got := s.failures[http.StatusInternalServerError]; got != 1 {
		t.Errorf("500 count: got %d, want 1", got)
	}
	if got := s.failures[http.StatusTooManyRequests]; got != 2 {
		t.Errorf("429 count: got %d, want 2", got)
	}
}

// TestRecordFailure_NilResponse verifies that a nil HTTP response (connection error)
// is recorded as status 0 without panicking.
func TestRecordFailure_NilResponse(t *testing.T) {
	s := newTestSubmitter(10, nil)

	s.recordFailure(nil, fmt.Errorf("connection refused"))

	s.failMu.Lock()
	defer s.failMu.Unlock()

	if got := s.failures[0]; got != 1 {
		t.Errorf("connection error count: got %d, want 1", got)
	}
}

// TestDoSubmit_Success verifies that a 202 response increments successCount and
// leaves the failures map empty.
func TestDoSubmit_Success(t *testing.T) {
	captureLog(t) // suppress output

	s := newTestSubmitter(10, func(_ context.Context, _ *Clients, _ submitParams) (string, *http.Response, error) {
		return "wl-id", resp(http.StatusAccepted), nil
	})

	s.doSubmit(context.Background(), 0)

	if got := s.successCount.Load(); got != 1 {
		t.Errorf("successCount: got %d, want 1", got)
	}
	s.failMu.Lock()
	defer s.failMu.Unlock()
	if len(s.failures) != 0 {
		t.Errorf("expected no failures, got %v", s.failures)
	}
}

// TestDoSubmit_Failure500 verifies that a 500 response is recorded as a failure.
func TestDoSubmit_Failure500(t *testing.T) {
	captureLog(t)

	s := newTestSubmitter(10, func(_ context.Context, _ *Clients, _ submitParams) (string, *http.Response, error) {
		return "", resp(http.StatusInternalServerError), fmt.Errorf("server error")
	})

	s.doSubmit(context.Background(), 0)

	if got := s.successCount.Load(); got != 0 {
		t.Errorf("expected no successes, got %d", got)
	}
	s.failMu.Lock()
	defer s.failMu.Unlock()
	if s.failures[http.StatusInternalServerError] != 1 {
		t.Errorf("expected failures[500]=1, got %v", s.failures)
	}
}

// TestLatencyWarning verifies that a slow submitFn triggers a latency warning.
// latencyWarnThreshold is lowered to 50ms so the test runs in ~100ms.
func TestLatencyWarning(t *testing.T) {
	logs := captureLog(t)

	s := newTestSubmitter(10, func(_ context.Context, _ *Clients, _ submitParams) (string, *http.Response, error) {
		time.Sleep(100 * time.Millisecond)
		return "", resp(http.StatusAccepted), nil
	})
	s.latencyWarnThreshold = 50 * time.Millisecond

	s.doSubmit(context.Background(), 0)

	if !strings.Contains(logs(), "[WARNING] slow submission") {
		t.Errorf("expected slow submission warning; log:\n%s", logs())
	}
}

// TestNoLatencyWarning verifies that a fast submitFn does not trigger a latency warning.
func TestNoLatencyWarning(t *testing.T) {
	logs := captureLog(t)

	s := newTestSubmitter(10, func(_ context.Context, _ *Clients, _ submitParams) (string, *http.Response, error) {
		return "", resp(http.StatusAccepted), nil
	})
	s.latencyWarnThreshold = 50 * time.Millisecond

	s.doSubmit(context.Background(), 0)

	if strings.Contains(logs(), "[WARNING] slow submission") {
		t.Errorf("unexpected slow submission warning; log:\n%s", logs())
	}
}

// TestQueueDepthWarning verifies that when more goroutines are in-flight than
// MaxConcurrent, a queue depth warning is logged, and that a recovery message
// is logged once the queue clears.
func TestQueueDepthWarning(t *testing.T) {
	logs := captureLog(t)

	// block controls when submitFn returns; inFn counts goroutines inside submitFn.
	block := make(chan struct{})
	var inFn atomic.Int32
	ready := make(chan struct{}) // closed when 2 goroutines are blocked inside submitFn

	s := newTestSubmitter(2, func(_ context.Context, _ *Clients, _ submitParams) (string, *http.Response, error) {
		if inFn.Add(1) == 2 {
			close(ready)
		}
		<-block
		return "", resp(http.StatusAccepted), nil
	})

	ctx := context.Background()
	var wg sync.WaitGroup

	// Launch 2 goroutines that will fill the semaphore and block inside submitFn.
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			s.launchSubmit(ctx, int64(idx))
		}(i)
	}

	// Wait until both are blocked inside submitFn (both semaphore slots taken).
	<-ready

	// Launch a 3rd goroutine: pending becomes 3 > MaxConcurrent=2 → warning.
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.launchSubmit(ctx, 2)
	}()

	// Give the 3rd goroutine time to increment pending and log the warning
	// before we check. 50ms is well above goroutine scheduling latency.
	time.Sleep(50 * time.Millisecond)

	if !strings.Contains(logs(), "[WARNING] submission queue depth") {
		t.Fatalf("expected queue depth warning; log:\n%s", logs())
	}

	// Unblock all submitFns; goroutines exit → pending drops → recovery fires.
	close(block)
	wg.Wait()

	if !strings.Contains(logs(), "[INFO] submission queue cleared") {
		t.Errorf("expected recovery log after queue cleared; log:\n%s", logs())
	}
}
