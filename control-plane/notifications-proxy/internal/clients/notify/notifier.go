/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package notify

import (
	"context"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/clients/subscription"
	"time"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/clients"
	subpkg "github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/subscription"
	notifyclient "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/pkg/client/types"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

const (
	NotificationCategory        = "Notifications system events"
	NotificationTypeWelcome     = "Welcome"
	deleteWelcomeUserRouteDelay = 15 * time.Second
)

type UserRoutesClienter interface {
	GetRoutes(ctx context.Context, tenantId, userId string) ([]*types.UserRouteConfig, error)
	CreateRoute(ctx context.Context, tenantId, userId string, route *types.UserRouteConfig) error
	UpdateRoute(ctx context.Context, tenantId, userId string, route *types.UserRouteConfig) error
	DeleteRoute(ctx context.Context, tenantId, userId, routeId string) error
}

type notifier struct {
	userRoutesClient            UserRoutesClienter
	deleteWelcomeUserRouteDelay time.Duration
}

func NewNotifier() *notifier {
	userRoutesClient := clients.NewUserRoutes()
	return &notifier{
		userRoutesClient:            userRoutesClient,
		deleteWelcomeUserRouteDelay: deleteWelcomeUserRouteDelay,
	}
}

func (wn *notifier) Welcome(ctx context.Context, tenantId string, email string, channelId string) error {
	routeId := uuid.NewString()
	if err := wn.createWelcomeRoute(ctx, tenantId, email, routeId, channelId); err != nil {
		return err
	}

	wn.notifyWelcomeMessage(tenantId, email)

	wn.deleteWelcomeRouteWithDelay(context.WithoutCancel(ctx), tenantId, email, routeId)

	logging.Infof("Slack app authorized successfully for tenant %s by user %s", tenantId, email)

	return nil
}

func (wn *notifier) notifyWelcomeMessage(tenantId string, email string) {
	notifyclient.BuildNotification(NotificationCategory, NotificationTypeWelcome).
		WithUser(email).
		WithTenant(tenantId).
		WithTitle("Welcome to Our Slack App!").
		WithDescription("You're now successfully connected and will start receiving notifications on Slack.").
		Notify()
}

func (wn *notifier) createWelcomeRoute(ctx context.Context, tenantId, userId, routeId, sinkId string) error {
	route := subscription.BuildRoute(tenantId, userId, &subpkg.Subscription{
		Id: routeId,
		Events: []subpkg.SubscriptionEvent{
			{
				Category: NotificationCategory,
				Types:    []string{NotificationTypeWelcome},
			},
		},
		Channels: []string{sinkId},
	})

	route.Metadata = map[string]string{subscription.InternalRouteFlag: "true"}
	err := wn.userRoutesClient.CreateRoute(ctx, tenantId, userId, route)
	if err == subpkg.ErrSubscriptionAlreadyExists {
		return wn.userRoutesClient.UpdateRoute(ctx, tenantId, userId, route)
	}

	return err
}

func (wn *notifier) deleteWelcomeRouteWithDelay(ctx context.Context, tenantId string, userId string, routeId string) {
	go func() {
		time.Sleep(wn.deleteWelcomeUserRouteDelay)
		err := wn.userRoutesClient.DeleteRoute(ctx, tenantId, userId, routeId)
		if err != nil {
			logging.Warnf("Failed to delete temporary route %s for user %s, tenant %s: %v", routeId, userId, tenantId, err)
		}
	}()
}
