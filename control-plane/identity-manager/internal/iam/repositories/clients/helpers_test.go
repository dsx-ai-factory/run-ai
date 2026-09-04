/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"strconv"
	"testing"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/utils"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/utils/pointer"
	"github.com/stretchr/testify/assert"
)

func TestFromClientModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		client         *models.Client
		expectedResult *gocloak.Client
		nilResult      bool
	}{
		{
			name: "successful conversion with service account",
			client: &models.Client{
				ID:           pointer.ToString("client-id-123"),
				Name:         pointer.ToString("test-client"),
				ClientID:     pointer.ToString("test-client-id"),
				Enabled:      pointer.ToBool(true),
				TenantId:     pointer.ToString("tenant-1"),
				RedirectURIs: &[]string{"https://example.com/callback"},
				WebOrigins:   &[]string{"https://example.com"},
				ClientConfig: &models.ClientConfig{
					PublicClient:              pointer.ToBool(false),
					ServiceAccountsEnabled:    pointer.ToBool(true),
					StandardFlowEnabled:       pointer.ToBool(true),
					DirectAccessGrantsEnabled: pointer.ToBool(false),
				},
				Attributes: &models.ClientAttributes{
					IsExternal: pointer.ToBool(false),
					ClientType: pointer.ToString("agent"),
				},
			},
			expectedResult: &gocloak.Client{
				ID:       pointer.ToString("client-id-123"),
				Name:     pointer.ToString("test-client"),
				ClientID: pointer.ToString("test-client-id"),
				Enabled:  pointer.ToBool(true),
			},
			nilResult: false,
		},
		{
			name: "client without name uses clientID",
			client: &models.Client{
				ClientID: pointer.ToString("test-client-id"),
				Enabled:  pointer.ToBool(true),
				TenantId: pointer.ToString("tenant-1"),
				ClientConfig: &models.ClientConfig{
					PublicClient:              pointer.ToBool(false),
					ServiceAccountsEnabled:    pointer.ToBool(true),
					StandardFlowEnabled:       pointer.ToBool(true),
					DirectAccessGrantsEnabled: pointer.ToBool(false),
				},
				Attributes: &models.ClientAttributes{
					IsExternal: pointer.ToBool(false),
					ClientType: pointer.ToString("agent"),
				},
			},
			expectedResult: &gocloak.Client{
				Name:     pointer.ToString("test-client-id"),
				ClientID: pointer.ToString("test-client-id"),
				Enabled:  pointer.ToBool(true),
			},
			nilResult: false,
		},
		{
			name: "client without client type returns nil",
			client: &models.Client{
				ClientID: pointer.ToString("test-client-id"),
				Enabled:  pointer.ToBool(true),
				TenantId: pointer.ToString("tenant-1"),
				ClientConfig: &models.ClientConfig{
					PublicClient:              pointer.ToBool(false),
					ServiceAccountsEnabled:    pointer.ToBool(true),
					StandardFlowEnabled:       pointer.ToBool(true),
					DirectAccessGrantsEnabled: pointer.ToBool(false),
				},
				Attributes: &models.ClientAttributes{
					IsExternal: pointer.ToBool(false),
				},
			},
			expectedResult: nil,
			nilResult:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := fromClientModel(tt.client)

			if tt.nilResult {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.NotNil(t, result.Attributes)
				assert.NotNil(t, result.ProtocolMappers)
				if tt.expectedResult != nil {
					assert.Equal(t, tt.expectedResult.ID, result.ID)
					assert.Equal(t, tt.expectedResult.ClientID, result.ClientID)
					assert.Equal(t, tt.expectedResult.Enabled, result.Enabled)
				}
			}
		})
	}
}

func TestToClientModel(t *testing.T) {
	t.Parallel()

	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now().Add(-1 * time.Hour)
	lastLogin := time.Now().Add(-30 * time.Minute)

	tests := []struct {
		name           string
		kcClient       *gocloak.Client
		expectedClient *models.Client
	}{
		{
			name: "successful conversion with all attributes",
			kcClient: &gocloak.Client{
				ID:                        pointer.ToString("client-id-123"),
				Name:                      pointer.ToString("test-client"),
				ClientID:                  pointer.ToString("test-client-id"),
				Enabled:                   pointer.ToBool(true),
				Secret:                    pointer.ToString("secret-value"),
				RedirectURIs:              &[]string{"https://example.com/callback"},
				WebOrigins:                &[]string{"https://example.com"},
				PublicClient:              pointer.ToBool(false),
				ServiceAccountsEnabled:    pointer.ToBool(true),
				StandardFlowEnabled:       pointer.ToBool(true),
				DirectAccessGrantsEnabled: pointer.ToBool(false),
				Attributes: &map[string]string{
					utils.ClientTypeAttr: "agent",
					utils.IsExternal:     "false",
					utils.CreatedBy:      "user@example.com",
					utils.CreationTime:   strconv.FormatInt(createdAt.Unix(), 10),
					utils.UpdateTime:     strconv.FormatInt(updatedAt.Unix(), 10),
					utils.LastLoginTime:  strconv.FormatInt(lastLogin.Unix(), 10),
				},
			},
			expectedClient: &models.Client{
				ID:           pointer.ToString("client-id-123"),
				Name:         pointer.ToString("test-client"),
				ClientID:     pointer.ToString("test-client-id"),
				Enabled:      pointer.ToBool(true),
				Secret:       pointer.ToString("secret-value"),
				RedirectURIs: &[]string{"https://example.com/callback"},
				WebOrigins:   &[]string{"https://example.com"},
				ClientConfig: &models.ClientConfig{
					PublicClient:              pointer.ToBool(false),
					ServiceAccountsEnabled:    pointer.ToBool(true),
					StandardFlowEnabled:       pointer.ToBool(true),
					DirectAccessGrantsEnabled: pointer.ToBool(false),
				},
				Attributes: &models.ClientAttributes{
					ClientType: pointer.ToString("agent"),
					IsExternal: pointer.ToBool(false),
					CreatedBy:  pointer.ToString("user@example.com"),
				},
			},
		},
		{
			name: "conversion with minimal attributes",
			kcClient: &gocloak.Client{
				ID:       pointer.ToString("client-id-456"),
				ClientID: pointer.ToString("minimal-client"),
				Attributes: &map[string]string{
					utils.ClientTypeAttr: "cli",
					utils.IsExternal:     "true",
				},
			},
			expectedClient: &models.Client{
				ID:           pointer.ToString("client-id-456"),
				ClientID:     pointer.ToString("minimal-client"),
				ClientConfig: &models.ClientConfig{},
				Attributes: &models.ClientAttributes{
					ClientType: pointer.ToString("cli"),
					IsExternal: pointer.ToBool(true),
				},
			},
		},
		{
			name: "conversion with invalid isExternal value",
			kcClient: &gocloak.Client{
				ID:       pointer.ToString("client-id-789"),
				ClientID: pointer.ToString("invalid-client"),
				Attributes: &map[string]string{
					utils.ClientTypeAttr: "test",
					utils.IsExternal:     "not-a-boolean",
				},
			},
			expectedClient: &models.Client{
				ID:           pointer.ToString("client-id-789"),
				ClientID:     pointer.ToString("invalid-client"),
				ClientConfig: &models.ClientConfig{},
				Attributes: &models.ClientAttributes{
					ClientType: pointer.ToString("test"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := toClientModel(tt.kcClient)

			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedClient.ID, result.ID)
			assert.Equal(t, tt.expectedClient.ClientID, result.ClientID)
			assert.Equal(t, tt.expectedClient.Name, result.Name)
			assert.Equal(t, tt.expectedClient.Enabled, result.Enabled)
			assert.Equal(t, tt.expectedClient.Secret, result.Secret)
			assert.Equal(t, tt.expectedClient.Attributes.ClientType, result.Attributes.ClientType)
			assert.Equal(t, tt.expectedClient.Attributes.IsExternal, result.Attributes.IsExternal)

			if tt.expectedClient.Attributes.CreatedBy != nil {
				assert.Equal(t, tt.expectedClient.Attributes.CreatedBy, result.Attributes.CreatedBy)
			}
		})
	}
}

func TestToKcFilters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		filter         *models.ClientFilter
		expectedSearch *bool
		checkClientID  bool
		expectedMax    *int
	}{
		{
			name:           "nil filter",
			filter:         nil,
			expectedSearch: pointer.ToBool(false),
			checkClientID:  false,
		},
		{
			name: "filter with clientID",
			filter: &models.ClientFilter{
				ClientID: pointer.ToString("test-client"),
			},
			expectedSearch: pointer.ToBool(true),
			checkClientID:  true,
		},
		{
			name: "filter with max results",
			filter: &models.ClientFilter{
				MaxResults: func() *int { i := 10; return &i }(),
			},
			expectedSearch: pointer.ToBool(true),
			checkClientID:  false,
			expectedMax:    func() *int { i := 10; return &i }(),
		},
		{
			name: "filter with isExternal",
			filter: &models.ClientFilter{
				IsExternal: pointer.ToBool(true),
			},
			expectedSearch: pointer.ToBool(true),
			checkClientID:  false,
		},
		{
			name: "filter with clientType",
			filter: &models.ClientFilter{
				ClientType: pointer.ToString("agent"),
			},
			expectedSearch: pointer.ToBool(true),
			checkClientID:  false,
		},
		{
			name: "filter with clusterId",
			filter: &models.ClientFilter{
				ClusterId: pointer.ToString("cluster-123"),
			},
			expectedSearch: pointer.ToBool(true),
			checkClientID:  false,
		},
		{
			name: "filter with multiple fields",
			filter: &models.ClientFilter{
				ClientID:   pointer.ToString("test-client"),
				MaxResults: func() *int { i := 5; return &i }(),
				IsExternal: pointer.ToBool(false),
				ClientType: pointer.ToString("agent"),
			},
			expectedSearch: pointer.ToBool(true),
			checkClientID:  true,
			expectedMax:    func() *int { i := 5; return &i }(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := toKcFilters(tt.filter)

			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedSearch, result.Search)

			if tt.checkClientID {
				assert.Equal(t, tt.filter.ClientID, result.ClientID)
			}

			if tt.expectedMax != nil {
				assert.Equal(t, tt.expectedMax, result.Max)
			}
		})
	}
}

func TestSetIamClientAttribute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		kcClient *gocloak.Client
		key      string
		value    string
	}{
		{
			name: "set attribute on client with existing attributes",
			kcClient: &gocloak.Client{
				Attributes: &map[string]string{
					"existing-key": "existing-value",
				},
			},
			key:   "new-key",
			value: "new-value",
		},
		{
			name:     "set attribute on client with nil attributes",
			kcClient: &gocloak.Client{},
			key:      "test-key",
			value:    "test-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			setIamClientAttribute(tt.kcClient, tt.key, tt.value)

			assert.NotNil(t, tt.kcClient.Attributes)
			assert.Equal(t, tt.value, (*tt.kcClient.Attributes)[tt.key])
		})
	}
}

func TestAddKcFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		initialChain      string
		filterKey         string
		filterValue       string
		expectedSubstring string
	}{
		{
			name:              "add filter to empty chain",
			initialChain:      "",
			filterKey:         "key1",
			filterValue:       "value1",
			expectedSubstring: "key1:value1",
		},
		{
			name:              "add filter to existing chain",
			initialChain:      " key1:value1",
			filterKey:         "key2",
			filterValue:       "value2",
			expectedSubstring: "key2:value2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			chain := tt.initialChain
			addKcFilter(&chain, tt.filterKey, tt.filterValue)

			assert.Contains(t, chain, tt.expectedSubstring)
		})
	}
}

func TestSetKcDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		client   *models.Client
		validate func(*testing.T, *gocloak.Client)
	}{
		{
			name: "confidential client sets authenticator type",
			client: &models.Client{
				ClientID: pointer.ToString("test-client"),
				TenantId: pointer.ToString("tenant-1"),
				ClientConfig: &models.ClientConfig{
					PublicClient: pointer.ToBool(false),
				},
				Attributes: &models.ClientAttributes{
					IsExternal: pointer.ToBool(false),
					ClientType: pointer.ToString("agent"),
				},
			},
			validate: func(t *testing.T, kc *gocloak.Client) {
				assert.Equal(t, utils.AuthTypeClientSecret, *kc.ClientAuthenticatorType)
				assert.Equal(t, utils.AuthProtocolOidc, *kc.Protocol)
				assert.Equal(t, false, *kc.FullScopeAllowed)
				assert.Equal(t, false, *kc.FrontChannelLogout)
			},
		},
		{
			name: "public client does not set authenticator type",
			client: &models.Client{
				ClientID: pointer.ToString("test-client"),
				TenantId: pointer.ToString("tenant-1"),
				ClientConfig: &models.ClientConfig{
					PublicClient: pointer.ToBool(true),
				},
				Attributes: &models.ClientAttributes{
					IsExternal: pointer.ToBool(true),
					ClientType: pointer.ToString("cli"),
				},
			},
			validate: func(t *testing.T, kc *gocloak.Client) {
				assert.Nil(t, kc.ClientAuthenticatorType)
				assert.Equal(t, utils.AuthProtocolOidc, *kc.Protocol)
				assert.Equal(t, false, *kc.FullScopeAllowed)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kcClient := &gocloak.Client{
				Attributes: &map[string]string{},
			}

			setKcDefaults(kcClient, tt.client)

			tt.validate(t, kcClient)
		})
	}
}

func TestSetProtocolMappers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		client           *models.Client
		minMapperCount   int
		expectedClaims   []string
		unexpectedClaims []string
	}{
		{
			name: "service account enabled user app",
			client: &models.Client{
				ClientID: pointer.ToString("test-client"),
				TenantId: pointer.ToString("tenant-1"),
				ClientConfig: &models.ClientConfig{
					ServiceAccountsEnabled: pointer.ToBool(true),
				},
				Attributes: &models.ClientAttributes{
					ClientType: pointer.ToString(models.ClientTypeUserApp),
					CreatedBy:  pointer.ToString("user@example.com"),
				},
			},
			minMapperCount: 5,
			expectedClaims: []string{"groups", "uid", "gid", "supplementarygroups"},
		},
		{
			name: "service account enabled non-user app",
			client: &models.Client{
				ClientID: pointer.ToString("test-client"),
				TenantId: pointer.ToString("tenant-1"),
				ClientConfig: &models.ClientConfig{
					ServiceAccountsEnabled: pointer.ToBool(true),
				},
				Attributes: &models.ClientAttributes{
					ClientType: pointer.ToString("agent"),
				},
			},
			minMapperCount:   5,
			unexpectedClaims: []string{"groups", "uid", "gid", "supplementarygroups"},
		},
		{
			name: "service account disabled",
			client: &models.Client{
				ClientID: pointer.ToString("test-client"),
				TenantId: pointer.ToString("tenant-1"),
				ClientConfig: &models.ClientConfig{
					ServiceAccountsEnabled: pointer.ToBool(false),
				},
				Attributes: &models.ClientAttributes{
					ClientType: pointer.ToString("cli"),
				},
			},
			minMapperCount: 5,
			expectedClaims: []string{"uid", "gid", "supplementarygroups"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kcClient := &gocloak.Client{
				ProtocolMappers: &[]gocloak.ProtocolMapperRepresentation{},
			}

			setProtocolMappers(kcClient, tt.client)

			assert.NotNil(t, kcClient.ProtocolMappers)
			assert.GreaterOrEqual(t, len(*kcClient.ProtocolMappers), tt.minMapperCount)

			claimNames := make(map[string]struct{})
			for _, mapper := range *kcClient.ProtocolMappers {
				if mapper.Config == nil {
					continue
				}
				if claim, ok := (*mapper.Config)["claim.name"]; ok {
					claimNames[claim] = struct{}{}
				}
			}
			for _, claim := range tt.expectedClaims {
				_, ok := claimNames[claim]
				assert.True(t, ok, "expected claim %q to be mapped", claim)
			}
			for _, claim := range tt.unexpectedClaims {
				_, ok := claimNames[claim]
				assert.False(t, ok, "did not expect claim %q to be mapped", claim)
			}
		})
	}
}

func TestSetAttributes(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name     string
		client   *models.Client
		validate func(*testing.T, *gocloak.Client)
	}{
		{
			name: "all attributes set",
			client: &models.Client{
				Attributes: &models.ClientAttributes{
					ClientType: pointer.ToString("agent"),
					IsExternal: pointer.ToBool(true),
					CreatedBy:  pointer.ToString("user@example.com"),
					CreatedAt:  pointer.ToString(now.Format(time.RFC3339)),
					UpdatedAt:  pointer.ToString(now.Format(time.RFC3339)),
					LastLogin:  pointer.ToString(now.Format(time.RFC3339)),
				},
			},
			validate: func(t *testing.T, kc *gocloak.Client) {
				assert.NotNil(t, kc.Attributes)
				attrs := *kc.Attributes
				assert.Equal(t, "agent", attrs[utils.ClientTypeAttr])
				assert.Equal(t, "true", attrs[utils.IsExternal])
				assert.Equal(t, "user@example.com", attrs[utils.CreatedBy])
				assert.NotEmpty(t, attrs[utils.CreationTime])
				assert.NotEmpty(t, attrs[utils.UpdateTime])
				assert.NotEmpty(t, attrs[utils.LastLoginTime])
			},
		},
		{
			name: "minimal attributes set",
			client: &models.Client{
				Attributes: &models.ClientAttributes{
					ClientType: pointer.ToString("cli"),
				},
			},
			validate: func(t *testing.T, kc *gocloak.Client) {
				assert.NotNil(t, kc.Attributes)
				attrs := *kc.Attributes
				assert.Equal(t, "cli", attrs[utils.ClientTypeAttr])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kcClient := &gocloak.Client{
				Attributes: &map[string]string{},
			}

			setAttributes(kcClient, tt.client)

			tt.validate(t, kcClient)
		})
	}
}
