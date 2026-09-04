/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package filters

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
)

type DummyFilter struct{}

func (df DummyFilter) Filter(ctx context.Context, event *events.Event) (bool, error) {
	return false, nil
}
