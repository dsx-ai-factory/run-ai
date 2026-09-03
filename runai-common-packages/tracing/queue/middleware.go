/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package queue

import (
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

func QueueMiddleware(redisClient *redis.Client, tp tracing.TracerProvider) error {
	switch t := tp.(type) {
	case *tracing.OtelTracerProvider:
		opts := redisotel.WithTracerProvider(t.Provider)
		return redisotel.InstrumentTracing(redisClient, opts)
	default:
		return nil
	}
}
