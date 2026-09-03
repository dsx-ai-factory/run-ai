/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package filter

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
)

type Filter interface {
	// Filter - should return true if event dropped the filter
	Filter(ctx context.Context, event *events.Event) (bool, error)
}

type Factory interface {
	CreateFilter(fc *FilterConfig) (Filter, error)
}
