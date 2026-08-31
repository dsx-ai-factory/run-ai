/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	serviceUtils "github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
)

func TestServiceAccount_MintInstallerBootstrapToken_RotatesSecretEachCall(t *testing.T) {
	ctx := newTestContext()

	before, err := IamServices.ClientService.GetClientByClientId(ctx, BaseRealmName, serviceUtils.RunaiInstallerClientId)
	require.NoError(t, err)
	require.NotNil(t, before.ID)

	firstSecret, err := IamServices.ClientService.GetClientSecret(ctx, BaseRealmName, *before.ID)
	require.NoError(t, err)

	resp1, httpResp1, err := OpenapiClient.ServiceAccountsAPI.MintInstallerBootstrapToken(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp1.StatusCode)
	require.NotEmpty(t, resp1.Token)

	secondSecret, err := IamServices.ClientService.GetClientSecret(ctx, BaseRealmName, *before.ID)
	require.NoError(t, err)
	require.NotEqual(t, firstSecret, secondSecret)

	resp2, httpResp2, err := OpenapiClient.ServiceAccountsAPI.MintInstallerBootstrapToken(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp2.StatusCode)
	require.NotEmpty(t, resp2.Token)

	thirdSecret, err := IamServices.ClientService.GetClientSecret(ctx, BaseRealmName, *before.ID)
	require.NoError(t, err)
	require.NotEqual(t, secondSecret, thirdSecret)
}

func TestServiceAccount_MintInstallerBootstrapToken_OldSecretInvalidatedAfterMint(t *testing.T) {
	ctx := newTestContext()

	client, err := IamServices.ClientService.GetClientByClientId(ctx, BaseRealmName, serviceUtils.RunaiInstallerClientId)
	require.NoError(t, err)
	require.NotNil(t, client.ID)
	require.NotNil(t, client.ClientID)

	oldSecret, err := IamServices.ClientService.GetClientSecret(ctx, BaseRealmName, *client.ID)
	require.NoError(t, err)

	_, err = IamServices.AuthService.AuthenticateClient(ctx, BaseRealmName, *client.ClientID, oldSecret)
	require.NoError(t, err)

	_, httpResp, err := OpenapiClient.ServiceAccountsAPI.MintInstallerBootstrapToken(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)

	_, err = IamServices.AuthService.AuthenticateClient(ctx, BaseRealmName, *client.ClientID, oldSecret)
	require.Error(t, err)
}

func TestServiceAccount_MintInstallerBootstrapToken_TokenIsUsable(t *testing.T) {
	ctx := newTestContext()

	client, err := IamServices.ClientService.GetClientByClientId(ctx, BaseRealmName, serviceUtils.RunaiInstallerClientId)
	require.NoError(t, err)
	require.NotNil(t, client.ClientID)

	resp, httpResp, err := OpenapiClient.ServiceAccountsAPI.MintInstallerBootstrapToken(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.NotEmpty(t, resp.Token)

	parser := jwt.NewParser()
	parsedToken, _, err := parser.ParseUnverified(resp.Token, &jwt.MapClaims{})
	require.NoError(t, err)

	claims, ok := parsedToken.Claims.(*jwt.MapClaims)
	require.True(t, ok)

	azp, _ := (*claims)["azp"].(string)
	require.Equal(t, *client.ClientID, azp)

	exp, err := claims.GetExpirationTime()
	require.NoError(t, err)
	require.NotNil(t, exp)
	require.True(t, exp.Time.After(time.Now()))
	require.True(t, exp.Time.Before(time.Now().Add(3*time.Hour)))
}

func TestServiceAccount_MintInstallerBootstrapToken_ConcurrentMintsAllSucceed(t *testing.T) {
	ctx := newTestContext()

	const concurrentMints = 5
	var wg sync.WaitGroup
	errs := make([]error, concurrentMints)
	statusCodes := make([]int, concurrentMints)
	tokens := make([]string, concurrentMints)

	for i := 0; i < concurrentMints; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			resp, httpResp, err := OpenapiClient.ServiceAccountsAPI.MintInstallerBootstrapToken(ctx).Execute()
			errs[idx] = err
			if httpResp != nil {
				statusCodes[idx] = httpResp.StatusCode
			}
			if resp != nil {
				tokens[idx] = resp.Token
			}
		}(i)
	}
	wg.Wait()

	for i := range errs {
		require.NoError(t, errs[i], "mint %d failed", i)
		require.Equal(t, http.StatusOK, statusCodes[i], "mint %d returned unexpected status", i)
		require.NotEmpty(t, tokens[i], "mint %d returned an empty token", i)
	}
}
