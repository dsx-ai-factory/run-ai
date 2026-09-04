/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package enrich

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
)

type Enricher interface {
	Enrich(ctx context.Context, event *events.Event) error
}

type Factory interface {
	CreateEnricher(ctx context.Context, name string, configProvider EnrichingConfigProvider) (Enricher, error)
}
