/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/runaicloak/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
)

func newTestIamServiceContainer(runaicloakClient *mocks.RunAICloakClient) *iam.IamServiceContainer {
	return iam.NewTestIamServiceContainer(runaicloakClient)
}

func TestGetOpenidConfiguration_HappyFlow(t *testing.T) {
	keycloakConfig := map[string]interface{}{
		"issuer":                 "https://example.com/auth/realms/test",
		"authorization_endpoint": "https://example.com/auth/realms/test/protocol/openid-connect/auth",
		"token_endpoint":         "https://example.com/auth/realms/test/protocol/openid-connect/token",
		"jwks_uri":               "https://example.com/auth/realms/test/protocol/openid-connect/certs",
		"mtls_endpoint_aliases": map[string]interface{}{
			"token_endpoint": "https://example.com/auth/realms/test/protocol/openid-connect/token",
		},
	}

	tenantConfig := &config.TenantRealm{
		AppHost:       "https://app.run.ai",
		IsMultiTenant: false,
	}

	mockClient := mocks.NewRunAICloakClient(t)
	mockClient.On("FetchKeycloakConfiguration", mock.Anything, "test").Return(keycloakConfig, nil)

	iamServices := newTestIamServiceContainer(mockClient)
	service := NewOpenIDConfigurationService(tenantConfig, iamServices).(*OpenIDConfigurationService)

	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, "test")
	result, err := service.GetOpenidConfiguration(ctx)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, result.Code)

	configMap, ok := result.Body.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "https://app.run.ai/api/v2/token", configMap["token_endpoint"])

	mtlsAliases, ok := configMap["mtls_endpoint_aliases"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "https://app.run.ai/api/v2/token", mtlsAliases["token_endpoint"])

	mockClient.AssertExpectations(t)
}

func TestGetOpenidConfiguration_MultiTenant(t *testing.T) {
	keycloakConfig := map[string]interface{}{
		"issuer":         "https://example.com/auth/realms/tenant1",
		"token_endpoint": "https://example.com/auth/realms/tenant1/protocol/openid-connect/token",
	}

	tenantConfig := &config.TenantRealm{
		AppHost:              "https://app.run.ai",
		TenantDomainTemplate: "https://{{.}}.run.ai",
		IsMultiTenant:        true,
	}

	mockClient := mocks.NewRunAICloakClient(t)
	mockClient.On("FetchKeycloakConfiguration", mock.Anything, "tenant1").Return(keycloakConfig, nil)

	iamServices := newTestIamServiceContainer(mockClient)
	service := NewOpenIDConfigurationService(tenantConfig, iamServices).(*OpenIDConfigurationService)

	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, "tenant1")
	result, err := service.GetOpenidConfiguration(ctx)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, result.Code)

	configMap, ok := result.Body.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "https://tenant1.run.ai/api/v2/token", configMap["token_endpoint"])

	mockClient.AssertExpectations(t)
}

func TestGetOpenidConfiguration_NoRealm_ReturnsBadRequest(t *testing.T) {
	tenantConfig := &config.TenantRealm{AppHost: "https://app.run.ai"}

	mockClient := mocks.NewRunAICloakClient(t)
	iamServices := newTestIamServiceContainer(mockClient)

	service := NewOpenIDConfigurationService(tenantConfig, iamServices).(*OpenIDConfigurationService)

	ctx := context.Background()
	result, err := service.GetOpenidConfiguration(ctx)

	require.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, result.Code)
}

func TestGetOpenidConfiguration_KeycloakError_ReturnsInternalError(t *testing.T) {
	tenantConfig := &config.TenantRealm{AppHost: "https://app.run.ai"}

	mockClient := mocks.NewRunAICloakClient(t)
	mockClient.On("FetchKeycloakConfiguration", mock.Anything, "test").
		Return(nil, assert.AnError)

	iamServices := newTestIamServiceContainer(mockClient)
	service := NewOpenIDConfigurationService(tenantConfig, iamServices).(*OpenIDConfigurationService)

	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, "test")
	result, err := service.GetOpenidConfiguration(ctx)

	require.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, result.Code)

	mockClient.AssertExpectations(t)
}

func TestModifyConfiguration_MultiTenant_RendersTemplate(t *testing.T) {
	tenantConfig := &config.TenantRealm{
		AppHost:              "https://app.example.com",
		TenantDomainTemplate: "https://{{.}}.staging.example.com",
		IsMultiTenant:        true,
	}

	mockClient := mocks.NewRunAICloakClient(t)
	iamServices := newTestIamServiceContainer(mockClient)
	service := NewOpenIDConfigurationService(tenantConfig, iamServices).(*OpenIDConfigurationService)

	modified := service.modifyConfiguration(context.Background(), map[string]interface{}{}, "mytenant")

	assert.Equal(t, "https://mytenant.staging.example.com/api/v2/token", modified["token_endpoint"])
}

func TestModifyConfiguration_ReplacesTokenEndpoint(t *testing.T) {
	tenantConfig := &config.TenantRealm{
		AppHost:       "https://app.run.ai",
		IsMultiTenant: false,
	}

	mockClient := mocks.NewRunAICloakClient(t)
	iamServices := newTestIamServiceContainer(mockClient)

	service := NewOpenIDConfigurationService(tenantConfig, iamServices).(*OpenIDConfigurationService)

	keycloakConfig := map[string]interface{}{
		"token_endpoint": "https://keycloak/auth/realms/test/protocol/openid-connect/token",
		"mtls_endpoint_aliases": map[string]interface{}{
			"token_endpoint": "https://keycloak/auth/realms/test/protocol/openid-connect/token",
			"other_endpoint": "https://keycloak/other",
		},
	}

	modified := service.modifyConfiguration(context.Background(), keycloakConfig, "test")

	assert.Equal(t, "https://app.run.ai/api/v2/token", modified["token_endpoint"])

	mtlsAliases, ok := modified["mtls_endpoint_aliases"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "https://app.run.ai/api/v2/token", mtlsAliases["token_endpoint"])
	assert.Equal(t, "https://keycloak/other", mtlsAliases["other_endpoint"])
}
