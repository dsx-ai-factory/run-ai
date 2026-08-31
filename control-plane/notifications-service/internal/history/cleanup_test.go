/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package history_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/history"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/history/mocks"
	"github.com/stretchr/testify/mock"
)

func notifyOnce(called chan bool) func(mock.Arguments) {
	return func(_ mock.Arguments) {
		called <- true
	}
}

func TestSchedule_TicksAndStopsOnStop(t *testing.T) {
	called := make(chan bool)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	repo := mocks.NewEventsRepository(t)
	repo.On("Cleanup", mock.AnythingOfType("history.CleanupFilter")).Return(nil).Run(notifyOnce(called))

	cleaner := history.NewEventsCleaner(repo, history.CleanupConfig{Interval: 5 * time.Millisecond})
	cleaner.Schedule(ctx)

	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("Cleanup was not invoked by the scheduler")
	}
}

func TestSchedule_ContinuesAfterRepositoryError(t *testing.T) {
	called := make(chan bool)
	repo := mocks.NewEventsRepository(t)
	repo.On("Cleanup", mock.AnythingOfType("history.CleanupFilter")).
		Return(errors.New("cleanup failed")).Run(notifyOnce(called))

	cleaner := history.NewEventsCleaner(repo, history.CleanupConfig{Interval: 5 * time.Millisecond})
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	cleaner.Schedule(ctx)

	for i := 0; i < 2; i++ {
		select {
		case <-called:
		case <-time.After(time.Second):
			t.Fatalf("Cleanup was not invoked %d times", i+1)
		}
	}
}

func TestSchedule_StopsOnExternalContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	repo := mocks.NewEventsRepository(t)
	cleaner := history.NewEventsCleaner(repo, history.CleanupConfig{Interval: time.Hour})

	cleaner.Schedule(ctx)
	cancel()
}
