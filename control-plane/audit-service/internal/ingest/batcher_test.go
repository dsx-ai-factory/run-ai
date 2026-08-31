/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package ingest

import (
	"context"
	"testing"
	"time"

	"github.com/run-ai/runai/control-plane/audit-service/internal/config"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/control-plane/audit-service/internal/ingest/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var conf = &config.BatcherConfig{
	MaxSize:     4,
	IntervalSec: 2,
}

type TestItem struct {
	f string
}

func NewTestItem(s string) *TestItem {
	return &TestItem{s}
}

func TestFlushDueToInterval(t *testing.T) {
	defer requireTimeIntervalElapsed(t, time.Now())

	expectedCallCount := 1
	batcher, done, cleanup := buildBatcherWithMockedHandler(t, context.Background(), expectedCallCount)
	addItems(batcher, 1) // adding only 1 item, which is less than MaxSize
	<-done
	cleanup()
}

func TestFlushDueToCtx(t *testing.T) {
	defer requireWithinTimeInterval(t, time.Now())

	expectedCallCount := 1
	ctx, cancel := context.WithCancel(context.Background())

	batcher, done, cleanup := buildBatcherWithMockedHandler(t, ctx, expectedCallCount)
	addItems(batcher, 1) // adding only 1 item, which is less than MaxSize
	time.Sleep(1 * time.Second)

	cancel()
	<-done
	cleanup()
}

func TestFlushDueToMaxSize(t *testing.T) {
	defer requireWithinTimeInterval(t, time.Now())

	expectedCallCount := 2
	batcher, done, cleanup := buildBatcherWithMockedHandler(t, context.Background(), expectedCallCount)
	addItems(batcher, conf.MaxSize*expectedCallCount)
	<-done
	cleanup()
}

func buildBatcherWithMockedHandler(t *testing.T, ctx context.Context, expectedCallCount int) (batcher *Batcher[TestItem], done chan bool, cleanup func()) {
	require.NoError(t, runai_config.Set(conf))
	bh := mocks.NewBatchHandler[TestItem](t)

	done = make(chan bool)
	actualCallCount := 0
	bh.On("Execute", mock.Anything, mock.Anything).Return(nil).
		Times(expectedCallCount).
		// after expectedCallCount calls - signal "done"
		Run(func(args mock.Arguments) {
			actualCallCount++
			log.Info().Msgf("handler mock called %d times", actualCallCount)
			if actualCallCount >= expectedCallCount {
				done <- true
			}
		})

	batcher = NewBatcher[TestItem](ctx, bh.Execute)

	return batcher, done, runai_config.FlushAll
}

func addItems(batcher *Batcher[TestItem], itemCount int) {
	for i := 0; i < itemCount; i++ {
		_ = batcher.Add(NewTestItem("1"))
	}
}

func requireWithinTimeInterval(t *testing.T, startTime time.Time) {
	require.True(t, time.Now().Sub(startTime) < (time.Duration(conf.IntervalSec)*time.Second), "time interval shouldn't have reached yet!")
}

func requireTimeIntervalElapsed(t *testing.T, startTime time.Time) {
	require.True(t, time.Now().Sub(startTime) > (time.Duration(conf.IntervalSec)*time.Second), "time interval shouldn have been reached!")
}
