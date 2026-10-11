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

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/utils"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	realm_client "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/realm/client"
	serviceUtils "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
)

func TestCreateRealm_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	realmName := GenerateRandomString(DefaultRealmPrefix)
	tenantId := GenerateRandomTenantID()
	realmData := openapi.RealmCreationRequest{
		RealmName: realmName,
		TenantId:  tenantId,
	}
	body, resp, err := OpenapiClient.RealmAPI.CreateRealm(ctx).RealmCreationRequest(realmData).Execute()
	// API Assertions
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, body)
	// Iam Assertions
	realm, err := IamServices.RealmService.GetRealm(ctx, realmName)
	require.NoError(t, err)
	require.NotNil(t, realm)
	require.Equal(t, realmName, *realm.Name)
	require.True(t, *realm.Enabled)
	// DB Assertions
	realmDB, err := RealmDal.GetRealmByName(ctx, realmName)
	require.NoError(t, err)
	require.NotNil(t, realmDB)
	require.Equal(t, realmName, realmDB.Realm)
}

func TestCreateRealm_DuplicateRealmName(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	realmData := openapi.RealmCreationRequest{
		RealmName: BaseRealmName,
		TenantId:  BaseTenantId,
	}
	body, resp, err := OpenapiClient.RealmAPI.CreateRealm(ctx).RealmCreationRequest(realmData).Execute()
	require.Error(t, err)
	require.Nil(t, body)
	require.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestCreateRealm_InvalidRealmName(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	realmData := openapi.RealmCreationRequest{
		RealmName: "a*a",
		TenantId:  BaseTenantId,
	}
	body, resp, err := OpenapiClient.RealmAPI.CreateRealm(ctx).RealmCreationRequest(realmData).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Nil(t, body)
}

func TestCreateRealm_InvalidTenantId(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	realmData := openapi.RealmCreationRequest{
		RealmName: "abc12",
		TenantId:  "abc",
	}
	body, resp, err := OpenapiClient.RealmAPI.CreateRealm(ctx).RealmCreationRequest(realmData).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Nil(t, body)
}

func TestCreateRealm_EmptyRealmName(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	realmData := openapi.RealmCreationRequest{
		RealmName: "",
		TenantId:  BaseTenantId,
	}
	body, resp, err := OpenapiClient.RealmAPI.CreateRealm(ctx).RealmCreationRequest(realmData).Execute()
	require.Error(t, err)
	require.Nil(t, body)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateRealm_EmptyTenantId(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	realmData := openapi.RealmCreationRequest{
		RealmName: "abcd",
		TenantId:  "",
	}
	body, resp, err := OpenapiClient.RealmAPI.CreateRealm(ctx).RealmCreationRequest(realmData).Execute()
	require.Error(t, err)
	require.Nil(t, body)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestDeleteRealmHardDelete_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	// Prerequisites
	name, _, err := TestStepsClient.CreateRealm(ctx, nil, nil)
	require.NoError(t, err)
	// Delete Realm
	resp, err := OpenapiClient.RealmAPI.DeleteRealm(ctx, name).IsHardDelete(true).Execute()
	// API Assertions
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	// DB Assertions
	realmDB, err := RealmDal.GetRealmByNameWithDeleted(ctx, name)
	require.Error(t, err)
	require.Nil(t, realmDB)
	// Iam Assertions
	realm, err := IamServices.RealmService.GetRealm(ctx, name)
	require.Error(t, err)
	require.Nil(t, realm)
}

func TestDeleteRealmSoftDelete_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	// Prerequisites
	name, _, err := TestStepsClient.CreateRealm(ctx, nil, nil)
	require.NoError(t, err)
	// Delete Realm
	resp, err := OpenapiClient.RealmAPI.DeleteRealm(ctx, name).IsHardDelete(false).Execute()
	// API Assertions
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	// DB Assertions
	realmDB, err := RealmDal.GetRealmByNameWithDeleted(ctx, name)
	require.NoError(t, err)
	require.NotNil(t, realmDB)
	require.NotNil(t, realmDB.DeletedAt)
	// Iam Assertions
	realm, err := IamServices.RealmService.GetRealm(ctx, name)
	require.Nil(t, err)
	require.False(t, *realm.Enabled)
}

func TestDeleteRealmSoftDelete_RealmDoesNotExist(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	name := GenerateRandomString(DefaultRealmPrefix)
	resp, err := OpenapiClient.RealmAPI.DeleteRealm(ctx, name).IsHardDelete(false).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestReconcileRealm_HappyFlow_ClientsFlow(t *testing.T) {
	ctx := context.Background()
	resp, err := OpenapiClient.RealmAPI.ReconcileRealm(ctx, BaseRealmName).RealmReconcileRequest(openapi.RealmReconcileRequest{Force: ptr.To(false)}).Execute()

	require.Equal(t, http.StatusAccepted, resp.StatusCode)
	require.NoError(t, err)

	time.Sleep(10 * time.Second)

	realmDB, err := RealmDal.GetRealmByName(ctx, BaseRealmName)
	require.NoError(t, err)
	require.NotNil(t, realmDB)

	realm, err := IamServices.RealmService.GetRealm(ctx, BaseRealmName)
	require.NoError(t, err)
	require.NotNil(t, realm)
	require.True(t, *realm.Enabled)

	checkCliClient(t, ctx)
	checkRunaiClientDeleted(t, ctx)
	checkRunaiCliClient(t, ctx)
	checkRunaiAdminUiClient(t, ctx)
	checkRunaiTokenExchangeClient(t, ctx)
	checkRunaiInternalInstallerClient(t, ctx)
	checkMcpClient(t, ctx)
}

func TestReconcile_Realms_HappyFlow(t *testing.T) {
	ctx := context.Background()

	realmName := GenerateRandomString(DefaultRealmPrefix)
	tenantId := GenerateRandomTenantID()
	realmData := openapi.RealmCreationRequest{
		RealmName: realmName,
		TenantId:  tenantId,
	}
	body, resp, err := OpenapiClient.RealmAPI.CreateRealm(ctx).RealmCreationRequest(realmData).Execute()
	defer realmCleanup(t, ctx, realmName, true)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, body)

	reconcilerType, err := openapi.NewReconcilerTypeFromValue("realms")
	require.NoError(t, err)
	err = TestStepsClient.ReconcileRealm(ctx, realmName, reconcilerType)
	require.NoError(t, err)

	checkRealmDefaults(t, ctx, realmName)
	checkAuthFlow(t, ctx, realmName)
	checkRequiredActions(t, ctx, realmName)
}

func TestReconcile_ServiceAccount_HappyFlow_ValidateFields(t *testing.T) {
	ctx := context.Background()

	realmName, tenantId, err := TestStepsClient.CreateRealm(ctx, nil, nil)
	defer realmCleanup(t, ctx, realmName, true)
	require.NoError(t, err)

	serviceAccountId, err := TestStepsClient.CreateServiceAccount(ctx, realmName, tenantId, nil)
	require.NoError(t, err)
	require.NotNil(t, serviceAccountId)

	reconcilerType, err := openapi.NewReconcilerTypeFromValue("apps")
	require.NoError(t, err)
	err = TestStepsClient.ReconcileRealm(ctx, realmName, reconcilerType)
	require.NoError(t, err)

	serviceAccount, err := IamServices.ClientService.GetClient(ctx, realmName, serviceAccountId)
	require.NoError(t, err)
	require.NotNil(t, serviceAccount)

	require.Equal(t, serviceAccount.Name, serviceAccount.ClientID)
	require.True(t, *serviceAccount.Attributes.IsExternal)
	require.NotNil(t, serviceAccount.Attributes.UpdatedAt)
	require.NotNil(t, serviceAccount.Attributes.CreatedAt)
	require.Nil(t, serviceAccount.Attributes.LastLogin)
	require.Equal(t, realm_client.ClientTypeTenantApp, *serviceAccount.Attributes.ClientType)
	require.Equal(t, createdByUser, *serviceAccount.Attributes.CreatedBy)
}

func TestReconcile_ServiceAccount_NoServiceAccounts(t *testing.T) {
	ctx := context.Background()

	realmName, _, err := TestStepsClient.CreateRealm(ctx, nil, nil)
	defer realmCleanup(t, ctx, realmName, true)
	require.NoError(t, err)

	reconcilerType, err := openapi.NewReconcilerTypeFromValue("apps")
	require.NoError(t, err)
	err = TestStepsClient.ReconcileRealm(ctx, realmName, reconcilerType)
	require.NoError(t, err)

	serviceAccounts, err := IamServices.ClientService.GetClients(ctx, realmName, &models.ClientFilter{ClientType: ptr.To(realm_client.ClientTypeTenantApp), IsTenantAdmin: true})
	require.NoError(t, err)
	require.Empty(t, serviceAccounts)
}

func TestReconcile_ServiceAccount_RunReconcileAfterUpdate(t *testing.T) {
	ctx := context.Background()

	realmName, tenantId, err := TestStepsClient.CreateRealm(ctx, nil, nil)
	defer realmCleanup(t, ctx, realmName, true)
	require.NoError(t, err)

	serviceAccountId, err := TestStepsClient.CreateServiceAccount(ctx, realmName, tenantId, nil)
	require.NoError(t, err)
	require.NotNil(t, serviceAccountId)

	serviceAccount, err := IamServices.ClientService.GetClient(ctx, realmName, serviceAccountId)
	require.NoError(t, err)
	require.Equal(t, createdByUser, *serviceAccount.Attributes.CreatedBy)

	updatedServiceAccountClient := &models.Client{
		ID:       ptr.To(serviceAccountId),
		Enabled:  ptr.To(true),
		Secret:   ptr.To("secret"),
		TenantId: ptr.To(tenantId),
		Attributes: &models.ClientAttributes{
			IsExternal: nil,
			ClientType: ptr.To(realm_client.ClientTypeTenantApp),
			CreatedBy:  ptr.To("test1@run.ai"),
		},
	}
	err = IamServices.ClientService.UpdateClient(ctx, realmName, updatedServiceAccountClient)
	require.NoError(t, err)

	reconcilerType, err := openapi.NewReconcilerTypeFromValue("apps")
	require.NoError(t, err)
	err = TestStepsClient.ReconcileRealm(ctx, realmName, reconcilerType)
	require.NoError(t, err)

	updatedServiceAccount, err := IamServices.ClientService.GetClient(ctx, realmName, serviceAccountId)
	require.NoError(t, err)
	require.True(t, *updatedServiceAccount.Attributes.IsExternal)
	require.Equal(t, realm_client.ClientTypeTenantApp, *updatedServiceAccount.Attributes.ClientType)
	require.Equal(t, "test1@run.ai", *updatedServiceAccount.Attributes.CreatedBy)
}

func TestReconcile_InternalClients_HappyFlow_ValidateFields(t *testing.T) {
	ctx := context.Background()

	realmName, tenantId, err := TestStepsClient.CreateRealm(ctx, nil, nil)
	defer realmCleanup(t, ctx, realmName, true)
	require.NoError(t, err)

	ctx = TestStepsClient.SetContextValues(context.Background(), tenantId)
	clientId := GenerateRandomString(clientNamePrefix)
	createAgentClientWithTenant(t, ctx, clientId, ptr.To("clusterId"), tenantId)

	internalClients, err := IamServices.ClientService.GetClients(ctx, realmName, &models.ClientFilter{IsExternal: ptr.To(false)})
	require.NoError(t, err)
	require.NotNil(t, internalClients)
	require.Len(t, internalClients, 2)

	reconcilerType, err := openapi.NewReconcilerTypeFromValue("internalApps")
	require.NoError(t, err)
	err = TestStepsClient.ReconcileRealm(ctx, realmName, reconcilerType)
	require.NoError(t, err)

	internalClients, err = IamServices.ClientService.GetClients(ctx, realmName, &models.ClientFilter{IsExternal: ptr.To(false)})
	require.NoError(t, err)
	require.NotNil(t, internalClients)

	require.Len(t, internalClients, 2)
	for _, client := range internalClients {
		clientType := getClientType(client)
		if clientType == realm_client.ClientTypeInstaller {
			checkInstallerClientFields(t, client)
		} else if clientType == realm_client.ClientTypeAgent {
			checkAgentClientFields(t, client)
		} else {
			require.Fail(t, "unsupported client type for internal clients reconciler")
		}
	}
}

func realmCleanup(t *testing.T, ctx context.Context, realmName string, isHardDelete bool) {
	err := TestStepsClient.DeleteRealm(ctx, realmName, isHardDelete)
	require.NoError(t, err)
}

func checkRequiredActions(t *testing.T, ctx context.Context, realmName string) {
	action, err := IamServices.UserService.GetRequiredAction(ctx, realmName, string(iamModels.RequiredActionVerifyProfile))
	require.NoError(t, err)
	require.Equal(t, string(iamModels.RequiredActionVerifyProfile), *action.Alias)
	require.False(t, *action.Enabled)
}

func checkAuthFlow(t *testing.T, ctx context.Context, realmName string) {
	authFlow, err := IamServices.AuthFlowService.GetAuthFlow(ctx, realmName, serviceUtils.LOGIN_FLOW_NAME)
	require.NoError(t, err)
	require.NotNil(t, authFlow)
	require.Equal(t, serviceUtils.LOGIN_FLOW_NAME, *authFlow.Alias)
	require.True(t, *authFlow.TopLevel)
	require.Equal(t, "basic-flow", *authFlow.ProviderID)
}

func checkRealmDefaults(t *testing.T, ctx context.Context, realmName string) {
	realm, err := IamServices.RealmService.GetRealm(ctx, realmName)
	require.NoError(t, err)
	require.NotNil(t, realm)

	require.NotNil(t, realm.ID)
	require.Equal(t, realmName, *realm.Name)
	require.True(t, *realm.Enabled)
	require.True(t, *realm.RememberMe)
	require.True(t, *realm.EventsEnabled)
	require.Equal(t, "runai-theme", *realm.LoginTheme)
	require.Equal(t, "runai-theme", *realm.EmailTheme)
	require.Contains(t, *realm.EnabledEventTypes, "LOGIN")
	require.Contains(t, *realm.EnabledEventTypes, "CLIENT_LOGIN")
}

func checkRunaiAdminUiClient(t *testing.T, ctx context.Context) {
	client, err := IamServices.ClientService.GetClientByClientId(ctx, BaseRealmName, serviceUtils.RUNAI_ADMIN_UI_CLIENT_ID)
	require.NoError(t, err)
	require.NotEmpty(t, client)
	require.Equal(t, serviceUtils.RUNAI_ADMIN_UI_CLIENT_ID, *client.Name)
	require.Equal(t, serviceUtils.RUNAI_ADMIN_UI_CLIENT_ID, *client.ClientID)
	require.True(t, *client.Enabled)
	require.True(t, client.IsClientAuthenticationEnabled())
	require.NotNil(t, client.Attributes)
}

func checkRunaiTokenExchangeClient(t *testing.T, ctx context.Context) {
	client, err := IamServices.ClientService.GetClientByClientId(ctx, BaseRealmName, serviceUtils.RUNAI_TOKEN_EXCHANGE_CLIENT_ID)
	require.NoError(t, err)
	require.NotEmpty(t, client)
	require.Equal(t, serviceUtils.RUNAI_TOKEN_EXCHANGE_CLIENT_ID, *client.Name)
	require.Equal(t, serviceUtils.RUNAI_TOKEN_EXCHANGE_CLIENT_ID, *client.ClientID)
	require.True(t, *client.Enabled)
	require.True(t, client.IsClientAuthenticationEnabled())
	require.NotNil(t, client.Attributes)
}

func checkRunaiClientDeleted(t *testing.T, ctx context.Context) {
	client, err := IamServices.ClientService.GetClientByClientId(ctx, BaseRealmName, serviceUtils.RUNAI_CLIENT_ID)
	require.NoError(t, err)
	require.Nil(t, client)
}

func checkRunaiCliClient(t *testing.T, ctx context.Context) {
	client, err := IamServices.ClientService.GetClientByClientId(ctx, BaseRealmName, serviceUtils.RUNAI_CLI_CLIENT_ID)
	require.NoError(t, err)
	require.NotEmpty(t, client)
	require.Equal(t, serviceUtils.RUNAI_CLI_CLIENT_ID, *client.Name)
	require.Equal(t, serviceUtils.RUNAI_CLI_CLIENT_ID, *client.ClientID)
	require.True(t, *client.Enabled)
	require.False(t, client.IsClientAuthenticationEnabled())
	require.NotNil(t, client.Attributes)
}

func checkRunaiInternalInstallerClient(t *testing.T, ctx context.Context) {
	client, err := IamServices.ClientService.GetClientByClientId(ctx, BaseRealmName, serviceUtils.RunaiInstallerClientId)
	require.NoError(t, err)
	require.NotEmpty(t, client)
	require.Equal(t, serviceUtils.RunaiInstallerClientId, *client.Name)
	require.Equal(t, serviceUtils.RunaiInstallerClientId, *client.ClientID)
	require.True(t, *client.Enabled)
	require.True(t, client.IsClientAuthenticationEnabled())
	require.NotNil(t, client.Attributes)

	require.NotNil(t, client.Attributes.ClientType)
	require.Equal(t, realm_client.ClientTypeInstaller, *client.Attributes.ClientType)
}

func checkCliClient(t *testing.T, ctx context.Context) {
	client, err := IamServices.ClientService.GetClientByClientId(ctx, BaseRealmName, serviceUtils.CLI_CLIENT_ID)
	require.NoError(t, err)
	require.NotEmpty(t, client)
	require.Equal(t, serviceUtils.CLI_CLIENT_ID, *client.Name)
	require.Equal(t, serviceUtils.CLI_CLIENT_ID, *client.ClientID)
	require.True(t, *client.Enabled)
	require.True(t, client.IsClientAuthenticationEnabled())
	require.NotNil(t, client.Attributes)
}

func checkMcpClient(t *testing.T, ctx context.Context) {
	client, err := IamServices.ClientService.GetClientByClientId(ctx, BaseRealmName, serviceUtils.RUNAI_MCP_CLIENT_ID)
	require.NoError(t, err)
	require.NotEmpty(t, client)
	require.Equal(t, serviceUtils.RUNAI_MCP_CLIENT_ID, *client.Name)
	require.Equal(t, serviceUtils.RUNAI_MCP_CLIENT_ID, *client.ClientID)
	require.True(t, *client.Enabled)
	require.True(t, client.IsClientAuthenticationEnabled())
	require.NotNil(t, client.Attributes)
	require.NotNil(t, client.Attributes.PkceCodeChallengeMethod)
	require.Equal(t, "S256", *client.Attributes.PkceCodeChallengeMethod)
	require.Contains(t, *client.RedirectURIs, "cursor://anysphere.cursor-mcp/oauth/callback")
	require.Contains(t, *client.RedirectURIs, "http://localhost/callback")
	require.Contains(t, *client.RedirectURIs, "http://localhost/oauth/callback")
	require.Contains(t, *client.RedirectURIs, "http://localhost/oauth/callback/debug")
}

func getClientType(client iamModels.Client) string {
	if client.Attributes == nil || client.Attributes.ClientType == nil {
		return "unsupported type"
	}
	return *client.Attributes.ClientType
}

func checkInstallerClientFields(t *testing.T, client models.Client) {
	require.Equal(t, "runai-internal.installer", *client.ClientID, "client ID should be runai-internal.installer")
	require.Equal(t, "runai-internal.installer", *client.Name, "client name should be runai-internal.installer")
	require.True(t, *client.Enabled, "client should be enabled")
	require.NotNil(t, client.Attributes, "client attributes should not be empty")
	require.True(t, client.IsClientAuthenticationEnabled(), "client authentication should be enabled")
	require.NotNil(t, client.Secret, "client secret should not be nil")
	require.Equal(t, "false", *client.Attributes.Oauth2AuthGrantEnabled)
	require.Equal(t, "false", *client.Attributes.OidcGrantEnabled, "oidc grant should be disabled")
	require.True(t, utils.IsValidLogoutRedirectUris(*client.Attributes.LogoutRedirectUris), "logout redirect uris should start with +")
	require.False(t, *client.Attributes.IsExternal, "is external should be false")
	require.Equal(t, realm_client.ClientTypeInstaller, *client.Attributes.ClientType, "client type should be installer")
	require.NotNil(t, client.Attributes.UpdatedAt, "update time should not be nil")
	require.NotNil(t, client.Attributes.CreatedAt, "creation time should not be nil")
}

func checkAgentClientFields(t *testing.T, client models.Client) {
	require.Equal(t, *client.Name, *client.ClientID, "client name should be equal to client ID")
	require.True(t, *client.Enabled, "client should be enabled")
	require.NotNil(t, client.Attributes, "client attributes should not be empty")
	require.True(t, client.IsClientAuthenticationEnabled(), "client authentication should be enabled")
	require.NotNil(t, client.Secret, "client secret should not be nil")
	require.Equal(t, "false", *client.Attributes.Oauth2AuthGrantEnabled)
	require.Equal(t, "false", *client.Attributes.OidcGrantEnabled, "oidc grant should be disabled")
	require.True(t, utils.IsValidLogoutRedirectUris(*client.Attributes.LogoutRedirectUris), "logout redirect uris should start with +")
	require.False(t, *client.Attributes.IsExternal, "is external should be false")
	require.Equal(t, realm_client.ClientTypeAgent, *client.Attributes.ClientType, "client type should be agent")
	require.NotNil(t, client.Attributes.UpdatedAt, "update time should not be nil")
	require.NotNil(t, client.Attributes.CreatedAt, "creation time should not be nil")
}
