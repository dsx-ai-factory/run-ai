/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package notify

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/sinks"
	"github.com/stretchr/testify/require"
)

func TestNewSink_HistoryDisabled_ReturnsNoOpHistory(t *testing.T) {
	sc := &sinks.SinkConfig{History: &sinks.HistorySinkConfig{HistoryEnabled: false}}

	s, err := newSink("my_history", sc)
	require.NoError(t, err)
	require.NotNil(t, s)
	require.Equal(t, sinks.HistorySinkType, s.Type())

	_, ok := s.(*sinks.HistorySink)
	require.True(t, ok)

	require.NoError(t, s.Verify())
	require.NoError(t, s.Sink(&events.Event{UID: "uid-1"}))
}

func TestNewSink_PostgresUnreachable_ReturnsError(t *testing.T) {
	sc := &sinks.SinkConfig{Postgres: &sinks.PostgresSinkConfig{
		Host:     "127.0.0.1",
		Port:     "1",
		Database: "notifications",
		Schema:   "public",
		Table:    "events",
		User:     "user",
		Password: "<password>",
		SslMode:  "disable",
	}}

	s, err := newSink("my_postgres", sc)
	require.Error(t, err)
	require.Contains(t, err.Error(), "postgres sink my_postgres failed to create event dao")
	// A typed nil inside the interface would pass a plain nil check and panic on first use.
	require.True(t, s == nil, "expected a nil Sinker, got %#v", s)
}
