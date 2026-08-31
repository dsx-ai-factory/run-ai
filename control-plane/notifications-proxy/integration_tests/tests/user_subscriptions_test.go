//go:build integration
// +build integration


/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/go-viper/mapstructure/v2"

	"github.com/google/uuid"
	openapi "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/types"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/require"
)

func TestCreateUserSubscription_Succeed(t *testing.T) {
	routeId := "user_route" + uuid.NewString()
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "1")
	category := "Workloads"
	subscription, _, err := client.SubscriptionsAPI.CreateSubscription(ctx).Subscription(openapi.Subscription{
		Events: []openapi.SubscriptionEvent{
			{
				Category: category,
				Types:    []string{"All"},
			},
		},
		Channels: []string{"some-channel"},
		Id:       &routeId,
	}).Execute()
	require.NoError(t, err)
	defer client.SubscriptionsAPI.DeleteSubscription(ctx, routeId).Execute()

	require.NotNil(t, subscription)
	require.Equal(t, *subscription.Id, routeId)
	require.Equal(t, subscription.Channels, []string{"some-channel"})

	routeData := getSubscriptionConfigFromService(t, ctx, "1", "1", routeId)
	require.NotNil(t, routeData)
	require.Equal(t, routeData.ListenerId, "http-runai")
	require.Equal(t, routeData.NotifyIds, []string{"some-channel"})
	require.Equal(t, routeData.BasicRouteConfig.Filtering.Policy, "any")
	require.Equal(t, *routeData.BasicRouteConfig.Filtering.Filters[0].Matcher, types.MatcherFilterConfig{
		{
			Key:       "category",
			Value:     "Workloads",
			MatchType: "==",
		},
		{
			Key:       "type",
			Value:     "All",
			MatchType: "=~",
		},
	})
}

func TestCreateUserSubscription_NoId_Succeed(t *testing.T) {
	channelId := "some-channel-" + uuid.New().String()
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "123")
	category := "Workloads"
	subscription, _, err := client.SubscriptionsAPI.CreateSubscription(ctx).Subscription(openapi.Subscription{
		Events: []openapi.SubscriptionEvent{
			{
				Category: category,
				Types:    []string{"All"},
			},
		},
		Channels: []string{channelId},
	}).Execute()
	require.NoError(t, err)
	defer client.SubscriptionsAPI.DeleteSubscription(ctx, *subscription.Id).Execute()
	require.NotEmpty(t, *subscription.Id)
	require.Equal(t, subscription.Channels, []string{channelId})

	resp := getSubscriptionConfigFromService(t, ctx, "1", "123", *subscription.Id)
	require.Equal(t, resp.ListenerId, "http-runai")
	require.NotEmpty(t, resp.Filtering)
	require.Equal(t, resp.NotifyIds, []string{channelId})
}

func TestCreateUserSubscription_EmptyRoute_Succeed(t *testing.T) {
	routeId := "user_route" + uuid.NewString()
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "1")
	subscription, _, err := client.SubscriptionsAPI.CreateSubscription(ctx).Subscription(openapi.Subscription{
		Events:   []openapi.SubscriptionEvent{},
		Channels: []string{},
		Id:       &routeId,
	}).Execute()
	require.NoError(t, err)
	defer client.SubscriptionsAPI.DeleteSubscription(ctx, routeId).Execute()
	require.Equal(t, subscription.Channels, []string{})

	resp := getSubscriptionConfigFromService(t, ctx, "1", "1", *subscription.Id)
	require.Equal(t, resp.ListenerId, "http-runai")
	require.Equal(t, resp.NotifyIds, []string{})
}

func TestGetUserSubscription_NotExist(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1001")
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "123")
	subscriptions, res, err := client.SubscriptionsAPI.GetUserSubscriptions(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, res.StatusCode, http.StatusOK)
	require.Empty(t, subscriptions)
}

func TestGetUserSubscription_Exist(t *testing.T) {
	routeId := "test-get-routes-" + uuid.NewString()
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "1")
	category := "Workloads"
	events := []openapi.SubscriptionEvent{
		{
			Category: category,
			Types:    []string{"All"},
		},
	}
	_, _, err := client.SubscriptionsAPI.CreateSubscription(ctx).Subscription(openapi.Subscription{
		Events:   events,
		Channels: []string{"some-channel"},
		Id:       &routeId,
	}).Execute()
	require.NoError(t, err)
	defer client.SubscriptionsAPI.DeleteSubscription(ctx, routeId).Execute()

	subscriptions, res, err := client.SubscriptionsAPI.GetUserSubscriptions(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, res.StatusCode, http.StatusOK)
	require.NotEmpty(t, subscriptions)
	require.NotEmpty(t, *subscriptions[0].Id)
	require.Equal(t, subscriptions[0].Events, events)
}

func TestGetUserSubscription_ShouldFailIfOtherNonInternalUserRouteAlreadyExists(t *testing.T) {
	routeId := "test-get-routes-" + uuid.NewString()
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "1")
	category := "Workloads"
	events := []openapi.SubscriptionEvent{
		{
			Category: category,
			Types:    []string{"All"},
		},
	}
	_, _, err := client.SubscriptionsAPI.CreateSubscription(ctx).Subscription(openapi.Subscription{
		Events:   events,
		Channels: []string{"some-channel"},
		Id:       &routeId,
	}).Execute()
	require.NoError(t, err)
	defer client.SubscriptionsAPI.DeleteSubscription(ctx, routeId).Execute()

	routeId2 := "test-get-routes-" + uuid.NewString()
	_, res, err := client.SubscriptionsAPI.CreateSubscription(ctx).Subscription(openapi.Subscription{
		Events:   events,
		Channels: []string{"some-channel"},
		Id:       &routeId2,
	}).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusConflict, res.StatusCode)
	body, _ := io.ReadAll(res.Body)
	require.Contains(t, string(body), "user already has subscription")

	subscriptions, res, err := client.SubscriptionsAPI.GetUserSubscriptions(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, res.StatusCode, http.StatusOK)
	require.NotEmpty(t, subscriptions)
	require.NotEmpty(t, *subscriptions[0].Id)
	require.Equal(t, subscriptions[0].Events, events)
}

func getSubscriptionConfigFromService(t *testing.T, ctx context.Context, tenantId, userId, routeId string) *types.UserRouteConfig {
	routes, _, err := notificationsClient.UserRouteAPI.GetUserRoutes(ctx, tenantId, userId).Execute()
	require.NoError(t, err)
	for _, r := range routes {
		if r.RouteId == routeId {
			cfg := types.BasicRouteConfig{}
			mapstructure.Decode(r.RouteConfig, &cfg)
			return &types.UserRouteConfig{UserId: userId, BasicRouteConfig: cfg}
		}
	}
	return nil
}
