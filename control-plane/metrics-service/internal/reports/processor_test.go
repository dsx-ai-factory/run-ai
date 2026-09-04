/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reports_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/reports"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/reports/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStart_FailedToGetReports(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	prg := mocks.NewPendingReportsGetter(t)
	wg := mocks.NewWaitGrouper(t)
	processor := &reports.Processor{
		FetchLimit:         10,
		RunPeriod:          1,
		ConcurrencyLimitCh: make(chan struct{}, 10),
		WaitGroup:          wg,
		ReportsGetter:      prg,
		Collector:          mocks.NewCollectorer(t),
	}

	prg.On("GetPendingReports", ctx, 10).Return(nil, &api.Error{Code: 500})
	prg.On("GetTotalReportsByStatus", ctx).Return(map[string]int{}, nil)
	wg.On("Wait")

	go processor.Start(ctx)
	defer processor.WaitStop()
	defer cancel()

	time.Sleep(time.Millisecond * 1_100)
	assert.Equal(t, len(processor.ConcurrencyLimitCh), 0)
}

func TestStart_NoReport(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	prg := mocks.NewPendingReportsGetter(t)
	wg := mocks.NewWaitGrouper(t)
	processor := &reports.Processor{
		FetchLimit:         10,
		RunPeriod:          1,
		ConcurrencyLimitCh: make(chan struct{}, 10),
		WaitGroup:          wg,
		ReportsGetter:      prg,
		Collector:          mocks.NewCollectorer(t),
	}

	prg.On("GetPendingReports", ctx, 10).Return([]models.Report{}, nil)
	prg.On("GetTotalReportsByStatus", ctx).Return(map[string]int{}, nil)
	wg.On("Wait")

	go processor.Start(ctx)
	defer processor.WaitStop()
	defer cancel()

	time.Sleep(time.Millisecond * 1_100)
	assert.Equal(t, len(processor.ConcurrencyLimitCh), 0)
}

func TestStart_WithReport(t *testing.T) {
	called := make(chan bool, 1)
	ctx, cancel := context.WithCancel(context.Background())
	prg := mocks.NewPendingReportsGetter(t)
	wg := mocks.NewWaitGrouper(t)
	c := mocks.NewCollectorer(t)
	processor := &reports.Processor{
		FetchLimit:         10,
		RunPeriod:          1,
		ConcurrencyLimitCh: make(chan struct{}, 10),
		WaitGroup:          wg,
		ReportsGetter:      prg,
		Collector:          c,
	}

	prg.On("GetPendingReports", ctx, 10).Return([]models.Report{{
		ID:       uuid.New(),
		TenantId: 1,
	}}, nil).Once()
	prg.On("GetTotalReportsByStatus", ctx).Return(map[string]int{}, nil)

	c.On("Collect", mock.Anything).Once()
	wg.On("Add", 1).Once()
	wg.On("Done").Run(
		func(args mock.Arguments) {
			called <- true
		},
	).Once()
	wg.On("Wait")

	go processor.Start(ctx)
	defer processor.WaitStop()
	defer cancel()

	<-called
	assert.Equal(t, len(processor.ConcurrencyLimitCh), 0)
}

func TestStart_WithAsLimitReports(t *testing.T) {
	called := make(chan bool, 2)
	ctx, cancel := context.WithCancel(context.Background())
	prg := mocks.NewPendingReportsGetter(t)
	wg := mocks.NewWaitGrouper(t)
	c := mocks.NewCollectorer(t)
	processor := &reports.Processor{
		FetchLimit:         2,
		RunPeriod:          2,
		ConcurrencyLimitCh: make(chan struct{}, 2),
		WaitGroup:          wg,
		ReportsGetter:      prg,
		Collector:          c,
	}

	prg.On("GetPendingReports", ctx, 2).Return([]models.Report{{
		ID:       uuid.New(),
		TenantId: 1,
	}, {
		ID:       uuid.New(),
		TenantId: 1,
	}}, nil).Once()
	prg.On("GetTotalReportsByStatus", ctx).Return(map[string]int{}, nil)
	c.On("Collect", mock.Anything)
	wg.On("Add", 1)
	wg.On("Done").Run(
		func(args mock.Arguments) {
			called <- true
		},
	).Twice()
	wg.On("Wait")

	go processor.Start(ctx)
	defer processor.WaitStop()

	<-called
	<-called
	cancel()
}

func TestStart_GettingToConcurrencyLimit(t *testing.T) {
	called := make(chan bool, 2)
	ctx, cancel := context.WithCancel(context.Background())
	prg := mocks.NewPendingReportsGetter(t)
	wg := mocks.NewWaitGrouper(t)
	processor := &reports.Processor{
		FetchLimit:         2,
		RunPeriod:          2,
		ConcurrencyLimitCh: make(chan struct{}, 1),
		WaitGroup:          wg,
		ReportsGetter:      prg,
		Collector: NewTestCollector(map[int]time.Duration{
			0: time.Millisecond * 1000,
		}),
	}

	prg.On("GetPendingReports", ctx, 2).Return([]models.Report{{
		ID:       uuid.New(),
		TenantId: 1,
	}, {
		ID:       uuid.New(),
		TenantId: 1,
	}}, nil)
	prg.On("GetTotalReportsByStatus", ctx).Return(map[string]int{}, nil)

	wg.On("Add", 1)
	wg.On("Done").Run(
		func(args mock.Arguments) {
			called <- true
		},
	).Times(1)
	wg.On("Wait")

	go processor.Start(ctx)
	defer cancel()
	defer processor.WaitStop()

	<-called // we expect the first report to be processed
	assert.Equal(t, 0, len(called))
	assert.Equal(t, 0, len(processor.ConcurrencyLimitCh))
}

type TestCollector struct {
	runSleepDuration map[int]time.Duration
	run              int
	lock             sync.Mutex
}

func NewTestCollector(rm map[int]time.Duration) reports.Collectorer {
	return &TestCollector{run: 0, runSleepDuration: rm, lock: sync.Mutex{}}
}

func (c *TestCollector) Collect(_ models.Report) {
	c.lock.Lock()
	r := c.run
	c.run++
	c.lock.Unlock()

	time.Sleep(c.runSleepDuration[r])
}
