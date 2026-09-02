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
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/email"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/notification_state"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/services/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

const emailStateTestTenantId = "test-tenant-id"

func TestGetEmail_Succeed(t *testing.T) {
	ctx := context.Background()
	mockManager := mocks.NewEmailManager(t)
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{manager: mockManager, authorizer: mockAuthorizer}

	emailChannelConfig := &email.ChannelConfig{
		Username: "user",
	}

	mockAuthorizer.On("IsSystem", ctx).Return(true)
	mockManager.On("GetEmail", ctx).Return(emailChannelConfig, nil)

	res, err := eas.GetEmail(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, emailChannelConfigToAPI(emailChannelConfig), res.Body)
}

func TestGetEmail_Forbidden(t *testing.T) {
	ctx := context.Background()
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{authorizer: mockAuthorizer}

	mockAuthorizer.On("IsSystem", ctx).Return(false)

	res, err := eas.GetEmail(ctx)
	require.EqualError(t, err, "access denied")
	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestGetEmail_NotFound(t *testing.T) {
	ctx := context.Background()
	mockManager := mocks.NewEmailManager(t)
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{manager: mockManager, authorizer: mockAuthorizer}

	mockAuthorizer.On("IsSystem", ctx).Return(true)
	mockManager.On("GetEmail", ctx).Return(nil, email.ErrNotFound)

	res, err := eas.GetEmail(ctx)
	require.EqualError(t, err, "email not found")
	require.Equal(t, http.StatusNotFound, res.Code)
	require.Nil(t, res.Body)
}

func TestGetEmail_Failed(t *testing.T) {
	ctx := context.Background()
	mockManager := mocks.NewEmailManager(t)
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{manager: mockManager, authorizer: mockAuthorizer}

	mockAuthorizer.On("IsSystem", ctx).Return(true)
	mockManager.On("GetEmail", ctx).Return(nil, errors.New("error"))

	res, err := eas.GetEmail(ctx)
	require.EqualError(t, err, "failed to get email")
	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.Nil(t, res.Body)
}

func TestUpdateEmail_Succeed(t *testing.T) {
	ctx := context.Background()
	mockManager := mocks.NewEmailManager(t)
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{manager: mockManager, authorizer: mockAuthorizer}

	apiEmail := api.EmailChannel{
		Config: api.EmailChannelConfig{
			Username: "user",
			Password: "pass",
		},
	}
	domainEmail := emailChannelFromAPI(apiEmail)

	mockAuthorizer.On("IsSystem", ctx).Return(true)
	mockManager.On("UpdateEmail", ctx, domainEmail).Return(domainEmail, nil)

	res, err := eas.UpdateEmail(ctx, apiEmail)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, apiEmail, res.Body)
}

func TestUpdateEmail_Forbidden(t *testing.T) {
	ctx := context.Background()
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{authorizer: mockAuthorizer}

	mockAuthorizer.On("IsSystem", ctx).Return(false)

	res, err := eas.UpdateEmail(ctx, api.EmailChannel{})
	require.EqualError(t, err, "access denied")
	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestUpdateEmail_NotFound(t *testing.T) {
	ctx := context.Background()
	mockManager := mocks.NewEmailManager(t)
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{manager: mockManager, authorizer: mockAuthorizer}

	apiEmail := api.EmailChannel{
		Config: api.EmailChannelConfig{
			Username: "user",
			Password: "pass",
		},
	}
	domainEmail := emailChannelFromAPI(apiEmail)

	mockAuthorizer.On("IsSystem", ctx).Return(true)
	mockManager.On("UpdateEmail", ctx, domainEmail).Return(nil, email.ErrNotFound)

	res, err := eas.UpdateEmail(ctx, apiEmail)
	require.EqualError(t, err, "email not found")
	require.Equal(t, http.StatusNotFound, res.Code)
	require.Nil(t, res.Body)
}

func TestUpdateEmail_Failed(t *testing.T) {
	ctx := context.Background()
	mockManager := mocks.NewEmailManager(t)
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{manager: mockManager, authorizer: mockAuthorizer}

	apiEmail := api.EmailChannel{
		Config: api.EmailChannelConfig{
			Username: "user",
			Password: "pass",
		},
	}
	domainEmail := emailChannelFromAPI(apiEmail)

	mockAuthorizer.On("IsSystem", ctx).Return(true)
	mockManager.On("UpdateEmail", ctx, domainEmail).Return(nil, errors.New("error"))

	res, err := eas.UpdateEmail(ctx, apiEmail)
	require.EqualError(t, err, "failed to update email")
	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.Nil(t, res.Body)
}

func TestCreateEmail_Succeed(t *testing.T) {
	ctx := context.Background()
	mockManager := mocks.NewEmailManager(t)
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{manager: mockManager, authorizer: mockAuthorizer}

	apiEmail := api.EmailChannel{
		Config: api.EmailChannelConfig{
			Username: "user",
			Password: "pass",
		},
	}
	domainEmail := emailChannelFromAPI(apiEmail)

	mockAuthorizer.On("IsSystem", ctx).Return(true)
	mockManager.On("CreateEmail", ctx, domainEmail).Return(domainEmail, nil)

	res, err := eas.CreateEmail(ctx, apiEmail)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, apiEmail, res.Body)
}

func TestCreateEmail_Forbidden(t *testing.T) {
	ctx := context.Background()
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{authorizer: mockAuthorizer}

	mockAuthorizer.On("IsSystem", ctx).Return(false)

	res, err := eas.CreateEmail(ctx, api.EmailChannel{})
	require.EqualError(t, err, "access denied")
	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestCreateEmail_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	mockManager := mocks.NewEmailManager(t)
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{manager: mockManager, authorizer: mockAuthorizer}

	apiEmail := api.EmailChannel{
		Config: api.EmailChannelConfig{
			Username: "user",
			Password: "pass",
		},
	}
	domainEmail := emailChannelFromAPI(apiEmail)

	mockAuthorizer.On("IsSystem", ctx).Return(true)
	mockManager.On("CreateEmail", ctx, domainEmail).Return(nil, email.ErrAlreadyExists)

	res, err := eas.CreateEmail(ctx, apiEmail)
	require.EqualError(t, err, "email already exists")
	require.Equal(t, http.StatusConflict, res.Code)
	require.Nil(t, res.Body)
}

func TestCreateEmail_Failed(t *testing.T) {
	ctx := context.Background()
	mockManager := mocks.NewEmailManager(t)
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{manager: mockManager, authorizer: mockAuthorizer}

	apiEmail := api.EmailChannel{
		Config: api.EmailChannelConfig{
			Username: "user",
			Password: "pass",
		},
	}
	domainEmail := emailChannelFromAPI(apiEmail)

	mockAuthorizer.On("IsSystem", ctx).Return(true)
	mockManager.On("CreateEmail", ctx, domainEmail).Return(nil, errors.New("error"))

	res, err := eas.CreateEmail(ctx, apiEmail)
	require.EqualError(t, err, "failed to create email")
	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.Nil(t, res.Body)
}

func TestDeleteEmail_Succeed(t *testing.T) {
	ctx := context.Background()
	mockManager := mocks.NewEmailManager(t)
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{manager: mockManager, authorizer: mockAuthorizer}

	mockAuthorizer.On("IsSystem", ctx).Return(true)
	mockManager.On("DeleteEmail", ctx).Return(nil)

	res, err := eas.DeleteEmail(ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, res.Code)
}

func TestDeleteEmail_Forbidden(t *testing.T) {
	ctx := context.Background()
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{authorizer: mockAuthorizer}

	mockAuthorizer.On("IsSystem", ctx).Return(false)

	res, err := eas.DeleteEmail(ctx)
	require.EqualError(t, err, "access denied")
	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestDeleteEmail_NotFound(t *testing.T) {
	ctx := context.Background()
	mockManager := mocks.NewEmailManager(t)
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{manager: mockManager, authorizer: mockAuthorizer}

	mockAuthorizer.On("IsSystem", ctx).Return(true)
	mockManager.On("DeleteEmail", ctx).Return(email.ErrNotFound)

	res, err := eas.DeleteEmail(ctx)
	require.EqualError(t, err, "email not found")
	require.Equal(t, http.StatusNotFound, res.Code)
	require.Nil(t, res.Body)
}

func TestDeleteEmail_Failed(t *testing.T) {
	ctx := context.Background()
	mockManager := mocks.NewEmailManager(t)
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{manager: mockManager, authorizer: mockAuthorizer}

	mockAuthorizer.On("IsSystem", ctx).Return(true)
	mockManager.On("DeleteEmail", ctx).Return(errors.New("error"))

	res, err := eas.DeleteEmail(ctx)
	require.EqualError(t, err, "failed to delete email")
	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.Nil(t, res.Body)
}

func TestPatchEmail_Succeed(t *testing.T) {
	ctx := context.Background()
	mockManager := mocks.NewEmailManager(t)
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{manager: mockManager, authorizer: mockAuthorizer}

	apiPatch := api.EmailChannelForPatch{
		Config: api.EmailChannelConfigForPatch{
			Username: "user",
		},
	}

	expectedPatch := emailChannelForPatchFromAPI(apiPatch)

	mockAuthorizer.On("IsSystem", ctx).Return(true)
	mockManager.On("PatchEmail", ctx, mock.MatchedBy(func(e *email.ChannelConfig) bool {
		return e.Username == expectedPatch.Username
	})).Return(expectedPatch, nil)

	res, err := eas.PatchEmail(ctx, apiPatch)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, emailChannelConfigToAPI(expectedPatch), res.Body)
}

func TestPatchEmail_Forbidden(t *testing.T) {
	ctx := context.Background()
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{authorizer: mockAuthorizer}

	mockAuthorizer.On("IsSystem", ctx).Return(false)

	res, err := eas.PatchEmail(ctx, api.EmailChannelForPatch{})
	require.EqualError(t, err, "access denied")
	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestPatchEmail_NotFound(t *testing.T) {
	ctx := context.Background()
	mockManager := mocks.NewEmailManager(t)
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{manager: mockManager, authorizer: mockAuthorizer}

	apiPatch := api.EmailChannelForPatch{
		Config: api.EmailChannelConfigForPatch{
			Username: "user",
		},
	}

	mockAuthorizer.On("IsSystem", ctx).Return(true)
	mockManager.On("PatchEmail", ctx, mock.Anything).Return(nil, email.ErrNotFound)

	res, err := eas.PatchEmail(ctx, apiPatch)
	require.EqualError(t, err, "email not found")
	require.Equal(t, http.StatusNotFound, res.Code)
	require.Nil(t, res.Body)
}

func TestPatchEmail_Failed(t *testing.T) {
	ctx := context.Background()
	mockManager := mocks.NewEmailManager(t)
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{manager: mockManager, authorizer: mockAuthorizer}

	apiPatch := api.EmailChannelForPatch{
		Config: api.EmailChannelConfigForPatch{
			Username: "user",
		},
	}

	mockAuthorizer.On("IsSystem", ctx).Return(true)
	mockManager.On("PatchEmail", ctx, mock.Anything).Return(nil, errors.New("error"))

	res, err := eas.PatchEmail(ctx, apiPatch)
	require.EqualError(t, err, "failed to patch email")
	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.Nil(t, res.Body)
}

func TestEmailChannelFromAPI(t *testing.T) {
	useAttachments := true
	apiChannel := api.EmailChannel{
		Config: api.EmailChannelConfig{
			Username:            "user",
			Password:            "pass",
			SmtpHost:            "host",
			SmtpPort:            587,
			From:                "from",
			AuthType:            api.EMAILAUTHTYPE_AUTH_LOGIN,
			SmtpTlsEnabled:      nil,
			Recipients:          []string{"r1", "r2"},
			FromDisplayName:     "display",
			DirectNotifications: nil,
			UseAttachments:      &useAttachments,
			Subject:             "subject",
		},
	}

	result := emailChannelFromAPI(apiChannel)

	require.NotNil(t, result)
	require.Equal(t, "user", result.Username)
	require.Equal(t, "pass", result.Password)
	require.Equal(t, "host", result.SmtpHost)
	require.Equal(t, int32(587), result.SmtpPort)
	require.Equal(t, "from", result.From)
	require.Equal(t, "auth_login", result.AuthType)
	require.Nil(t, result.SmtpTlsEnabled)
	require.Equal(t, []string{"r1", "r2"}, result.Recipients)
	require.Equal(t, "display", result.FromDisplayName)
	require.Nil(t, result.DirectNotifications)
	require.True(t, *result.UseAttachments)
	require.Equal(t, "subject", result.Subject)
}

func TestEmailChannelForPatchFromAPI(t *testing.T) {
	useAttachments := true
	apiPatch := api.EmailChannelForPatch{
		Config: api.EmailChannelConfigForPatch{
			Username:            "user",
			Password:            "pass",
			SmtpHost:            "host",
			SmtpPort:            587,
			From:                "from",
			AuthType:            api.EMAILAUTHTYPE_AUTH_LOGIN,
			SmtpTlsEnabled:      nil,
			Recipients:          []string{"r1", "r2"},
			FromDisplayName:     "display",
			DirectNotifications: nil,
			UseAttachments:      &useAttachments,
			Subject:             "subject",
		},
	}

	result := emailChannelForPatchFromAPI(apiPatch)

	require.NotNil(t, result)
	require.Equal(t, "user", result.Username)
	require.Equal(t, "pass", result.Password)
	require.Equal(t, "host", result.SmtpHost)
	require.Equal(t, int32(587), result.SmtpPort)
	require.Equal(t, "from", result.From)
	require.Equal(t, "auth_login", result.AuthType)
	require.Nil(t, result.SmtpTlsEnabled)
	require.Equal(t, []string{"r1", "r2"}, result.Recipients)
	require.Equal(t, "display", result.FromDisplayName)
	require.Nil(t, result.DirectNotifications)
	require.True(t, *result.UseAttachments)
	require.Equal(t, "subject", result.Subject)
}

func TestEmailChannelToAPI(t *testing.T) {
	useAttachments := true
	domainChannelConfig := &email.ChannelConfig{
		Username:            "user",
		Password:            "pass",
		SmtpHost:            "host",
		SmtpPort:            587,
		From:                "from",
		AuthType:            "auth_login",
		SmtpTlsEnabled:      nil,
		Recipients:          []string{"r1", "r2"},
		FromDisplayName:     "display",
		DirectNotifications: nil,
		UseAttachments:      &useAttachments,
		Subject:             "subject",
	}

	result := emailChannelConfigToAPI(domainChannelConfig)

	require.Equal(t, "user", result.Config.Username)
	require.Equal(t, "pass", result.Config.Password)
	require.Equal(t, "host", result.Config.SmtpHost)
	require.Equal(t, int32(587), result.Config.SmtpPort)
	require.Equal(t, "from", result.Config.From)
	require.Equal(t, api.EMAILAUTHTYPE_AUTH_LOGIN, result.Config.AuthType)
	require.Nil(t, result.Config.SmtpTlsEnabled)
	require.Equal(t, []string{"r1", "r2"}, result.Config.Recipients)
	require.Equal(t, "display", result.Config.FromDisplayName)
	require.Nil(t, result.Config.DirectNotifications)
	require.True(t, *result.Config.UseAttachments)
	require.Equal(t, "subject", result.Config.Subject)
}

func TestEmailChannelConfigToAPI_Nil(t *testing.T) {
	result := emailChannelConfigToAPI(nil)
	require.Equal(t, api.EmailChannel{}, result)
}

func TestEmailChannelConfigToAPI(t *testing.T) {
	useAttachments := true
	domainConfig := email.ChannelConfig{
		Username:            "user",
		Password:            "pass",
		SmtpHost:            "host",
		SmtpPort:            587,
		From:                "from",
		AuthType:            "auth_login",
		SmtpTlsEnabled:      nil,
		Recipients:          []string{"r1", "r2"},
		FromDisplayName:     "display",
		DirectNotifications: nil,
		UseAttachments:      &useAttachments,
		Subject:             "subject",
	}

	result := emailChannelConfigToAPI(&domainConfig)

	require.Equal(t, "user", result.Config.Username)
	require.Equal(t, "pass", result.Config.Password)
	require.Equal(t, "host", result.Config.SmtpHost)
	require.Equal(t, int32(587), result.Config.SmtpPort)
	require.Equal(t, "from", result.Config.From)
	require.Equal(t, api.EMAILAUTHTYPE_AUTH_LOGIN, result.Config.AuthType)
	require.Nil(t, result.Config.SmtpTlsEnabled)
	require.Equal(t, []string{"r1", "r2"}, result.Config.Recipients)
	require.Equal(t, "display", result.Config.FromDisplayName)
	require.Nil(t, result.Config.DirectNotifications)
	require.True(t, *result.Config.UseAttachments)
	require.Equal(t, "subject", result.Config.Subject)
}

func TestEmailChannelConfigFromAPI(t *testing.T) {
	useAttachments := true
	apiConfig := api.EmailChannelConfig{
		Username:            "user",
		Password:            "pass",
		SmtpHost:            "host",
		SmtpPort:            587,
		From:                "from",
		AuthType:            api.EMAILAUTHTYPE_AUTH_LOGIN,
		SmtpTlsEnabled:      nil,
		Recipients:          []string{"r1", "r2"},
		FromDisplayName:     "display",
		DirectNotifications: nil,
		UseAttachments:      &useAttachments,
		Subject:             "subject",
	}

	result := emailChannelConfigFromAPI(apiConfig)

	require.Equal(t, "user", result.Username)
	require.Equal(t, "pass", result.Password)
	require.Equal(t, "host", result.SmtpHost)
	require.Equal(t, int32(587), result.SmtpPort)
	require.Equal(t, "from", result.From)
	require.Equal(t, "auth_login", result.AuthType)
	require.Nil(t, result.SmtpTlsEnabled)
	require.Equal(t, []string{"r1", "r2"}, result.Recipients)
	require.Equal(t, "display", result.FromDisplayName)
	require.Nil(t, result.DirectNotifications)
	require.True(t, *result.UseAttachments)
	require.Equal(t, "subject", result.Subject)
}

func TestValidateEmail(t *testing.T) {
	ctx := context.Background()
	mockManager := mocks.NewEmailManager(t)
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{manager: mockManager, authorizer: mockAuthorizer}

	apiConfig := api.EmailChannelConfig{
		Username:   "user",
		SmtpHost:   "smtp.example.com",
		SmtpPort:   587,
		AuthType:   api.EMAILAUTHTYPE_AUTH_LOGIN,
		Recipients: []string{"me"},
	}
	domainConfig := emailChannelConfigFromAPI(apiConfig)

	mockAuthorizer.On("IsSystem", ctx).Return(true)
	mockManager.On("ValidateEmail", ctx, mock.MatchedBy(func(e *email.ChannelConfig) bool {
		return e.Username == domainConfig.Username && e.SmtpHost == domainConfig.SmtpHost
	})).Return(email.ValidationResponse{Status: false}, nil)

	res, err := eas.ValidateEmail(ctx, apiConfig)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	body := res.Body.(api.ValidateSinkResponse)
	require.False(t, body.Status)
}

func TestValidateEmail_Forbidden(t *testing.T) {
	ctx := context.Background()
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{authorizer: mockAuthorizer}

	mockAuthorizer.On("IsSystem", ctx).Return(false)

	res, err := eas.ValidateEmail(ctx, api.EmailChannelConfig{})
	require.EqualError(t, err, "access denied")
	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestValidateEmail_Failed(t *testing.T) {
	ctx := context.Background()
	mockManager := mocks.NewEmailManager(t)
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{manager: mockManager, authorizer: mockAuthorizer}

	apiConfig := api.EmailChannelConfig{
		Username: "user",
	}

	mockAuthorizer.On("IsSystem", ctx).Return(true)
	mockManager.On("ValidateEmail", ctx, mock.Anything).Return(email.ValidationResponse{Status: false}, email.ErrClient)

	res, err := eas.ValidateEmail(ctx, apiConfig)
	require.EqualError(t, err, "email validation failed")
	require.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestValidateEmail_ValidationFailed(t *testing.T) {
	ctx := context.Background()
	mockManager := mocks.NewEmailManager(t)
	mockAuthorizer := mocks.NewAuthorizer(t)
	eas := &EmailAPIService{manager: mockManager, authorizer: mockAuthorizer}

	apiConfig := api.EmailChannelConfig{
		Username: "user",
		SmtpHost: "smtp.example.com",
	}
	domainConfig := emailChannelConfigFromAPI(apiConfig)

	mockAuthorizer.On("IsSystem", ctx).Return(true)
	mockManager.On("ValidateEmail", ctx, mock.MatchedBy(func(e *email.ChannelConfig) bool {
		return e.Username == domainConfig.Username && e.SmtpHost == domainConfig.SmtpHost
	})).Return(email.ValidationResponse{Status: false, Message: "bad credentials"}, errors.Join(email.ErrValidationFailed, errors.New("bad credentials")))

	res, err := eas.ValidateEmail(ctx, apiConfig)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	body := res.Body.(api.ValidateSinkResponse)
	require.False(t, body.Status)
	require.Equal(t, "bad credentials", body.Message)
}

func TestGetEmailNotificationState_MissingTenantId(t *testing.T) {
	ctx := context.Background()
	stateMock := mocks.NewEmailNotificationStateManager(t)
	eas := &EmailAPIService{stateManager: stateMock}

	res, err := eas.GetEmailNotificationState(ctx)
	require.EqualError(t, err, "missing tenant id in request context")
	require.NotNil(t, res)
	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestGetEmailNotificationState_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, emailStateTestTenantId)
	stateMock := mocks.NewEmailNotificationStateManager(t)
	eas := &EmailAPIService{stateManager: stateMock}
	emailState := &state.EmailState{Enabled: ptr.To(true)}
	stateMock.On("GetEmailNotificationState", ctx, emailStateTestTenantId).Return(emailState, nil)

	res, err := eas.GetEmailNotificationState(ctx)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, getAPIState(emailState), res.Body)
}

func TestGetEmailNotificationState_SuccessWhenErrFailedToRemoveSinkFromUserRoutes(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, emailStateTestTenantId)
	stateMock := mocks.NewEmailNotificationStateManager(t)
	eas := &EmailAPIService{stateManager: stateMock}
	emailState := &state.EmailState{Enabled: ptr.To(false)}
	stateMock.On("GetEmailNotificationState", ctx, emailStateTestTenantId).Return(emailState, notification_state.ErrFailedToRemoveSinkFromUserRoutes)

	res, err := eas.GetEmailNotificationState(ctx)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, getAPIState(emailState), res.Body)
}

func TestGetEmailNotificationState_ErrNoSinkConfigured(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, emailStateTestTenantId)
	stateMock := mocks.NewEmailNotificationStateManager(t)
	eas := &EmailAPIService{stateManager: stateMock}
	stateMock.On("GetEmailNotificationState", ctx, emailStateTestTenantId).Return(nil, notification_state.ErrNoSinkConfigured)

	res, err := eas.GetEmailNotificationState(ctx)
	require.EqualError(t, err, "email app not configured to update state")
	require.NotNil(t, res)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestGetEmailNotificationState_InternalError(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, emailStateTestTenantId)
	stateMock := mocks.NewEmailNotificationStateManager(t)
	eas := &EmailAPIService{stateManager: stateMock}
	stateMock.On("GetEmailNotificationState", ctx, emailStateTestTenantId).Return(nil, errors.New("backend error"))

	res, err := eas.GetEmailNotificationState(ctx)
	require.EqualError(t, err, "failed to get email state")
	require.NotNil(t, res)
	require.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestUpdateEmailNotificationState_MissingTenantId(t *testing.T) {
	ctx := context.Background()
	stateMock := mocks.NewEmailNotificationStateManager(t)
	eas := &EmailAPIService{stateManager: stateMock}

	res, err := eas.UpdateEmailNotificationState(ctx, api.NotificationStateStatus{Enabled: ptr.To(true)})
	require.EqualError(t, err, "missing tenant id in request context")
	require.NotNil(t, res)
	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestUpdateEmailNotificationState_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, emailStateTestTenantId)
	stateMock := mocks.NewEmailNotificationStateManager(t)
	eas := &EmailAPIService{stateManager: stateMock}
	updatedState := &state.EmailState{Enabled: ptr.To(true)}
	stateMock.On("UpdateEmailNotificationState", ctx, emailStateTestTenantId, mock.MatchedBy(func(s notification_state.Stater) bool {
		es, ok := s.(*state.EmailState)
		return ok && es != nil && es.Enabled != nil && *es.Enabled
	})).Return(updatedState, nil)

	res, err := eas.UpdateEmailNotificationState(ctx, api.NotificationStateStatus{Enabled: ptr.To(true)})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, getAPIState(updatedState), res.Body)
}

func TestUpdateEmailNotificationState_InternalError(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, emailStateTestTenantId)
	stateMock := mocks.NewEmailNotificationStateManager(t)
	eas := &EmailAPIService{stateManager: stateMock}
	stateMock.On("UpdateEmailNotificationState", ctx, emailStateTestTenantId, mock.Anything).Return(nil, errors.New("backend error"))

	res, err := eas.UpdateEmailNotificationState(ctx, api.NotificationStateStatus{Enabled: ptr.To(false)})
	require.EqualError(t, err, "failed to update email notifications state")
	require.NotNil(t, res)
	require.Equal(t, http.StatusInternalServerError, res.Code)
}
