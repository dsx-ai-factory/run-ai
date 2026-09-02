/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package queue

import (
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing"
)

// TestQueueMiddleware tests the QueueMiddleware function.
func TestQueueMiddleware(t *testing.T) {
	// Create a new Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: ":6379",
	})
	tracing.Init(&tracing.Config{
		ServiceName:       "test-service",
		UnderlineTracer:   "otel",
		CollectorEndpoint: "localhost:4317",
		Enabled:           true,
	})
	tp := tracing.GetTracerProvider()
	// Call the QueueMiddleware function
	err := QueueMiddleware(rdb, tp)
	if err != nil {
		t.Fatalf("Expected nil error, but got %v", err)
	}
}
