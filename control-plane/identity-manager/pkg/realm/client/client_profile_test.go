/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"

	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
)

func TestApplyClientProfile(t *testing.T) {
	tests := []struct {
		name                    string
		clientType              string
		expectedConfig          *iamModels.ClientConfig
		expectedExtern          *bool
		expectedEmailScopeDefault bool
	}{
		{
			name:       "RunaiAdminUI",
			clientType: ClientTypeRunaiAdminUI,
			expectedConfig: &iamModels.ClientConfig{
				PublicClient:              ptr.To(false),
				DirectAccessGrantsEnabled: ptr.To(false),
				StandardFlowEnabled:       ptr.To(true),
				ServiceAccountsEnabled:    ptr.To(false),
			},
			expectedEmailScopeDefault: true,
		},
		{
			name:       "RunaiCLI",
			clientType: ClientTypeRunaiCLI,
			expectedConfig: &iamModels.ClientConfig{
				PublicClient:              ptr.To(true),
				DirectAccessGrantsEnabled: ptr.To(true),
				StandardFlowEnabled:       ptr.To(true),
				ServiceAccountsEnabled:    ptr.To(false),
			},
			expectedEmailScopeDefault: true,
		},
		{
			name:       "CLI",
			clientType: ClientTypeCLI,
			expectedConfig: &iamModels.ClientConfig{
				PublicClient:              ptr.To(false),
				DirectAccessGrantsEnabled: ptr.To(true),
				StandardFlowEnabled:       ptr.To(true),
				ServiceAccountsEnabled:    ptr.To(false),
			},
			expectedEmailScopeDefault: true,
		},
		{
			name:       "WorkloadsAuth",
			clientType: ClientTypeWorkloadsAuth,
			expectedConfig: &iamModels.ClientConfig{
				PublicClient:              ptr.To(false),
				DirectAccessGrantsEnabled: ptr.To(false),
				StandardFlowEnabled:       ptr.To(true),
				ServiceAccountsEnabled:    ptr.To(false),
			},
			expectedEmailScopeDefault: true,
		},
		{
			name:       "Runai",
			clientType: ClientTypeRunai,
			expectedConfig: &iamModels.ClientConfig{
				PublicClient:              ptr.To(true),
				DirectAccessGrantsEnabled: ptr.To(true),
				StandardFlowEnabled:       ptr.To(false),
				ServiceAccountsEnabled:    ptr.To(false),
			},
			expectedEmailScopeDefault: true,
		},
		{
			name:       "TokenExchange",
			clientType: ClientTypeTokenExchange,
			expectedConfig: &iamModels.ClientConfig{
				PublicClient:              ptr.To(false),
				DirectAccessGrantsEnabled: ptr.To(false),
				StandardFlowEnabled:       ptr.To(true),
				ServiceAccountsEnabled:    ptr.To(false),
			},
			expectedEmailScopeDefault: false,
		},
		{
			name:       "Agent",
			clientType: ClientTypeAgent,
			expectedConfig: &iamModels.ClientConfig{
				PublicClient:              ptr.To(false),
				DirectAccessGrantsEnabled: ptr.To(false),
				StandardFlowEnabled:       ptr.To(true),
				ServiceAccountsEnabled:    ptr.To(true),
			},
			expectedExtern:            ptr.To(false),
			expectedEmailScopeDefault: false,
		},
		{
			name:       "Installer",
			clientType: ClientTypeInstaller,
			expectedConfig: &iamModels.ClientConfig{
				PublicClient:              ptr.To(false),
				DirectAccessGrantsEnabled: ptr.To(false),
				StandardFlowEnabled:       ptr.To(true),
				ServiceAccountsEnabled:    ptr.To(true),
			},
			expectedExtern:            ptr.To(false),
			expectedEmailScopeDefault: false,
		},
		{
			name:       "AccessKey",
			clientType: ClientTypeUserApp,
			expectedConfig: &iamModels.ClientConfig{
				PublicClient:              ptr.To(false),
				DirectAccessGrantsEnabled: ptr.To(false),
				StandardFlowEnabled:       ptr.To(true),
				ServiceAccountsEnabled:    ptr.To(true),
			},
			expectedEmailScopeDefault: false,
		},
		{
			name:       "TenantApp",
			clientType: ClientTypeTenantApp,
			expectedConfig: &iamModels.ClientConfig{
				PublicClient:              ptr.To(false),
				DirectAccessGrantsEnabled: ptr.To(false),
				StandardFlowEnabled:       ptr.To(true),
				ServiceAccountsEnabled:    ptr.To(true),
			},
			expectedExtern:            ptr.To(true),
			expectedEmailScopeDefault: false,
		},
		{
			name:       "MCP",
			clientType: ClientTypeMCP,
			expectedConfig: &iamModels.ClientConfig{
				PublicClient:              ptr.To(false),
				DirectAccessGrantsEnabled: ptr.To(false),
				StandardFlowEnabled:       ptr.To(true),
				ServiceAccountsEnabled:    ptr.To(false),
			},
			expectedEmailScopeDefault: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := &iamModels.Client{}

			ApplyClientProfile(client, tt.clientType)

			assert.Equal(t, tt.clientType, *client.Attributes.ClientType)
			assert.Equal(t, tt.expectedConfig, client.ClientConfig)
			if tt.expectedExtern == nil {
				assert.Nil(t, client.Attributes.IsExternal)
			} else {
				assert.Equal(t, *tt.expectedExtern, *client.Attributes.IsExternal)
			}
			assert.Equal(t, tt.expectedEmailScopeDefault, ProfileEmailScopeDefault(tt.clientType))
		})
	}
}

func TestApplyClientProfile_TokenLifespan(t *testing.T) {
	t.Run("service-account types get the default lifespan", func(t *testing.T) {
		t.Parallel()
		for _, clientType := range []string{ClientTypeUserApp, ClientTypeTenantApp} {
			client := &iamModels.Client{}
			ApplyClientProfile(client, clientType)
			assert.NotNil(t, client.Attributes.TokenLifespan, clientType)
			assert.Equal(t, DefaultServiceAccountTokenLifespan, *client.Attributes.TokenLifespan, clientType)
		}
	})

	t.Run("types without a profile lifespan stay nil", func(t *testing.T) {
		t.Parallel()
		client := &iamModels.Client{}
		ApplyClientProfile(client, ClientTypeRunaiAdminUI)
		assert.Nil(t, client.Attributes.TokenLifespan)
	})

	t.Run("profile lifespan is authoritative (overrides any existing value)", func(t *testing.T) {
		t.Parallel()
		custom := 5 * time.Minute
		client := &iamModels.Client{Attributes: &iamModels.ClientAttributes{TokenLifespan: &custom}}
		ApplyClientProfile(client, ClientTypeUserApp)
		assert.Equal(t, DefaultServiceAccountTokenLifespan, *client.Attributes.TokenLifespan)
	})
}

func TestApplyClientProfile_ExtraAttributes(t *testing.T) {
	t.Parallel()

	t.Run("type with extra attributes gets them", func(t *testing.T) {
		client := &iamModels.Client{}
		ApplyClientProfile(client, ClientTypeWorkloadsAuth)
		assert.Equal(t, workloadsAuthAttributes, client.Attributes.Extra)
	})

	t.Run("type without extra attributes leaves Extra nil", func(t *testing.T) {
		for _, clientType := range []string{ClientTypeUserApp, ClientTypeRunai} {
			client := &iamModels.Client{}
			ApplyClientProfile(client, clientType)
			assert.Nil(t, client.Attributes.Extra, clientType)
		}
	})
}

func TestApplyClientProfile_UnknownType(t *testing.T) {
	t.Parallel()
	client := &iamModels.Client{}

	ApplyClientProfile(client, "some-unknown-type")

	assert.Equal(t, "some-unknown-type", *client.Attributes.ClientType)
	assert.Nil(t, client.ClientConfig)
	assert.Nil(t, client.Attributes.IsExternal)
}
