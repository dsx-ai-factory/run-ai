/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"

	subscriptionsClient "github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/clients/subscription"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/subscription"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/pkg/errors"
)

type SubscriptionsManager interface {
	CreateSubscription(ctx context.Context, tenantId, userId string, sub *subscription.Subscription) (*subscription.Subscription, error)
	DeleteSubscription(ctx context.Context, tenantId, userId, subscriptionId string) error
	GetUserSubscriptions(ctx context.Context, tenantId, userId string) ([]subscription.Subscription, error)
	UpdateSubscription(ctx context.Context, tenantId, userId, subscriptionId string, subscriptionForPut *subscription.Subscription) error
}

type SubscriptionsAPIService struct {
	subscriptionsManager SubscriptionsManager
}

func NewSubscriptionsAPIService() *SubscriptionsAPIService {
	return &SubscriptionsAPIService{
		subscriptionsManager: subscription.NewManager(subscriptionsClient.NewClient()),
	}
}

func (s *SubscriptionsAPIService) CreateSubscription(ctx context.Context, userSubscription api.Subscription) (api.ImplResponse, error) {
	tenantId, userId, err := getTenantAndUserIds(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("parameter not in context err: %v", err)
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id or user id in request context")
	}

	created, err := s.subscriptionsManager.CreateSubscription(ctx, tenantId, userId, subscriptionFromAPI(&userSubscription))
	if err == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("subscription %s created, tenant %s", created.Id, tenantId)
		return api.ImplResponse{
			Code: http.StatusOK, Body: subscriptionToAPI(created),
		}, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to create subscription %s err: %v", userSubscription.Id, err)
	if errors.Is(err, subscription.ErrClient) {
		return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to create subscription")
	}
	if errors.Is(err, subscription.ErrSubscriptionAlreadyExists) {
		return api.ImplResponse{Code: http.StatusConflict}, fmt.Errorf("subscription with same name already exist")
	}
	if errors.Is(err, subscription.ErrCreatingForbidden) {
		return api.ImplResponse{Code: http.StatusConflict}, fmt.Errorf("user already has subscription")
	}
	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to create subscription")
}

func (s *SubscriptionsAPIService) DeleteSubscription(ctx context.Context, subscriptionId string) (api.ImplResponse, error) {
	tenantId, userId, err := getTenantAndUserIds(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("parameter not in context err: %v", err)
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id or user id in request context")
	}

	if err = s.subscriptionsManager.DeleteSubscription(ctx, tenantId, userId, subscriptionId); err == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("subscription %s deleted, tenant %s", subscriptionId, tenantId)
		return api.ImplResponse{Code: http.StatusNoContent}, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to delete route %s err: %v", subscriptionId, err)
	if errors.Is(err, subscription.ErrNotFound) {
		return api.ImplResponse{Code: http.StatusNotFound}, fmt.Errorf("subscription not found")
	}
	if errors.Is(err, subscription.ErrClient) {
		return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to delete subscription")
	}
	return api.ImplResponse{Code: http.StatusInternalServerError}, nil
}

func (s *SubscriptionsAPIService) GetUserSubscriptions(ctx context.Context) (api.ImplResponse, error) {
	tenantId, userId, err := getTenantAndUserIds(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("parameter not in context err: %v", err)
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id or user id in request context")
	}

	userSubscriptions, err := s.subscriptionsManager.GetUserSubscriptions(ctx, tenantId, userId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get subscriptions, err: %v", err)
		return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to get subscriptions")
	}

	return api.ImplResponse{Code: http.StatusOK, Body: subscriptionToAPIList(userSubscriptions)}, nil
}

func (s *SubscriptionsAPIService) UpdateSubscription(ctx context.Context, subscriptionId string, subscriptionForPut api.SubscriptionForPut) (api.ImplResponse, error) {
	tenantId, userId, err := getTenantAndUserIds(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("parameter not in context err: %v", err)
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id or user id in request context")
	}

	userSubscription := api.Subscription{Events: subscriptionForPut.Events, Channels: subscriptionForPut.Channels, Id: subscriptionId}
	if err = s.subscriptionsManager.UpdateSubscription(ctx, tenantId, userId, subscriptionId, subscriptionFromAPI(&userSubscription)); err == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("subscription %s updated, tenant %s", subscriptionId, tenantId)
		return api.ImplResponse{
			Code: http.StatusOK, Body: userSubscription,
		}, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to update user subscription, err: %v", err)
	if errors.Is(err, subscription.ErrClient) {
		return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to update subscription")
	}
	if errors.Is(err, subscription.ErrNotFound) {
		return api.ImplResponse{Code: http.StatusNotFound}, fmt.Errorf("subscription not found")
	}
	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to update subscription")
}

func getTenantAndUserIds(ctx context.Context) (string, string, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		return "", "", err
	}

	userId, err := runaictx.ParseSubjectId(ctx)
	return tenantId, userId, err
}

func subscriptionToAPIList(subscriptions []subscription.Subscription) []api.Subscription {
	parsedSubscriptions := []api.Subscription{}
	for _, s := range subscriptions {
		parsedSubscriptions = append(parsedSubscriptions, subscriptionToAPI(&s))
	}

	return parsedSubscriptions
}

func subscriptionToAPI(s *subscription.Subscription) api.Subscription {
	return api.Subscription{
		Id:       s.Id,
		Channels: s.Channels,
		Events:   subscriptionEventToAPI(&s.Events),
	}
}

func subscriptionFromAPI(apiSub *api.Subscription) *subscription.Subscription {
	if apiSub == nil {
		return nil
	}

	return &subscription.Subscription{Events: subscriptionEventFromAPI(apiSub.Events), Channels: apiSub.Channels, Id: apiSub.Id}
}

func subscriptionEventToAPI(subscriptionEvents *subscription.SubscriptionEvents) []api.SubscriptionEvent {
	events := []api.SubscriptionEvent{}
	if subscriptionEvents == nil {
		return events
	}

	for _, e := range *subscriptionEvents {
		events = append(events, api.SubscriptionEvent{Category: e.Category, Types: e.Types})
	}
	return events
}

func subscriptionEventFromAPI(events []api.SubscriptionEvent) subscription.SubscriptionEvents {
	if events == nil {
		return nil
	}

	e := subscription.SubscriptionEvents{}
	for _, event := range events {
		e = append(e, subscription.SubscriptionEvent{Category: event.Category, Types: event.Types})
	}
	return e
}
