/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package state

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/clients"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/notification_state"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/slack"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/pkg/client/types"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type TenantRoutesClienter interface {
	GetRoute(ctx context.Context, tenantId string) (*types.BasicRouteConfig, error)
	CreateOrUpdateRoute(ctx context.Context, route *types.BasicRouteConfig) error
}

type SlackCredentialsGetter interface {
	Get(ctx context.Context) (*slack.SlackCredentials, error)
}

type Client struct {
	tenantRoutesClient     TenantRoutesClienter
	slackCredentialsGetter SlackCredentialsGetter
}

func NewClient() *Client {
	return &Client{
		tenantRoutesClient:     clients.NewTenantRoutes(),
		slackCredentialsGetter: dao.NewSlackCredentials(),
	}
}

// ------------------------------------------- START Implements NotificationState's StateClienter --------------------------

func (cs *Client) Get(ctx context.Context, tenantId string) ([]notification_state.Stater, error) {
	route, defaultStates, err := cs.fetchRoute(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	if route == nil {
		return defaultStates, nil
	}

	notificationState := ToState(route.Metadata)

	return notificationState, nil
}

func (cs *Client) GetDetails(ctx context.Context, tenantId string) ([]notification_state.Stater, error) {
	route, defaultStates, err := cs.fetchRoute(ctx, tenantId)
	if err != nil {
		return nil, err
	}

	slackAppConfigured := cs.isSlackAppConfigured(ctx)

	if route == nil {
		return ToDetailedState(FromState(defaultStates), slackAppConfigured), nil
	}

	return ToDetailedState(route.Metadata, slackAppConfigured), nil
}

func (cs *Client) UpdateState(ctx context.Context, tenantId string, notificationState []notification_state.Stater) ([]notification_state.Stater, error) {
	route, err := cs.getRouteOrDefault(ctx, tenantId)
	if err != nil {
		return nil, errors.Join(err, notification_state.ErrClientFailure)
	}

	newNotificationState, err := cs.updateNotificationState(ctx, route, notificationState)
	if err != nil {
		return nil, errors.Join(err, notification_state.ErrFailedToUpdateNotificationState)
	}

	return newNotificationState, nil
}

// ------------------------------------------- END Implements NotificationState's StateClienter ----------------------------

// ------------------------------------------- START Implements Email's NotificationStateClienter --------------------------

func (cs *Client) GetEmail(ctx context.Context, tenantId string) (notification_state.Stater, error) {
	route, _, err := cs.fetchRoute(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	if route == nil {
		return &EmailState{Enabled: ptr.To(false)}, nil
	}

	return ToEmailState(route.Metadata), nil
}

func (cs *Client) UpdateEmail(ctx context.Context, tenantId string, emailState notification_state.Stater) (notification_state.Stater, error) {
	route, err := cs.getRouteOrDefault(ctx, tenantId)
	if err != nil {
		return nil, errors.Join(err, notification_state.ErrClientFailure)
	}

	if _, err = cs.updateNotificationState(ctx, route, []notification_state.Stater{emailState}); err != nil {
		return nil, errors.Join(err, notification_state.ErrFailedToUpdateNotificationState)
	}

	return emailState, nil
}

// ------------------------------------------- END Implements Email's NotificationStateClienter --------------------------

// ------------------------------------------- START Implements Slack's NotificationStateClienter --------------------------

func (cs *Client) UpdateSlackNotificationState(ctx context.Context, slackState slack.NotificationState) error {
	route, err := cs.getRouteOrDefault(ctx, slackState.TenantId)
	if err != nil {
		return fmt.Errorf("failed to retrieve route of tenant %s: %w", slackState.TenantId, err)
	}

	metadataUpdate := map[string]string{}
	if slackState.Enabled != nil {
		metadataUpdate[SlackEnabled] = fmt.Sprintf("%t", *slackState.Enabled)
	}
	if slackState.WorkspaceName != nil {
		metadataUpdate[SlackWorkspaceName] = *slackState.WorkspaceName
	}

	route.Metadata = MergeMetadata(getRouteMetadata(route), metadataUpdate)

	if err := cs.tenantRoutesClient.CreateOrUpdateRoute(ctx, route); err != nil {
		return err
	}

	return nil
}

func (cs *Client) GetSlack(ctx context.Context, tenantId string) (notification_state.Stater, error) {
	route, _, err := cs.fetchRoute(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	if route == nil {
		return &SlackState{Enabled: ptr.To(false)}, nil
	}

	return ToSlackState(route.Metadata), nil
}

func (cs *Client) GetSlackDetails(ctx context.Context, tenantId string) (notification_state.Stater, error) {
	route, defaultStates, err := cs.fetchRoute(ctx, tenantId)
	if err != nil {
		return nil, err
	}

	slackAppConfigured := cs.isSlackAppConfigured(ctx)

	if route == nil {
		return ToDetailedSlackState(FromState(defaultStates), slackAppConfigured), nil
	}

	return ToDetailedSlackState(route.Metadata, slackAppConfigured), nil
}

func (cs *Client) UpdateSlack(ctx context.Context, tenantId string, slackState notification_state.Stater) (notification_state.Stater, error) {
	route, err := cs.getRouteOrDefault(ctx, tenantId)
	if err != nil {
		return nil, errors.Join(err, notification_state.ErrClientFailure)
	}

	if _, err = cs.updateNotificationState(ctx, route, []notification_state.Stater{slackState}); err != nil {
		return nil, errors.Join(err, notification_state.ErrFailedToUpdateNotificationState)
	}

	return slackState, nil
}

// ------------------------------------------- END Implements Slack's NotificationStateClienter -------------------

// ------------------------------------------- START Private Helpers ----------------------------------------------

func (cs *Client) fetchRoute(ctx context.Context, tenantId string) (*types.BasicRouteConfig, []notification_state.Stater, error) {
	route, err := cs.tenantRoutesClient.GetRoute(ctx, tenantId)
	if err != nil {
		return nil, nil, errors.Join(err, notification_state.ErrClientFailure)
	}

	if route == nil {
		defaultStates := []notification_state.Stater{
			&EmailState{Enabled: ptr.To(false)},
			&SlackState{Enabled: ptr.To(false)},
		}
		return nil, defaultStates, nil
	}

	return route, nil, nil
}

func (cs *Client) getRouteOrDefault(ctx context.Context, tenantId string) (*types.BasicRouteConfig, error) {
	route, err := cs.tenantRoutesClient.GetRoute(ctx, tenantId)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve route of tenant %s: %w", tenantId, err)
	}
	if route == nil {
		return getDefaultTenantRoute(tenantId), nil
	}

	return route, nil
}

func (cs *Client) updateNotificationState(ctx context.Context, route *types.BasicRouteConfig, notificationState []notification_state.Stater) ([]notification_state.Stater, error) {
	metadataToUpdate := FromState(notificationState)
	route.Metadata = MergeMetadata(getRouteMetadata(route), metadataToUpdate)

	err := cs.tenantRoutesClient.CreateOrUpdateRoute(ctx, route)
	if err != nil {
		return nil, err
	}

	return ToState(route.Metadata), nil
}

func (cs *Client) isSlackAppConfigured(ctx context.Context) bool {
	creds, err := cs.slackCredentialsGetter.Get(ctx)
	if errors.Is(err, slack.ErrMissingAppCredentials) {
		return false
	}
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to check if slack credentials configured: %v", err)
		return false
	}
	return creds != nil
}

func getDefaultTenantRoute(tenantId string) *types.BasicRouteConfig {
	return &types.BasicRouteConfig{
		RouteId:    TenantRouteId(tenantId),
		TenantId:   tenantId,
		ListenerId: "http-runai",
		Filtering: types.FilteringConfig{
			Policy: "drop",
		},
	}
}

// ------------------------------------------- END Private Helpers ------------------------------------------------

func TenantRouteId(tenantId string) string {
	return fmt.Sprintf("default-route-%s", tenantId)
}
