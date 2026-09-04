/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	realm_client "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/realm/client"
)

func TestServiceAccount_CreateServiceAccount_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	name := GenerateRandomString(serviceAccountNamePrefix)
	serviceAccountRequest := openapi.ServiceAccountCreationRequest{
		Name: name,
	}
	resp, httpResp, err := OpenapiClient.ServiceAccountsAPI.CreateServiceAccount(ctx).ServiceAccountCreationRequest(serviceAccountRequest).Execute()
	defer serviceAccountCleanup(t, ctx, resp.Id)

	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, httpResp.StatusCode)
	require.NotNil(t, resp)

	require.NotNil(t, resp.Id)
	require.NotNil(t, resp.ClientSecret)
	require.NotNil(t, resp.ClientId)
	require.Equal(t, name, resp.Name)
	require.Equal(t, *resp.ClientId, resp.Name)

	newServiceAccount, err := IamServices.ClientService.GetClient(ctx, BaseRealmName, resp.Id)
	require.NoError(t, err)
	require.NotNil(t, newServiceAccount)
	require.True(t, *newServiceAccount.Attributes.IsExternal)
	require.NotNil(t, newServiceAccount.Attributes.UpdatedAt)
	require.NotNil(t, newServiceAccount.Attributes.CreatedAt)
	require.Nil(t, newServiceAccount.Attributes.LastLogin)
	require.Equal(t, 30*time.Minute, *newServiceAccount.Attributes.TokenLifespan)
	require.Equal(t, realm_client.ClientTypeTenantApp, *newServiceAccount.Attributes.ClientType)
	require.Equal(t, createdByUser, *newServiceAccount.Attributes.CreatedBy)
}

func TestServiceAccount_CreateServiceAccount_InvalidName(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	serviceAccountRequest := openapi.ServiceAccountCreationRequest{
		Name: "",
	}
	resp, httpResp, err := OpenapiClient.ServiceAccountsAPI.CreateServiceAccount(ctx).ServiceAccountCreationRequest(serviceAccountRequest).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestServiceAccount_Delete_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	serviceAccountId, _ := TestStepsClient.CreateServiceAccount(ctx, BaseRealmName, BaseTenantId, nil)
	httpResp, err := OpenapiClient.ServiceAccountsAPI.DeleteServiceAccountById(ctx, serviceAccountId).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, httpResp.StatusCode)
}

func TestServiceAccount_Delete_InvalidID(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	httpResp, err := OpenapiClient.ServiceAccountsAPI.DeleteServiceAccountById(ctx, "invalid").Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode)
}

func TestServiceAccount_GetServiceAccountById_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	name := GenerateRandomString(serviceAccountNamePrefix)
	serviceAccountId, err := TestStepsClient.CreateServiceAccount(ctx, BaseRealmName, BaseTenantId, ptr.To(name))
	require.NoError(t, err)
	defer serviceAccountCleanup(t, ctx, serviceAccountId)

	resp, httpResp, err := OpenapiClient.ServiceAccountsAPI.GetServiceAccountById(ctx, serviceAccountId).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.NotNil(t, resp)

	require.Equal(t, serviceAccountId, resp.Id)
	require.Equal(t, name, resp.Name)
	require.NotNil(t, resp.ClientId)
	require.Equal(t, resp.CreatedBy, createdByUser)
	require.Equal(t, *resp.TenantId, BaseTenantId)
	require.True(t, resp.Enabled)
	require.NotNil(t, resp.CreatedAt)
	require.NotNil(t, resp.UpdatedAt)
}

func TestServiceAccount_GetServiceAccountById_InvalidID(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	resp, httpResp, err := OpenapiClient.ServiceAccountsAPI.GetServiceAccountById(ctx, "").Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestServiceAccount_GetServiceAccounts_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	name1 := GenerateRandomString(serviceAccountNamePrefix)
	name2 := GenerateRandomString(serviceAccountNamePrefix)
	serviceAccountId1, err := TestStepsClient.CreateServiceAccount(ctx, BaseRealmName, BaseTenantId, ptr.To(name1))
	require.NoError(t, err)
	serviceAccountId2, err := TestStepsClient.CreateServiceAccount(ctx, BaseRealmName, BaseTenantId, ptr.To(name2))
	require.NoError(t, err)
	defer serviceAccountCleanup(t, ctx, serviceAccountId1)
	defer serviceAccountCleanup(t, ctx, serviceAccountId2)

	resp, httpResp, err := OpenapiClient.ServiceAccountsAPI.GetServiceAccounts(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.NotEmpty(t, resp)

	ids := make(map[string]bool, len(resp))
	for _, sa := range resp {
		ids[sa.Id] = true
	}
	require.True(t, ids[serviceAccountId1], "created service account %s should appear in list", name1)
	require.True(t, ids[serviceAccountId2], "created service account %s should appear in list", name2)
}

func TestServiceAccount_UpdateServiceAccount_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	serviceAccountId, _ := TestStepsClient.CreateServiceAccount(ctx, BaseRealmName, BaseTenantId, nil)
	defer serviceAccountCleanup(t, ctx, serviceAccountId)

	httpResp, err := OpenapiClient.ServiceAccountsAPI.UpdateServiceAccountById(ctx, serviceAccountId).ServiceAccountPatchRequest(openapi.ServiceAccountPatchRequest{
		Enabled: false,
	}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)

	updatedServiceAccount, err := IamServices.ClientService.GetClient(ctx, BaseRealmName, serviceAccountId)
	require.NoError(t, err)
	require.False(t, *updatedServiceAccount.Enabled)
}

func TestServiceAccount_UpdateServiceAccount_InvalidID(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	httpResp, err := OpenapiClient.ServiceAccountsAPI.UpdateServiceAccountById(ctx, "").ServiceAccountPatchRequest(openapi.ServiceAccountPatchRequest{
		Enabled: false,
	}).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode)
}

func TestServiceAccount_RegenerateServiceAccountSecret_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	serviceAccountId, _ := TestStepsClient.CreateServiceAccount(ctx, BaseRealmName, BaseTenantId, nil)
	defer serviceAccountCleanup(t, ctx, serviceAccountId)
	serviceAccount, err := IamServices.ClientService.GetClient(ctx, BaseRealmName, serviceAccountId)
	require.NoError(t, err)
	require.NotNil(t, serviceAccount.Secret)

	resp, httpResp, err := OpenapiClient.ServiceAccountsAPI.RegenerateServiceAccountSecret(ctx, serviceAccountId).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.NotNil(t, resp)
	require.NotNil(t, resp.ClientSecret)

	require.NotEqual(t, serviceAccount.Secret, resp.ClientSecret)
}

func TestServiceAccount_RegenerateServiceAccountSecret_InvalidID(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	resp, httpResp, err := OpenapiClient.ServiceAccountsAPI.RegenerateServiceAccountSecret(ctx, "").Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestServiceAccount_RegenerateServiceAccountSecret_NonExistentServiceAccount(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	nonExistentServiceAccountId := "non-existent-service-account-id-12345"
	resp, httpResp, err := OpenapiClient.ServiceAccountsAPI.RegenerateServiceAccountSecret(ctx, nonExistentServiceAccountId).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestServiceAccount_RegenerateServiceAccountSecret_Forbidden(t *testing.T) {
	t.Parallel()
	ctx := context.Background() // Using unauthorized context to trigger privilege escalation failure
	serviceAccountId, _ := TestStepsClient.CreateServiceAccount(newTestContext(), BaseRealmName, BaseTenantId, nil)
	defer serviceAccountCleanup(t, newTestContext(), serviceAccountId)

	resp, httpResp, err := OpenapiClient.ServiceAccountsAPI.RegenerateServiceAccountSecret(ctx, serviceAccountId).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestServiceAccount_Delete_VerifyGone(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	serviceAccountId, err := TestStepsClient.CreateServiceAccount(ctx, BaseRealmName, BaseTenantId, nil)
	require.NoError(t, err)

	httpResp, err := OpenapiClient.ServiceAccountsAPI.DeleteServiceAccountById(ctx, serviceAccountId).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, httpResp.StatusCode)

	_, httpResp, err = OpenapiClient.ServiceAccountsAPI.GetServiceAccountById(ctx, serviceAccountId).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode)
}

func TestServiceAccount_DoubleDelete(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	serviceAccountId, err := TestStepsClient.CreateServiceAccount(ctx, BaseRealmName, BaseTenantId, nil)
	require.NoError(t, err)

	httpResp, err := OpenapiClient.ServiceAccountsAPI.DeleteServiceAccountById(ctx, serviceAccountId).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, httpResp.StatusCode)

	httpResp, err = OpenapiClient.ServiceAccountsAPI.DeleteServiceAccountById(ctx, serviceAccountId).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode)
}

func TestServiceAccount_GetAll_Forbidden(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	resp, httpResp, err := OpenapiClient.ServiceAccountsAPI.GetServiceAccounts(ctx).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestServiceAccount_Create_Forbidden(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	resp, httpResp, err := OpenapiClient.ServiceAccountsAPI.CreateServiceAccount(ctx).ServiceAccountCreationRequest(openapi.ServiceAccountCreationRequest{
		Name: GenerateRandomString(serviceAccountNamePrefix),
	}).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestServiceAccount_GetById_Forbidden(t *testing.T) {
	t.Parallel()
	adminCtx := newTestContext()
	serviceAccountId, err := TestStepsClient.CreateServiceAccount(adminCtx, BaseRealmName, BaseTenantId, nil)
	require.NoError(t, err)
	defer serviceAccountCleanup(t, adminCtx, serviceAccountId)

	ctx := context.Background()
	resp, httpResp, err := OpenapiClient.ServiceAccountsAPI.GetServiceAccountById(ctx, serviceAccountId).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestServiceAccount_CreateInvalidName_LongOrSymbols(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	resp, httpResp, err := OpenapiClient.ServiceAccountsAPI.CreateServiceAccount(ctx).ServiceAccountCreationRequest(openapi.ServiceAccountCreationRequest{
		Name: "sdfsdfsdfsdfsdfsdfsdfdsf!@@$@$454364ehfhghjfgjfjfgjfjfgjfgjfjfjfgjgfbdfgbcdb",
	}).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestServiceAccount_CreateInvalidName_CapitalLetters(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	resp, httpResp, err := OpenapiClient.ServiceAccountsAPI.CreateServiceAccount(ctx).ServiceAccountCreationRequest(openapi.ServiceAccountCreationRequest{
		Name: "AAAA",
	}).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestServiceAccount_CreateInvalidName_StartsWithNumber(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	resp, httpResp, err := OpenapiClient.ServiceAccountsAPI.CreateServiceAccount(ctx).ServiceAccountCreationRequest(openapi.ServiceAccountCreationRequest{
		Name: "2serviceaccount",
	}).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestServiceAccount_CreateInvalidName_XSS(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	resp, httpResp, err := OpenapiClient.ServiceAccountsAPI.CreateServiceAccount(ctx).ServiceAccountCreationRequest(openapi.ServiceAccountCreationRequest{
		Name: "<script>alert(123)</script>",
	}).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
	require.Nil(t, resp)
}

func serviceAccountCleanup(t *testing.T, ctx context.Context, serviceAccountId string) {
	if serviceAccountId == "" {
		return
	}
	err := TestStepsClient.DeleteClient(ctx, BaseRealmName, serviceAccountId)
	require.NoError(t, err)
}
