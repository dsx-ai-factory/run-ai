/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package http_listener

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/advanced/http_listener/encryption"
	api "github.com/run-ai/runai/control-plane/notifications-service/internal/config/advanced/http_listener/openapi/generated"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/postgres/models"
)

func dbRoutesToApiRoutes(routes []*models.RouteGlobalConfig, secretHandler encryption.SecretFieldHandlerer) ([]api.RouteData, error) {
	parsed := []api.RouteData{}
	for _, r := range routes {
		withoutSecrets, err := secretHandler.RemoveSecrets(r.RouteConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to remove secrets route: %s tenant: %s", r.RouteId, r.TenantId)
		}
		parsed = append(parsed, api.RouteData{
			RouteId:     r.RouteId,
			TenantId:    r.TenantId,
			RouteConfig: withoutSecrets,
		})
	}
	return parsed, nil
}

func dbSinksToApiSinks(dbSinks []*models.SinkConfig, secretHandler encryption.SecretFieldHandlerer) ([]api.SinkData, error) {
	parsed := []api.SinkData{}
	for _, dbSink := range dbSinks {
		withoutSecrets, err := secretHandler.RemoveSecrets(dbSink.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to remove secrets for sinkId: %s", dbSink.SinkId)
		}
		parsed = append(parsed, api.SinkData{
			Id:     dbSink.SinkId,
			Type:   dbSink.Type,
			Config: withoutSecrets,
		})
	}
	return parsed, nil
}

func notificationToApiNotificationKind(an []*models.Notification) []api.NotificationKind {
	all := []api.NotificationKind{}
	for _, n := range an {
		all = append(all, api.NotificationKind{
			Category:    n.Category,
			Type:        n.Type,
			Description: n.Description,
			Template:    n.Template,
		})
	}

	return all
}

func dbSinkToApiSink(dbSink *models.SinkConfig, secretHandler encryption.SecretFieldHandlerer) (*api.SinkData, error) {
	withoutSecrets, err := secretHandler.RemoveSecrets(dbSink.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to remove secrets for sinkId: %s", dbSink.SinkId)
	}

	return &api.SinkData{
		Id:     dbSink.SinkId,
		Type:   dbSink.Type,
		Config: withoutSecrets,
	}, nil
}
