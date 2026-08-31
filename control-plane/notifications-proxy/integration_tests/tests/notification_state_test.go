//go:build integration
// +build integration


/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	openapi "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/generated/config_client"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"
)

var defaultTenantRouteID = fmt.Sprintf("default-route-%s", tenantIdWithSystemScope)

func TestGetNotificationState_ValidTenantId_ReturnsStateAndStatusOK(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)
	notificationState, response, err := client.NotificationStateAPI.GetNotificationState(ctx).Execute()

	require.NoError(t, err)
	require.NotNil(t, notificationState, 0)
	require.Equal(t, http.StatusOK, response.StatusCode)
}

func TestGetNotificationState_MissingTenantId(t *testing.T) {
	ctx := context.Background()
	data, response, err := client.NotificationStateAPI.GetNotificationState(ctx).Execute()

	require.Nil(t, data)
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestPatchNotificationState_EnablesStateSuccessfully_Email(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "test@run.ai")
	response, err := client.NotificationChannelsAPI.CreateNotificationChannel(ctx).NotificationChannel(openapi.NotificationChannel{
		Type:   "email",
		Config: *openapi.NewNotificationChannelConfigWithDefaults(),
	}).Execute()
	defer deleteSink(ctx, "runai-email", t)

	notificationState, response, err := client.NotificationStateAPI.UpdateNotificationState(ctx).NotificationState(openapi.NotificationState{
		Email: &openapi.NotificationStateStatus{
			Enabled: *openapi.NewNullableBool(ptr.To(true)),
		},
	}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	defer deleteRoute(ctx, t)

	require.True(t, *notificationState.Email.Enabled.Get())
	require.False(t, *notificationState.Slack.Enabled.Get())
}

func TestPatchNotificationState_EnablesStateFailed_SinkDoesntExist(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "345") // no slack sinks for this tenant
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "test@run.ai")

	_, response, err := client.NotificationStateAPI.UpdateNotificationState(ctx).NotificationState(openapi.NotificationState{
		Slack: &openapi.NotificationStateStatus{
			Enabled: *openapi.NewNullableBool(ptr.To(true)),
		},
	}).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, response.StatusCode)

	_, response, err = client.NotificationStateAPI.GetNotificationState(ctx).Execute()
	require.NoError(t, err)
}

func TestPatchNotificationState_EnablesStateSuccessfully_Slack(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithSystemScope)
	conf := *openapi.NewNotificationChannelConfigWithDefaults()
	conf.ApiToken = ptr.To("xoxb-" + uuid.NewString())

	response, err := client.NotificationChannelsAPI.CreateNotificationChannel(ctx).NotificationChannel(openapi.NotificationChannel{
		Type:   "slack",
		Config: conf,
	}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	defer deleteSink(ctx, "slack-1", t)

	notificationState, response, err := client.NotificationStateAPI.UpdateNotificationState(ctx).NotificationState(openapi.NotificationState{
		Slack: &openapi.NotificationStateStatus{
			Enabled: *openapi.NewNullableBool(ptr.To(true)),
		},
	}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	defer deleteRoute(ctx, t)
	require.True(t, *notificationState.Slack.Enabled.Get())
}

func TestPatchNotificationState_EnableEmailAndThenDisable(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithSystemScope)
	response, err := client.NotificationChannelsAPI.CreateNotificationChannel(ctx).NotificationChannel(openapi.NotificationChannel{
		Type:   "email",
		Config: *openapi.NewNotificationChannelConfigWithDefaults(),
	}).Execute()
	defer deleteSink(ctx, "runai-email", t)

	notificationState, response, err := client.NotificationStateAPI.UpdateNotificationState(ctx).NotificationState(openapi.NotificationState{
		Email: &openapi.NotificationStateStatus{
			Enabled: *openapi.NewNullableBool(ptr.To(true)),
		},
	}).Execute()
	defer deleteRoute(ctx, t)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)

	require.True(t, *notificationState.Email.Enabled.Get())

	notificationState, response, err = client.NotificationStateAPI.UpdateNotificationState(ctx).NotificationState(openapi.NotificationState{
		Email: &openapi.NotificationStateStatus{
			Enabled: *openapi.NewNullableBool(ptr.To(false)),
		},
	}).Execute()

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.False(t, *notificationState.Email.Enabled.Get())
}

func TestGetNotificationState_TenantRouteIsMigrated(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)

	_, response, err := notificationsClient.TenantRouteAPI.CreateRoute(ctx, tenantIdWithSystemScope).RouteData(config_client.RouteData{
		TenantId: tenantIdWithSystemScope,
		RouteId:  defaultTenantRouteID,
		RouteConfig: map[string]interface{}{ // no metadata in RouteConfig
			"listener_id": "http-runai",
		},
	}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)

	defer deleteRoute(ctx, t)

	notificationState, response, err := client.NotificationStateAPI.GetNotificationState(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.True(t, *notificationState.Email.Enabled.Get()) // email should be enabled because of the migration
}

func TestPatchNotificationState_TenantRouteWithoutMetadata_IsMigrated(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)

	_, response, err := notificationsClient.TenantRouteAPI.CreateRoute(ctx, tenantIdWithSystemScope).RouteData(config_client.RouteData{
		TenantId: tenantIdWithSystemScope,
		RouteId:  defaultTenantRouteID,
		RouteConfig: map[string]interface{}{ // no metadata in RouteConfig
			"listener_id": "http-runai",
		},
	}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	defer deleteRoute(ctx, t)

	notificationState, response, err := client.NotificationStateAPI.UpdateNotificationState(ctx).NotificationState(openapi.NotificationState{
		Slack: &openapi.NotificationStateStatus{
			Enabled: *openapi.NewNullableBool(ptr.To(false)),
		},
	}).Execute()

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.False(t, *notificationState.Slack.Enabled.Get())
	require.True(t, *notificationState.Email.Enabled.Get()) // email should also be enabled because of the migration
}

func deleteSink(ctx context.Context, sinkId string, t *testing.T) {
	_, err := client.NotificationChannelsAPI.DeleteNotificationChannel(ctx, sinkId).Execute()
	require.NoError(t, err)
}

func deleteRoute(ctx context.Context, t *testing.T) {
	_, _, err := notificationsClient.TenantRouteAPI.DeleteRoute(ctx, tenantIdWithSystemScope, defaultTenantRouteID).Execute()
	require.NoError(t, err)
}
