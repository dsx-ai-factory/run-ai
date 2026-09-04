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
	"testing"

	subpkg "github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/subscription"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/services/mocks"
	"github.com/stretchr/testify/require"
)

var user = "test@runai.com"

func TestCreateSubscription_NoDataInContext(t *testing.T) {
	subscriptionsManager := mocks.NewSubscriptionsManager(t)
	subscriptionsService := SubscriptionsAPIService{subscriptionsManager}

	res, err := subscriptionsService.CreateSubscription(context.Background(), api.Subscription{})
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, res.Code)
	require.Equal(t, err.Error(), "missing tenant id or user id in request context")
}

func TestCreateSubscription_Failed(t *testing.T) {
	ctx := getInitializedContext()
	subscriptionsManager := mocks.NewSubscriptionsManager(t)
	subscriptionsService := SubscriptionsAPIService{subscriptionsManager}
	subscriptionsManager.On("CreateSubscription", ctx, "1", user, &subpkg.Subscription{Events: subpkg.SubscriptionEvents{subpkg.SubscriptionEvent{Category: "Workloads", Types: []string{"All"}}}, Channels: []string{"email"}}).Return(nil, subpkg.ErrCreatingForbidden)

	res, err := subscriptionsService.CreateSubscription(ctx, api.Subscription{Channels: []string{"email"}, Events: []api.SubscriptionEvent{{Category: "Workloads", Types: []string{"All"}}}})
	require.Error(t, err)
	require.Equal(t, http.StatusConflict, res.Code)
	require.Equal(t, err.Error(), "user already has subscription")
}

func TestCreateSubscription_Succeed(t *testing.T) {
	ctx := getInitializedContext()
	subscriptionsManager := mocks.NewSubscriptionsManager(t)
	subscriptionsService := SubscriptionsAPIService{subscriptionsManager}

	route := &subpkg.Subscription{Id: uuid.NewString()}
	subscriptionsManager.On("CreateSubscription", ctx, "1", user, &subpkg.Subscription{Events: nil, Channels: nil, Id: ""}).Return(route, nil)

	res, err := subscriptionsService.CreateSubscription(ctx, api.Subscription{})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, route.Id, res.Body.(api.Subscription).Id)
}

func TestDeleteSubscription_NoDataInContext(t *testing.T) {
	subscriptionsManager := mocks.NewSubscriptionsManager(t)
	subscriptionsService := SubscriptionsAPIService{subscriptionsManager}

	res, err := subscriptionsService.DeleteSubscription(context.Background(), "some_route_id")
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, res.Code)
	require.Equal(t, err.Error(), "missing tenant id or user id in request context")
}

func TestDeleteSubscription_Failed(t *testing.T) {
	ctx := getInitializedContext()
	subscriptionsManager := mocks.NewSubscriptionsManager(t)
	subscriptionsService := SubscriptionsAPIService{subscriptionsManager}
	subscriptionsManager.On("DeleteSubscription", ctx, "1", user, "some_route_id").Return(errors.Join(fmt.Errorf("failed"), subpkg.ErrNotFound))

	res, err := subscriptionsService.DeleteSubscription(ctx, "some_route_id")
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, res.Code)
	require.Equal(t, err.Error(), "subscription not found")
}

func TestDeleteSubscription_ClientFailed(t *testing.T) {
	ctx := getInitializedContext()
	subscriptionsManager := mocks.NewSubscriptionsManager(t)
	subscriptionsService := SubscriptionsAPIService{subscriptionsManager}
	subscriptionsManager.On("DeleteSubscription", ctx, "1", user, "some_route_id").Return(errors.Join(fmt.Errorf("failed"), subpkg.ErrClient))

	res, err := subscriptionsService.DeleteSubscription(ctx, "some_route_id")
	require.Error(t, err)
	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.Equal(t, err.Error(), "failed to delete subscription")
}

func TestDeleteSubscription_Succeed(t *testing.T) {
	ctx := getInitializedContext()
	subscriptionsManager := mocks.NewSubscriptionsManager(t)
	subscriptionsService := SubscriptionsAPIService{subscriptionsManager}
	subscriptionsManager.On("DeleteSubscription", ctx, "1", user, "some_route_id").Return(nil)

	res, err := subscriptionsService.DeleteSubscription(ctx, "some_route_id")
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, res.Code)
}

func TestUpdateSubscription_EmptyContext(t *testing.T) {
	subscriptionsManager := mocks.NewSubscriptionsManager(t)
	subscriptionsService := SubscriptionsAPIService{subscriptionsManager}

	res, err := subscriptionsService.UpdateSubscription(context.Background(), "some_route_id", api.SubscriptionForPut{})
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, res.Code)
	require.Equal(t, err.Error(), "missing tenant id or user id in request context")
}

func TestUpdateSubscription_Failed(t *testing.T) {
	ctx := getInitializedContext()
	subscriptionsManager := mocks.NewSubscriptionsManager(t)
	subscriptionsService := SubscriptionsAPIService{subscriptionsManager}
	subscriptionsManager.On("UpdateSubscription", ctx, "1", user, "some_route_id", &subpkg.Subscription{Id: "some_route_id"}).Return(errors.Join(fmt.Errorf("failed"), subpkg.ErrClient))

	res, err := subscriptionsService.UpdateSubscription(ctx, "some_route_id", api.SubscriptionForPut{})
	require.Error(t, err)
	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.Equal(t, err.Error(), "failed to update subscription")
}

func TestUpdateSubscription_FailedNotFound(t *testing.T) {
	ctx := getInitializedContext()
	subscriptionsManager := mocks.NewSubscriptionsManager(t)
	subscriptionsService := SubscriptionsAPIService{subscriptionsManager}
	subscriptionsManager.On("UpdateSubscription", ctx, "1", user, "some_route_id", &subpkg.Subscription{Id: "some_route_id"}).Return(errors.Join(fmt.Errorf("failed"), subpkg.ErrNotFound))

	res, err := subscriptionsService.UpdateSubscription(ctx, "some_route_id", api.SubscriptionForPut{})
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, res.Code)
	require.Equal(t, err.Error(), "subscription not found")
}

func TestUpdateSubscription_Succeed(t *testing.T) {
	ctx := getInitializedContext()
	subscriptionsManager := mocks.NewSubscriptionsManager(t)
	subscriptionsService := SubscriptionsAPIService{subscriptionsManager}
	subscriptionsManager.On("UpdateSubscription", ctx, "1", user, "some_route_id", &subpkg.Subscription{Id: "some_route_id"}).Return(nil)

	res, err := subscriptionsService.UpdateSubscription(ctx, "some_route_id", api.SubscriptionForPut{})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
}

func TestGetUserSubscriptions_EmptyContext(t *testing.T) {
	subscriptionsManager := mocks.NewSubscriptionsManager(t)
	subscriptionsService := SubscriptionsAPIService{subscriptionsManager}

	res, err := subscriptionsService.GetUserSubscriptions(context.Background())
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, res.Code)
	require.Equal(t, err.Error(), "missing tenant id or user id in request context")
}

func TestGetUserSubscriptions_Failed(t *testing.T) {
	ctx := getInitializedContext()
	subscriptionsManager := mocks.NewSubscriptionsManager(t)
	subscriptionsService := SubscriptionsAPIService{subscriptionsManager}
	subscriptionsManager.On("GetUserSubscriptions", ctx, "1", user).Return(nil, errors.Join(fmt.Errorf("some error"), subpkg.ErrFailedToParse))

	res, err := subscriptionsService.GetUserSubscriptions(ctx)
	require.Error(t, err)
	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.Equal(t, err.Error(), "failed to get subscriptions")
}

func TestGetUserSubscriptions_Succeed(t *testing.T) {
	ctx := getInitializedContext()
	subscriptionsManager := mocks.NewSubscriptionsManager(t)
	subscriptionsService := SubscriptionsAPIService{subscriptionsManager}

	route := subpkg.Subscription{Id: uuid.NewString(), Events: []subpkg.SubscriptionEvent{}, Channels: []string{}}
	subscriptionsManager.On("GetUserSubscriptions", ctx, "1", user).Return([]subpkg.Subscription{route}, nil)
	res, err := subscriptionsService.GetUserSubscriptions(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, api.Subscription{Id: route.Id, Events: []api.SubscriptionEvent{}, Channels: []string{}}, res.Body.([]api.Subscription)[0])
}

func getInitializedContext() context.Context {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	return context.WithValue(ctx, runaictx.CfSubjectID, user)
}
