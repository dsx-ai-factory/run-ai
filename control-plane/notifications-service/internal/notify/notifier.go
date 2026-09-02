/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package notify

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/sinks"
)

func NewEventNotifier() Notifier {
	return &EventNotifier{
		sinks: []sinks.Sinker{},
	}
}

type EventNotifier struct {
	sinks []sinks.Sinker
}

func (en *EventNotifier) Register(sink sinks.Sinker) {
	en.sinks = append(en.sinks, sink)
}

func (en *EventNotifier) Notify(events.Event) error {
	return nil
}
