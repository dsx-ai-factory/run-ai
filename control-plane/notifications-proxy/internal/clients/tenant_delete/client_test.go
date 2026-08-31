/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tenant_delete

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/clients/channel"
	channelmocks "github.com/run-ai/runai/control-plane/notifications-proxy/internal/clients/channel/mocks"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/clients/tenant_delete/mocks"
	proxyapi "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/types"
	"github.com/stretchr/testify/require"
)

const tenantId = "1"

func TestDelete_Success(t *testing.T) {
	ctx := context.Background()
	urc := mocks.NewUserRoutesClienter(t)
	trc := mocks.NewTenantRoutesClienter(t)
	sc := channelmocks.NewSinksClienter(t)
	l := &Client{
		userRoutesClienter: urc,
		tenantRoutesClient: trc,
		sinkClient:         sc,
	}

	sc.On("DeleteSink", ctx, channel.SlackChannelName(tenantId)).Return(nil)
	trc.On("GetRoute", ctx, tenantId).Return(&types.BasicRouteConfig{RouteId: "route-id"}, nil)
	trc.On("DeleteRoute", ctx, tenantId, "route-id").Return(nil)
	urc.On("DeleteAllUserRoutes", ctx, tenantId).Return(nil)

	err := l.Delete(ctx, tenantId)
	require.NoError(t, err)
}

func TestDelete_FailedToGetTenantRoute(t *testing.T) {
	ctx := context.Background()
	urc := mocks.NewUserRoutesClienter(t)
	trc := mocks.NewTenantRoutesClienter(t)
	sc := channelmocks.NewSinksClienter(t)
	l := &Client{
		userRoutesClienter: urc,
		tenantRoutesClient: trc,
		sinkClient:         sc,
	}

	sc.On("DeleteSink", ctx, channel.SlackChannelName(tenantId)).Return(nil)
	trc.On("GetRoute", ctx, tenantId).Return(nil, fmt.Errorf("failed"))
	urc.On("DeleteAllUserRoutes", ctx, tenantId).Return(nil)

	err := l.Delete(ctx, tenantId)
	require.ErrorContains(t, err, "failed to get tenant route")
}

func TestDelete_NoTenantRoute(t *testing.T) {
	ctx := context.Background()
	urc := mocks.NewUserRoutesClienter(t)
	trc := mocks.NewTenantRoutesClienter(t)
	sc := channelmocks.NewSinksClienter(t)
	l := &Client{
		userRoutesClienter: urc,
		tenantRoutesClient: trc,
		sinkClient:         sc,
	}

	sc.On("DeleteSink", ctx, channel.SlackChannelName(tenantId)).Return(nil)
	trc.On("GetRoute", ctx, tenantId).Return(nil, nil)
	urc.On("DeleteAllUserRoutes", ctx, tenantId).Return(nil)

	err := l.Delete(ctx, tenantId)
	require.NoError(t, err)
}

func TestDelete_NoTenantRouteError(t *testing.T) {
	ctx := context.Background()
	urc := mocks.NewUserRoutesClienter(t)
	trc := mocks.NewTenantRoutesClienter(t)
	sc := channelmocks.NewSinksClienter(t)
	l := &Client{
		userRoutesClienter: urc,
		tenantRoutesClient: trc,
		sinkClient:         sc,
	}

	sc.On("DeleteSink", ctx, channel.SlackChannelName(tenantId)).Return(nil)
	trc.On("GetRoute", ctx, tenantId).Return(nil, &proxyapi.Error{Code: http.StatusNotFound})
	urc.On("DeleteAllUserRoutes", ctx, tenantId).Return(nil)

	err := l.Delete(ctx, tenantId)
	require.NoError(t, err)
}

func TestDelete_FailedToDeleteTenantRoute(t *testing.T) {
	ctx := context.Background()
	urc := mocks.NewUserRoutesClienter(t)
	trc := mocks.NewTenantRoutesClienter(t)
	sc := channelmocks.NewSinksClienter(t)
	l := &Client{
		userRoutesClienter: urc,
		tenantRoutesClient: trc,
		sinkClient:         sc,
	}

	sc.On("DeleteSink", ctx, channel.SlackChannelName(tenantId)).Return(nil)
	trc.On("GetRoute", ctx, tenantId).Return(&types.BasicRouteConfig{RouteId: "route-id"}, nil)
	trc.On("DeleteRoute", ctx, tenantId, "route-id").Return(fmt.Errorf("failed"))
	urc.On("DeleteAllUserRoutes", ctx, tenantId).Return(nil)

	err := l.Delete(ctx, tenantId)
	require.ErrorContains(t, err, "failed")
}

func TestDelete_FailedToDeleteAllUserRoutes(t *testing.T) {
	ctx := context.Background()
	urc := mocks.NewUserRoutesClienter(t)
	trc := mocks.NewTenantRoutesClienter(t)
	sc := channelmocks.NewSinksClienter(t)
	l := &Client{
		userRoutesClienter: urc,
		tenantRoutesClient: trc,
		sinkClient:         sc,
	}

	sc.On("DeleteSink", ctx, channel.SlackChannelName(tenantId)).Return(nil)
	trc.On("GetRoute", ctx, tenantId).Return(&types.BasicRouteConfig{RouteId: "route-id"}, nil)
	trc.On("DeleteRoute", ctx, tenantId, "route-id").Return(nil)
	urc.On("DeleteAllUserRoutes", ctx, tenantId).Return(fmt.Errorf("failed"))

	err := l.Delete(ctx, tenantId)
	require.ErrorContains(t, err, "failed")
}

func TestDelete_SinkNotFoundIgnored(t *testing.T) {
	ctx := context.Background()
	urc := mocks.NewUserRoutesClienter(t)
	trc := mocks.NewTenantRoutesClienter(t)
	sc := channelmocks.NewSinksClienter(t)
	l := &Client{
		userRoutesClienter: urc,
		tenantRoutesClient: trc,
		sinkClient:         sc,
	}

	sc.On("DeleteSink", ctx, channel.SlackChannelName(tenantId)).Return(proxyapi.NewError(http.StatusNotFound, "not found"))
	trc.On("GetRoute", ctx, tenantId).Return(nil, nil)
	urc.On("DeleteAllUserRoutes", ctx, tenantId).Return(nil)

	err := l.Delete(ctx, tenantId)
	require.NoError(t, err)
}

func TestDelete_FailedToDeleteSink(t *testing.T) {
	ctx := context.Background()
	urc := mocks.NewUserRoutesClienter(t)
	trc := mocks.NewTenantRoutesClienter(t)
	sc := channelmocks.NewSinksClienter(t)
	l := &Client{
		userRoutesClienter: urc,
		tenantRoutesClient: trc,
		sinkClient:         sc,
	}

	sc.On("DeleteSink", ctx, channel.SlackChannelName(tenantId)).Return(fmt.Errorf("failed"))
	trc.On("GetRoute", ctx, tenantId).Return(nil, nil)
	urc.On("DeleteAllUserRoutes", ctx, tenantId).Return(nil)

	err := l.Delete(ctx, tenantId)
	require.ErrorContains(t, err, "failed")
}
