/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package observability

import (
	"context"
	"fmt"
	"time"

	cache_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing"
)

const (
	OTGet        OperationType = "Get"
	OTSet        OperationType = "Set"
	OTDel        OperationType = "Del"
	OTStore      OperationType = "Store"
	OTClearStore OperationType = "ClearStore"
	OTClearCache OperationType = "ClearCache"
	OTLock       OperationType = "Lock"
	OTUnlock     OperationType = "Unlock"
)

type OperationType string

type MetricsWriterer interface {
	RecordOperation(cacheType string, storeName string, operationType OperationType, duration time.Duration, err error, miss bool)
}

type observabilityHandler struct {
	cacheType     string
	metricsWriter MetricsWriterer
}

func NewObservabilityHandler() *observabilityHandler {
	return &observabilityHandler{
		cacheType:     string(config.Get[*cache_config.Cache]().Type),
		metricsWriter: newMetricsWriter(),
	}
}

func (h *observabilityHandler) WithObservability(ctx context.Context, storeName, key string, operationType OperationType, operation func(ctx context.Context) (bool, error)) error {
	spanName := fmt.Sprintf("cache.%s", operationType)
	ctx, span := tracing.GetTracer().StartSpan(ctx, spanName)
	defer span.End()

	span.SetAttribute("cache.type", h.cacheType).
		SetAttribute("cache.store", storeName).
		SetAttribute("cache.operation", string(operationType))

	if key != "" {
		span.SetAttribute("cache.key", key)
	}

	start := time.Now()
	miss, err := operation(ctx)
	duration := time.Since(start)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(tracing.SpanStatusError, fmt.Sprintf("%s failed", operationType))
	} else {
		span.SetStatus(tracing.SpanStatusOK, "")
	}

	h.metricsWriter.RecordOperation(h.cacheType, storeName, operationType, duration, err, miss)

	return err
}
