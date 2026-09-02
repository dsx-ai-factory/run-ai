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
	"strings"
	"unicode"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/authorization"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/channel"
	channelClient "github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/clients/channel"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/email"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/slack"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type NotificationChannelsAPIService struct {
	authorizer      Authorizer
	channelsManager ChannelsManager
}

type ChannelsManager interface {
	GetNotificationsChannels(ctx context.Context, tenantId string) ([]channel.NotificationChannelMetadata, error)
	CreateChannel(ctx context.Context, tenantId string, config channel.NotificationChannel) error
	DeleteChannel(ctx context.Context, tenantId, sinkId string) error
	GetChannel(ctx context.Context, tenantId, channelId string) (*channel.NotificationChannel, error)
	ValidateChannel(ctx context.Context, tenantId string, config channel.NotificationChannel) (channel.ValidationResponse, error)
	UpdateChannel(ctx context.Context, tenantId, channelId string, config channel.NotificationChannel) error
	PatchChannel(ctx context.Context, tenantId, channelId string, config channel.NotificationChannel) error
}

type Authorizer interface {
	IsSystem(ctx context.Context) bool
	IsTenant(ctx context.Context) bool
}

func NewNotificationChannelsAPIService() api.NotificationChannelsAPIServicer {
	return &NotificationChannelsAPIService{
		authorizer:      authorization.NewAuthorizer(),
		channelsManager: channel.NewManager(channelClient.NewClient()),
	}
}

func (ncs *NotificationChannelsAPIService) CreateNotificationChannel(ctx context.Context, nc api.NotificationChannel) (api.ImplResponse, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id in request context")
	}

	if ok, err := ncs.isAccessPermitted(ctx, nc.Type); err != nil {
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("invalid channel type")
	} else if !ok {
		return api.ImplResponse{Code: http.StatusForbidden}, fmt.Errorf("access denied")
	}

	ncs.addDataToAudit(ctx, string(nc.Type), channelName(nc.Type, tenantId))
	if err := validateChannelConfig(nc.Config, nc.Type); err != nil {
		return api.ImplResponse{Code: http.StatusBadRequest}, err
	}

	if err = ncs.channelsManager.CreateChannel(ctx, tenantId, fromApiToChannel(nc)); err == nil {
		return api.ImplResponse{Code: http.StatusOK}, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to create channel. error %v", err)
	if errors.Is(err, channel.ErrNotSupportedType) {
		return api.ImplResponse{Code: http.StatusBadRequest}, err
	}
	if errors.Is(err, channel.ErrAlreadyExists) {
		return api.ImplResponse{Code: http.StatusConflict}, fmt.Errorf("channel already exist")
	}
	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to create")
}

func (ncs *NotificationChannelsAPIService) DeleteNotificationChannel(ctx context.Context, channelId string) (api.ImplResponse, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to delete channel %s. error %v", channelId, err)
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id in request context")
	}

	ncs.addDataToAudit(ctx, channelType(tenantId, channelId), channelId)
	if ok, err := ncs.isAccessPermittedByChannelId(ctx, tenantId, channelId); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to delete channel %s. error %v", channelId, err)
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("invalid channel id")
	} else if !ok {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("access denied to delete channel %s. error %v", channelId, err)
		return api.ImplResponse{Code: http.StatusForbidden}, fmt.Errorf("access denied")
	}

	err = ncs.channelsManager.DeleteChannel(ctx, tenantId, channelId)
	if err == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("successfully deleted channel %s", channelId)
		return api.ImplResponse{Code: http.StatusNoContent}, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to delete channel %s. error %v", channelId, err)
	if errors.Is(err, channel.ErrNotSupportedType) {
		return api.ImplResponse{Code: http.StatusBadRequest}, err
	}
	if errors.Is(err, channel.ErrClientFailure) {
		return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to delete channel")
	}
	return api.ImplResponse{Code: http.StatusInternalServerError}, nil
}

func (ncs *NotificationChannelsAPIService) GetNotificationChannel(ctx context.Context, channelId string) (api.ImplResponse, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get channel %s. error %v", channelId, err)
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id in request context")
	}

	if ok, err := ncs.isAccessPermittedByChannelId(ctx, tenantId, channelId); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get channel %s. error %v", channelId, err)
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("invalid channel id")
	} else if !ok {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("access denied to get channel %s. error %v", channelId, err)
		return api.ImplResponse{Code: http.StatusForbidden}, fmt.Errorf("access denied")
	}

	cfg, err := ncs.channelsManager.GetChannel(ctx, tenantId, channelId)
	if err == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("successfully got channel %s", channelId)
		return api.ImplResponse{Code: http.StatusOK, Body: &api.NotificationChannelResponse{
			Type:   api.NotificationChannelType(cfg.Type),
			Config: channelConfigToAPI(cfg)}}, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get channel. error %v", err)
	if errors.Is(err, channel.ErrNotSupportedType) {
		return api.ImplResponse{Code: http.StatusBadRequest}, err
	}
	if errors.Is(err, channel.ErrNotFound) {
		return api.ImplResponse{Code: http.StatusNotFound}, fmt.Errorf("channel not found")
	}
	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to get channel")
}

func (ncs *NotificationChannelsAPIService) GetNotificationChannels(ctx context.Context) (api.ImplResponse, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to fetch all available channel. error %v", err)
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id in request context")
	}

	configs, err := ncs.channelsManager.GetNotificationsChannels(ctx, tenantId)
	if err == nil {
		return api.ImplResponse{
			Code: http.StatusOK, Body: channelsMetadatToApi(configs),
		}, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to fetch all available channel. error %v", err)
	if errors.Is(err, channel.ErrNotSupportedType) {
		return api.ImplResponse{Code: http.StatusBadRequest}, err
	}
	if errors.Is(err, channel.ErrClientFailure) {
		return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to fetch all available channel")
	}

	return api.ImplResponse{Code: http.StatusInternalServerError}, nil
}

func (ncs *NotificationChannelsAPIService) UpdateNotificationChannel(ctx context.Context, channelId string, nc api.NotificationChannel) (api.ImplResponse, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to update channel %s, request messing tenant id . error %v", channelId, err)
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id in request context")
	}

	ncs.addDataToAudit(ctx, string(nc.Type), channelId)
	if ok, err := ncs.isAccessPermittedByChannelId(ctx, tenantId, channelId); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to update channel %s. error %v", channelId, err)
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("invalid channel id")
	} else if !ok {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("access denied to update channel %s. error %v", channelId, err)
		return api.ImplResponse{Code: http.StatusForbidden}, fmt.Errorf("access denied")
	}

	if err := validateChannelConfig(nc.Config, nc.Type); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to validate channel %s. error %v", channelId, err)
		return api.ImplResponse{Code: http.StatusBadRequest}, err
	}

	err = ncs.channelsManager.UpdateChannel(ctx, tenantId, channelId, fromApiToChannel(nc))
	if err == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("successfully updated channel %s", channelId)
		return api.ImplResponse{Code: http.StatusOK}, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to update channel. error %v", err)
	if errors.Is(err, channel.ErrNotFound) {
		return api.ImplResponse{Code: http.StatusNotFound}, fmt.Errorf("channel not found")
	}
	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to update channel")
}

func (ncs *NotificationChannelsAPIService) PatchNotificationChannel(ctx context.Context, channelId string, nc api.NotificationChannelForPatch) (api.ImplResponse, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to patch channel %s, request messing tenant id . error %v", channelId, err)
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id in request context")
	}

	ncs.addDataToAudit(ctx, string(nc.Type), channelId)
	if ok, err := ncs.isAccessPermittedByChannelId(ctx, tenantId, channelId); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to patch channel %s . error %v", channelId, err)
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("invalid channel id")
	} else if !ok {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("access denied to patch channel %s. error %v", channelId, err)
		return api.ImplResponse{Code: http.StatusForbidden}, fmt.Errorf("access denied")
	}

	err = ncs.channelsManager.PatchChannel(ctx, tenantId, channelId, fromApiToChannel(api.NotificationChannel(nc)))
	if err == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("successfully patched channel %s", channelId)
		return api.ImplResponse{Code: http.StatusOK}, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to patch channel. error %v", err)
	if errors.Is(err, channel.ErrNotFound) {
		return api.ImplResponse{Code: http.StatusNotFound}, fmt.Errorf("channel not found")
	}

	return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to patch channel")
}

func (ncs *NotificationChannelsAPIService) ValidateNotificationChannel(ctx context.Context, nc api.NotificationChannelForValidate) (api.ImplResponse, error) {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to validate channel, request messing tenant id . error %v", err)
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id in request context")
	}

	resp, err := ncs.channelsManager.ValidateChannel(ctx, tenantId, fromApiToChannel(api.NotificationChannel(nc)))
	if err == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("successfully validated channel")
		return api.ImplResponse{Code: http.StatusOK, Body: resp.Status}, nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("channel validation failed %v", err)
	if errors.Is(err, channel.ErrNotSupportedType) {
		return api.ImplResponse{Code: http.StatusBadRequest}, err
	}
	if errors.Is(err, channel.ErrValidationFailed) {
		return api.ImplResponse{Code: http.StatusOK, Body: resp.Status}, nil
	}
	return api.ImplResponse{Code: http.StatusInternalServerError}, err
}

func (ncs *NotificationChannelsAPIService) isAccessPermitted(ctx context.Context, channelType api.NotificationChannelType) (bool, error) {
	if channelType == api.NOTIFICATIONCHANNELTYPE_EMAIL {
		return ncs.authorizer.IsSystem(ctx), nil
	}
	if channelType == api.NOTIFICATIONCHANNELTYPE_SLACK {
		return ncs.authorizer.IsTenant(ctx), nil
	}

	return false, fmt.Errorf("unsupported channel type %s", channelType)
}

func (ncs *NotificationChannelsAPIService) isAccessPermittedByChannelId(ctx context.Context, tenantId, channelId string) (bool, error) {
	if channelId == channelClient.EmailChannelName() {
		return ncs.authorizer.IsSystem(ctx), nil
	}
	if channelId == channelClient.SlackChannelName(tenantId) {
		return ncs.authorizer.IsTenant(ctx), nil
	}
	return false, fmt.Errorf("unsupported channel id %s", channelId)
}

func (*NotificationChannelsAPIService) addDataToAudit(ctx context.Context, typee string, channelId string) *audit.AuditLog {
	return audit.AuditFromCtx(ctx).WithEntityType("Notification channel").WithEntityName(typee).WithEntityId(channelId)
}

func channelsMetadatToApi(sinks []channel.NotificationChannelMetadata) []*api.NotificationChannelMeta {
	res := []*api.NotificationChannelMeta{}
	for _, s := range sinks {
		res = append(res, &api.NotificationChannelMeta{
			Name: s.Id,
			Type: s.Type,
		})
	}
	return res
}

func validateChannelConfig(channel api.NotificationChannelConfig, channelType api.NotificationChannelType) error {
	if channelType == api.NOTIFICATIONCHANNELTYPE_SLACK && !isApiTokenValid(channel.ApiToken) {
		return fmt.Errorf("invalid API token; must start with 'xoxp-' or 'xoxb-'")
	}

	return nil
}

func isApiTokenValid(apiToken string) bool {
	return strings.HasPrefix(apiToken, "xoxp-") || strings.HasPrefix(apiToken, "xoxb-")
}

func fromApiToChannel(nc api.NotificationChannel) channel.NotificationChannel {
	if nc.Type == api.NOTIFICATIONCHANNELTYPE_SLACK {
		return channel.NotificationChannel{
			Type: string(nc.Type),
			Slack: slack.ChannelConfig{
				Type:                string(nc.Type),
				DirectNotifications: nc.Config.DirectNotifications,
				Recipients:          nc.Config.Recipients,
				ApiToken:            nc.Config.ApiToken,
			},
		}
	}

	return channel.NotificationChannel{
		Type: string(nc.Type),
		Email: email.ChannelConfig{
			Username:            nc.Config.Username,
			Password:            nc.Config.Password,
			SmtpHost:            nc.Config.SmtpHost,
			SmtpPort:            nc.Config.SmtpPort,
			From:                nc.Config.From,
			AuthType:            string(nc.Config.AuthType),
			SmtpTlsEnabled:      nc.Config.SmtpTlsEnabled,
			Recipients:          nc.Config.Recipients,
			FromDisplayName:     nc.Config.FromDisplayName,
			DirectNotifications: nc.Config.DirectNotifications,
			UseAttachments:      nc.Config.UseAttachments,
			Subject:             nc.Config.Subject,
		},
	}
}

func channelConfigToAPI(cfg *channel.NotificationChannel) api.NotificationChannelConfig {
	if cfg == nil {
		return api.NotificationChannelConfig{}
	}

	if cfg.Type == channelClient.NotificationChannelTypeSlack {
		return api.NotificationChannelConfig{
			ApiToken:            cfg.Slack.ApiToken,
			Recipients:          cfg.Slack.Recipients,
			DirectNotifications: cfg.Slack.DirectNotifications,
		}
	}

	return api.NotificationChannelConfig{
		Username:            cfg.Email.Username,
		Password:            cfg.Email.Password,
		SmtpHost:            cfg.Email.SmtpHost,
		SmtpPort:            cfg.Email.SmtpPort,
		From:                cfg.Email.From,
		AuthType:            api.EmailAuthType(cfg.Email.AuthType),
		SmtpTlsEnabled:      cfg.Email.SmtpTlsEnabled,
		FromDisplayName:     cfg.Email.FromDisplayName,
		UseAttachments:      cfg.Email.UseAttachments,
		Subject:             cfg.Email.Subject,
		Recipients:          cfg.Email.Recipients,
		DirectNotifications: cfg.Email.DirectNotifications,
	}
}

func channelName(t api.NotificationChannelType, tenantId string) string {
	if t == api.NOTIFICATIONCHANNELTYPE_EMAIL {
		return channelClient.EmailChannelName()
	}
	if t == api.NOTIFICATIONCHANNELTYPE_SLACK {
		return channelClient.SlackChannelName(tenantId)
	}
	return "undefined"
}

func channelType(tenantId, channelId string) string {
	if channelId == channelClient.EmailChannelName() {
		return capitalize(channelClient.NotificationChannelTypeEmail)
	}
	if channelId == channelClient.SlackChannelName(tenantId) {
		return capitalize(channelClient.NotificationChannelTypeSlack)
	}
	return "undefined"
}

func capitalize(s string) string {
	if s == "" {
		return s
	}

	r := []rune(s)
	return string(unicode.ToUpper(r[0])) + string(r[1:])
}
