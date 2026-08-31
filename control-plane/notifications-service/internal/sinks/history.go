/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sinks

import (
	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/postgres/pgsink"
	"github.com/spf13/viper"
)

const HistorySinkType = "history"

type HistorySinkConfig struct {
	HistoryEnabled bool `mapstructure:"history_enabled" json:"history_enabled"`
}

func NewHistorySinkConfig(viper *viper.Viper, id string) *HistorySinkConfig {
	return &HistorySinkConfig{
		HistoryEnabled: viper.GetBool("history_enabled"),
	}
}

type HistorySink struct {
	*PostgresSink
	enabled bool
}

func NewHistorySink(sinkId string, config *HistorySinkConfig) (*HistorySink, error) {
	if !config.HistoryEnabled {
		return &HistorySink{
			PostgresSink: &PostgresSink{typ: HistorySinkType, sinkId: sinkId},
			enabled:      false,
		}, nil
	}

	eventDao, err := pgsink.NewDefaultEventDao()
	if err != nil {
		return nil, err
	}

	return &HistorySink{
		PostgresSink: &PostgresSink{typ: HistorySinkType, sinkId: sinkId, eventDao: eventDao},
		enabled:      true,
	}, nil
}

func (h *HistorySink) Sink(event *events.Event) error {
	if !h.enabled {
		return nil
	}

	return h.PostgresSink.Sink(event)
}

func (h *HistorySink) Verify() error {
	if !h.enabled {
		return nil
	}
	return h.PostgresSink.Verify()
}
