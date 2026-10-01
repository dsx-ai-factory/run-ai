/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestPacedTransportSpacesRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	t.Cleanup(server.Close)
	client := &http.Client{Transport: &pacedTransport{limiter: rate.NewLimiter(100, 1), next: http.DefaultTransport}}
	if burstOf(25) != 100 {
		t.Fatalf("burst must allow four seconds of requests, got %d", burstOf(25))
	}

	start := time.Now()
	for range 5 {
		response, err := client.Get(server.URL)
		if err != nil {
			t.Fatal(err)
		}
		_ = response.Body.Close()
	}

	if elapsed := time.Since(start); elapsed < 35*time.Millisecond {
		t.Fatalf("5 requests at 100/s with burst 1 must take at least 40ms, took %s", elapsed)
	}
}

type mustNotBeCalled struct{ t *testing.T }

func (m mustNotBeCalled) RoundTrip(*http.Request) (*http.Response, error) {
	m.t.Fatal("the request must be rejected by the limiter before reaching the network")
	return nil, nil
}

func TestPacedTransportStopsOnCancelledContext(t *testing.T) {
	transport := &pacedTransport{limiter: rate.NewLimiter(1, 1), next: mustNotBeCalled{t}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:1/", nil)
	_ = transport.limiter.Wait(context.Background())

	start := time.Now()
	_, err := transport.RoundTrip(request)
	if err == nil {
		t.Fatal("a cancelled request must not wait in the limiter")
	}
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("a cancelled request must return at once, waited %s for a token", elapsed)
	}
}
