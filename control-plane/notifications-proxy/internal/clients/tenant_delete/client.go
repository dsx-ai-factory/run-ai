/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tenant_delete

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/clients"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/clients/channel"
	proxyapi "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/types"
)

type TenantRoutesClienter interface {
	DeleteRoute(ctx context.Context, tenantId string, routeId string) error
	GetRoute(ctx context.Context, tenantId string) (*types.BasicRouteConfig, error)
}

type UserRoutesClienter interface {
	DeleteAllUserRoutes(ctx context.Context, tenantId string) error
}

type SinksClienter interface {
	DeleteSink(ctx context.Context, channelId string) error
}

type Client struct {
	tenantRoutesClient TenantRoutesClienter
	userRoutesClienter UserRoutesClienter
	sinkClient         channel.SinksClienter
}

func NewClient() *Client {
	return &Client{tenantRoutesClient: clients.NewTenantRoutes(), userRoutesClienter: clients.NewUserRoutes(), sinkClient: clients.NewSinks()}
}

func (c *Client) Delete(ctx context.Context, tenantId string) error {
	return errors.Join(
		c.deleteTenantChannels(ctx, tenantId),
		c.deleteTenantSubscription(ctx, tenantId),
		c.deleteUserSubscriptions(ctx, tenantId),
	)
}

func (c *Client) deleteTenantChannels(ctx context.Context, tenantId string) error {
	err := c.sinkClient.DeleteSink(ctx, channel.SlackChannelName(tenantId))
	if err == nil || errNotFound(err) {
		return nil
	}

	return err
}

func (c *Client) deleteTenantSubscription(ctx context.Context, tenantId string) error {
	route, err := c.tenantRoutesClient.GetRoute(ctx, tenantId)
	if err != nil {
		if errNotFound(err) {
			return nil
		}

		return errors.Join(fmt.Errorf("failed to get tenant route: %w", err))
	}
	if route == nil {
		return nil
	}

	err = c.tenantRoutesClient.DeleteRoute(ctx, tenantId, route.RouteId)
	if err == nil || errNotFound(err) {
		return nil
	}

	return err
}

func (c *Client) deleteUserSubscriptions(ctx context.Context, tenantId string) error {
	err := c.userRoutesClienter.DeleteAllUserRoutes(ctx, tenantId)
	if err == nil || errNotFound(err) {
		return nil
	}

	return err
}

func errNotFound(err error) bool {
	apiErr, ok := err.(*proxyapi.Error)
	return ok && apiErr.Code == http.StatusNotFound
}
