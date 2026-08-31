/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package history_test

import (
	"errors"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/history"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/history/mocks"
	"github.com/stretchr/testify/require"
)

func TestGetEvents_DelegatesToRepository(t *testing.T) {
	read := true
	want := []*history.Event{{Id: "e1"}, {Id: "e2"}}
	ef := history.EventFilter{TenantId: "t1", UserId: "u1", Read: &read}

	repo := mocks.NewEventsRepository(t)
	svc := history.NewEventsHistoryService(repo)

	repo.On("GetEvents", ef).Return(want, nil).Once()
	got, err := svc.GetEvents(ef)

	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestGetEvents_PropagatesError(t *testing.T) {
	ef := history.EventFilter{TenantId: "t1", UserId: "u1"}

	repo := mocks.NewEventsRepository(t)
	repo.On("GetEvents", ef).Return(nil, errors.New("boom")).Once()

	svc := history.NewEventsHistoryService(repo)
	got, err := svc.GetEvents(ef)

	require.EqualError(t, err, "boom")
	require.Nil(t, got)
}

func TestDeleteEvent_Success(t *testing.T) {
	repo := mocks.NewEventsRepository(t)
	repo.On("DeleteEvent", "event-1").Return(nil).Once()

	svc := history.NewEventsHistoryService(repo)
	require.NoError(t, svc.DeleteEvent("event-1"))
}

func TestDeleteEvent_Failed(t *testing.T) {
	repo := mocks.NewEventsRepository(t)
	repo.On("DeleteEvent", "event-1").Return(errors.New("failed")).Once()

	svc := history.NewEventsHistoryService(repo)
	require.EqualError(t, svc.DeleteEvent("event-1"), "failed")
}

func TestDeleteEvents_DelegatesWithFilter(t *testing.T) {
	read := false
	ef := history.EventFilter{TenantId: "t1", UserId: "u1", Read: &read}
	repo := mocks.NewEventsRepository(t)
	repo.On("DeleteEvents", ef).Return(nil).Once()

	svc := history.NewEventsHistoryService(repo)
	require.NoError(t, svc.DeleteEvents(ef))
}

func TestUpdateEventRead_True_CallsMarkRead(t *testing.T) {
	repo := mocks.NewEventsRepository(t)
	repo.On("MarkRead", "event-1").Return(nil).Once()

	svc := history.NewEventsHistoryService(repo)
	require.NoError(t, svc.UpdateEventRead("event-1", true))
}

func TestUpdateEventRead_False_CallsMarkUnread(t *testing.T) {
	repo := mocks.NewEventsRepository(t)
	repo.On("MarkUnread", "event-1").Return(nil).Once()

	svc := history.NewEventsHistoryService(repo)
	require.NoError(t, svc.UpdateEventRead("event-1", false))
}

func TestUpdateEventRead_PropagatesError(t *testing.T) {
	repo := mocks.NewEventsRepository(t)
	repo.On("MarkRead", "event-1").Return(errors.New("boom")).Once()

	svc := history.NewEventsHistoryService(repo)
	require.EqualError(t, svc.UpdateEventRead("event-1", true), "boom")
}
