/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/history"
	api "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/history/openapi/generated"
)

type EventsHistoryServicer interface {
	GetEvents(ef history.EventFilter) ([]*history.Event, error)
	DeleteEvent(eventId string) error
	DeleteEvents(ef history.EventFilter) error
	UpdateEventRead(eventId string, read bool) error
}

type EventsAPIService struct {
	service EventsHistoryServicer
}

func NewEventsAPIService(service EventsHistoryServicer) *EventsAPIService {
	return &EventsAPIService{
		service: service,
	}
}

func (s *EventsAPIService) GetEvents(ctx context.Context, tenantId string, userId string, read *bool) (api.ImplResponse, error) {
	events, err := s.service.GetEvents(history.EventFilter{
		TenantId: tenantId,
		UserId:   userId,
		Read:     read,
	})
	if err == nil {
		return api.ImplResponse{Code: http.StatusOK, Body: toApiEvents(events)}, nil
	}

	log.Err(err).Msgf("failed to get events for tenant %s user %s", tenantId, userId)
	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to get events")
}

func (s *EventsAPIService) UpdateEventRead(ctx context.Context, eventId string, eventReadUpdate api.EventReadUpdate) (api.ImplResponse, error) {
	if eventReadUpdate.Read == nil {
		return api.ImplResponse{Code: http.StatusUnprocessableEntity, Body: api.ApiError{Code: http.StatusUnprocessableEntity, Message: "read is required"}}, nil
	}

	err := s.service.UpdateEventRead(eventId, *eventReadUpdate.Read)
	if err == nil {
		return api.ImplResponse{Code: http.StatusOK}, nil
	}

	if errors.Is(err, history.ErrEventNotFound) {
		return api.ImplResponse{Code: http.StatusNotFound}, fmt.Errorf("event not found")
	}

	log.Err(err).Msgf("failed to update event read status for event %s", eventId)
	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to update event read status")
}

func (s *EventsAPIService) DeleteEvent(ctx context.Context, eventId string) (api.ImplResponse, error) {
	err := s.service.DeleteEvent(eventId)
	if err == nil {
		return api.ImplResponse{Code: http.StatusOK}, nil
	}

	if errors.Is(err, history.ErrEventNotFound) {
		return api.ImplResponse{Code: http.StatusNotFound}, fmt.Errorf("event not found")
	}

	log.Err(err).Msgf("failed to delete event %s", eventId)
	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to delete event")
}

func (s *EventsAPIService) DeleteEvents(ctx context.Context, tenantId string, userId string, read *bool) (api.ImplResponse, error) {
	err := s.service.DeleteEvents(history.EventFilter{
		TenantId: tenantId,
		UserId:   userId,
		Read:     read,
	})

	if err == nil {
		return api.ImplResponse{Code: http.StatusOK}, nil
	}

	log.Err(err).Msgf("failed to delete events for tenant %s user %s", tenantId, userId)
	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to delete events")
}

func toApiEvents(events []*history.Event) []api.Event {
	apiEvents := []api.Event{}
	for _, event := range events {
		apiEvents = append(apiEvents, api.Event{
			Id:             event.Id,
			SourceListener: event.SourceListener,
			Properties:     event.Properties,
			CreatedAt:      event.CreatedAt,
			Read:           event.Read,
		})
	}
	return apiEvents
}
