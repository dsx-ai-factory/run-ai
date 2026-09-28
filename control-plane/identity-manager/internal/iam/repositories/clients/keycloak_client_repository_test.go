/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/Nerzal/gocloak/v13"
	authmocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/auth/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/errors"
	gocloakmocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/gocloak/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	runaicloakmocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/runaicloak/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/utils"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/utils/pointer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewKeycloakClientRepository(t *testing.T) {
	t.Parallel()

	mockAuth := authmocks.NewAuth(t)
	mockGoCloakClient := gocloakmocks.NewGoCloak(t)
	mockRunaiCloak := runaicloakmocks.NewRunAICloakClient(t)

	repo := &KeycloakClientRepository{
		auth:             mockAuth,
		kcClient:         mockGoCloakClient,
		runaicloakClient: mockRunaiCloak,
	}

	assert.NotNil(t, repo)
	assert.Equal(t, mockAuth, repo.auth)
	assert.Equal(t, mockGoCloakClient, repo.kcClient)
	assert.Equal(t, mockRunaiCloak, repo.runaicloakClient)
}

func TestKeycloakClientRepository_CreateClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		realm         string
		client        *models.Client
		mockSetup     func(*authmocks.Auth, *gocloakmocks.GoCloak)
		expectedID    string
		expectedError bool
	}{
		{
			name:  "successful client creation",
			realm: "test-realm",
			client: &models.Client{
				Name:     pointer.ToString("test-client"),
				ClientID: pointer.ToString("test-client"),
				TenantId: pointer.ToString("tenant-1"),
				Enabled:  pointer.ToBool(true),
				ClientConfig: &models.ClientConfig{
					PublicClient:              pointer.ToBool(false),
					ServiceAccountsEnabled:    pointer.ToBool(true),
					StandardFlowEnabled:       pointer.ToBool(true),
					DirectAccessGrantsEnabled: pointer.ToBool(false),
				},
				Attributes: &models.ClientAttributes{
					IsExternal: pointer.ToBool(false),
					ClientType: pointer.ToString("test-type"),
				},
			},
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcMock.On("CreateClient", mock.Anything, token, "test-realm", mock.AnythingOfType("gocloak.Client")).
					Return("client-id-123", nil)
			},
			expectedID:    "client-id-123",
			expectedError: false,
		},
		{
			name:  "auth token retrieval fails",
			realm: "test-realm",
			client: &models.Client{
				Name:     pointer.ToString("test-client"),
				ClientID: pointer.ToString("test-client"),
				TenantId: pointer.ToString("tenant-1"),
				Enabled:  pointer.ToBool(true),
				ClientConfig: &models.ClientConfig{
					PublicClient:              pointer.ToBool(false),
					ServiceAccountsEnabled:    pointer.ToBool(true),
					StandardFlowEnabled:       pointer.ToBool(true),
					DirectAccessGrantsEnabled: pointer.ToBool(false),
				},
				Attributes: &models.ClientAttributes{
					IsExternal: pointer.ToBool(false),
					ClientType: pointer.ToString("test-type"),
				},
			},
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				authMock.On("GetAdminToken", mock.Anything).Return(nil, fmt.Errorf("auth error"))
			},
			expectedID:    "",
			expectedError: true,
		},
		{
			name:  "keycloak create client fails",
			realm: "test-realm",
			client: &models.Client{
				Name:     pointer.ToString("test-client"),
				ClientID: pointer.ToString("test-client"),
				TenantId: pointer.ToString("tenant-1"),
				Enabled:  pointer.ToBool(true),
				ClientConfig: &models.ClientConfig{
					PublicClient:              pointer.ToBool(false),
					ServiceAccountsEnabled:    pointer.ToBool(true),
					StandardFlowEnabled:       pointer.ToBool(true),
					DirectAccessGrantsEnabled: pointer.ToBool(false),
				},
				Attributes: &models.ClientAttributes{
					IsExternal: pointer.ToBool(false),
					ClientType: pointer.ToString("test-type"),
				},
			},
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcMock.On("CreateClient", mock.Anything, token, "test-realm", mock.AnythingOfType("gocloak.Client")).
					Return("", &gocloak.APIError{Code: http.StatusConflict, Message: "client already exists"})
			},
			expectedID:    "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockAuth := authmocks.NewAuth(t)
			mockGoCloakClient := gocloakmocks.NewGoCloak(t)
			mockRunaiCloak := runaicloakmocks.NewRunAICloakClient(t)

			tt.mockSetup(mockAuth, mockGoCloakClient)

			repo := &KeycloakClientRepository{
				auth:             mockAuth,
				kcClient:         mockGoCloakClient,
				runaicloakClient: mockRunaiCloak,
			}

			id, err := repo.CreateClient(context.Background(), tt.realm, tt.client)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
				assert.NotNil(t, tt.client.Attributes.CreatedAt)
				assert.NotNil(t, tt.client.Attributes.UpdatedAt)
			}
		})
	}
}

func TestKeycloakClientRepository_UpdateClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		realm         string
		client        *models.Client
		mockSetup     func(*authmocks.Auth, *gocloakmocks.GoCloak)
		expectedError bool
	}{
		{
			name:  "successful client update",
			realm: "test-realm",
			client: &models.Client{
				ID:       pointer.ToString("client-id-123"),
				Name:     pointer.ToString("test-client"),
				ClientID: pointer.ToString("test-client"),
				TenantId: pointer.ToString("tenant-1"),
				Enabled:  pointer.ToBool(true),
				ClientConfig: &models.ClientConfig{
					PublicClient:              pointer.ToBool(false),
					ServiceAccountsEnabled:    pointer.ToBool(true),
					StandardFlowEnabled:       pointer.ToBool(true),
					DirectAccessGrantsEnabled: pointer.ToBool(false),
				},
				Attributes: &models.ClientAttributes{
					IsExternal: pointer.ToBool(false),
					ClientType: pointer.ToString("test-type"),
				},
			},
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcMock.On("UpdateClient", mock.Anything, token, "test-realm", mock.AnythingOfType("gocloak.Client")).
					Return(nil)
			},
			expectedError: false,
		},
		{
			name:  "auth token retrieval fails",
			realm: "test-realm",
			client: &models.Client{
				ID:       pointer.ToString("client-id-123"),
				Name:     pointer.ToString("test-client"),
				ClientID: pointer.ToString("test-client"),
				TenantId: pointer.ToString("tenant-1"),
				Enabled:  pointer.ToBool(true),
				ClientConfig: &models.ClientConfig{
					PublicClient:              pointer.ToBool(false),
					ServiceAccountsEnabled:    pointer.ToBool(true),
					StandardFlowEnabled:       pointer.ToBool(true),
					DirectAccessGrantsEnabled: pointer.ToBool(false),
				},
				Attributes: &models.ClientAttributes{
					IsExternal: pointer.ToBool(false),
					ClientType: pointer.ToString("test-type"),
				},
			},
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				authMock.On("GetAdminToken", mock.Anything).Return(nil, fmt.Errorf("auth error"))
			},
			expectedError: true,
		},
		{
			name:  "keycloak update client fails",
			realm: "test-realm",
			client: &models.Client{
				ID:       pointer.ToString("client-id-123"),
				Name:     pointer.ToString("test-client"),
				ClientID: pointer.ToString("test-client"),
				TenantId: pointer.ToString("tenant-1"),
				Enabled:  pointer.ToBool(true),
				ClientConfig: &models.ClientConfig{
					PublicClient:              pointer.ToBool(false),
					ServiceAccountsEnabled:    pointer.ToBool(true),
					StandardFlowEnabled:       pointer.ToBool(true),
					DirectAccessGrantsEnabled: pointer.ToBool(false),
				},
				Attributes: &models.ClientAttributes{
					IsExternal: pointer.ToBool(false),
					ClientType: pointer.ToString("test-type"),
				},
			},
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcMock.On("UpdateClient", mock.Anything, token, "test-realm", mock.AnythingOfType("gocloak.Client")).
					Return(&gocloak.APIError{Code: http.StatusNotFound, Message: "client not found"})
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockAuth := authmocks.NewAuth(t)
			mockGoCloakClient := gocloakmocks.NewGoCloak(t)
			mockRunaiCloak := runaicloakmocks.NewRunAICloakClient(t)

			tt.mockSetup(mockAuth, mockGoCloakClient)

			repo := &KeycloakClientRepository{
				auth:             mockAuth,
				kcClient:         mockGoCloakClient,
				runaicloakClient: mockRunaiCloak,
			}

			err := repo.UpdateClient(context.Background(), tt.realm, tt.client)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tt.client.Attributes.UpdatedAt)
			}
		})
	}
}

func TestKeycloakClientRepository_GetClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		realm          string
		clientID       string
		mockSetup      func(*authmocks.Auth, *gocloakmocks.GoCloak)
		expectedClient *models.Client
		expectedError  bool
	}{
		{
			name:     "successful client retrieval",
			realm:    "test-realm",
			clientID: "client-id-123",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcClient := &gocloak.Client{
					ID:       pointer.ToString("client-id-123"),
					ClientID: pointer.ToString("test-client"),
					Name:     pointer.ToString("test-client"),
					Enabled:  pointer.ToBool(true),
					Attributes: &map[string]string{
						"runai.client_type": "test-type",
						"runai.is_external": "false",
					},
				}
				kcMock.On("GetClient", mock.Anything, token, "test-realm", "client-id-123").
					Return(kcClient, nil)
			},
			expectedClient: &models.Client{
				ID:           pointer.ToString("client-id-123"),
				ClientID:     pointer.ToString("test-client"),
				Name:         pointer.ToString("test-client"),
				Enabled:      pointer.ToBool(true),
				ClientConfig: &models.ClientConfig{},
				Attributes: &models.ClientAttributes{
					ClientType: pointer.ToString("test-type"),
					IsExternal: pointer.ToBool(false),
				},
			},
			expectedError: false,
		},
		{
			name:     "auth token retrieval fails",
			realm:    "test-realm",
			clientID: "client-id-123",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				authMock.On("GetAdminToken", mock.Anything).Return(nil, fmt.Errorf("auth error"))
			},
			expectedClient: nil,
			expectedError:  true,
		},
		{
			name:     "keycloak get client fails",
			realm:    "test-realm",
			clientID: "client-id-123",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcMock.On("GetClient", mock.Anything, token, "test-realm", "client-id-123").
					Return(nil, &gocloak.APIError{Code: http.StatusNotFound, Message: "client not found"})
			},
			expectedClient: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockAuth := authmocks.NewAuth(t)
			mockGoCloakClient := gocloakmocks.NewGoCloak(t)
			mockRunaiCloak := runaicloakmocks.NewRunAICloakClient(t)

			tt.mockSetup(mockAuth, mockGoCloakClient)

			repo := &KeycloakClientRepository{
				auth:             mockAuth,
				kcClient:         mockGoCloakClient,
				runaicloakClient: mockRunaiCloak,
			}

			client, err := repo.GetClient(context.Background(), tt.realm, tt.clientID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.expectedClient.ID, client.ID)
				assert.Equal(t, tt.expectedClient.ClientID, client.ClientID)
				assert.Equal(t, tt.expectedClient.Name, client.Name)
				assert.Equal(t, tt.expectedClient.Enabled, client.Enabled)
				assert.NotNil(t, client.Attributes)
				assert.Equal(t, tt.expectedClient.Attributes.ClientType, client.Attributes.ClientType)
				assert.Equal(t, tt.expectedClient.Attributes.IsExternal, client.Attributes.IsExternal)
			}
		})
	}
}

func TestKeycloakClientRepository_DeleteClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		realm         string
		clientID      string
		mockSetup     func(*authmocks.Auth, *gocloakmocks.GoCloak)
		expectedError bool
	}{
		{
			name:     "successful client deletion",
			realm:    "test-realm",
			clientID: "client-id-123",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcMock.On("DeleteClient", mock.Anything, token, "test-realm", "client-id-123").
					Return(nil)
			},
			expectedError: false,
		},
		{
			name:     "auth token retrieval fails",
			realm:    "test-realm",
			clientID: "client-id-123",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				authMock.On("GetAdminToken", mock.Anything).Return(nil, fmt.Errorf("auth error"))
			},
			expectedError: true,
		},
		{
			name:     "keycloak delete client fails",
			realm:    "test-realm",
			clientID: "client-id-123",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcMock.On("DeleteClient", mock.Anything, token, "test-realm", "client-id-123").
					Return(&gocloak.APIError{Code: http.StatusNotFound, Message: "client not found"})
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockAuth := authmocks.NewAuth(t)
			mockGoCloakClient := gocloakmocks.NewGoCloak(t)
			mockRunaiCloak := runaicloakmocks.NewRunAICloakClient(t)

			tt.mockSetup(mockAuth, mockGoCloakClient)

			repo := &KeycloakClientRepository{
				auth:             mockAuth,
				kcClient:         mockGoCloakClient,
				runaicloakClient: mockRunaiCloak,
			}

			err := repo.DeleteClient(context.Background(), tt.realm, tt.clientID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestKeycloakClientRepository_GetClients(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		realm           string
		filter          *models.ClientFilter
		mockSetup       func(*authmocks.Auth, *gocloakmocks.GoCloak)
		expectedClients []models.Client
		expectedError   bool
	}{
		{
			name:  "successful clients retrieval",
			realm: "test-realm",
			filter: &models.ClientFilter{
				ClientID: pointer.ToString("test-client"),
			},
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcClients := []*gocloak.Client{
					{
						ID:       pointer.ToString("client-id-123"),
						ClientID: pointer.ToString("test-client"),
						Name:     pointer.ToString("test-client"),
						Enabled:  pointer.ToBool(true),
						Attributes: &map[string]string{
							"runai.client_type": "test-type",
							"runai.is_external": "false",
						},
					},
				}
				kcMock.On("GetClients", mock.Anything, token, "test-realm", mock.AnythingOfType("gocloak.GetClientsParams")).
					Return(kcClients, nil)
			},
			expectedClients: []models.Client{
				{
					ID:           pointer.ToString("client-id-123"),
					ClientID:     pointer.ToString("test-client"),
					Name:         pointer.ToString("test-client"),
					Enabled:      pointer.ToBool(true),
					ClientConfig: &models.ClientConfig{},
					Attributes: &models.ClientAttributes{
						ClientType: pointer.ToString("test-type"),
						IsExternal: pointer.ToBool(false),
					},
				},
			},
			expectedError: false,
		},
		{
			name:   "auth token retrieval fails",
			realm:  "test-realm",
			filter: nil,
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				authMock.On("GetAdminToken", mock.Anything).Return(nil, fmt.Errorf("auth error"))
			},
			expectedClients: nil,
			expectedError:   true,
		},
		{
			name:   "keycloak get clients fails",
			realm:  "test-realm",
			filter: nil,
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcMock.On("GetClients", mock.Anything, token, "test-realm", mock.AnythingOfType("gocloak.GetClientsParams")).
					Return(nil, &gocloak.APIError{Code: http.StatusInternalServerError, Message: "internal error"})
			},
			expectedClients: nil,
			expectedError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockAuth := authmocks.NewAuth(t)
			mockGoCloakClient := gocloakmocks.NewGoCloak(t)
			mockRunaiCloak := runaicloakmocks.NewRunAICloakClient(t)

			tt.mockSetup(mockAuth, mockGoCloakClient)

			repo := &KeycloakClientRepository{
				auth:             mockAuth,
				kcClient:         mockGoCloakClient,
				runaicloakClient: mockRunaiCloak,
			}

			clients, err := repo.GetClients(context.Background(), tt.realm, tt.filter)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, clients)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, clients)
				assert.Len(t, clients, len(tt.expectedClients))
				for i := range tt.expectedClients {
					assert.NotNil(t, clients[i].Attributes)
					assert.Equal(t, tt.expectedClients[i].Attributes.ClientType, clients[i].Attributes.ClientType)
					assert.Equal(t, tt.expectedClients[i].Attributes.IsExternal, clients[i].Attributes.IsExternal)
				}
			}
		})
	}
}

func TestKeycloakClientRepository_GetClientSecret(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		realm          string
		clientID       string
		mockSetup      func(*authmocks.Auth, *gocloakmocks.GoCloak)
		expectedSecret string
		expectedError  bool
	}{
		{
			name:     "successful secret retrieval",
			realm:    "test-realm",
			clientID: "client-id-123",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				credentialData := &gocloak.CredentialRepresentation{
					Value: pointer.ToString("super-secret"),
				}
				kcMock.On("GetClientSecret", mock.Anything, token, "test-realm", "client-id-123").
					Return(credentialData, nil)
			},
			expectedSecret: "super-secret",
			expectedError:  false,
		},
		{
			name:     "auth token retrieval fails",
			realm:    "test-realm",
			clientID: "client-id-123",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				authMock.On("GetAdminToken", mock.Anything).Return(nil, fmt.Errorf("auth error"))
			},
			expectedSecret: "",
			expectedError:  true,
		},
		{
			name:     "keycloak get client secret fails",
			realm:    "test-realm",
			clientID: "client-id-123",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcMock.On("GetClientSecret", mock.Anything, token, "test-realm", "client-id-123").
					Return(nil, &gocloak.APIError{Code: http.StatusNotFound, Message: "client not found"})
			},
			expectedSecret: "",
			expectedError:  true,
		},
		{
			name:     "client has no secret credential yet",
			realm:    "test-realm",
			clientID: "client-id-123",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				credentialData := &gocloak.CredentialRepresentation{
					Value: nil,
				}
				kcMock.On("GetClientSecret", mock.Anything, token, "test-realm", "client-id-123").
					Return(credentialData, nil)
			},
			expectedSecret: "",
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockAuth := authmocks.NewAuth(t)
			mockGoCloakClient := gocloakmocks.NewGoCloak(t)
			mockRunaiCloak := runaicloakmocks.NewRunAICloakClient(t)

			tt.mockSetup(mockAuth, mockGoCloakClient)

			repo := &KeycloakClientRepository{
				auth:             mockAuth,
				kcClient:         mockGoCloakClient,
				runaicloakClient: mockRunaiCloak,
			}

			secret, err := repo.GetClientSecret(context.Background(), tt.realm, tt.clientID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSecret, secret)
			}
		})
	}
}

func TestKeycloakClientRepository_GetClientByClientId(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		realm          string
		clientID       string
		mockSetup      func(*authmocks.Auth, *gocloakmocks.GoCloak)
		expectedClient *models.Client
		expectedError  bool
	}{
		{
			name:     "successful client retrieval by client id",
			realm:    "test-realm",
			clientID: "test-client",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcClients := []*gocloak.Client{
					{
						ID:       pointer.ToString("client-id-123"),
						ClientID: pointer.ToString("test-client"),
						Name:     pointer.ToString("test-client"),
						Enabled:  pointer.ToBool(true),
						Attributes: &map[string]string{
							"runai.client_type": "test-type",
							"runai.is_external": "false",
						},
					},
				}
				kcMock.On("GetClients", mock.Anything, token, "test-realm", mock.AnythingOfType("gocloak.GetClientsParams")).
					Return(kcClients, nil)
			},
			expectedClient: &models.Client{
				ID:       pointer.ToString("client-id-123"),
				ClientID: pointer.ToString("test-client"),
				Name:     pointer.ToString("test-client"),
				Enabled:  pointer.ToBool(true),
			},
			expectedError: false,
		},
		{
			name:     "no client found",
			realm:    "test-realm",
			clientID: "non-existent-client",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcClients := []*gocloak.Client{}
				kcMock.On("GetClients", mock.Anything, token, "test-realm", mock.AnythingOfType("gocloak.GetClientsParams")).
					Return(kcClients, nil)
			},
			expectedClient: nil,
			expectedError:  false,
		},
		{
			name:     "more than one client found",
			realm:    "test-realm",
			clientID: "duplicate-client",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcClients := []*gocloak.Client{
					{
						ID:       pointer.ToString("client-id-123"),
						ClientID: pointer.ToString("duplicate-client"),
						Attributes: &map[string]string{
							"runai.client_type": "test-type",
							"runai.is_external": "false",
						},
					},
					{
						ID:       pointer.ToString("client-id-456"),
						ClientID: pointer.ToString("duplicate-client"),
						Attributes: &map[string]string{
							"runai.client_type": "test-type",
							"runai.is_external": "false",
						},
					},
				}
				kcMock.On("GetClients", mock.Anything, token, "test-realm", mock.AnythingOfType("gocloak.GetClientsParams")).
					Return(kcClients, nil)
			},
			expectedClient: nil,
			expectedError:  true,
		},
		{
			name:     "auth token retrieval fails",
			realm:    "test-realm",
			clientID: "test-client",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				authMock.On("GetAdminToken", mock.Anything).Return(nil, fmt.Errorf("auth error"))
			},
			expectedClient: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockAuth := authmocks.NewAuth(t)
			mockGoCloakClient := gocloakmocks.NewGoCloak(t)
			mockRunaiCloak := runaicloakmocks.NewRunAICloakClient(t)

			tt.mockSetup(mockAuth, mockGoCloakClient)

			repo := &KeycloakClientRepository{
				auth:             mockAuth,
				kcClient:         mockGoCloakClient,
				runaicloakClient: mockRunaiCloak,
			}

			client, err := repo.GetClientByClientId(context.Background(), tt.realm, tt.clientID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expectedClient != nil {
					assert.NotNil(t, client)
					assert.Equal(t, tt.expectedClient.ID, client.ID)
					assert.Equal(t, tt.expectedClient.ClientID, client.ClientID)
				} else {
					assert.Nil(t, client)
				}
			}
		})
	}
}

func TestKeycloakClientRepository_GetClientsCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		realm         string
		filter        *models.ClientFilter
		mockSetup     func(*authmocks.Auth, *gocloakmocks.GoCloak)
		expectedCount int
		expectedError bool
	}{
		{
			name:  "successful count retrieval",
			realm: "test-realm",
			filter: &models.ClientFilter{
				IsExternal: pointer.ToBool(false),
			},
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcClients := []*gocloak.Client{
					{ID: pointer.ToString("client-1"), Attributes: &map[string]string{"runai.client_type": "test-type", "runai.is_external": "false"}},
					{ID: pointer.ToString("client-2"), Attributes: &map[string]string{"runai.client_type": "test-type", "runai.is_external": "false"}},
					{ID: pointer.ToString("client-3"), Attributes: &map[string]string{"runai.client_type": "test-type", "runai.is_external": "false"}},
				}
				kcMock.On("GetClients", mock.Anything, token, "test-realm", mock.AnythingOfType("gocloak.GetClientsParams")).
					Return(kcClients, nil)
			},
			expectedCount: 3,
			expectedError: false,
		},
		{
			name:   "auth token retrieval fails",
			realm:  "test-realm",
			filter: nil,
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				authMock.On("GetAdminToken", mock.Anything).Return(nil, fmt.Errorf("auth error"))
			},
			expectedCount: 0,
			expectedError: true,
		},
		{
			name:   "keycloak get clients fails",
			realm:  "test-realm",
			filter: nil,
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcMock.On("GetClients", mock.Anything, token, "test-realm", mock.AnythingOfType("gocloak.GetClientsParams")).
					Return(nil, &gocloak.APIError{Code: http.StatusInternalServerError, Message: "internal error"})
			},
			expectedCount: 0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockAuth := authmocks.NewAuth(t)
			mockGoCloakClient := gocloakmocks.NewGoCloak(t)
			mockRunaiCloak := runaicloakmocks.NewRunAICloakClient(t)

			tt.mockSetup(mockAuth, mockGoCloakClient)

			repo := &KeycloakClientRepository{
				auth:             mockAuth,
				kcClient:         mockGoCloakClient,
				runaicloakClient: mockRunaiCloak,
			}

			count, err := repo.GetClientsCount(context.Background(), tt.realm, tt.filter)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
			}
		})
	}
}

func TestKeycloakClientRepository_RegenerateClientSecret(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		realm          string
		clientID       string
		mockSetup      func(*authmocks.Auth, *gocloakmocks.GoCloak)
		expectedSecret string
		expectedError  bool
	}{
		{
			name:     "successful secret regeneration",
			realm:    "test-realm",
			clientID: "client-id-123",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				credentialData := &gocloak.CredentialRepresentation{
					Value: pointer.ToString("new-super-secret"),
				}
				kcMock.On("RegenerateClientSecret", mock.Anything, token, "test-realm", "client-id-123").
					Return(credentialData, nil)
			},
			expectedSecret: "new-super-secret",
			expectedError:  false,
		},
		{
			name:     "auth token retrieval fails",
			realm:    "test-realm",
			clientID: "client-id-123",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				authMock.On("GetAdminToken", mock.Anything).Return(nil, fmt.Errorf("auth error"))
			},
			expectedSecret: "",
			expectedError:  true,
		},
		{
			name:     "keycloak regenerate secret fails",
			realm:    "test-realm",
			clientID: "client-id-123",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcMock.On("RegenerateClientSecret", mock.Anything, token, "test-realm", "client-id-123").
					Return(nil, &gocloak.APIError{Code: http.StatusNotFound, Message: "client not found"})
			},
			expectedSecret: "",
			expectedError:  true,
		},
		{
			name:     "credential data is nil",
			realm:    "test-realm",
			clientID: "client-id-123",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcMock.On("RegenerateClientSecret", mock.Anything, token, "test-realm", "client-id-123").
					Return(nil, nil)
			},
			expectedSecret: "",
			expectedError:  true,
		},
		{
			name:     "credential value is nil",
			realm:    "test-realm",
			clientID: "client-id-123",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				credentialData := &gocloak.CredentialRepresentation{
					Value: nil,
				}
				kcMock.On("RegenerateClientSecret", mock.Anything, token, "test-realm", "client-id-123").
					Return(credentialData, nil)
			},
			expectedSecret: "",
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockAuth := authmocks.NewAuth(t)
			mockGoCloakClient := gocloakmocks.NewGoCloak(t)
			mockRunaiCloak := runaicloakmocks.NewRunAICloakClient(t)

			tt.mockSetup(mockAuth, mockGoCloakClient)

			repo := &KeycloakClientRepository{
				auth:             mockAuth,
				kcClient:         mockGoCloakClient,
				runaicloakClient: mockRunaiCloak,
			}

			secret, err := repo.RegenerateClientSecret(context.Background(), tt.realm, tt.clientID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSecret, secret)
			}
		})
	}
}

func TestKeycloakClientRepository_ToggleClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		realm         string
		clientID      string
		enabled       bool
		mockSetup     func(*authmocks.Auth, *gocloakmocks.GoCloak)
		expectedError bool
	}{
		{
			name:     "successful client enable",
			realm:    "test-realm",
			clientID: "client-id-123",
			enabled:  true,
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcClient := &gocloak.Client{
					ID:       pointer.ToString("client-id-123"),
					ClientID: pointer.ToString("test-client"),
					Enabled:  pointer.ToBool(false),
					Attributes: &map[string]string{
						"runai.client_type": "test-type",
						"runai.is_external": "false",
					},
				}
				kcMock.On("GetClient", mock.Anything, token, "test-realm", "client-id-123").
					Return(kcClient, nil)
				kcMock.On("UpdateClient", mock.Anything, token, "test-realm", mock.AnythingOfType("gocloak.Client")).
					Return(nil)
			},
			expectedError: false,
		},
		{
			name:     "successful client disable",
			realm:    "test-realm",
			clientID: "client-id-123",
			enabled:  false,
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcClient := &gocloak.Client{
					ID:       pointer.ToString("client-id-123"),
					ClientID: pointer.ToString("test-client"),
					Enabled:  pointer.ToBool(true),
					Attributes: &map[string]string{
						"runai.client_type": "test-type",
						"runai.is_external": "false",
					},
				}
				kcMock.On("GetClient", mock.Anything, token, "test-realm", "client-id-123").
					Return(kcClient, nil)
				kcMock.On("UpdateClient", mock.Anything, token, "test-realm", mock.AnythingOfType("gocloak.Client")).
					Return(nil)
			},
			expectedError: false,
		},
		{
			name:     "auth token retrieval fails",
			realm:    "test-realm",
			clientID: "client-id-123",
			enabled:  true,
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				authMock.On("GetAdminToken", mock.Anything).Return(nil, fmt.Errorf("auth error"))
			},
			expectedError: true,
		},
		{
			name:     "get client fails",
			realm:    "test-realm",
			clientID: "client-id-123",
			enabled:  true,
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcMock.On("GetClient", mock.Anything, token, "test-realm", "client-id-123").
					Return(nil, &gocloak.APIError{Code: http.StatusNotFound, Message: "client not found"})
			},
			expectedError: true,
		},
		{
			name:     "update client fails",
			realm:    "test-realm",
			clientID: "client-id-123",
			enabled:  true,
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcClient := &gocloak.Client{
					ID:       pointer.ToString("client-id-123"),
					ClientID: pointer.ToString("test-client"),
					Enabled:  pointer.ToBool(false),
					Attributes: &map[string]string{
						"runai.client_type": "test-type",
						"runai.is_external": "false",
					},
				}
				kcMock.On("GetClient", mock.Anything, token, "test-realm", "client-id-123").
					Return(kcClient, nil)
				kcMock.On("UpdateClient", mock.Anything, token, "test-realm", mock.AnythingOfType("gocloak.Client")).
					Return(&gocloak.APIError{Code: http.StatusInternalServerError, Message: "update failed"})
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockAuth := authmocks.NewAuth(t)
			mockGoCloakClient := gocloakmocks.NewGoCloak(t)
			mockRunaiCloak := runaicloakmocks.NewRunAICloakClient(t)

			tt.mockSetup(mockAuth, mockGoCloakClient)

			repo := &KeycloakClientRepository{
				auth:             mockAuth,
				kcClient:         mockGoCloakClient,
				runaicloakClient: mockRunaiCloak,
			}

			err := repo.ToggleClient(context.Background(), tt.realm, tt.clientID, tt.enabled)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestKeycloakClientRepository_UpdateRedirectUris(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		realm         string
		clientID      string
		redirectUris  []string
		mockSetup     func(*authmocks.Auth, *gocloakmocks.GoCloak)
		expectedError bool
	}{
		{
			name:         "successful redirect URIs update",
			realm:        "test-realm",
			clientID:     "client-id-123",
			redirectUris: []string{"https://example.com/callback", "https://example.com/callback2"},
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcClient := &gocloak.Client{
					ID:           pointer.ToString("client-id-123"),
					ClientID:     pointer.ToString("test-client"),
					RedirectURIs: &[]string{"https://old.com/callback"},
					Attributes: &map[string]string{
						"runai.client_type": "test-type",
						"runai.is_external": "false",
					},
				}
				kcMock.On("GetClient", mock.Anything, token, "test-realm", "client-id-123").
					Return(kcClient, nil)
				kcMock.On("UpdateClient", mock.Anything, token, "test-realm", mock.AnythingOfType("gocloak.Client")).
					Return(nil)
			},
			expectedError: false,
		},
		{
			name:         "auth token retrieval fails",
			realm:        "test-realm",
			clientID:     "client-id-123",
			redirectUris: []string{"https://example.com/callback"},
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				authMock.On("GetAdminToken", mock.Anything).Return(nil, fmt.Errorf("auth error"))
			},
			expectedError: true,
		},
		{
			name:         "get client fails",
			realm:        "test-realm",
			clientID:     "client-id-123",
			redirectUris: []string{"https://example.com/callback"},
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcMock.On("GetClient", mock.Anything, token, "test-realm", "client-id-123").
					Return(nil, &gocloak.APIError{Code: http.StatusNotFound, Message: "client not found"})
			},
			expectedError: true,
		},
		{
			name:         "update client fails",
			realm:        "test-realm",
			clientID:     "client-id-123",
			redirectUris: []string{"https://example.com/callback"},
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcClient := &gocloak.Client{
					ID:           pointer.ToString("client-id-123"),
					ClientID:     pointer.ToString("test-client"),
					RedirectURIs: &[]string{"https://old.com/callback"},
					Attributes: &map[string]string{
						"runai.client_type": "test-type",
						"runai.is_external": "false",
					},
				}
				kcMock.On("GetClient", mock.Anything, token, "test-realm", "client-id-123").
					Return(kcClient, nil)
				kcMock.On("UpdateClient", mock.Anything, token, "test-realm", mock.AnythingOfType("gocloak.Client")).
					Return(&gocloak.APIError{Code: http.StatusBadRequest, Message: "invalid redirect URIs"})
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockAuth := authmocks.NewAuth(t)
			mockGoCloakClient := gocloakmocks.NewGoCloak(t)
			mockRunaiCloak := runaicloakmocks.NewRunAICloakClient(t)

			tt.mockSetup(mockAuth, mockGoCloakClient)

			repo := &KeycloakClientRepository{
				auth:             mockAuth,
				kcClient:         mockGoCloakClient,
				runaicloakClient: mockRunaiCloak,
			}

			err := repo.UpdateRedirectUris(context.Background(), tt.realm, tt.clientID, tt.redirectUris)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestKeycloakClientRepository_GetExchangePolicyId(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		realm           string
		adminToken      string
		mngIdOfClient   string
		mockSetup       func(*gocloakmocks.GoCloak)
		expectedID      string
		expectedError   bool
		expectedErrType error
	}{
		{
			name:          "successful policy id retrieval",
			realm:         "test-realm",
			adminToken:    "admin-token",
			mngIdOfClient: "mng-client-id",
			mockSetup: func(kcMock *gocloakmocks.GoCloak) {
				policies := []*gocloak.PolicyRepresentation{
					{
						ID:   pointer.ToString("policy-id-123"),
						Name: pointer.ToString(ExchangePolicyName),
					},
				}
				kcMock.On("GetPolicies", mock.Anything, "admin-token", "test-realm", "mng-client-id", mock.AnythingOfType("gocloak.GetPolicyParams")).
					Return(policies, nil)
			},
			expectedID:    "policy-id-123",
			expectedError: false,
		},
		{
			name:          "no policy found",
			realm:         "test-realm",
			adminToken:    "admin-token",
			mngIdOfClient: "mng-client-id",
			mockSetup: func(kcMock *gocloakmocks.GoCloak) {
				policies := []*gocloak.PolicyRepresentation{}
				kcMock.On("GetPolicies", mock.Anything, "admin-token", "test-realm", "mng-client-id", mock.AnythingOfType("gocloak.GetPolicyParams")).
					Return(policies, nil)
			},
			expectedID:      "",
			expectedError:   true,
			expectedErrType: errors.ErrInternal,
		},
		{
			name:          "more than one policy found",
			realm:         "test-realm",
			adminToken:    "admin-token",
			mngIdOfClient: "mng-client-id",
			mockSetup: func(kcMock *gocloakmocks.GoCloak) {
				policies := []*gocloak.PolicyRepresentation{
					{
						ID:   pointer.ToString("policy-id-123"),
						Name: pointer.ToString(ExchangePolicyName),
					},
					{
						ID:   pointer.ToString("policy-id-456"),
						Name: pointer.ToString(ExchangePolicyName),
					},
				}
				kcMock.On("GetPolicies", mock.Anything, "admin-token", "test-realm", "mng-client-id", mock.AnythingOfType("gocloak.GetPolicyParams")).
					Return(policies, nil)
			},
			expectedID:      "",
			expectedError:   true,
			expectedErrType: errors.ErrInternal,
		},
		{
			name:          "keycloak get policies fails",
			realm:         "test-realm",
			adminToken:    "admin-token",
			mngIdOfClient: "mng-client-id",
			mockSetup: func(kcMock *gocloakmocks.GoCloak) {
				kcMock.On("GetPolicies", mock.Anything, "admin-token", "test-realm", "mng-client-id", mock.AnythingOfType("gocloak.GetPolicyParams")).
					Return(nil, fmt.Errorf("keycloak error"))
			},
			expectedID:      "",
			expectedError:   true,
			expectedErrType: errors.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockAuth := authmocks.NewAuth(t)
			mockGoCloakClient := gocloakmocks.NewGoCloak(t)
			mockRunaiCloak := runaicloakmocks.NewRunAICloakClient(t)

			tt.mockSetup(mockGoCloakClient)

			repo := &KeycloakClientRepository{
				auth:             mockAuth,
				kcClient:         mockGoCloakClient,
				runaicloakClient: mockRunaiCloak,
			}

			id, err := repo.getExchangePolicyId(context.Background(), tt.realm, tt.adminToken, tt.mngIdOfClient)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}

func TestKeycloakClientRepository_SetEmailScopeOptional(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		realm         string
		clientID      string
		mockSetup     func(*authmocks.Auth, *gocloakmocks.GoCloak)
		expectedError bool
	}{
		{
			name:     "email scope already optional — no-op",
			realm:    "test-realm",
			clientID: "client-id-123",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				emailScopeID := "email-scope-id"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcMock.On("GetClientsOptionalScopes", mock.Anything, token, "test-realm", "client-id-123").
					Return([]*gocloak.ClientScope{{ID: &emailScopeID, Name: pointer.ToString(utils.EmailScope)}}, nil)
			},
			expectedError: false,
		},
		{
			name:     "email scope is default — moved to optional",
			realm:    "test-realm",
			clientID: "client-id-123",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				emailScopeID := "email-scope-id"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcMock.On("GetClientsOptionalScopes", mock.Anything, token, "test-realm", "client-id-123").
					Return([]*gocloak.ClientScope{}, nil)
				kcMock.On("GetClientScopes", mock.Anything, token, "test-realm").
					Return([]*gocloak.ClientScope{{ID: &emailScopeID, Name: pointer.ToString(utils.EmailScope)}}, nil)
				kcMock.On("RemoveDefaultScopeFromClient", mock.Anything, token, "test-realm", "client-id-123", emailScopeID).
					Return(nil)
				kcMock.On("AddOptionalScopeToClient", mock.Anything, token, "test-realm", "client-id-123", emailScopeID).
					Return(nil)
			},
			expectedError: false,
		},
		{
			name:     "auth token retrieval fails",
			realm:    "test-realm",
			clientID: "client-id-123",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				authMock.On("GetAdminToken", mock.Anything).Return(nil, fmt.Errorf("auth error"))
			},
			expectedError: true,
		},
		{
			name:     "email scope not found in realm scopes",
			realm:    "test-realm",
			clientID: "client-id-123",
			mockSetup: func(authMock *authmocks.Auth, kcMock *gocloakmocks.GoCloak) {
				token := "admin-token"
				authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
				kcMock.On("GetClientsOptionalScopes", mock.Anything, token, "test-realm", "client-id-123").
					Return([]*gocloak.ClientScope{}, nil)
				kcMock.On("GetClientScopes", mock.Anything, token, "test-realm").
					Return([]*gocloak.ClientScope{}, nil)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockAuth := authmocks.NewAuth(t)
			mockGoCloakClient := gocloakmocks.NewGoCloak(t)
			mockRunaiCloak := runaicloakmocks.NewRunAICloakClient(t)

			tt.mockSetup(mockAuth, mockGoCloakClient)

			repo := &KeycloakClientRepository{
				auth:             mockAuth,
				kcClient:         mockGoCloakClient,
				runaicloakClient: mockRunaiCloak,
			}

			err := repo.SetEmailScopeOptional(context.Background(), tt.realm, tt.clientID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
