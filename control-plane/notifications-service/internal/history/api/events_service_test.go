/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/history"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/history/api/mocks"
	api "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/history/openapi/generated"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

func TestDeleteEvent(t *testing.T) {
	servicer := mocks.NewEventsHistoryServicer(t)
	servicer.On("DeleteEvent", "event-1").Return(nil).Once()

	resp, err := NewEventsAPIService(servicer).DeleteEvent(context.Background(), "event-1")

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.Code)
}

func TestDeleteEvent_NotFound(t *testing.T) {
	servicer := mocks.NewEventsHistoryServicer(t)
	servicer.On("DeleteEvent", "event-1").Return(history.ErrEventNotFound).Once()

	resp, err := NewEventsAPIService(servicer).DeleteEvent(context.Background(), "event-1")

	require.EqualError(t, err, "event not found")
	require.Equal(t, http.StatusNotFound, resp.Code)
}

func TestDeleteEvent_InternalError(t *testing.T) {
	servicer := mocks.NewEventsHistoryServicer(t)
	servicer.On("DeleteEvent", "event-1").Return(errors.New("boom")).Once()

	resp, err := NewEventsAPIService(servicer).DeleteEvent(context.Background(), "event-1")

	require.EqualError(t, err, "failed to delete event")
	require.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestDeleteEvents(t *testing.T) {
	read := true
	servicer := mocks.NewEventsHistoryServicer(t)
	servicer.On("DeleteEvents", history.EventFilter{TenantId: "t1", UserId: "u1", Read: &read}).
		Return(nil).Once()

	resp, err := NewEventsAPIService(servicer).DeleteEvents(context.Background(), "t1", "u1", &read)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.Code)
}

func TestDeleteEvents_Failed(t *testing.T) {
	servicer := mocks.NewEventsHistoryServicer(t)
	servicer.On("DeleteEvents", history.EventFilter{TenantId: "t1", UserId: "u1", Read: nil}).
		Return(errors.New("db down")).Once()

	resp, err := NewEventsAPIService(servicer).DeleteEvents(context.Background(), "t1", "u1", nil)

	require.EqualError(t, err, "failed to delete events")
	require.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestGetEvents(t *testing.T) {
	createdAt := time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC)
	read := false
	events := []*history.Event{
		{
			Id:             "event-1",
			SourceListener: "kube",
			Properties:     map[string]interface{}{"k": "v"},
			CreatedAt:      createdAt,
			Read:           false,
		},
	}
	servicer := mocks.NewEventsHistoryServicer(t)
	servicer.On("GetEvents", history.EventFilter{TenantId: "t1", UserId: "u1", Read: &read}).
		Return(events, nil).Once()

	resp, err := NewEventsAPIService(servicer).GetEvents(context.Background(), "t1", "u1", &read)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.Code)
	body, ok := resp.Body.([]api.Event)
	require.True(t, ok)
	require.Len(t, body, 1)
	require.Equal(t, "event-1", body[0].Id)
	require.Equal(t, "kube", body[0].SourceListener)
	require.Equal(t, createdAt, body[0].CreatedAt)
	require.False(t, body[0].Read)
}

func TestGetEvents_EmptySliceWhenNil(t *testing.T) {
	servicer := mocks.NewEventsHistoryServicer(t)
	servicer.On("GetEvents", history.EventFilter{TenantId: "t1", UserId: "u1"}).Return(nil, nil).Once()

	resp, err := NewEventsAPIService(servicer).GetEvents(context.Background(), "t1", "u1", nil)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.Code)
	body, ok := resp.Body.([]api.Event)
	require.True(t, ok)
	require.Empty(t, body)
}

func TestGetEvents_Failed(t *testing.T) {
	servicer := mocks.NewEventsHistoryServicer(t)
	servicer.On("GetEvents", history.EventFilter{TenantId: "t1", UserId: "u1"}).
		Return(nil, errors.New("db down")).Once()

	resp, err := NewEventsAPIService(servicer).GetEvents(context.Background(), "t1", "u1", nil)

	require.EqualError(t, err, "failed to get events")
	require.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestUpdateEventRead(t *testing.T) {
	servicer := mocks.NewEventsHistoryServicer(t)
	servicer.On("UpdateEventRead", "event-1", true).Return(nil).Once()

	resp, err := NewEventsAPIService(servicer).UpdateEventRead(context.Background(), "event-1", api.EventReadUpdate{Read: ptr.To(true)})

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.Code)
}

func TestUpdateEventRead_NoReadStatusChange(t *testing.T) {
	servicer := mocks.NewEventsHistoryServicer(t)
	resp, err := NewEventsAPIService(servicer).UpdateEventRead(context.Background(), "event-1", api.EventReadUpdate{})

	require.NoError(t, err)
	require.Equal(t, http.StatusUnprocessableEntity, resp.Code)
	body, ok := resp.Body.(api.ApiError)
	require.True(t, ok)
	require.Equal(t, int32(http.StatusUnprocessableEntity), int32(body.Code))
	require.Equal(t, "read is required", body.Message)
}

func TestUpdateEventRead_NotFound(t *testing.T) {
	servicer := mocks.NewEventsHistoryServicer(t)
	servicer.On("UpdateEventRead", "event-1", true).Return(history.ErrEventNotFound).Once()

	resp, err := NewEventsAPIService(servicer).UpdateEventRead(context.Background(), "event-1", api.EventReadUpdate{Read: ptr.To(true)})

	require.EqualError(t, err, "event not found")
	require.Equal(t, http.StatusNotFound, resp.Code)
}

func TestUpdateEventRead_Failed(t *testing.T) {
	servicer := mocks.NewEventsHistoryServicer(t)
	servicer.On("UpdateEventRead", "event-1", false).Return(errors.New("boom")).Once()

	resp, err := NewEventsAPIService(servicer).UpdateEventRead(context.Background(), "event-1", api.EventReadUpdate{Read: ptr.To(false)})

	require.EqualError(t, err, "failed to update event read status")
	require.Equal(t, http.StatusInternalServerError, resp.Code)
}
