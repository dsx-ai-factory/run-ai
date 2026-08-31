/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth_flow

import (
	"context"
	"testing"

	"github.com/Nerzal/gocloak/v13"
	authMocks "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/auth/mocks"
	kcMocks "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/gocloak/mocks"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	helpers "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/repositories/auth_flow/helpers"
	"github.com/stretchr/testify/assert"
)

const (
	realm = "runai"
)

func initTest(t *testing.T) (authMock *authMocks.Auth, goCloakMock *kcMocks.GoCloak, testSubject *KeycloakAuthFlowRepository) {
	goCloak := kcMocks.NewGoCloak(t)
	auth := authMocks.NewAuth(t)
	testSubject = &KeycloakAuthFlowRepository{
		auth:     auth,
		kcClient: goCloak,
	}
	return auth, goCloak, testSubject
}

func TestCreateAuthFlow_HappyFlow(t *testing.T) {
	authMock, gocloakMock, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	flow := models.AuthenticationFlowRepresentation{
		Alias:       gocloak.StringP("Alias"),
		Description: gocloak.StringP("Description"),
		ID:          gocloak.StringP("ID"),
		ProviderID:  gocloak.StringP("ProviderID"),
		BuiltIn:     gocloak.BoolP(true),
		TopLevel:    gocloak.BoolP(true),
	}
	kcFlow := helpers.FromAuthFlowRepModel(&flow)
	authMock.On("GetAdminToken", context.Background()).Return(jwt, nil)
	gocloakMock.On("CreateAuthenticationFlow", context.Background(), *jwt, realm, *kcFlow).Return(nil)
	err := testSubject.CreateAuthFlow(context.Background(), realm, flow)
	assert.Nil(t, err)
}

func TestCreateAuthFlow_AdminTokenError(t *testing.T) {
	authMock, _, testSubject := initTest(t)
	flow := models.AuthenticationFlowRepresentation{
		Alias:       gocloak.StringP("Alias"),
		Description: gocloak.StringP("Description"),
		ID:          gocloak.StringP("ID"),
		ProviderID:  gocloak.StringP("ProviderID"),
		BuiltIn:     gocloak.BoolP(true),
		TopLevel:    gocloak.BoolP(true),
	}
	authMock.On("GetAdminToken", context.Background()).Return(nil, assert.AnError)
	err := testSubject.CreateAuthFlow(context.Background(), realm, flow)
	assert.NotNil(t, err)
}

func TestCreateAuthFlow_CreateAuthenticationFlow_KcError(t *testing.T) {
	authMock, gocloakMock, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	flow := models.AuthenticationFlowRepresentation{
		Alias:       gocloak.StringP("Alias"),
		Description: gocloak.StringP("Description"),
		ID:          gocloak.StringP("ID"),
		ProviderID:  gocloak.StringP("ProviderID"),
		BuiltIn:     gocloak.BoolP(true),
		TopLevel:    gocloak.BoolP(true),
	}
	kcFlow := helpers.FromAuthFlowRepModel(&flow)
	authMock.On("GetAdminToken", context.Background()).Return(jwt, nil)
	gocloakMock.On("CreateAuthenticationFlow", context.Background(), *jwt, realm, *kcFlow).Return(assert.AnError)
	err := testSubject.CreateAuthFlow(context.Background(), realm, flow)
	assert.NotNil(t, err)
}

func TestGetAuthFlow_HappyFlow(t *testing.T) {
	authMock, gocloakMock, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	flow := models.AuthenticationFlowRepresentation{
		Alias:       gocloak.StringP("Alias"),
		Description: gocloak.StringP("Description"),
		ID:          gocloak.StringP("ID"),
		ProviderID:  gocloak.StringP("ProviderID"),
		BuiltIn:     gocloak.BoolP(true),
		TopLevel:    gocloak.BoolP(true),
	}

	var authFlows []*gocloak.AuthenticationFlowRepresentation
	authFlows = append(authFlows, helpers.FromAuthFlowRepModel(&flow))

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("GetAuthenticationFlows", ctx, *jwt, realm).Return(authFlows, nil)
	resp, err := testSubject.GetAuthFlow(ctx, realm, "Alias")
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestGetAuthFlow_EmptyArray(t *testing.T) {
	authMock, gocloakMock, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()

	var authFlows []*gocloak.AuthenticationFlowRepresentation

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("GetAuthenticationFlows", ctx, *jwt, realm).Return(authFlows, nil)
	resp, err := testSubject.GetAuthFlow(ctx, realm, "Alias")
	assert.Nil(t, err)
	assert.Nil(t, resp)
}

func TestGetAuthFlow_GetAdminToken_Error(t *testing.T) {
	authMock, _, testSubject := initTest(t)

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	resp, err := testSubject.GetAuthFlow(ctx, realm, "Alias")
	assert.NotNil(t, err)
	assert.Nil(t, resp)
}

func TestGetAuthFlow_KcError(t *testing.T) {
	authMock, gocloakMock, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("GetAuthenticationFlows", ctx, *jwt, realm).Return(nil, assert.AnError)
	resp, err := testSubject.GetAuthFlow(ctx, realm, "Alias")
	assert.NotNil(t, err)
	assert.Nil(t, resp)
}

func TestGetAuthExecutions_HappyFlow(t *testing.T) {
	authMock, gocloakMock, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	flow := models.ModifyAuthenticationExecutionRepresentation{
		Alias:       gocloak.StringP("Alias"),
		Description: gocloak.StringP("Description"),
		ID:          gocloak.StringP("ID"),
		ProviderID:  gocloak.StringP("ProviderID"),
	}

	var authFlows []*gocloak.ModifyAuthenticationExecutionRepresentation
	authFlows = append(authFlows, helpers.FromModifyAutheExecutionRepModel(&flow))

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("GetAuthenticationExecutions", ctx, *jwt, realm, "Alias").Return(authFlows, nil)
	resp, err := testSubject.GetAuthExecutions(ctx, realm, "Alias")
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestGetAuthExecutions_AdminTokenError(t *testing.T) {
	authMock, _, testSubject := initTest(t)

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	resp, err := testSubject.GetAuthExecutions(ctx, realm, "Alias")
	assert.NotNil(t, err)
	assert.Nil(t, resp)
}

func TestGetAuthExecutions_KcError(t *testing.T) {
	authMock, gocloakMock, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("GetAuthenticationExecutions", ctx, *jwt, realm, "Alias").Return(nil, assert.AnError)
	resp, err := testSubject.GetAuthExecutions(ctx, realm, "Alias")
	assert.NotNil(t, err)
	assert.Nil(t, resp)
}

func TestCreateAuthExecution_HappyFlow(t *testing.T) {
	authMock, gocloakMock, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("CreateAuthenticationExecution", ctx, *jwt, realm, "Alias", gocloak.CreateAuthenticationExecutionRepresentation{}).Return(nil)
	err := testSubject.CreateAuthExecution(ctx, realm, "Alias", models.CreateAuthenticationExecutionRepresentation{})
	assert.Nil(t, err)
}

func TestCreateAuthExecution_AdminTokenError(t *testing.T) {
	authMock, _, testSubject := initTest(t)

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	err := testSubject.CreateAuthExecution(ctx, realm, "Alias", models.CreateAuthenticationExecutionRepresentation{})
	assert.NotNil(t, err)
}

func TestCreateAuthExecution_KcError(t *testing.T) {
	authMock, gocloakMock, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("CreateAuthenticationExecution", ctx, *jwt, realm, "Alias", gocloak.CreateAuthenticationExecutionRepresentation{}).Return(assert.AnError)
	err := testSubject.CreateAuthExecution(ctx, realm, "Alias", models.CreateAuthenticationExecutionRepresentation{})
	assert.NotNil(t, err)
}

func TestUpdateAuthExecution_HappyFlow(t *testing.T) {
	authMock, gocloakMock, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("UpdateAuthenticationExecution", ctx, *jwt, realm, "Alias", gocloak.ModifyAuthenticationExecutionRepresentation{}).Return(nil)
	err := testSubject.UpdateAuthExecution(ctx, realm, "Alias", models.ModifyAuthenticationExecutionRepresentation{})
	assert.Nil(t, err)
}

func TestUpdateAuthExecution_AdminTokenError(t *testing.T) {
	authMock, _, testSubject := initTest(t)

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	err := testSubject.UpdateAuthExecution(ctx, realm, "Alias", models.ModifyAuthenticationExecutionRepresentation{})
	assert.NotNil(t, err)
}

func TestUpdateAuthExecution_KcError(t *testing.T) {
	authMock, gocloakMock, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("UpdateAuthenticationExecution", ctx, *jwt, realm, "Alias", gocloak.ModifyAuthenticationExecutionRepresentation{}).Return(assert.AnError)
	err := testSubject.UpdateAuthExecution(ctx, realm, "Alias", models.ModifyAuthenticationExecutionRepresentation{})
	assert.NotNil(t, err)
}

func TestCreateAuthExecutionFlow_HappyFlow(t *testing.T) {
	authMock, gocloakMock, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("CreateAuthenticationExecutionFlow", ctx, *jwt, realm, "Alias", gocloak.CreateAuthenticationExecutionFlowRepresentation{}).Return(nil)
	err := testSubject.CreateAuthExecutionFlow(ctx, realm, "Alias", models.CreateAuthenticationExecutionFlowRepresentation{})
	assert.Nil(t, err)
}

func TestCreateAuthExecutionFlow_AdminTokenError(t *testing.T) {
	authMock, _, testSubject := initTest(t)

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	err := testSubject.CreateAuthExecutionFlow(ctx, realm, "Alias", models.CreateAuthenticationExecutionFlowRepresentation{})
	assert.NotNil(t, err)
}

func TestCreateAuthExecutionFlow_KcError(t *testing.T) {
	authMock, gocloakMock, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("CreateAuthenticationExecutionFlow", ctx, *jwt, realm, "Alias", gocloak.CreateAuthenticationExecutionFlowRepresentation{}).Return(assert.AnError)
	err := testSubject.CreateAuthExecutionFlow(ctx, realm, "Alias", models.CreateAuthenticationExecutionFlowRepresentation{})
	assert.NotNil(t, err)
}
