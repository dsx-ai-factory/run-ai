/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/internal/generated/pkg/api"
	auth_manager "github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/authorization"
	system_notification "github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/system-notification"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
)

type SystemNotificationManager interface {
	Get(ctx context.Context) (*system_notification.Message, error)
	Set(ctx context.Context, message *system_notification.Message) error
	Invalidate(ctx context.Context) error
}

type SystemNotificationAPIService struct {
	systemNotificationManager SystemNotificationManager
	authManager               auth_manager.AuthManagerInterface
}

func NewSystemNotificationAPIService() *SystemNotificationAPIService {
	return &SystemNotificationAPIService{
		systemNotificationManager: system_notification.NewSystemNotificationManager(),
		authManager:               auth_manager.NewAuthManager(),
	}
}

func (s *SystemNotificationAPIService) GetSystemNotification(ctx context.Context) (api.ImplResponse, error) {
	msg, err := s.systemNotificationManager.Get(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get system notification, error %+v", err)
		return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to get system notification")
	}
	return api.Response(http.StatusOK, toApi(msg)), nil
}

func (s *SystemNotificationAPIService) UpdateSystemNotification(ctx context.Context, updateRequest api.UpdateSystemNotification) (api.ImplResponse, error) {
	if err := s.authManager.HandleSystemPermissions(ctx); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("forbidden to update system notification, error %+v", err)
		return api.ImplResponse{Code: http.StatusForbidden}, fmt.Errorf("forbidden to update system notification")
	}

	msg := system_notification.Message{
		Msg: *updateRequest.Msg,
		ID:  uuid.New().String(),
	}
	err := s.systemNotificationManager.Set(ctx, &msg)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to update system notification, error %+v", err)
		return api.Response(http.StatusInternalServerError, nil), fmt.Errorf("failed to update system notification")
	}
	return api.Response(http.StatusOK, updateRequest), nil
}

func (s *SystemNotificationAPIService) DeleteSystemNotification(ctx context.Context) (api.ImplResponse, error) {
	if err := s.authManager.HandleSystemPermissions(ctx); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("forbidden to delete system notification, error %+v", err)
		return api.ImplResponse{Code: http.StatusForbidden}, fmt.Errorf("forbidden to delete system notification")
	}

	err := s.systemNotificationManager.Invalidate(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to delete system notification, error %+v", err)
		return api.Response(http.StatusInternalServerError, nil), fmt.Errorf("failed to delete system notification")
	}
	return api.Response(http.StatusOK, nil), nil
}

func toApi(msg *system_notification.Message) api.GetSystemNotification {
	if msg == nil {
		return api.GetSystemNotification{}
	}

	return api.GetSystemNotification{
		Id:          msg.ID,
		Msg:         msg.Msg,
		CreatedAt:   msg.CreatedAt,
		DisplayName: "System notification",
	}
}
