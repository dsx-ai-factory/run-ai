/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"context"
	"sync"
	"testing"
	"time"
)

type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Sleep(_ context.Context, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
	return nil
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func TestPacerSubmitsAtTheTargetRateForTheDuration(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)}
	pacer := Pacer{Rate: 2, Duration: 10 * time.Second, Users: 4, Now: clock.Now, Sleep: clock.Sleep}

	var mu sync.Mutex
	var sent []time.Time
	pacing := pacer.Run(context.Background(), func(context.Context, int) {
		mu.Lock()
		sent = append(sent, clock.Now())
		mu.Unlock()
	})

	if pacing.Submitted != 20 || len(sent) != 20 {
		t.Fatalf("2 per second for 10 seconds must submit 20 times, got %d dispatched and %d calls", pacing.Submitted, len(sent))
	}
	if got := pacing.Rate(); got < 1.9 || got > 2.1 {
		t.Fatalf("achieved rate must be about 2 per second, got %.2f", got)
	}
}

func TestPacerSlowWorkersLowerTheAchievedRateInsteadOfQueueingWork(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)}
	pacer := Pacer{Rate: 10, Duration: 2 * time.Second, Users: 1, Now: clock.Now, Sleep: clock.Sleep}

	var inFlight, peak int
	var mu sync.Mutex
	pacing := pacer.Run(context.Background(), func(context.Context, int) {
		mu.Lock()
		inFlight++
		peak = max(peak, inFlight)
		mu.Unlock()
		clock.Advance(500 * time.Millisecond)
		mu.Lock()
		inFlight--
		mu.Unlock()
	})

	if peak != 1 {
		t.Fatalf("one user means one submission in flight, got %d", peak)
	}
	if pacing.Submitted >= 20 {
		t.Fatalf("a worker that needs 500ms cannot sustain 10 per second, yet %d were submitted", pacing.Submitted)
	}
	if got := pacing.Rate(); got > 2.5 {
		t.Fatalf("achieved rate must reflect the slow worker, got %.2f", got)
	}
}

func TestPacerStopsWhenTheContextIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	pacer := Pacer{Rate: 1000, Duration: time.Hour, Users: 2}

	var calls int
	var mu sync.Mutex
	pacing := pacer.Run(ctx, func(context.Context, int) {
		mu.Lock()
		calls++
		if calls == 5 {
			cancel()
		}
		mu.Unlock()
	})

	if pacing.Submitted < 5 || pacing.Submitted > 20 {
		t.Fatalf("the pacer must stop right after cancellation, submitted %d", pacing.Submitted)
	}
}

func TestPacingRateNeedsTwoSubmissions(t *testing.T) {
	if got := (Pacing{Submitted: 1}).Rate(); got != 0 {
		t.Fatalf("one submission is not a rate, got %.2f", got)
	}
}
