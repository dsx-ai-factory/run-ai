//go:build integration
// +build integration

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"github.com/mitchellh/mapstructure"
	"io"
	"net/http"
	"testing"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/pkg/client/types"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

func TestCreateChannel_Email_Succeed(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithSystemScope)

	conf := openapi.NewNotificationChannelConfigWithDefaults()
	conf.SetUseAttachments(true)

	response, err := client.NotificationChannelsAPI.CreateNotificationChannel(ctx).NotificationChannel(openapi.NotificationChannel{
		Type:   "email",
		Config: *conf,
	}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	defer deleteChannel(t, ctx, "runai-email")

	bodyBytes, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Empty(t, string(bodyBytes))
}

func TestCreateChannel_Slack_Succeed(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithSystemScope)

	conf := openapi.NewNotificationChannelConfigWithDefaults()
	conf.SetApiToken("xoxb-1234567890")
	response, err := client.NotificationChannelsAPI.CreateNotificationChannel(ctx).NotificationChannel(openapi.NotificationChannel{
		Type:   "slack",
		Config: *conf,
	}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	defer deleteChannel(t, ctx, "slack-1")
}

func TestCreateChannel_Slack_Conflict(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithSystemScope)

	conf := openapi.NewNotificationChannelConfigWithDefaults()
	conf.SetApiToken("xoxb-1234567890")
	response, err := client.NotificationChannelsAPI.CreateNotificationChannel(ctx).NotificationChannel(openapi.NotificationChannel{
		Type:   "slack",
		Config: *conf,
	}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	defer deleteChannel(t, ctx, "slack-1")

	response, err = client.NotificationChannelsAPI.CreateNotificationChannel(ctx).NotificationChannel(openapi.NotificationChannel{
		Type:   "slack",
		Config: *conf,
	}).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusConflict, response.StatusCode)
}

func TestCreateChannel_FailedBadType(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "1")

	conf := openapi.NewNotificationChannelConfigWithDefaults()

	response, err := client.NotificationChannelsAPI.CreateNotificationChannel(ctx).NotificationChannel(openapi.NotificationChannel{
		Type:   "smtp",
		Config: *conf,
	}).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestGetChannel_Succeed(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithSystemScope)

	conf := openapi.NewNotificationChannelConfigWithDefaults()

	_, err := client.NotificationChannelsAPI.CreateNotificationChannel(ctx).NotificationChannel(openapi.NotificationChannel{
		Type:   "email",
		Config: *conf,
	}).Execute()
	require.NoError(t, err)
	defer deleteChannel(t, ctx, "runai-email")

	channel, _, err := client.NotificationChannelsAPI.GetNotificationChannel(ctx, "runai-email").Execute()
	require.NoError(t, err)
	require.NotNil(t, channel)
	require.Equal(t, channel.Type, openapi.NotificationChannelType("email"))

	resp := getSinkConfigFromService(t, ctx, "runai-email")
	require.Equal(t, resp.Id, "runai-email")
	require.Equal(t, resp.Type, "email")
	require.Equal(t, resp.Config.AuthType, "auth_login")
	require.Nil(t, resp.Config.DirectNotifications)
	require.Equal(t, resp.Config.From, "")
	require.Equal(t, resp.Config.FromDisplayName, "Run:ai")
	require.Equal(t, resp.Config.Password, "")
	require.Nil(t, resp.Config.SmtpTlsEnabled)
	require.Equal(t, resp.Config.Subject, "Run:ai notification - {{ .category }} {{ .reason }}")
}

func TestGetChannel_NotExist(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithSystemScope)

	conf, res, err := client.NotificationChannelsAPI.GetNotificationChannel(ctx, "runai-email").Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, res.StatusCode)
	require.Nil(t, conf)
}

func TestGetAllChannel_GetDefaultChannelFromConfig(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithoutSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithoutSystemScope)

	conf, _, err := client.NotificationChannelsAPI.GetNotificationChannels(ctx).Execute()
	require.NoError(t, err)
	require.NotEmpty(t, conf)
	require.Equal(t, "slack-3", conf[0].Name)
	require.Equal(t, openapi.NotificationChannelType("slack"), openapi.NotificationChannelType(conf[0].Type))
}

func TestGetAllChannels_NoPermissions(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithSystemScope)

	conf := openapi.NewNotificationChannelConfigWithDefaults()

	response, err := client.NotificationChannelsAPI.CreateNotificationChannel(ctx).NotificationChannel(openapi.NotificationChannel{
		Type:   "email",
		Config: *conf,
	}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	defer deleteChannel(t, ctx, "runai-email")

	ctxWithoutPermissions := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithoutPermissions)
	ctxWithoutPermissions = context.WithValue(ctxWithoutPermissions, runaictx.CfSubjectID, tenantIdWithoutPermissions)
	confs, _, err := client.NotificationChannelsAPI.GetNotificationChannels(ctxWithoutPermissions).Execute()
	require.NoError(t, err)
	require.NotEmpty(t, confs)
	require.Len(t, confs, 1)
	require.Equal(t, "runai-email", confs[0].Name)
	require.Equal(t, openapi.NotificationChannelType("email"), openapi.NotificationChannelType(confs[0].Type))
}

func TestCreateChannel_NoSystemScope(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithoutSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithoutSystemScope)

	conf := openapi.NewNotificationChannelConfigWithDefaults()
	response, err := client.NotificationChannelsAPI.CreateNotificationChannel(ctx).NotificationChannel(openapi.NotificationChannel{
		Type:   "email",
		Config: *conf,
	}).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, response.StatusCode)
}

func TestDeleteChannel_NoSystemScope(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithoutSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithoutSystemScope)

	response, err := client.NotificationChannelsAPI.DeleteNotificationChannel(ctx, "runai-email").Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, response.StatusCode)
}

func TestGetChannel_NoSystemScope(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithoutSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithoutSystemScope)

	_, response, err := client.NotificationChannelsAPI.GetNotificationChannel(ctx, "runai-email").Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, response.StatusCode)
}

func TestUpdateChannel_NoSystemScope(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithoutSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithoutSystemScope)

	conf := openapi.NewNotificationChannelConfigWithDefaults()
	response, err := client.NotificationChannelsAPI.UpdateNotificationChannel(ctx, "runai-email").NotificationChannel(openapi.NotificationChannel{
		Type:   "email",
		Config: *conf,
	}).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, response.StatusCode)
}

func TestUpdateChannel_NotExist(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithSystemScope)

	conf := openapi.NewNotificationChannelConfigWithDefaults()
	conf.SetApiToken("xoxb-1234567890")

	response, err := client.NotificationChannelsAPI.UpdateNotificationChannel(ctx, "slack-1").NotificationChannel(openapi.NotificationChannel{
		Type:   "slack",
		Config: *conf,
	}).Execute()
	require.EqualError(t, err, "404 Not Found")
	require.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestPatchChannel_NoSystemScope(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithoutSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithoutSystemScope)

	conf := openapi.NewNotificationChannelConfigWithDefaults()
	response, err := client.NotificationChannelsAPI.PatchNotificationChannel(ctx, "runai-email").NotificationChannelForPatch(openapi.NotificationChannelForPatch{
		Type:   ptr.To(openapi.NotificationChannelType("email")),
		Config: conf,
	}).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, response.StatusCode)
}

func TestPatchChannel_Succeed(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithSystemScope)
	conf := openapi.NewNotificationChannelConfigWithDefaults()
	response, err := client.NotificationChannelsAPI.CreateNotificationChannel(ctx).NotificationChannel(openapi.NotificationChannel{
		Type:   "email",
		Config: *conf,
	}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	defer deleteChannel(t, ctx, "runai-email")

	channel, _, err := client.NotificationChannelsAPI.GetNotificationChannel(ctx, "runai-email").Execute()
	require.NoError(t, err)
	require.NotNil(t, channel)
	conf.SetFromDisplayName("test")
	response, err = client.NotificationChannelsAPI.PatchNotificationChannel(ctx, "runai-email").NotificationChannelForPatch(openapi.NotificationChannelForPatch{
		Type:   ptr.To(openapi.NotificationChannelType("email")),
		Config: conf,
	}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)

	channel, _, err = client.NotificationChannelsAPI.GetNotificationChannel(ctx, "runai-email").Execute()
	require.NoError(t, err)
	require.NotNil(t, channel)
	require.Equal(t, *channel.Config.FromDisplayName, "test")
}

func getSinkConfigFromService(t *testing.T, ctx context.Context, channelId string) *types.SinkConfig {
	sinkData, _, err := notificationsClient.SinksAPI.GetSink(ctx, channelId).Execute()
	require.NoError(t, err)

	config := types.SinkConfigDetails{}
	require.NoError(t, mapstructure.Decode(sinkData.Config, &config))
	return &types.SinkConfig{
		Id:     sinkData.Id,
		Type:   sinkData.Type,
		Config: config,
	}
}

func deleteChannel(t *testing.T, ctx context.Context, channelId string) {
	response, err := client.NotificationChannelsAPI.DeleteNotificationChannel(ctx, channelId).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, response.StatusCode)
}
