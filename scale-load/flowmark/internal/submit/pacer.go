/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"context"
	"sync"
	"time"
)

type Pacer struct {
	Rate     int
	Duration time.Duration
	Users    int
	Now      func() time.Time
	Sleep    func(context.Context, time.Duration) error
}

type Pacing struct {
	Submitted     int
	FirstDispatch time.Time
	LastDispatch  time.Time
}

func (p Pacing) Rate() float64 {
	window := p.LastDispatch.Sub(p.FirstDispatch)
	if p.Submitted < 2 || window <= 0 {
		return 0
	}
	return float64(p.Submitted-1) / window.Seconds()
}

func (p Pacer) Run(ctx context.Context, submit func(ctx context.Context, index int)) Pacing {
	now := p.Now
	if now == nil {
		now = time.Now
	}
	sleep := p.Sleep
	if sleep == nil {
		sleep = sleepWithContext
	}

	jobs := make(chan int)
	var workers sync.WaitGroup
	for range max(p.Users, 1) {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for index := range jobs {
				submit(ctx, index)
			}
		}()
	}

	var pacing Pacing
	start := now()
	interval := time.Second / time.Duration(max(p.Rate, 1))
	next := start
	for index := 0; ctx.Err() == nil && next.Sub(start) < p.Duration; index++ {
		if wait := next.Sub(now()); wait > 0 {
			if err := sleep(ctx, wait); err != nil {
				break
			}
		}
		select {
		case jobs <- index:
			dispatched := now()
			if pacing.Submitted == 0 {
				pacing.FirstDispatch = dispatched
			}
			pacing.LastDispatch = dispatched
			pacing.Submitted++
			next = dispatched.Add(interval)
		case <-ctx.Done():
		}
	}
	close(jobs)
	workers.Wait()
	return pacing
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
