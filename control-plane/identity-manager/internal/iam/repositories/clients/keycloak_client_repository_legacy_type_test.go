/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"context"
	"strings"
	"testing"

	"github.com/Nerzal/gocloak/v13"
	authmocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/auth/mocks"
	gocloakmocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/gocloak/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	runaicloakmocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/runaicloak/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/utils/pointer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

const legacyTestToken = "admin-token"

func searchesFor(clientType string) interface{} {
	return mock.MatchedBy(func(p gocloak.GetClientsParams) bool {
		return p.SearchableAttributes != nil && strings.Contains(*p.SearchableAttributes, "runai.client_type:"+clientType)
	})
}

func newLegacyTypeTestRepo(t *testing.T) (*KeycloakClientRepository, *gocloakmocks.GoCloak) {
	t.Helper()
	authMock := authmocks.NewAuth(t)
	token := legacyTestToken
	authMock.On("GetAdminToken", mock.Anything).Return(&token, nil)
	kcMock := gocloakmocks.NewGoCloak(t)
	return &KeycloakClientRepository{
		auth:             authMock,
		kcClient:         kcMock,
		runaicloakClient: runaicloakmocks.NewRunAICloakClient(t),
	}, kcMock
}

func kcClientOfType(id, clientType string) *gocloak.Client {
	return &gocloak.Client{
		ID:       pointer.ToString(id),
		ClientID: pointer.ToString(id),
		Attributes: &map[string]string{
			"runai.client_type": clientType,
			"runai.created.by":  "creator",
		},
	}
}

func TestKeycloakClientRepository_GetClients_ReturnsBothTypeFormats(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		newVal string
		oldVal string
	}{
		{name: "access keys", newVal: models.ClientTypeAccessKey, oldVal: "user-application"},
		{name: "service accounts", newVal: models.ClientTypeServiceAccount, oldVal: "tenant-application"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo, kcMock := newLegacyTypeTestRepo(t)
			kcMock.On("GetClients", mock.Anything, legacyTestToken, "realm", searchesFor(tc.newVal)).
				Return([]*gocloak.Client{kcClientOfType("migrated", tc.newVal)}, nil).Once()
			kcMock.On("GetClients", mock.Anything, legacyTestToken, "realm", searchesFor(tc.oldVal)).
				Return([]*gocloak.Client{kcClientOfType("legacy", tc.oldVal)}, nil).Once()

			clients, err := repo.GetClients(context.Background(), "realm", &models.ClientFilter{
				ClientType:    pointer.ToString(tc.newVal),
				IsTenantAdmin: true,
			})

			require.NoError(t, err)
			require.Len(t, clients, 2)
			ids := []string{*clients[0].ID, *clients[1].ID}
			assert.ElementsMatch(t, []string{"migrated", "legacy"}, ids)
		})
	}
}

func TestKeycloakClientRepository_GetClients_DeduplicatesAcrossFormats(t *testing.T) {
	t.Parallel()
	repo, kcMock := newLegacyTypeTestRepo(t)
	kcMock.On("GetClients", mock.Anything, legacyTestToken, "realm", searchesFor(models.ClientTypeAccessKey)).
		Return([]*gocloak.Client{kcClientOfType("same", models.ClientTypeAccessKey)}, nil).Once()
	kcMock.On("GetClients", mock.Anything, legacyTestToken, "realm", searchesFor("user-application")).
		Return([]*gocloak.Client{kcClientOfType("same", "user-application")}, nil).Once()

	clients, err := repo.GetClients(context.Background(), "realm", &models.ClientFilter{
		ClientType:    pointer.ToString(models.ClientTypeAccessKey),
		IsTenantAdmin: true,
	})

	require.NoError(t, err)
	assert.Len(t, clients, 1)
}

func TestKeycloakClientRepository_GetClients_RespectsMaxResultsAfterMerge(t *testing.T) {
	t.Parallel()
	repo, kcMock := newLegacyTypeTestRepo(t)
	kcMock.On("GetClients", mock.Anything, legacyTestToken, "realm", searchesFor(models.ClientTypeServiceAccount)).
		Return([]*gocloak.Client{kcClientOfType("a", models.ClientTypeServiceAccount)}, nil).Once()
	kcMock.On("GetClients", mock.Anything, legacyTestToken, "realm", searchesFor("tenant-application")).
		Return([]*gocloak.Client{kcClientOfType("b", "tenant-application")}, nil).Once()

	clients, err := repo.GetClients(context.Background(), "realm", &models.ClientFilter{
		ClientType: pointer.ToString(models.ClientTypeServiceAccount),
		MaxResults: ptr.To(1),
	})

	require.NoError(t, err)
	assert.Len(t, clients, 1)
}

func TestKeycloakClientRepository_GetClients_UnrenamedTypeQueriesOnce(t *testing.T) {
	t.Parallel()
	repo, kcMock := newLegacyTypeTestRepo(t)
	kcMock.On("GetClients", mock.Anything, legacyTestToken, "realm", searchesFor("agent")).
		Return([]*gocloak.Client{kcClientOfType("agent-1", "agent")}, nil).Once()

	clients, err := repo.GetClients(context.Background(), "realm", &models.ClientFilter{ClientType: pointer.ToString("agent")})

	require.NoError(t, err)
	assert.Len(t, clients, 1)
}

func TestKeycloakClientRepository_GetClientsCount_CountsBothTypeFormats(t *testing.T) {
	t.Parallel()
	repo, kcMock := newLegacyTypeTestRepo(t)
	kcMock.On("GetClients", mock.Anything, legacyTestToken, "realm", searchesFor(models.ClientTypeAccessKey)).
		Return([]*gocloak.Client{kcClientOfType("new-1", models.ClientTypeAccessKey)}, nil).Once()
	kcMock.On("GetClients", mock.Anything, legacyTestToken, "realm", searchesFor("user-application")).
		Return([]*gocloak.Client{kcClientOfType("old-1", "user-application"), kcClientOfType("old-2", "user-application")}, nil).Once()

	count, err := repo.GetClientsCount(context.Background(), "realm", &models.ClientFilter{
		ClientType: pointer.ToString(models.ClientTypeAccessKey),
		CreatedBy:  pointer.ToString("creator"),
	})

	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestKeycloakClientRepository_GetAccessKeyClientById_FindsLegacyAccessKey(t *testing.T) {
	t.Parallel()
	repo, kcMock := newLegacyTypeTestRepo(t)
	kcMock.On("GetClients", mock.Anything, legacyTestToken, "realm", searchesFor(models.ClientTypeAccessKey)).
		Return([]*gocloak.Client{}, nil).Once()
	kcMock.On("GetClients", mock.Anything, legacyTestToken, "realm", searchesFor("user-application")).
		Return([]*gocloak.Client{kcClientOfType("legacy-key", "user-application")}, nil).Once()

	client, err := repo.GetAccessKeyClientById(context.Background(), "realm", "creator", "legacy-key")

	require.NoError(t, err)
	assert.Equal(t, "legacy-key", *client.ID)
}

func TestKeycloakClientRepository_GetAccessKeyClientById_FindsMigratedAccessKey(t *testing.T) {
	t.Parallel()
	repo, kcMock := newLegacyTypeTestRepo(t)
	kcMock.On("GetClients", mock.Anything, legacyTestToken, "realm", searchesFor(models.ClientTypeAccessKey)).
		Return([]*gocloak.Client{kcClientOfType("new-key", models.ClientTypeAccessKey)}, nil).Once()
	kcMock.On("GetClients", mock.Anything, legacyTestToken, "realm", searchesFor("user-application")).
		Return([]*gocloak.Client{}, nil).Once()

	client, err := repo.GetAccessKeyClientById(context.Background(), "realm", "creator", "new-key")

	require.NoError(t, err)
	assert.Equal(t, "new-key", *client.ID)
}
