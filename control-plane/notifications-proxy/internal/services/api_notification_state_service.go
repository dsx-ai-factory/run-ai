/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/clients/channel"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/clients/state"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/clients/subscription"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/clients/tenant_migration"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/notification_state"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type NotificationStateManager interface {
	GetNotificationState(ctx context.Context, tenantId string) ([]notification_state.Stater, error)
	GetDetailedNotificationState(ctx context.Context, tenantId string) ([]notification_state.Stater, error)
	UpdateNotificationState(ctx context.Context, tenantId string, notificationState []notification_state.Stater) ([]notification_state.Stater, error)
}

type NotificationStateAPIService struct {
	manager NotificationStateManager
}

// NewNotificationStateAPIService creates a default api service
func NewNotificationStateAPIService() api.NotificationStateAPIServicer {
	return &NotificationStateAPIService{
		manager: notification_state.NewNotificationStateManager(state.NewClient(), channel.NewClient(), subscription.NewClient(), tenant_migration.NewClient()),
	}
}

func (nss *NotificationStateAPIService) GetNotificationState(ctx context.Context) (api.ImplResponse, error) {
	return nss.getNotificationState(ctx, false)
}

func (nss *NotificationStateAPIService) GetDetailedNotificationState(ctx context.Context) (api.ImplResponse, error) {
	return nss.getNotificationState(ctx, true)
}

func (nss *NotificationStateAPIService) getNotificationState(ctx context.Context, withDetails bool) (api.ImplResponse, error) {
	tenantId := runaictx.GetString(ctx, runaictx.CfTenantId)
	if tenantId == nil {
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id in request context")
	}

	getNotificationState := nss.manager.GetNotificationState
	if withDetails {
		getNotificationState = nss.manager.GetDetailedNotificationState
	}

	nstate, err := getNotificationState(ctx, *tenantId)
	if err == nil {
		if withDetails {
			return api.Response(http.StatusOK, getAPIDetailedNotificationsState(nstate)), nil
		}
		return api.Response(http.StatusOK, getAPINotificationsStatePlain(nstate)), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get tenant route from server for tenant: %+v", err)
	if errors.Is(err, notification_state.ErrClientFailure) {
		return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("request to get notification state failed, internal error")
	}
	return api.ImplResponse{Code: http.StatusInternalServerError}, nil
}

func (nss *NotificationStateAPIService) UpdateNotificationState(ctx context.Context, notificationState api.NotificationState) (api.ImplResponse, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id in request context")
	}

	domainState := toDomainState(notificationState, tenantId)
	channel, action := getChannelTypeAndAction(domainState)
	nss.auditUpdate(ctx, tenantId, channel, action)
	newNotificationState, err := nss.manager.UpdateNotificationState(ctx, tenantId, domainState)
	if err == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("notifications state updated, state: %+v", newNotificationState)
		return api.Response(http.StatusOK, getAPINotificationsStatePlain(newNotificationState)), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to update notifications state %+v", err)
	if errors.Is(err, notification_state.ErrFailedToRemoveSinkFromUserRoutes) {
		return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to remove sink from user routes")
	}
	if errors.Is(err, notification_state.ErrFailedToUpdateNotificationState) {
		return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to update notification state")
	}
	if errors.Is(err, notification_state.ErrNoSinkConfigured) {
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("no sinks server configured, bad request")
	}
	if errors.Is(err, notification_state.ErrIncompatibleSinkType) {
		return api.ImplResponse{Code: http.StatusConflict}, fmt.Errorf("incompatible sink type")
	}

	return api.ImplResponse{Code: http.StatusInternalServerError}, nil
}

func (*NotificationStateAPIService) auditUpdate(ctx context.Context, tenantId, channel, action string) {
	audit.AuditFromCtx(ctx).WithEntityType("Notifications").WithEntityName(channel).WithEntityId(state.TenantRouteId(tenantId)).WithAction(action)
}

func getAPINotificationsStatePlain(staters []notification_state.Stater) api.NotificationState {
	if len(staters) == 0 {
		return api.NotificationState{}
	}
	r := api.NotificationState{}
	for _, s := range staters {
		switch s.GetType() {
		case "email":
			r.Email = getAPIState(s)
		case "slack":
			r.Slack = getAPIState(s)
		}
	}
	return r
}

func getAPIDetailedNotificationsState(staters []notification_state.Stater) api.DetailedNotificationState {
	if len(staters) == 0 {
		return api.DetailedNotificationState{}
	}
	r := api.DetailedNotificationState{}
	for _, s := range staters {
		switch s.GetType() {
		case "email":
			r.Email = getAPIState(s)
		case "slack":
			r.Slack = getAPISlackStateWithDetails(s)
		}
	}
	return r
}

func getAPISlackStateWithDetails(s notification_state.Stater) api.DetailedNotificationStateSlack {
	v := s.IsEnabled()
	md, ok := s.GetMetadata().(*state.SlackMetadata)
	if !ok || md == nil {
		return api.DetailedNotificationStateSlack{Enabled: &v}
	}

	return api.DetailedNotificationStateSlack{
		Enabled: &v,
		Settings: api.SlackDetailsSettings{
			WorkspaceName:            md.WorkspaceName,
			AppCredentialsConfigured: md.AppCredentialsConfigured,
		},
	}
}

func getAPIState(s notification_state.Stater) api.NotificationStateStatus {
	v := s.IsEnabled()
	return api.NotificationStateStatus{Enabled: &v}
}

func toDomainState(nstate api.NotificationState, tenantId string) []notification_state.Stater {
	var parsed []notification_state.Stater
	if nstate.Email.Enabled != nil {
		parsed = append(parsed, &state.EmailState{Enabled: nstate.Email.Enabled})
	}
	if nstate.Slack.Enabled != nil {
		parsed = append(parsed, &state.SlackState{TenantId: tenantId, Enabled: nstate.Slack.Enabled})
	}
	return parsed
}

func getChannelTypeAndAction(domainState []notification_state.Stater) (string, string) {
	for _, ds := range domainState {
		if ds.IsEnabled() {
			return capitalize(ds.GetType()), "Enable"
		}
	}

	if len(domainState) > 0 {
		return capitalize(domainState[0].GetType()), "Disable"
	}

	return "", "Disable"
}
