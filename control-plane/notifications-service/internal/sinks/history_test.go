/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sinks

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/sinks/mocks"
	"github.com/stretchr/testify/require"
)

var sinkId = uuid.New().String()

func TestHistorySink_Type(t *testing.T) {
	m := mocks.NewEventDaoer(t)
	s := HistorySink{PostgresSink: &PostgresSink{typ: HistorySinkType, sinkId: sinkId, eventDao: m}, enabled: true}
	require.Equal(t, HistorySinkType, s.Type())
}

func TestHistorySink_Verify(t *testing.T) {
	m := mocks.NewEventDaoer(t)
	s := HistorySink{PostgresSink: &PostgresSink{typ: HistorySinkType, sinkId: sinkId, eventDao: m}, enabled: true}

	m.On("IsHealthy").Return(true, nil).Once()

	require.NoError(t, s.Verify())
}

func TestHistorySink_Verify_Error(t *testing.T) {
	m := mocks.NewEventDaoer(t)
	s := HistorySink{PostgresSink: &PostgresSink{typ: HistorySinkType, sinkId: sinkId, eventDao: m}, enabled: true}

	want := errors.New("db unreachable")
	m.On("IsHealthy").Return(false, want).Once()

	err := s.Verify()
	require.ErrorIs(t, err, want)
}

func TestHistorySink_Verify_NotHealthy(t *testing.T) {
	m := mocks.NewEventDaoer(t)
	s := HistorySink{PostgresSink: &PostgresSink{typ: HistorySinkType, sinkId: sinkId, eventDao: m}, enabled: true}

	m.On("IsHealthy").Return(false, nil).Once()

	err := s.Verify()
	require.Error(t, err)
	require.Contains(t, err.Error(), "history sink "+sinkId+" dao is not healthy")
}

func TestHistorySink_Sink(t *testing.T) {
	m := mocks.NewEventDaoer(t)
	ev := &events.Event{
		UID:        "uid-1",
		ListenerId: "listener-1",
		Properties: map[string]interface{}{
			"tenant_id":   "t1",
			"user_id":     "u1",
			"description": "d1",
		},
	}
	s := HistorySink{PostgresSink: &PostgresSink{typ: HistorySinkType, sinkId: sinkId, eventDao: m}, enabled: true}

	m.On("Store", *ev).Return(nil).Once()

	require.NoError(t, s.Sink(ev))
}

func TestHistorySink_Sink_Error(t *testing.T) {
	m := mocks.NewEventDaoer(t)
	ev := &events.Event{Properties: map[string]interface{}{}}
	s := HistorySink{PostgresSink: &PostgresSink{typ: HistorySinkType, sinkId: sinkId, eventDao: m}, enabled: true}

	want := errors.New("write failed")
	m.On("Store", *ev).Return(want).Once()

	err := s.Sink(ev)
	require.ErrorIs(t, err, want)
}

func TestHistorySink_Disabled_Sink_NoOp(t *testing.T) {
	s, err := NewHistorySink(sinkId, &HistorySinkConfig{HistoryEnabled: false})
	require.NoError(t, err)
	require.NotNil(t, s)

	ev := &events.Event{UID: "uid-1"}
	require.NoError(t, s.Sink(ev))
}

func TestHistorySink_Disabled_Verify_NoOp(t *testing.T) {
	s, err := NewHistorySink(sinkId, &HistorySinkConfig{HistoryEnabled: false})
	require.NoError(t, err)
	require.NotNil(t, s)

	require.NoError(t, s.Verify())
}

func TestHistorySink_Disabled_Type(t *testing.T) {
	s, err := NewHistorySink(sinkId, &HistorySinkConfig{HistoryEnabled: false})
	require.NoError(t, err)
	require.Equal(t, HistorySinkType, s.Type())
}
