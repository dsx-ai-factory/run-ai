/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package http

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
	api "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/listen/http/openapi/generated"
)

type HttpService struct {
	listenerId string
	eventCh    chan *events.Event
}

func (hs *HttpService) Notify(_ context.Context, notification api.Notification) (api.ImplResponse, error) {
	event := hs.NotificationToEvent(&notification)
	log.Debug().Msgf("event received in http listener %s", event.UID)
	select {
	case hs.eventCh <- event:
		return api.Response(http.StatusOK, "OK"), nil
	default:
		return api.Response(http.StatusTooManyRequests, api.ApiError{Code: http.StatusTooManyRequests, Message: "QueueIsFull"}), nil
	}
}

func (hs *HttpService) NotifyBatch(_ context.Context, notifications []api.Notification) (api.ImplResponse, error) {
	done := 0
	for _, notification := range notifications {
		event := hs.NotificationToEvent(&notification)
		log.Debug().Msgf("event received in http listener %s", event.UID)
		select {
		case hs.eventCh <- event:
			done++
			continue
		default:
			return api.Response(http.StatusTooManyRequests, api.NotifyBatch200Response{
				Succeed: int32(done),
				Failed:  int32(len(notifications) - done),
				Total:   int32(len(notifications)),
			}), nil
		}
	}

	return api.Response(http.StatusOK, api.NotifyBatch200Response{
		Succeed: int32(done),
		Failed:  int32(len(notifications) - done),
		Total:   int32(len(notifications)),
	}), nil
}

func (hs *HttpService) NotificationToEvent(notification *api.Notification) *events.Event {
	return &events.Event{
		RawEvent:   notification,
		UID:        events.UID(uuid.New().String()),
		ListenerId: hs.listenerId,
		Properties: notification.Properties,
	}
}
