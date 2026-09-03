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

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/clients/state"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/clients/subscription"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/clients/tenant_migration"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/notification_state"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/authorization"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/clients/channel"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/email"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type EmailManager interface {
	GetEmail(ctx context.Context) (*email.ChannelConfig, error)
	UpdateEmail(ctx context.Context, emailChannel *email.ChannelConfig) (*email.ChannelConfig, error)
	CreateEmail(ctx context.Context, emailChannel *email.ChannelConfig) (*email.ChannelConfig, error)
	DeleteEmail(ctx context.Context) error
	PatchEmail(ctx context.Context, emailChannel *email.ChannelConfig) (*email.ChannelConfig, error)
	ValidateEmail(ctx context.Context, emailChannel *email.ChannelConfig) (email.ValidationResponse, error)
}

type EmailNotificationStateManager interface {
	GetEmailNotificationState(ctx context.Context, tenantId string) (notification_state.Stater, error)
	UpdateEmailNotificationState(ctx context.Context, tenantId string, notificationState notification_state.Stater) (notification_state.Stater, error)
}

type EmailAPIService struct {
	manager      EmailManager
	stateManager EmailNotificationStateManager
	authorizer   Authorizer
}

func NewEmailAPIService() api.EmailAPIServicer {
	return &EmailAPIService{
		manager:      email.NewEmailManager(channel.NewClient()),
		authorizer:   authorization.NewAuthorizer(),
		stateManager: notification_state.NewNotificationStateManager(state.NewClient(), channel.NewClient(), subscription.NewClient(), tenant_migration.NewClient()),
	}
}

func (eas *EmailAPIService) GetEmail(ctx context.Context) (api.ImplResponse, error) {
	if !eas.authorizer.IsSystem(ctx) {
		return api.ImplResponse{Code: http.StatusForbidden}, errors.New("access denied")
	}

	emailChannel, err := eas.manager.GetEmail(ctx)
	if err == nil {
		return api.ImplResponse{Code: http.StatusOK, Body: emailChannelConfigToAPI(emailChannel)}, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get email: %v", err)
	if errors.Is(err, email.ErrNotFound) {
		return api.ImplResponse{Code: http.StatusNotFound}, errors.New("email not found")
	}
	return api.ImplResponse{Code: http.StatusInternalServerError}, errors.New("failed to get email")
}

func (eas *EmailAPIService) UpdateEmail(ctx context.Context, emailChannel api.EmailChannel) (api.ImplResponse, error) {
	if !eas.authorizer.IsSystem(ctx) {
		return api.ImplResponse{Code: http.StatusForbidden}, errors.New("access denied")
	}

	eas.addDataToAudit(ctx)
	updatedEmail, err := eas.manager.UpdateEmail(ctx, emailChannelFromAPI(emailChannel))
	if err == nil {
		return api.ImplResponse{Code: http.StatusOK, Body: emailChannelConfigToAPI(updatedEmail)}, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to update email: %v", err)
	if errors.Is(err, email.ErrNotFound) {
		return api.ImplResponse{Code: http.StatusNotFound}, errors.New("email not found")
	}
	return api.ImplResponse{Code: http.StatusInternalServerError}, errors.New("failed to update email")
}

func (eas *EmailAPIService) CreateEmail(ctx context.Context, emailChannel api.EmailChannel) (api.ImplResponse, error) {
	if !eas.authorizer.IsSystem(ctx) {
		return api.ImplResponse{Code: http.StatusForbidden}, errors.New("access denied")
	}

	eas.addDataToAudit(ctx)
	createdEmail, err := eas.manager.CreateEmail(ctx, emailChannelFromAPI(emailChannel))
	if err == nil {
		return api.ImplResponse{Code: http.StatusOK, Body: emailChannelConfigToAPI(createdEmail)}, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to create email: %v", err)
	if errors.Is(err, email.ErrAlreadyExists) {
		return api.ImplResponse{Code: http.StatusConflict}, errors.New("email already exists")
	}
	return api.ImplResponse{Code: http.StatusInternalServerError}, errors.New("failed to create email")
}

func (eas *EmailAPIService) DeleteEmail(ctx context.Context) (api.ImplResponse, error) {
	if !eas.authorizer.IsSystem(ctx) {
		return api.ImplResponse{Code: http.StatusForbidden}, errors.New("access denied")
	}

	eas.addDataToAudit(ctx)
	err := eas.manager.DeleteEmail(ctx)
	if err == nil {
		return api.ImplResponse{Code: http.StatusNoContent}, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to delete email. error %v", err)
	if errors.Is(err, email.ErrNotFound) {
		return api.ImplResponse{Code: http.StatusNotFound}, errors.New("email not found")
	}
	return api.ImplResponse{Code: http.StatusInternalServerError}, errors.New("failed to delete email")
}

func (eas *EmailAPIService) PatchEmail(ctx context.Context, emailChannel api.EmailChannelForPatch) (api.ImplResponse, error) {
	if !eas.authorizer.IsSystem(ctx) {
		return api.ImplResponse{Code: http.StatusForbidden}, errors.New("access denied")
	}

	eas.addDataToAudit(ctx)
	patchedEmail, err := eas.manager.PatchEmail(ctx, emailChannelForPatchFromAPI(emailChannel))
	if err == nil {
		return api.ImplResponse{Code: http.StatusOK, Body: emailChannelConfigToAPI(patchedEmail)}, nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to patch email. error %v", err)
	if errors.Is(err, email.ErrNotFound) {
		return api.ImplResponse{Code: http.StatusNotFound}, errors.New("email not found")
	}
	return api.ImplResponse{Code: http.StatusInternalServerError}, errors.New("failed to patch email")
}

func (eas *EmailAPIService) ValidateEmail(ctx context.Context, request api.EmailChannelConfig) (api.ImplResponse, error) {
	if !eas.authorizer.IsSystem(ctx) {
		return api.ImplResponse{Code: http.StatusForbidden}, errors.New("access denied")
	}

	resp, err := eas.manager.ValidateEmail(ctx, emailChannelConfigFromAPI(request))
	if err == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("successfully validated email")
		return api.ImplResponse{Code: http.StatusOK, Body: api.ValidateSinkResponse{Status: resp.Status}}, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("email validation failed %v", err)
	if errors.Is(err, email.ErrValidationFailed) {
		return api.ImplResponse{Code: http.StatusOK, Body: api.ValidateSinkResponse{Status: resp.Status, Message: resp.Message}}, nil
	}
	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("email validation failed")
}

func (eas *EmailAPIService) GetEmailNotificationState(ctx context.Context) (api.ImplResponse, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id in request context")
	}

	s, err := eas.stateManager.GetEmailNotificationState(ctx, tenantId)
	if err == nil {
		return api.ImplResponse{Code: http.StatusOK, Body: getAPIState(s)}, nil
	}

	if s != nil {
		return api.ImplResponse{Code: http.StatusOK, Body: getAPIState(s)}, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get email notifications state: %v", err)
	if errors.Is(err, notification_state.ErrNoSinkConfigured) {
		return api.ImplResponse{Code: http.StatusNotFound}, fmt.Errorf("email app not configured to update state")
	}
	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to get email state")
}

func (eas *EmailAPIService) UpdateEmailNotificationState(ctx context.Context, status api.NotificationStateStatus) (api.ImplResponse, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id in request context")
	}

	res, err := eas.stateManager.UpdateEmailNotificationState(ctx, tenantId, &state.EmailState{Enabled: status.Enabled})
	if err == nil {
		return api.ImplResponse{Code: http.StatusOK, Body: getAPIState(res)}, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update email notifications state: %v", err)
	if errors.Is(err, notification_state.ErrFailedToRemoveSinkFromUserRoutes) {
		return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to remove email sink from user routes")
	}
	if errors.Is(err, notification_state.ErrNoSinkConfigured) {
		return api.ImplResponse{Code: http.StatusNotFound}, fmt.Errorf("email not configured to update state")
	}
	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to update email notifications state")
}

func (*EmailAPIService) addDataToAudit(ctx context.Context) *audit.AuditLog {
	return audit.AuditFromCtx(ctx).WithEntityType("Notification channel").WithEntityName("Email").WithEntityId("runai-email")
}

func emailChannelForPatchFromAPI(patch api.EmailChannelForPatch) *email.ChannelConfig {
	return emailChannelConfigFromAPI(api.EmailChannelConfig(patch.Config))
}

func emailChannelFromAPI(emailChannel api.EmailChannel) *email.ChannelConfig {
	return emailChannelConfigFromAPI(emailChannel.Config)
}

func emailChannelConfigFromAPI(emailChannelConfig api.EmailChannelConfig) *email.ChannelConfig {
	return &email.ChannelConfig{
		Username:            emailChannelConfig.Username,
		Password:            emailChannelConfig.Password,
		SmtpHost:            emailChannelConfig.SmtpHost,
		SmtpPort:            emailChannelConfig.SmtpPort,
		From:                emailChannelConfig.From,
		AuthType:            string(emailChannelConfig.AuthType),
		SmtpTlsEnabled:      emailChannelConfig.SmtpTlsEnabled,
		Recipients:          emailChannelConfig.Recipients,
		FromDisplayName:     emailChannelConfig.FromDisplayName,
		DirectNotifications: emailChannelConfig.DirectNotifications,
		UseAttachments:      emailChannelConfig.UseAttachments,
		Subject:             emailChannelConfig.Subject,
	}
}

func emailChannelConfigToAPI(emailChannelConfig *email.ChannelConfig) api.EmailChannel {
	if emailChannelConfig == nil {
		return api.EmailChannel{}
	}
	return api.EmailChannel{
		Config: api.EmailChannelConfig{
			Username:            emailChannelConfig.Username,
			Password:            emailChannelConfig.Password,
			SmtpHost:            emailChannelConfig.SmtpHost,
			SmtpPort:            emailChannelConfig.SmtpPort,
			From:                emailChannelConfig.From,
			AuthType:            api.EmailAuthType(emailChannelConfig.AuthType),
			SmtpTlsEnabled:      emailChannelConfig.SmtpTlsEnabled,
			Recipients:          emailChannelConfig.Recipients,
			FromDisplayName:     emailChannelConfig.FromDisplayName,
			DirectNotifications: emailChannelConfig.DirectNotifications,
			UseAttachments:      emailChannelConfig.UseAttachments,
			Subject:             emailChannelConfig.Subject,
		},
	}
}
