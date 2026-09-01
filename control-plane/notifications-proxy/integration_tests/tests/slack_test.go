//go:build integration
// +build integration

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	openapi "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/require"
)

const (
	authAppEndpoint = "/api/v1/notification-channels/slack/auth-app"
	code            = "5439c659-19a5-5335-ca5b-bfbc83b1fd77"
)

var useLocalHostRedirect = GetEnvOrDefault("USE_LOCALHOST_REDIRECT", "true") == "true"

func TestAddSlackAppDeprecated(t *testing.T) {
	domain := GetEnvOrDefault("RUNAI_ENV_DOMAIN", "http://localhost:8080")
	accessToken := "xoxb-" + uuid.NewString()
	ctx := getInitializedContext()

	initProxyAPIClient()
	saveResponse(t, "save-authorize-response", map[string]interface{}{
		"code":         code,
		"redirect_uri": domain + authAppEndpoint,
	})

	saveResponse(t, "save-oauth-access-response", map[string]interface{}{
		"ok":           true,
		"access_token": accessToken,
		"team": map[string]interface{}{
			"name": "Test Workspace",
		},
	})

	res, httpRes, err := client.SlackAPI.AddSlackAppDeprecated(ctx).Execute()
	require.Equal(t, httpRes.StatusCode, http.StatusOK)
	require.NoError(t, err)
	require.NotEmpty(t, res.RedirectUrl)

	redirectUrl := getRedirectUrl(*res.RedirectUrl)

	redirectResp, err := http.Get(redirectUrl)
	require.NoError(t, err)
	defer redirectResp.Body.Close()

	bodyBytes, err := io.ReadAll(redirectResp.Body)
	require.NoError(t, err)

	require.Contains(t, string(bodyBytes), "Slack app authorized successfully. You can now return to the Runai platform.")

	channel, _, err := client.NotificationChannelsAPI.GetNotificationChannel(ctx, fmt.Sprintf("slack-%s", tenantIdWithoutSystemScope)).Execute()
	require.NoError(t, err)
	require.NotNil(t, channel)
	require.Equal(t, channel.Type, openapi.NotificationChannelType("slack"))

	notificationState, response, err := client.NotificationStateAPI.GetDetailedNotificationState(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.True(t, *notificationState.Slack.Enabled.Get())
	require.Equal(t, *notificationState.Slack.Settings.WorkspaceName.Get(), "Test Workspace")
	require.False(t, *notificationState.Email.Enabled.Get())
}

func TestAddSlackAppDeprecated_FailToRedirectToAuthorize(t *testing.T) {
	ctx := getInitializedContext()

	initProxyAPIClient()
	saveResponse(t, "save-authorize-response", map[string]interface{}{})

	res, httpRes, err := client.SlackAPI.AddSlackAppDeprecated(ctx).Execute()
	require.Equal(t, httpRes.StatusCode, http.StatusOK)
	require.NoError(t, err)
	require.NotEmpty(t, res.RedirectUrl)

	redirectUrl := getRedirectUrl(*res.RedirectUrl)

	redirectResp, err := http.Get(redirectUrl)
	require.NoError(t, err)
	defer redirectResp.Body.Close()

	require.Equal(t, redirectResp.StatusCode, http.StatusNotFound)
	require.Equal(t, redirectResp.Status, "404 Not Found")
}

func TestAddSlackAppDeprecated_FailToGetAccessCode(t *testing.T) {
	domain := GetEnvOrDefault("RUNAI_ENV_DOMAIN", "http://localhost:8080")
	ctx := getInitializedContext()

	initProxyAPIClient()
	saveResponse(t, "save-authorize-response", map[string]interface{}{
		"code":         code,
		"redirect_uri": domain + authAppEndpoint,
	})

	saveResponse(t, "save-oauth-access-response", map[string]interface{}{
		"ok":    false,
		"error": "failed to get authenticate code",
	})

	res, httpRes, err := client.SlackAPI.AddSlackAppDeprecated(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, httpRes.StatusCode, http.StatusOK)
	require.NotEmpty(t, res.RedirectUrl)

	redirectUrl := getRedirectUrl(*res.RedirectUrl)

	redirectResp, err := http.Get(redirectUrl)
	require.NoError(t, err)
	defer redirectResp.Body.Close()

	require.Equal(t, redirectResp.StatusCode, http.StatusInternalServerError)
	require.Equal(t, redirectResp.Status, "500 Internal Server Error")
}

func TestConnectSlackWorkspace(t *testing.T) {
	domain := GetEnvOrDefault("RUNAI_ENV_DOMAIN", "http://localhost:8080")
	accessToken := "xoxb-" + uuid.NewString()
	ctx := getInitializedContext()

	initProxyAPIClient()
	saveResponse(t, "save-authorize-response", map[string]interface{}{
		"code":         code,
		"redirect_uri": domain + authAppEndpoint,
	})

	saveResponse(t, "save-oauth-access-response", map[string]interface{}{
		"ok":           true,
		"access_token": accessToken,
		"team": map[string]interface{}{
			"name": "Test Workspace",
		},
	})

	res, httpRes, err := client.SlackAPI.ConnectSlackWorkspace(ctx).Execute()
	require.Equal(t, httpRes.StatusCode, http.StatusOK)
	require.NoError(t, err)
	require.NotEmpty(t, res.RedirectUrl)

	redirectUrl := getRedirectUrl(*res.RedirectUrl)

	redirectResp, err := http.Get(redirectUrl)
	require.NoError(t, err)
	defer redirectResp.Body.Close()

	bodyBytes, err := io.ReadAll(redirectResp.Body)
	require.NoError(t, err)

	require.Contains(t, string(bodyBytes), "Slack app authorized successfully. You can now return to the Runai platform.")

	channel, _, err := client.SlackAPI.GetSlack(ctx).Execute()
	require.NoError(t, err)
	require.NotNil(t, channel)

	notificationState, response, err := client.NotificationStateAPI.GetDetailedNotificationState(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.True(t, *notificationState.Slack.Enabled.Get())
	require.Equal(t, *notificationState.Slack.Settings.WorkspaceName.Get(), "Test Workspace")
	require.False(t, *notificationState.Email.Enabled.Get())
}

func TestConnectWorkspace_FailToRedirectToAuthorize(t *testing.T) {
	ctx := getInitializedContext()

	initProxyAPIClient()
	saveResponse(t, "save-authorize-response", map[string]interface{}{})

	res, httpRes, err := client.SlackAPI.ConnectSlackWorkspace(ctx).Execute()
	require.Equal(t, httpRes.StatusCode, http.StatusOK)
	require.NoError(t, err)
	require.NotEmpty(t, res.RedirectUrl)

	redirectUrl := getRedirectUrl(*res.RedirectUrl)

	redirectResp, err := http.Get(redirectUrl)
	require.NoError(t, err)
	defer redirectResp.Body.Close()

	require.Equal(t, redirectResp.StatusCode, http.StatusNotFound)
	require.Equal(t, redirectResp.Status, "404 Not Found")
}

func TestConnectWorkspace_FailToGetAccessCode(t *testing.T) {
	domain := GetEnvOrDefault("RUNAI_ENV_DOMAIN", "http://localhost:8080")
	ctx := getInitializedContext()

	initProxyAPIClient()
	saveResponse(t, "save-authorize-response", map[string]interface{}{
		"code":         code,
		"redirect_uri": domain + authAppEndpoint,
	})

	saveResponse(t, "save-oauth-access-response", map[string]interface{}{
		"ok":    false,
		"error": "failed to get authenticate code",
	})

	res, httpRes, err := client.SlackAPI.ConnectSlackWorkspace(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, httpRes.StatusCode, http.StatusOK)
	require.NotEmpty(t, res.RedirectUrl)

	redirectUrl := getRedirectUrl(*res.RedirectUrl)

	redirectResp, err := http.Get(redirectUrl)
	require.NoError(t, err)
	defer redirectResp.Body.Close()

	require.Equal(t, redirectResp.StatusCode, http.StatusInternalServerError)
	require.Equal(t, redirectResp.Status, "500 Internal Server Error")
}

func TestPatchSlack(t *testing.T) {
	domain := GetEnvOrDefault("RUNAI_ENV_DOMAIN", "http://localhost:8080")
	accessToken := "xoxb-" + uuid.NewString()
	ctx := getInitializedContext()

	initProxyAPIClient()
	saveResponse(t, "save-authorize-response", map[string]interface{}{
		"code":         code,
		"redirect_uri": domain + authAppEndpoint,
	})

	saveResponse(t, "save-oauth-access-response", map[string]interface{}{
		"ok":           true,
		"access_token": accessToken,
		"team": map[string]interface{}{
			"name": "Test Workspace",
		},
	})

	res, httpRes, err := client.SlackAPI.ConnectSlackWorkspace(ctx).Execute()
	require.Equal(t, httpRes.StatusCode, http.StatusOK)
	require.NoError(t, err)
	require.NotEmpty(t, res.RedirectUrl)

	redirectUrl := getRedirectUrl(*res.RedirectUrl)

	redirectResp, err := http.Get(redirectUrl)
	require.NoError(t, err)
	defer redirectResp.Body.Close()

	bodyBytes, err := io.ReadAll(redirectResp.Body)
	require.NoError(t, err)

	require.Contains(t, string(bodyBytes), "Slack app authorized successfully. You can now return to the Runai platform.")

	slackChannel, _, err := client.SlackAPI.GetSlack(ctx).Execute()
	require.NoError(t, err)
	require.NotNil(t, slackChannel)

	conf, httpRes, err := client.SlackAPI.PatchSlack(ctx).SlackNotificationChannelPatchRequest(
		openapi.SlackNotificationChannelPatchRequest{Config: openapi.SlackNotificationChannelPatchRequestConfig{Recipients: []string{"my-notifications-slackChannel"}}}).Execute()
	require.NoError(t, err)
	require.Equal(t, httpRes.StatusCode, http.StatusOK)
	require.Equal(t, conf.Config.Recipients, []string{"my-notifications-slackChannel"})
}

func TestDisconnectWorkspace(t *testing.T) {
	domain := GetEnvOrDefault("RUNAI_ENV_DOMAIN", "http://localhost:8080")
	ctx := getInitializedContext()

	initProxyAPIClient()
	saveResponse(t, "save-authorize-response", map[string]interface{}{
		"code":         code,
		"redirect_uri": domain + authAppEndpoint,
	})

	saveResponse(t, "save-oauth-access-response", map[string]interface{}{
		"ok":           true,
		"access_token": "xoxb-" + uuid.NewString(),
		"team": map[string]interface{}{
			"name": "Test Workspace",
		},
	})

	res, _, err := client.SlackAPI.ConnectSlackWorkspace(ctx).Execute()
	require.NoError(t, err)
	require.NotEmpty(t, res.RedirectUrl)

	redirectUrl := getRedirectUrl(*res.RedirectUrl)

	redirectResp, _ := http.Get(redirectUrl)
	defer redirectResp.Body.Close()

	bodyBytes, _ := io.ReadAll(redirectResp.Body)
	require.Contains(t, string(bodyBytes), "Slack app authorized successfully. You can now return to the Runai platform.")

	httpRes, err := client.SlackAPI.DisconnectSlackWorkspace(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, httpRes.StatusCode, http.StatusOK)

	_, httpRes, err = client.SlackAPI.GetSlack(ctx).Execute()
	require.Error(t, err)
	require.Equal(t, httpRes.StatusCode, http.StatusNotFound)

	notificationState, response, err := client.NotificationStateAPI.GetDetailedNotificationState(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.False(t, *notificationState.Slack.Enabled.Get())
}

func TestDeleteApp(t *testing.T) {
	domain := GetEnvOrDefault("RUNAI_ENV_DOMAIN", "http://localhost:8080")
	accessToken := "xoxb-" + uuid.NewString()
	ctx := getInitializedContext()

	initProxyAPIClient()
	saveResponse(t, "save-authorize-response", map[string]interface{}{
		"code":         code,
		"redirect_uri": domain + authAppEndpoint,
	})

	saveResponse(t, "save-oauth-access-response", map[string]interface{}{
		"ok":           true,
		"access_token": accessToken,
		"team": map[string]interface{}{
			"name": "Test Workspace",
		},
	})

	res, httpRes, err := client.SlackAPI.ConnectSlackWorkspace(ctx).Execute()
	require.Equal(t, httpRes.StatusCode, http.StatusOK)
	require.NoError(t, err)
	require.NotEmpty(t, res.RedirectUrl)

	redirectUrl := getRedirectUrl(*res.RedirectUrl)

	redirectResp, err := http.Get(redirectUrl)
	require.NoError(t, err)
	defer redirectResp.Body.Close()

	bodyBytes, err := io.ReadAll(redirectResp.Body)
	require.NoError(t, err)

	require.Contains(t, string(bodyBytes), "Slack app authorized successfully. You can now return to the Runai platform.")

	httpRes, err = client.SlackAPI.DeleteSlackApp(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, httpRes.StatusCode, http.StatusOK)

	res, httpRes, err = client.SlackAPI.ConnectSlackWorkspace(ctx).Execute()
	require.Error(t, err)
	require.NotNil(t, httpRes)
	require.Equal(t, httpRes.StatusCode, http.StatusBadRequest)
	bodyBytes, err = io.ReadAll(httpRes.Body)
	require.NoError(t, err)
	defer httpRes.Body.Close()
	require.Contains(t, string(bodyBytes), "missing slack app credentials")
}

func getInitializedContext() context.Context {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithoutSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "test@run.ai")
	return ctx
}

func saveResponse(t *testing.T, endpoint string, result map[string]interface{}) {
	slackDomain := GetEnvOrDefault("SLACK_BASE_URL", "http://localhost:8081")
	payload, err := json.Marshal(result)

	require.NoError(t, err)
	resp, err := http.Post(slackDomain+"/"+endpoint, "application/json", bytes.NewReader(payload))
	require.NoError(t, err)
	defer resp.Body.Close()
}

func getRedirectUrl(originalUrl string) string {
	if useLocalHostRedirect {
		return strings.Replace(originalUrl, "slack-mock", "localhost", 1)
	}
	return originalUrl
}
