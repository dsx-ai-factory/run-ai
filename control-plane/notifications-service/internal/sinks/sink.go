/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sinks

import (
	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
)

type Sinker interface {
	Sink(event *events.Event) error
	Verify() error
	Type() string
}

type SinkConfig struct {
	Email    *EmailSinkConfig    `json:"email"`
	Slack    *SlackSinkConfig    `json:"slack"`
	Postgres *PostgresSinkConfig `json:"postgres"`
	History  *HistorySinkConfig  `json:"history"`
}
