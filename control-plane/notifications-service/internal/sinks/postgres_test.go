/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sinks

import (
	"errors"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/sinks/mocks"
	"github.com/stretchr/testify/require"
)

func TestPostgresSink_Type(t *testing.T) {
	m := mocks.NewEventDaoer(t)
	s := PostgresSink{typ: PostgresSinkType, eventDao: m}
	require.Equal(t, PostgresSinkType, s.Type())
}

func TestPostgresSinkVerify(t *testing.T) {
	m := mocks.NewEventDaoer(t)
	s := PostgresSink{typ: PostgresSinkType, eventDao: m}

	m.On("IsHealthy").Return(true, nil).Once()

	require.NoError(t, s.Verify())
}

func TestPostgresSinkVerify_Error(t *testing.T) {
	m := mocks.NewEventDaoer(t)
	s := PostgresSink{typ: PostgresSinkType, eventDao: m}

	want := errors.New("db unreachable")
	m.On("IsHealthy").Return(false, want).Once()

	err := s.Verify()
	require.ErrorIs(t, err, want)
}

func TestPostgresSinkVerify_NotHealthy(t *testing.T) {
	m := mocks.NewEventDaoer(t)
	s := PostgresSink{typ: PostgresSinkType, sinkId: "test-pg", eventDao: m}

	m.On("IsHealthy").Return(false, nil).Once()

	err := s.Verify()
	require.Error(t, err)
	require.Contains(t, err.Error(), "postgres sink test-pg dao is not healthy")
}

func TestPostgresSinkSink(t *testing.T) {
	m := mocks.NewEventDaoer(t)
	s := PostgresSink{typ: PostgresSinkType, eventDao: m}

	in := &events.Event{
		UID:        "uid-99",
		ListenerId: "listener-z",
		Properties: map[string]interface{}{
			"tenant_id":   "tenant-a",
			"user_id":     "user-b",
			"description": "desc-c",
		},
	}

	m.On("Store", *in).Return(nil).Once()

	require.NoError(t, s.Sink(in))
}

func TestPostgresSinkSink_Error(t *testing.T) {
	m := mocks.NewEventDaoer(t)
	s := PostgresSink{typ: PostgresSinkType, eventDao: m}

	in := &events.Event{
		UID:        "id",
		ListenerId: "l",
		Properties: map[string]interface{}{
			"tenant_id":   "t",
			"user_id":     "u",
			"description": "d",
		},
	}

	want := errors.New("insert failed")
	m.On("Store", *in).Return(want).Once()

	require.ErrorIs(t, s.Sink(in), want)
}
