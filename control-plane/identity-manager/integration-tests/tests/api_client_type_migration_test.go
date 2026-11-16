/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	realm_client "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/realm/client"
)

// These tests cover the rename of runai.client_type values ("user-application" -> "access-key",
// "tenant-application" -> "service-account"). Legacy clients are created with the old strings and must be
// found by lookups and rewritten by the reconcilers, which can run more than once.

func clientTypeOf(t *testing.T, ctx context.Context, realmName, id string) string {
	t.Helper()
	client, err := IamServices.ClientService.GetClient(ctx, realmName, id)
	require.NoError(t, err)
	require.NotNil(t, client.Attributes)
	require.NotNil(t, client.Attributes.ClientType)
	return *client.Attributes.ClientType
}

func clientIdsOfType(t *testing.T, ctx context.Context, realmName, clientType string) []string {
	t.Helper()
	clients, err := IamServices.ClientService.GetClients(ctx, realmName, &models.ClientFilter{ClientType: ptr.To(clientType), IsTenantAdmin: true})
	require.NoError(t, err)
	ids := make([]string, 0, len(clients))
	for _, c := range clients {
		ids = append(ids, *c.ID)
	}
	return ids
}

func grantClientCredentialsClaims(t *testing.T, ctx context.Context, realmName, id string) jwt.MapClaims {
	t.Helper()
	client, err := IamServices.ClientService.GetClient(ctx, realmName, id)
	require.NoError(t, err)
	tokenCtx := TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, realmName)
	token, resp, err := OpenapiClient.TokensAPI.GrantToken(tokenCtx).TokenRequest(generateClientCredentialsTokenRequest(*client.ClientID, *client.Secret)).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	claims := jwt.MapClaims{}
	_, _, err = jwt.NewParser().ParseUnverified(*token.AccessToken, &claims)
	require.NoError(t, err)
	return claims
}

func TestReconcile_LegacyServiceAccount_RewrittenToCurrentType(t *testing.T) {
	ctx := context.Background()
	realmName, tenantId, err := TestStepsClient.CreateRealm(ctx, nil, nil)
	require.NoError(t, err)
	defer realmCleanup(t, ctx, realmName, true)

	id, err := TestStepsClient.CreateLegacyServiceAccount(ctx, realmName, tenantId)
	require.NoError(t, err)
	require.Equal(t, "tenant-application", clientTypeOf(t, ctx, realmName, id))

	// before the reconcile: lookups by the current type still find the legacy client
	require.Contains(t, clientIdsOfType(t, ctx, realmName, realm_client.ClientTypeServiceAccount), id)
	require.Equal(t, "tenant-application", grantClientCredentialsClaims(t, ctx, realmName, id)["client_type"])

	require.NoError(t, TestStepsClient.ReconcileRealmForced(ctx, realmName, ptr.To(openapi.RECONCILERTYPE_APPS)))

	require.Equal(t, "service-account", clientTypeOf(t, ctx, realmName, id))
	require.Equal(t, "service-account", grantClientCredentialsClaims(t, ctx, realmName, id)["client_type"])
	require.Contains(t, clientIdsOfType(t, ctx, realmName, realm_client.ClientTypeServiceAccount), id)
}

func TestReconcile_LegacyAccessKey_RewrittenToCurrentType(t *testing.T) {
	ctx := context.Background()
	realmName, tenantId, err := TestStepsClient.CreateRealm(ctx, nil, nil)
	require.NoError(t, err)
	defer realmCleanup(t, ctx, realmName, true)

	id, err := TestStepsClient.CreateLegacyAccessKey(ctx, realmName, tenantId, createdByUser)
	require.NoError(t, err)
	require.Equal(t, "user-application", clientTypeOf(t, ctx, realmName, id))

	// before the reconcile: lookups by the current type still find the legacy client, by id and by creator
	require.Contains(t, clientIdsOfType(t, ctx, realmName, realm_client.ClientTypeAccessKey), id)
	found, err := IamServices.ClientService.GetAccessKeyClientById(ctx, realmName, createdByUser, id)
	require.NoError(t, err)
	require.Equal(t, id, *found.ID)
	legacyClaims := grantClientCredentialsClaims(t, ctx, realmName, id)
	require.Equal(t, "user-application", legacyClaims["client_type"])
	require.Equal(t, createdByUser, legacyClaims["sub"])

	require.NoError(t, TestStepsClient.ReconcileRealmForced(ctx, realmName, ptr.To(openapi.RECONCILERTYPE_USER_APPS)))

	require.Equal(t, "access-key", clientTypeOf(t, ctx, realmName, id))
	claims := grantClientCredentialsClaims(t, ctx, realmName, id)
	require.Equal(t, "access-key", claims["client_type"])
	require.Equal(t, createdByUser, claims["sub"], "the access key must keep acting as its creator")
	found, err = IamServices.ClientService.GetAccessKeyClientById(ctx, realmName, createdByUser, id)
	require.NoError(t, err)
	require.Equal(t, id, *found.ID)
}

func TestReconcile_MixedLegacyAndMigratedClients_AllEndOnCurrentTypeAndRerunIsStable(t *testing.T) {
	ctx := context.Background()
	realmName, tenantId, err := TestStepsClient.CreateRealm(ctx, nil, nil)
	require.NoError(t, err)
	defer realmCleanup(t, ctx, realmName, true)

	legacyKeyId, err := TestStepsClient.CreateLegacyAccessKey(ctx, realmName, tenantId, createdByUser)
	require.NoError(t, err)
	migratedKeyId, err := TestStepsClient.CreateAccessKey(ctx, realmName, tenantId, nil, ptr.To(createdByUser))
	require.NoError(t, err)
	legacyAccountId, err := TestStepsClient.CreateLegacyServiceAccount(ctx, realmName, tenantId)
	require.NoError(t, err)
	migratedAccountId, err := TestStepsClient.CreateServiceAccount(ctx, realmName, tenantId, nil)
	require.NoError(t, err)

	// a partial earlier run leaves both formats side by side: the pull must return both
	require.ElementsMatch(t, []string{legacyKeyId, migratedKeyId}, clientIdsOfType(t, ctx, realmName, realm_client.ClientTypeAccessKey))
	require.ElementsMatch(t, []string{legacyAccountId, migratedAccountId}, clientIdsOfType(t, ctx, realmName, realm_client.ClientTypeServiceAccount))

	serviceAccountsReconciler := ptr.To(openapi.RECONCILERTYPE_APPS)
	accessKeysReconciler := ptr.To(openapi.RECONCILERTYPE_USER_APPS)

	// the reconcile can run more than once: the second pass must neither fail nor drop clients
	for run := 1; run <= 2; run++ {
		require.NoError(t, TestStepsClient.ReconcileRealmForced(ctx, realmName, serviceAccountsReconciler), "service accounts run %d", run)
		require.NoError(t, TestStepsClient.ReconcileRealmForced(ctx, realmName, accessKeysReconciler), "access keys run %d", run)

		require.Equal(t, "access-key", clientTypeOf(t, ctx, realmName, legacyKeyId), "run %d", run)
		require.Equal(t, "access-key", clientTypeOf(t, ctx, realmName, migratedKeyId), "run %d", run)
		require.Equal(t, "service-account", clientTypeOf(t, ctx, realmName, legacyAccountId), "run %d", run)
		require.Equal(t, "service-account", clientTypeOf(t, ctx, realmName, migratedAccountId), "run %d", run)
		require.ElementsMatch(t, []string{legacyKeyId, migratedKeyId}, clientIdsOfType(t, ctx, realmName, realm_client.ClientTypeAccessKey), "run %d", run)
		require.ElementsMatch(t, []string{legacyAccountId, migratedAccountId}, clientIdsOfType(t, ctx, realmName, realm_client.ClientTypeServiceAccount), "run %d", run)
	}
}
