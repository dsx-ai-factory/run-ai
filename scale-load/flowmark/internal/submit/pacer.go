/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"context"
	"math"
	"sync"
	"time"
)

type Pacer struct {
	Rate      int
	SpawnRate int
	Duration  time.Duration
	Users     int
	Now       func() time.Time
	Sleep     func(context.Context, time.Duration) error
}

type Pacing struct {
	Planned       int
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

	pacing := Pacing{Planned: p.planned()}
	start := now()
	for tick := 0; ctx.Err() == nil && p.tickAt(tick) < p.Duration; {
		due := start.Add(p.tickAt(tick))
		if wait := due.Sub(now()); wait > 0 {
			if err := sleep(ctx, wait); err != nil {
				break
			}
		}
		select {
		case jobs <- pacing.Submitted:
			dispatched := now()
			if pacing.Submitted == 0 {
				pacing.FirstDispatch = dispatched
			}
			pacing.LastDispatch = dispatched
			pacing.Submitted++
			for tick++; start.Add(p.tickAt(tick)).Before(dispatched); tick++ {
			}
		case <-ctx.Done():
		}
	}
	close(jobs)
	workers.Wait()
	return pacing
}

func (p Pacer) rampSeconds() float64 {
	if p.SpawnRate <= 0 {
		return 0
	}
	return float64(max(p.Users, 1)) / float64(p.SpawnRate)
}

func (p Pacer) tickAt(tick int) time.Duration {
	rate := float64(max(p.Rate, 1))
	ramp := p.rampSeconds()
	rampTicks := rate * ramp / 2
	seconds := ramp + (float64(tick)-rampTicks)/rate
	if float64(tick) < rampTicks {
		seconds = math.Sqrt(2 * ramp * float64(tick) / rate)
	}
	return time.Duration(seconds * float64(time.Second))
}

func (p Pacer) planned() int {
	rate := float64(max(p.Rate, 1))
	ramp := p.rampSeconds()
	duration := p.Duration.Seconds()
	ticks := rate * duration * duration / (2 * ramp)
	if ramp == 0 || duration >= ramp {
		ticks = rate*ramp/2 + rate*(duration-ramp)
	}
	return int(math.Ceil(ticks - 1e-9))
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
