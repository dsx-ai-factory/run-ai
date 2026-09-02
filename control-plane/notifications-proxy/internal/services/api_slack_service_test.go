/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/clients/state"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/notification_state"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/services/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/slack"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

const slackStateTestTenantId = "slack-tenant-id"

func TestAuthorizeSlackApp(t *testing.T) {
	ctx := context.Background()
	code := "1234567890"
	state := "state123"
	slackManager := mocks.NewSlackManager(t)
	slackManager.On("AuthorizeApp", ctx, code, state).Return(nil)

	slackService := &SlackAPIService{
		manager: slackManager,
	}
	res, err := slackService.AuthorizeSlackApp(ctx, code, state)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "Slack app authorized successfully. You can now return to the Runai platform.", res.Body)
}

func TestAuthorizeSlackApp_Failed(t *testing.T) {
	ctx := context.Background()
	code := "1234567890"
	state := "state123"
	slackManager := mocks.NewSlackManager(t)
	slackManager.On("AuthorizeApp", ctx, code, state).Return(errors.New("failed to authorize slack app"))

	slackService := &SlackAPIService{
		manager: slackManager,
	}
	res, err := slackService.AuthorizeSlackApp(ctx, code, state)
	require.Error(t, err)
	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.Equal(t, "failed to authorize slack app", err.Error())
}

func TestAuthorizeSlackApp_MissingAppCredentials(t *testing.T) {
	ctx := context.Background()
	code := "1234567890"
	state := "state123"
	slackManager := mocks.NewSlackManager(t)
	slackManager.On("AuthorizeApp", ctx, code, state).Return(slack.ErrMissingAppCredentials)

	slackService := &SlackAPIService{
		manager: slackManager,
	}
	res, err := slackService.AuthorizeSlackApp(ctx, code, state)
	require.EqualError(t, err, "missing slack app credentials")
	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestAuthorizeSlackApp_ExpiredState(t *testing.T) {
	ctx := context.Background()
	code := "1234567890"
	state := "state123"
	slackManager := mocks.NewSlackManager(t)
	slackManager.On("AuthorizeApp", ctx, code, state).Return(slack.ErrExpiredState)

	slackService := &SlackAPIService{
		manager: slackManager,
	}
	res, err := slackService.AuthorizeSlackApp(ctx, code, state)
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, res.Code)
	require.Equal(t, "slack state expired", err.Error())
}

func TestAuthorizeSlackApp_InvalidState(t *testing.T) {
	ctx := context.Background()
	code := "1234567890"
	state := "state123"
	slackManager := mocks.NewSlackManager(t)
	slackManager.On("AuthorizeApp", ctx, code, state).Return(slack.ErrInvalidState)

	slackService := &SlackAPIService{
		manager: slackManager,
	}
	res, err := slackService.AuthorizeSlackApp(ctx, code, state)
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, res.Code)
	require.Equal(t, "invalid slack state", err.Error())
}

func TestConnectSlackWorkspace(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "test@run.ai")
	slackManager := mocks.NewSlackManager(t)
	slackManager.On("ConnectWorkspace", ctx, "1", "test@run.ai").Return("redirect_url", nil)

	slackService := &SlackAPIService{
		manager: slackManager,
	}
	res, err := slackService.ConnectSlackWorkspace(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, res.Body, api.ConnectSlackWorkspace200Response{
		RedirectUrl: "redirect_url",
	})
}

func TestConnectSlackWorkspace_Failed(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "test@run.ai")
	slackManager := mocks.NewSlackManager(t)
	slackManager.On("ConnectWorkspace", ctx, "1", "test@run.ai").Return("", errors.New("failed to add slack app"))
	slackService := &SlackAPIService{
		manager: slackManager,
	}
	res, err := slackService.ConnectSlackWorkspace(ctx)
	require.Error(t, err)
	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.Nil(t, res.Body)
}

func TestConnectSlackWorkspace_MissingAppCredentials(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "test@run.ai")
	slackManager := mocks.NewSlackManager(t)
	slackManager.On("ConnectWorkspace", ctx, "1", "test@run.ai").Return("", slack.ErrMissingAppCredentials)
	slackService := &SlackAPIService{
		manager: slackManager,
	}
	res, err := slackService.ConnectSlackWorkspace(ctx)
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, res.Code)
	require.Nil(t, res.Body)
}

func TestConnectSlackWorkspace_MissingTenantId(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfSubjectID, "test@run.ai")
	slackService := &SlackAPIService{}
	res, err := slackService.ConnectSlackWorkspace(ctx)
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, res.Code)
	require.Equal(t, "missing tenant id in request context", err.Error())
}

func TestConnectSlackWorkspace_MissingSubjectId(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	slackService := &SlackAPIService{}

	res, err := slackService.ConnectSlackWorkspace(ctx)
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, res.Code)
	require.Equal(t, "missing subject id in request context", err.Error())
}

func TestCreateSlackApp(t *testing.T) {
	ctx := context.Background()
	accessToken := "xoxe.xoxp-1-token"
	slackManager := mocks.NewSlackManager(t)
	slackManager.On("CreateApp", ctx, accessToken).Return("app_url", nil)
	slackService := &SlackAPIService{
		manager: slackManager,
	}
	res, err := slackService.CreateSlackApp(ctx, api.CreateSlackAppRequestBody{AccessToken: accessToken})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, res.Body, api.CreateSlackAppResponse{SlackAppUrl: "app_url"})
}

func TestCreateSlackApp_Failed(t *testing.T) {
	ctx := context.Background()
	accessToken := "xoxe.xoxp-1-token"
	slackManager := mocks.NewSlackManager(t)
	slackManager.On("CreateApp", ctx, accessToken).Return("", errors.New("failed to create slack app"))
	slackService := &SlackAPIService{
		manager: slackManager,
	}
	res, err := slackService.CreateSlackApp(ctx, api.CreateSlackAppRequestBody{AccessToken: accessToken})
	require.Error(t, err)
	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.Equal(t, "failed to create slack app", err.Error())
}

func TestCreateSlackApp_InvalidAccessToken(t *testing.T) {
	ctx := context.Background()
	accessToken := "xoxb-1234567890"
	slackService := &SlackAPIService{}
	res, err := slackService.CreateSlackApp(ctx, api.CreateSlackAppRequestBody{AccessToken: accessToken})
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, res.Code)
	require.EqualError(t, err, "accessToken is invalid, must be a Slack token format (e.g., xoxp-123... or xoxe.xoxp-123...)")
}

func TestDeleteSlackApp(t *testing.T) {
	ctx := context.Background()
	slackManager := mocks.NewSlackManager(t)
	slackManager.On("DeleteApp", ctx).Return(nil)
	slackService := &SlackAPIService{
		manager: slackManager,
	}
	res, err := slackService.DeleteSlackApp(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
}

func TestDeleteSlackApp_Failed(t *testing.T) {
	ctx := context.Background()
	slackManager := mocks.NewSlackManager(t)
	slackManager.On("DeleteApp", ctx).Return(errors.New("failed to delete slack app"))
	slackService := &SlackAPIService{
		manager: slackManager,
	}
	res, err := slackService.DeleteSlackApp(ctx)
	require.Error(t, err)
	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.Equal(t, "failed to delete slack app", err.Error())
}

func TestDisconnectSlackWorkspace(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	tenantId := "1"
	slackManager := mocks.NewSlackManager(t)
	slackService := &SlackAPIService{
		manager: slackManager,
	}

	slackManager.On("DisconnectWorkspace", ctx, tenantId).Return(nil)
	res, err := slackService.DisconnectSlackWorkspace(ctx)
	require.NoError(t, err)
	require.Equal(t, res.Code, http.StatusOK)
}

func TestDisconnectSlackWorkspace_MissingTenantId(t *testing.T) {
	slackManager := mocks.NewSlackManager(t)
	slackService := &SlackAPIService{
		manager: slackManager,
	}

	res, err := slackService.DisconnectSlackWorkspace(context.Background())
	require.Error(t, err)
	require.Equal(t, res.Code, http.StatusBadRequest)
	require.Equal(t, err.Error(), "missing tenant id in request context")
}

func TestDisconnectSlackWorkspace_Failed(t *testing.T) {
	slackManager := mocks.NewSlackManager(t)
	slackService := &SlackAPIService{
		manager: slackManager,
	}

	slackManager.On("DisconnectWorkspace", ctx, tenantId).Return(errors.New("failed to disconnect slack workspace"))
	res, err := slackService.DisconnectSlackWorkspace(ctx)
	require.Error(t, err)
	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.Nil(t, res.Body)
}

func TestGetSlack(t *testing.T) {
	slackManager := mocks.NewSlackManager(t)
	slackService := &SlackAPIService{
		manager: slackManager,
	}

	slackManager.On("GetSlackConfig", ctx, tenantId).Return(&slack.ChannelConfig{
		ApiToken:            "xoxp-token",
		Recipients:          []string{"channel1"},
		DirectNotifications: ptr.To(true),
	}, nil)

	res, err := slackService.GetSlack(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, api.SlackNotificationChannel{
		Config: api.SlackNotificationChannelConfig{
			Recipients:          []string{"channel1"},
			DirectNotifications: ptr.To(true),
		},
	}, res.Body)
}

func TestGetSlack_MissingTenantId(t *testing.T) {
	slackService := &SlackAPIService{}

	res, err := slackService.GetSlack(context.Background())
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, res.Code)
	require.Equal(t, "missing tenant id in request context", err.Error())
}

func TestGetSlack_Failed(t *testing.T) {
	slackManager := mocks.NewSlackManager(t)
	slackService := &SlackAPIService{
		manager: slackManager,
	}

	slackManager.On("GetSlackConfig", ctx, tenantId).Return(nil, errors.New("failed to get slack config"))

	res, err := slackService.GetSlack(ctx)
	require.Error(t, err)
	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.Nil(t, res.Body)
}

func TestGetSlack_NotFound(t *testing.T) {
	slackManager := mocks.NewSlackManager(t)
	slackService := &SlackAPIService{
		manager: slackManager,
	}

	slackManager.On("GetSlackConfig", ctx, tenantId).Return(nil, slack.ErrNotFound)

	res, err := slackService.GetSlack(ctx)
	require.EqualError(t, err, "slack app not found")
	require.Equal(t, http.StatusNotFound, res.Code)
	require.Nil(t, res.Body)
}

func TestPatchSlack(t *testing.T) {
	slackManager := mocks.NewSlackManager(t)
	slackService := &SlackAPIService{
		manager: slackManager,
	}

	patchRequest := api.SlackNotificationChannelPatchRequest{
		Config: api.SlackNotificationChannelPatchRequestConfig{
			Recipients:          []string{"channel1"},
			DirectNotifications: ptr.To(true),
		},
	}

	slackManager.On("PatchSlackConfig", ctx, tenantId, slack.ChannelConfig{
		Recipients:          []string{"channel1"},
		DirectNotifications: ptr.To(true),
	}).Return(&slack.ChannelConfig{
		ApiToken:            "xoxp-token",
		Recipients:          []string{"channel1"},
		DirectNotifications: ptr.To(true),
	}, nil)

	res, err := slackService.PatchSlack(ctx, patchRequest)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, api.SlackNotificationChannel{Config: api.SlackNotificationChannelConfig{
		Recipients:          []string{"channel1"},
		DirectNotifications: ptr.To(true),
	},
	}, res.Body)
}

func TestPatchSlack_MissingTenantId(t *testing.T) {
	slackService := &SlackAPIService{}

	res, err := slackService.PatchSlack(context.Background(), api.SlackNotificationChannelPatchRequest{})
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, res.Code)
	require.Equal(t, "missing tenant id in request context", err.Error())
}

func TestPatchSlack_Failed(t *testing.T) {
	slackManager := mocks.NewSlackManager(t)
	slackService := &SlackAPIService{
		manager: slackManager,
	}

	slackManager.On("PatchSlackConfig", ctx, tenantId, slack.ChannelConfig{}).Return(nil, errors.New("failed to patch slack config"))

	res, err := slackService.PatchSlack(ctx, api.SlackNotificationChannelPatchRequest{})
	require.EqualError(t, err, "failed to patch slack app")
	require.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestPatchSlack_NotFound(t *testing.T) {
	slackManager := mocks.NewSlackManager(t)
	slackService := &SlackAPIService{
		manager: slackManager,
	}

	slackManager.On("PatchSlackConfig", ctx, tenantId, slack.ChannelConfig{}).Return(nil, slack.ErrNotFound)

	res, err := slackService.PatchSlack(ctx, api.SlackNotificationChannelPatchRequest{})
	require.EqualError(t, err, "slack app not found")
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestValidateSlack_Succeed(t *testing.T) {
	ctx := context.Background()
	slackManager := mocks.NewSlackManager(t)
	slackService := &SlackAPIService{manager: slackManager}

	request := api.SlackNotificationChannelConfig{
		ApiToken:            "xoxb-token",
		DirectNotifications: ptr.To(true),
		Recipients:          []string{"my-channel"},
	}
	slackManager.On("ValidateSlack", ctx, slack.ChannelConfig{
		ApiToken:            request.ApiToken,
		DirectNotifications: request.DirectNotifications,
		Recipients:          request.Recipients,
	}).Return(slack.ValidationResponse{Status: true}, nil)

	res, err := slackService.ValidateSlack(ctx, request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	body := res.Body.(api.ValidateSinkResponse)
	require.True(t, body.Status)
}

func TestValidateSlack_Failed(t *testing.T) {
	ctx := context.Background()
	slackManager := mocks.NewSlackManager(t)
	slackService := &SlackAPIService{manager: slackManager}

	request := api.SlackNotificationChannelConfig{ApiToken: "xoxb-token"}
	slackManager.On("ValidateSlack", ctx, mock.Anything).Return(slack.ValidationResponse{Status: false}, errors.New("invalid token"))

	res, err := slackService.ValidateSlack(ctx, request)
	require.EqualError(t, err, "slack validation failed")
	require.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestValidateSlack_ValidationFailed(t *testing.T) {
	ctx := context.Background()
	slackManager := mocks.NewSlackManager(t)
	slackService := &SlackAPIService{manager: slackManager}

	request := api.SlackNotificationChannelConfig{ApiToken: "xoxb-token"}
	slackManager.On("ValidateSlack", ctx, mock.Anything).Return(slack.ValidationResponse{Status: false}, slack.ErrValidationFailed)

	res, err := slackService.ValidateSlack(ctx, request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	body := res.Body.(api.ValidateSinkResponse)
	require.False(t, body.Status)
}

func TestIsAccessTokenValid(t *testing.T) {
	require.True(t, isAccessTokenValid("xoxp-1234567890abcd"))
	require.True(t, isAccessTokenValid("xoxe.xoxp-1234567890abcd"))

	require.False(t, isAccessTokenValid("1234567890abcd"))
	require.False(t, isAccessTokenValid("xoxb-1234567890abcd"))
	require.False(t, isAccessTokenValid("xoxe.1234567890abcd"))
}

func TestGetSlackNotificationState_MissingTenantId(t *testing.T) {
	ctx := context.Background()
	stateMock := mocks.NewSlackNotificationStateManager(t)
	sas := &SlackAPIService{stateManager: stateMock}

	res, err := sas.GetSlackNotificationState(ctx)
	require.EqualError(t, err, "missing tenant id in request context")
	require.NotNil(t, res)
	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestGetSlackNotificationState_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, slackStateTestTenantId)
	stateMock := mocks.NewSlackNotificationStateManager(t)
	sas := &SlackAPIService{stateManager: stateMock}
	slackState := &state.SlackState{Enabled: ptr.To(true)}
	stateMock.On("GetSlackNotificationState", ctx, slackStateTestTenantId).Return(slackState, nil)

	res, err := sas.GetSlackNotificationState(ctx)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, getAPIState(slackState), res.Body)
}

func TestGetSlackNotificationState_SuccessWhenErrFailedToRemoveSinkFromUserRoutes(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, slackStateTestTenantId)
	stateMock := mocks.NewSlackNotificationStateManager(t)
	sas := &SlackAPIService{stateManager: stateMock}
	slackState := &state.SlackState{Enabled: ptr.To(false)}
	stateMock.On("GetSlackNotificationState", ctx, slackStateTestTenantId).Return(slackState, notification_state.ErrFailedToRemoveSinkFromUserRoutes)

	res, err := sas.GetSlackNotificationState(ctx)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, getAPIState(slackState), res.Body)
}

func TestGetSlackNotificationState_ErrNoSinkConfigured(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, slackStateTestTenantId)
	stateMock := mocks.NewSlackNotificationStateManager(t)
	sas := &SlackAPIService{stateManager: stateMock}
	stateMock.On("GetSlackNotificationState", ctx, slackStateTestTenantId).Return(nil, notification_state.ErrNoSinkConfigured)

	res, err := sas.GetSlackNotificationState(ctx)
	require.EqualError(t, err, "slack app not configured to update state")
	require.NotNil(t, res)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestGetSlackNotificationState_InternalError(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, slackStateTestTenantId)
	stateMock := mocks.NewSlackNotificationStateManager(t)
	sas := &SlackAPIService{stateManager: stateMock}
	stateMock.On("GetSlackNotificationState", ctx, slackStateTestTenantId).Return(nil, errors.New("backend error"))

	res, err := sas.GetSlackNotificationState(ctx)
	require.EqualError(t, err, "failed to get slack state")
	require.NotNil(t, res)
	require.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestGetDetailedSlackNotificationState_MissingTenantId(t *testing.T) {
	ctx := context.Background()
	stateMock := mocks.NewSlackNotificationStateManager(t)
	sas := &SlackAPIService{stateManager: stateMock}

	res, err := sas.GetDetailedSlackNotificationState(ctx)
	require.EqualError(t, err, "missing tenant id in request context")
	require.NotNil(t, res)
	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestGetDetailedSlackNotificationState_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, slackStateTestTenantId)
	stateMock := mocks.NewSlackNotificationStateManager(t)
	sas := &SlackAPIService{stateManager: stateMock}
	slackState := &state.SlackState{
		Enabled:                  ptr.To(true),
		WorkspaceName:            ptr.To("my-workspace"),
		AppCredentialsConfigured: ptr.To(true),
	}
	stateMock.On("GetDetailedSlackNotificationState", ctx, slackStateTestTenantId).Return(slackState, nil)

	res, err := sas.GetDetailedSlackNotificationState(ctx)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, getAPISlackStateWithDetails(slackState), res.Body)
}

func TestGetDetailedSlackNotificationState_SuccessWhenErrFailedToRemoveSinkFromUserRoutes(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, slackStateTestTenantId)
	stateMock := mocks.NewSlackNotificationStateManager(t)
	sas := &SlackAPIService{stateManager: stateMock}
	slackState := &state.SlackState{Enabled: ptr.To(false)}
	stateMock.On("GetDetailedSlackNotificationState", ctx, slackStateTestTenantId).Return(slackState, notification_state.ErrFailedToRemoveSinkFromUserRoutes)

	res, err := sas.GetDetailedSlackNotificationState(ctx)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, getAPISlackStateWithDetails(slackState), res.Body)
}

func TestGetDetailedSlackNotificationState_ErrNoSinkConfigured(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, slackStateTestTenantId)
	stateMock := mocks.NewSlackNotificationStateManager(t)
	sas := &SlackAPIService{stateManager: stateMock}
	stateMock.On("GetDetailedSlackNotificationState", ctx, slackStateTestTenantId).Return(nil, notification_state.ErrNoSinkConfigured)

	res, err := sas.GetDetailedSlackNotificationState(ctx)
	require.EqualError(t, err, "slack app not configured to update state")
	require.NotNil(t, res)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestGetDetailedSlackNotificationState_InternalError(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, slackStateTestTenantId)
	stateMock := mocks.NewSlackNotificationStateManager(t)
	sas := &SlackAPIService{stateManager: stateMock}
	stateMock.On("GetDetailedSlackNotificationState", ctx, slackStateTestTenantId).Return(nil, errors.New("backend error"))

	res, err := sas.GetDetailedSlackNotificationState(ctx)
	require.EqualError(t, err, "failed to get slack state")
	require.NotNil(t, res)
	require.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestUpdateSlackNotificationState_MissingTenantId(t *testing.T) {
	ctx := context.Background()
	stateMock := mocks.NewSlackNotificationStateManager(t)
	sas := &SlackAPIService{stateManager: stateMock}

	res, err := sas.UpdateSlackNotificationState(ctx, api.NotificationStateStatus{Enabled: ptr.To(true)})
	require.EqualError(t, err, "missing tenant id in request context")
	require.NotNil(t, res)
	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestUpdateSlackNotificationState_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, slackStateTestTenantId)
	stateMock := mocks.NewSlackNotificationStateManager(t)
	sas := &SlackAPIService{stateManager: stateMock}
	updatedState := &state.SlackState{TenantId: slackStateTestTenantId, Enabled: ptr.To(true)}
	stateMock.On("UpdateSlackNotificationState", ctx, slackStateTestTenantId, mock.MatchedBy(func(s notification_state.Stater) bool {
		ss, ok := s.(*state.SlackState)
		return ok && ss != nil && ss.TenantId == slackStateTestTenantId && ss.Enabled != nil && *ss.Enabled
	})).Return(updatedState, nil)

	res, err := sas.UpdateSlackNotificationState(ctx, api.NotificationStateStatus{Enabled: ptr.To(true)})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, getAPIState(updatedState), res.Body)
}

func TestUpdateSlackNotificationState_InternalError(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, slackStateTestTenantId)
	stateMock := mocks.NewSlackNotificationStateManager(t)
	sas := &SlackAPIService{stateManager: stateMock}
	stateMock.On("UpdateSlackNotificationState", ctx, slackStateTestTenantId, mock.Anything).Return(nil, errors.New("backend error"))

	res, err := sas.UpdateSlackNotificationState(ctx, api.NotificationStateStatus{Enabled: ptr.To(false)})
	require.EqualError(t, err, "failed to update slack notifications state")
	require.NotNil(t, res)
	require.Equal(t, http.StatusInternalServerError, res.Code)
}
