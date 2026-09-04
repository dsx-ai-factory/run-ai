/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package subscription_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/subscription"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/subscription/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var user = "test@runai.com"

func TestCreateSubscription_Failed(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewUserRoutesClienter(t)
	manager := subscription.NewManager(client)

	client.On("GetUserSubscriptions", ctx, "1", user).Return([]subscription.Subscription{}, nil)
	client.On("Create", ctx, "1", user, mock.Anything).Return(nil, errors.Join(fmt.Errorf("failed to create route"), subscription.ErrClient))
	route, err := manager.CreateSubscription(ctx, "1", user, &subscription.Subscription{
		Id:       "some_id",
		Events:   []subscription.SubscriptionEvent{},
		Channels: []string{},
	})
	require.Error(t, err)
	require.ErrorContains(t, err, subscription.ErrClient.Error())
	require.ErrorContains(t, err, "failed to create route")
	require.Nil(t, route)
}

func TestCreateSubscription_Succeed(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewUserRoutesClienter(t)
	manager := subscription.NewManager(client)

	client.On("GetUserSubscriptions", ctx, "1", user).Return([]subscription.Subscription{}, nil)
	client.On("Create", ctx, "1", user, mock.Anything).Return(&subscription.Subscription{Id: "some_id"}, nil)
	res, err := manager.CreateSubscription(ctx, "1", user, &subscription.Subscription{
		Id:       "some_id",
		Events:   []subscription.SubscriptionEvent{},
		Channels: []string{},
	})
	require.NoError(t, err)
	require.Equal(t, "some_id", res.Id)
}

func TestCreateSubscription_FailedToValidate(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewUserRoutesClienter(t)
	manager := subscription.NewManager(client)

	client.On("GetUserSubscriptions", ctx, "1", user).Return(nil, errors.Join(fmt.Errorf("failed to get routes"), subscription.ErrClient))
	route, err := manager.CreateSubscription(ctx, "1", user, &subscription.Subscription{
		Id:       "some_id",
		Events:   []subscription.SubscriptionEvent{},
		Channels: []string{},
	})
	require.Error(t, err)
	require.ErrorContains(t, err, subscription.ErrClient.Error())
	require.ErrorContains(t, err, "failed to get routes")
	require.Nil(t, route)
}

func TestCreateSubscription_Forbidden(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewUserRoutesClienter(t)
	manager := subscription.NewManager(client)

	client.On("GetUserSubscriptions", ctx, "1", user).Return([]subscription.Subscription{{Id: "existing"}}, nil)
	route, err := manager.CreateSubscription(ctx, "1", user, &subscription.Subscription{
		Id:       "some_id",
		Events:   []subscription.SubscriptionEvent{},
		Channels: []string{},
	})
	require.Error(t, err)
	require.ErrorContains(t, err, subscription.ErrCreatingForbidden.Error())
	require.Nil(t, route)
}

func TestCreateSubscription_FailedIfUserTryingToUpdateInternalRouteToBeExternal(t *testing.T) {
	routeId := uuid.NewString()
	ctx := context.Background()
	client := mocks.NewUserRoutesClienter(t)
	manager := subscription.NewManager(client)

	client.On("GetUserSubscriptions", ctx, "1", user).Return([]subscription.Subscription{}, nil)
	client.On("Create", ctx, "1", user, mock.Anything).Return(nil, subscription.ErrSubscriptionAlreadyExists)
	route, err := manager.CreateSubscription(ctx, "1", user, &subscription.Subscription{
		Id:       routeId,
		Events:   []subscription.SubscriptionEvent{},
		Channels: []string{},
	})
	require.Error(t, err)
	require.Nil(t, route)
	require.ErrorContains(t, err, subscription.ErrSubscriptionAlreadyExists.Error())
}

func TestDeleteSubscription_Failed(t *testing.T) {
	ctx := context.Background()
	id, _ := uuid.NewUUID()
	client := mocks.NewUserRoutesClienter(t)
	manager := subscription.NewManager(client)

	client.On("Delete", ctx, "1", user, id.String()).Return(errors.Join(fmt.Errorf("failed to delete"), subscription.ErrClient))
	err := manager.DeleteSubscription(ctx, "1", user, id.String())
	require.Error(t, err)
	require.ErrorContains(t, err, subscription.ErrClient.Error())
}

func TestDeleteSubscription_NotFound(t *testing.T) {
	ctx := context.Background()
	id, _ := uuid.NewUUID()
	client := mocks.NewUserRoutesClienter(t)
	manager := subscription.NewManager(client)

	client.On("Delete", ctx, "1", user, id.String()).Return(subscription.ErrNotFound)
	err := manager.DeleteSubscription(ctx, "1", user, id.String())
	require.Error(t, err)
	require.ErrorContains(t, err, subscription.ErrNotFound.Error())
}

func TestDeleteSubscription_Succeed(t *testing.T) {
	ctx := context.Background()
	id, _ := uuid.NewUUID()
	client := mocks.NewUserRoutesClienter(t)
	manager := subscription.NewManager(client)

	client.On("Delete", ctx, "1", user, id.String()).Return(nil)
	err := manager.DeleteSubscription(ctx, "1", user, id.String())
	require.NoError(t, err)
}

func TestUpdateSubscription_Failed(t *testing.T) {
	ctx := context.Background()
	id, _ := uuid.NewUUID()
	client := mocks.NewUserRoutesClienter(t)
	manager := subscription.NewManager(client)

	client.On("Update", ctx, "1", user, mock.Anything).Return(errors.Join(fmt.Errorf("failed to update route"), subscription.ErrClient))
	err := manager.UpdateSubscription(ctx, "1", user, id.String(), &subscription.Subscription{})
	require.Error(t, err)
	require.ErrorContains(t, err, subscription.ErrClient.Error())
}

func TestUpdateSubscription_Succeed(t *testing.T) {
	ctx := context.Background()
	id, _ := uuid.NewUUID()
	client := mocks.NewUserRoutesClienter(t)
	manager := subscription.NewManager(client)

	client.On("Update", ctx, "1", user, mock.Anything).Return(nil)
	err := manager.UpdateSubscription(ctx, "1", user, id.String(), &subscription.Subscription{})
	require.NoError(t, err)
}

func TestUpdateSubscription_NotFound(t *testing.T) {
	ctx := context.Background()
	id, _ := uuid.NewUUID()
	client := mocks.NewUserRoutesClienter(t)
	manager := subscription.NewManager(client)

	client.On("Update", ctx, "1", user, mock.Anything).Return(subscription.ErrNotFound)
	err := manager.UpdateSubscription(ctx, "1", user, id.String(), &subscription.Subscription{})
	require.Error(t, err)
	require.ErrorContains(t, err, subscription.ErrNotFound.Error())
}

func TestGetUserSubscriptions_Failed(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewUserRoutesClienter(t)
	manager := subscription.NewManager(client)

	client.On("GetUserSubscriptions", ctx, "1", user).Return(nil, errors.Join(fmt.Errorf("failed to get routes"), subscription.ErrClient))
	routes, err := manager.GetUserSubscriptions(ctx, "1", user)
	require.Error(t, err)
	require.ErrorContains(t, err, subscription.ErrClient.Error())
	require.Nil(t, routes)
}

func TestGetUserSubscriptions_Succeed(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewUserRoutesClienter(t)
	manager := subscription.NewManager(client)

	client.On("GetUserSubscriptions", ctx, "1", user).Return([]subscription.Subscription{}, nil)
	routes, err := manager.GetUserSubscriptions(ctx, "1", user)
	require.NoError(t, err)
	require.Empty(t, routes)
}
