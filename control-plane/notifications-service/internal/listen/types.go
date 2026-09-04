/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package listen

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
)

type Factory interface {
	CreateListener(id string, eventCh chan *events.Event, configProvider ListeningConfigProvider) (Listener, error)
}

type Listener interface {
	Id() string
	Listen(ctx context.Context)
}
