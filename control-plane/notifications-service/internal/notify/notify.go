/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package notify

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/sinks"
)

type Factory interface {
	CreateSinks(configProvider ConfigProvider) ([]sinks.Sinker, error)
}

type Notifier interface {
	Register(sinks.Sinker)
	Notify(events.Event) error
}

type ConfigProvider interface {
	GetSinks(id ...string) (map[string]*sinks.SinkConfig, error)
}
