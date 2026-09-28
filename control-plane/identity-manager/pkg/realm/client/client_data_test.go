/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
)

const (
	testRealm        = "test-realm"
	testTenantId     = "1001"
	testAppHost      = "https://mock-q7z3k.example.com"
	testHostTemplate = "http://{{.}}.run.ai"
	testTenantUrl    = "http://test-realm.run.ai"
)

func defaultRedirectURIs() []string {
	return []string{
		testTenantUrl + "/login/callback",
		testTenantUrl + "/login/callback?" + utils.RedirectFlowQueryKey + "=" + string(utils.RedirectFlowInvite),
		testTenantUrl + "/oauth-code",
		testTenantUrl,
		testAppHost + "/login/callback",
		testAppHost + "/oauth-code",
		"http://localhost",
	}
}

func adminUiRedirectURIs() []string {
	return append(defaultRedirectURIs(), "http://localhost/login/callback")
}

// RunaiClient

func TestRunaiClient_GetClientID(t *testing.T) {
	c := NewRunaiClient(testRealm, testTenantId, testAppHost, testHostTemplate)
	assert.Equal(t, utils.RUNAI_CLIENT_ID, c.GetClientID())
}

func TestRunaiClient_GetClientData_NilExisting(t *testing.T) {
	c := NewRunaiClient(testRealm, testTenantId, testAppHost, testHostTemplate)
	result := c.GetClientData(context.Background(), nil)

	require.NotNil(t, result)
	assert.Equal(t, utils.RUNAI_CLIENT_ID, *result.ClientID)
	assert.Equal(t, testTenantId, *result.TenantId)
	assert.True(t, *result.Enabled)
	assert.Equal(t, ClientTypeRunai, *result.Attributes.ClientType)
	assert.Equal(t, []string{testTenantUrl, testAppHost}, *result.WebOrigins)
	assert.Equal(t, defaultRedirectURIs(), *result.RedirectURIs)
	assert.False(t, *result.ClientConfig.PublicClient)
	assert.False(t, *result.ClientConfig.DirectAccessGrantsEnabled)
	assert.False(t, *result.ClientConfig.StandardFlowEnabled)
	assert.True(t, *result.ClientConfig.ServiceAccountsEnabled)
}

func TestRunaiClient_GetClientData_ExistingClient(t *testing.T) {
	c := NewRunaiClient(testRealm, testTenantId, testAppHost, testHostTemplate)
	existing := &iamModels.Client{
		ID:     ptr.To("existing-id"),
		Secret: ptr.To("existing-secret"),
	}
	result := c.GetClientData(context.Background(), existing)

	require.NotNil(t, result)
	assert.Equal(t, "existing-id", *result.ID)
	assert.Equal(t, "existing-secret", *result.Secret)
	assert.Equal(t, utils.RUNAI_CLIENT_ID, *result.ClientID)
	assert.Equal(t, ClientTypeRunai, *result.Attributes.ClientType)
}

// AdminUiClient

func TestAdminUiClient_GetClientID(t *testing.T) {
	c := NewAdminUiClient(testRealm, testTenantId, testAppHost, testHostTemplate)
	assert.Equal(t, utils.RUNAI_ADMIN_UI_CLIENT_ID, c.GetClientID())
}

func TestAdminUiClient_GetClientData_NilExisting(t *testing.T) {
	c := NewAdminUiClient(testRealm, testTenantId, testAppHost, testHostTemplate)
	result := c.GetClientData(context.Background(), nil)

	require.NotNil(t, result)
	assert.Equal(t, utils.RUNAI_ADMIN_UI_CLIENT_ID, *result.ClientID)
	assert.Equal(t, testTenantId, *result.TenantId)
	assert.True(t, *result.Enabled)
	assert.Equal(t, ClientTypeRunaiAdminUI, *result.Attributes.ClientType)
	assert.Equal(t, "S256", *result.Attributes.PkceCodeChallengeMethod)
	assert.Equal(t, []string{testTenantUrl, testAppHost}, *result.WebOrigins)
	assert.Equal(t, adminUiRedirectURIs(), *result.RedirectURIs)
	assert.False(t, *result.ClientConfig.PublicClient)
	assert.False(t, *result.ClientConfig.DirectAccessGrantsEnabled)
	assert.True(t, *result.ClientConfig.StandardFlowEnabled)
	assert.False(t, *result.ClientConfig.ServiceAccountsEnabled)
}

func TestAdminUiClient_GetClientData_ExistingClient(t *testing.T) {
	c := NewAdminUiClient(testRealm, testTenantId, testAppHost, testHostTemplate)
	existing := &iamModels.Client{
		ID:     ptr.To("existing-id"),
		Secret: ptr.To("existing-secret"),
	}
	result := c.GetClientData(context.Background(), existing)

	require.NotNil(t, result)
	assert.Equal(t, "existing-id", *result.ID)
	assert.Equal(t, "existing-secret", *result.Secret)
	assert.Equal(t, utils.RUNAI_ADMIN_UI_CLIENT_ID, *result.ClientID)
	assert.Equal(t, ClientTypeRunaiAdminUI, *result.Attributes.ClientType)
}

// RunaiCliClient

func TestRunaiCliClient_GetClientID(t *testing.T) {
	c := NewRunaiCliClient(testRealm, testTenantId, testAppHost, testHostTemplate)
	assert.Equal(t, utils.RUNAI_CLI_CLIENT_ID, c.GetClientID())
}

func TestRunaiCliClient_GetClientData_NilExisting(t *testing.T) {
	c := NewRunaiCliClient(testRealm, testTenantId, testAppHost, testHostTemplate)
	result := c.GetClientData(context.Background(), nil)

	require.NotNil(t, result)
	assert.Equal(t, utils.RUNAI_CLI_CLIENT_ID, *result.ClientID)
	assert.Equal(t, testTenantId, *result.TenantId)
	assert.True(t, *result.Enabled)
	assert.Equal(t, ClientTypeRunaiCLI, *result.Attributes.ClientType)
	assert.Equal(t, []string{testTenantUrl, testAppHost}, *result.WebOrigins)
	assert.Equal(t, defaultRedirectURIs(), *result.RedirectURIs)
	assert.True(t, *result.ClientConfig.PublicClient)
	assert.True(t, *result.ClientConfig.DirectAccessGrantsEnabled)
	assert.True(t, *result.ClientConfig.StandardFlowEnabled)
	assert.False(t, *result.ClientConfig.ServiceAccountsEnabled)
}

func TestRunaiCliClient_GetClientData_ExistingClient(t *testing.T) {
	c := NewRunaiCliClient(testRealm, testTenantId, testAppHost, testHostTemplate)
	existing := &iamModels.Client{
		ID:     ptr.To("existing-id"),
		Secret: ptr.To("existing-secret"),
	}
	result := c.GetClientData(context.Background(), existing)

	require.NotNil(t, result)
	assert.Equal(t, "existing-id", *result.ID)
	assert.Equal(t, "existing-secret", *result.Secret)
	assert.Equal(t, utils.RUNAI_CLI_CLIENT_ID, *result.ClientID)
	assert.Equal(t, ClientTypeRunaiCLI, *result.Attributes.ClientType)
}

// CLIClient

func TestCLIClient_GetClientID(t *testing.T) {
	c := NewCLIClient(testRealm, testTenantId, testAppHost, testHostTemplate)
	assert.Equal(t, utils.CLI_CLIENT_ID, c.GetClientID())
}

func TestCLIClient_GetClientData_NilExisting(t *testing.T) {
	c := NewCLIClient(testRealm, testTenantId, testAppHost, testHostTemplate)
	result := c.GetClientData(context.Background(), nil)

	require.NotNil(t, result)
	assert.Equal(t, utils.CLI_CLIENT_ID, *result.ClientID)
	assert.Equal(t, testTenantId, *result.TenantId)
	assert.True(t, *result.Enabled)
	assert.Equal(t, ClientTypeCLI, *result.Attributes.ClientType)
	assert.Equal(t, []string{testTenantUrl, testAppHost}, *result.WebOrigins)
	assert.Equal(t, defaultRedirectURIs(), *result.RedirectURIs)
	assert.False(t, *result.ClientConfig.PublicClient)
	assert.True(t, *result.ClientConfig.DirectAccessGrantsEnabled)
	assert.True(t, *result.ClientConfig.StandardFlowEnabled)
	assert.False(t, *result.ClientConfig.ServiceAccountsEnabled)
}

func TestCLIClient_GetClientData_ExistingClient(t *testing.T) {
	c := NewCLIClient(testRealm, testTenantId, testAppHost, testHostTemplate)
	existing := &iamModels.Client{
		ID:     ptr.To("existing-id"),
		Secret: ptr.To("existing-secret"),
	}
	result := c.GetClientData(context.Background(), existing)

	require.NotNil(t, result)
	assert.Equal(t, "existing-id", *result.ID)
	assert.Equal(t, "existing-secret", *result.Secret)
	assert.Equal(t, utils.CLI_CLIENT_ID, *result.ClientID)
	assert.Equal(t, ClientTypeCLI, *result.Attributes.ClientType)
}

// TokenExchangeClient

func TestTokenExchangeClient_GetClientID(t *testing.T) {
	c := NewTokenExchangeClient(testRealm, testTenantId, testAppHost, testHostTemplate)
	assert.Equal(t, utils.RUNAI_TOKEN_EXCHANGE_CLIENT_ID, c.GetClientID())
}

func TestTokenExchangeClient_GetClientData_NilExisting(t *testing.T) {
	c := NewTokenExchangeClient(testRealm, testTenantId, testAppHost, testHostTemplate)
	result := c.GetClientData(context.Background(), nil)

	require.NotNil(t, result)
	assert.Equal(t, utils.RUNAI_TOKEN_EXCHANGE_CLIENT_ID, *result.ClientID)
	assert.Equal(t, testTenantId, *result.TenantId)
	assert.True(t, *result.Enabled)
	assert.Equal(t, ClientTypeTokenExchange, *result.Attributes.ClientType)
	assert.Equal(t, []string{testTenantUrl, testAppHost}, *result.WebOrigins)
	assert.Equal(t, defaultRedirectURIs(), *result.RedirectURIs)
	assert.False(t, *result.ClientConfig.PublicClient)
	assert.False(t, *result.ClientConfig.DirectAccessGrantsEnabled)
	assert.True(t, *result.ClientConfig.StandardFlowEnabled)
	assert.False(t, *result.ClientConfig.ServiceAccountsEnabled)
	assert.Equal(t, 10*time.Minute, *result.Attributes.TokenLifespan)
}

func TestTokenExchangeClient_GetClientData_ExistingClient(t *testing.T) {
	c := NewTokenExchangeClient(testRealm, testTenantId, testAppHost, testHostTemplate)
	existing := &iamModels.Client{
		ID:     ptr.To("existing-id"),
		Secret: ptr.To("existing-secret"),
	}
	result := c.GetClientData(context.Background(), existing)

	require.NotNil(t, result)
	assert.Equal(t, "existing-id", *result.ID)
	assert.Equal(t, "existing-secret", *result.Secret)
	assert.Equal(t, utils.RUNAI_TOKEN_EXCHANGE_CLIENT_ID, *result.ClientID)
	assert.Equal(t, ClientTypeTokenExchange, *result.Attributes.ClientType)
	assert.Equal(t, 10*time.Minute, *result.Attributes.TokenLifespan)
}

// WorkloadsAuthClient

func TestWorkloadsAuthClient_GetClientID(t *testing.T) {
	c := NewWorkloadsAuthClient("workloads-client", testTenantId, "cluster-1", testAppHost, testHostTemplate)
	assert.Equal(t, "workloads-client", c.GetClientID())
}

func TestWorkloadsAuthClient_GetClientData_NilExisting(t *testing.T) {
	// WorkloadsAuthClient has no realm parameter, so TemplateBaseUrl uses an empty realm.
	const workloadsTenantUrl = "http://.run.ai"
	c := NewWorkloadsAuthClient("workloads-client", testTenantId, "cluster-1", testAppHost, testHostTemplate)
	result := c.GetClientData(context.Background(), nil)

	require.NotNil(t, result)
	assert.Equal(t, "workloads-client", *result.ClientID)
	assert.Equal(t, testTenantId, *result.TenantId)
	assert.True(t, *result.Enabled)
	assert.Equal(t, ClientTypeWorkloadsAuth, *result.Attributes.ClientType)
	assert.Equal(t, "cluster-1", *result.Attributes.ClusterId)
	assert.False(t, *result.ClientConfig.PublicClient)
	assert.False(t, *result.ClientConfig.DirectAccessGrantsEnabled)
	assert.True(t, *result.ClientConfig.StandardFlowEnabled)
	assert.False(t, *result.ClientConfig.ServiceAccountsEnabled)

	baseRedirectURIs := *getDefaultRedirectUris(workloadsTenantUrl, testAppHost, nil)
	expectedRedirectURIs := append(baseRedirectURIs, "*")
	assert.Equal(t, expectedRedirectURIs, *result.RedirectURIs)
}

func TestWorkloadsAuthClient_GetClientData_ExistingClient(t *testing.T) {
	c := NewWorkloadsAuthClient("workloads-client", testTenantId, "cluster-1", testAppHost, testHostTemplate)
	existing := &iamModels.Client{
		ID:     ptr.To("existing-id"),
		Secret: ptr.To("existing-secret"),
	}
	result := c.GetClientData(context.Background(), existing)

	require.NotNil(t, result)
	assert.Equal(t, "existing-id", *result.ID)
	assert.Equal(t, "existing-secret", *result.Secret)
	assert.Equal(t, "workloads-client", *result.ClientID)
	assert.Equal(t, ClientTypeWorkloadsAuth, *result.Attributes.ClientType)
	assert.Equal(t, "cluster-1", *result.Attributes.ClusterId)
}

// RunaiApplicationClient

func TestRunaiApplicationClient_GetClientID(t *testing.T) {
	c := NewRunaiServiceAccountClient("my-app", testTenantId, nil, ClientTypeUserApp)
	assert.Equal(t, "my-app", c.GetClientID())
}

func TestRunaiApplicationClient_GetClientData_UserApp(t *testing.T) {
	c := NewRunaiServiceAccountClient("my-app", testTenantId, nil, ClientTypeUserApp)
	result := c.GetClientData(context.Background(), nil)

	require.NotNil(t, result)
	assert.Equal(t, "my-app", *result.ClientID)
	assert.Equal(t, testTenantId, *result.TenantId)
	assert.Equal(t, "my-app", *result.Name)
	assert.Equal(t, ClientTypeUserApp, *result.Attributes.ClientType)
	assert.False(t, *result.ClientConfig.PublicClient)
	assert.False(t, *result.ClientConfig.DirectAccessGrantsEnabled)
	assert.True(t, *result.ClientConfig.StandardFlowEnabled)
	assert.True(t, *result.ClientConfig.ServiceAccountsEnabled)
}

func TestRunaiApplicationClient_GetClientData_AgentApp(t *testing.T) {
	clusterId := ptr.To("cluster-1")
	c := NewRunaiServiceAccountClient("my-agent", testTenantId, clusterId, ClientTypeAgent)
	result := c.GetClientData(context.Background(), nil)

	require.NotNil(t, result)
	assert.Equal(t, "my-agent", *result.ClientID)
	assert.Equal(t, testTenantId, *result.TenantId)
	assert.Equal(t, ClientTypeAgent, *result.Attributes.ClientType)
	assert.Equal(t, "cluster-1", *result.Attributes.ClusterId)
	assert.False(t, *result.Attributes.IsExternal)
	assert.False(t, *result.ClientConfig.PublicClient)
	assert.False(t, *result.ClientConfig.DirectAccessGrantsEnabled)
	assert.True(t, *result.ClientConfig.StandardFlowEnabled)
	assert.True(t, *result.ClientConfig.ServiceAccountsEnabled)
}

func TestRunaiApplicationClient_GetClientData_AgentApp_NoCluster(t *testing.T) {
	c := NewRunaiServiceAccountClient("my-agent", testTenantId, nil, ClientTypeAgent)
	result := c.GetClientData(context.Background(), nil)

	require.NotNil(t, result)
	assert.Equal(t, ClientTypeAgent, *result.Attributes.ClientType)
	assert.Nil(t, result.Attributes.ClusterId)
}

func TestRunaiApplicationClient_GetClientData_InstallerApp(t *testing.T) {
	c := NewRunaiServiceAccountClient("my-installer", testTenantId, nil, ClientTypeInstaller)
	result := c.GetClientData(context.Background(), nil)

	require.NotNil(t, result)
	assert.Equal(t, "my-installer", *result.ClientID)
	assert.Equal(t, testTenantId, *result.TenantId)
	assert.Equal(t, ClientTypeInstaller, *result.Attributes.ClientType)
	assert.False(t, *result.Attributes.IsExternal)
	assert.False(t, *result.ClientConfig.PublicClient)
	assert.False(t, *result.ClientConfig.DirectAccessGrantsEnabled)
	assert.True(t, *result.ClientConfig.StandardFlowEnabled)
	assert.True(t, *result.ClientConfig.ServiceAccountsEnabled)
}

func TestRunaiApplicationClient_GetClientData_TenantApp(t *testing.T) {
	c := NewRunaiServiceAccountClient("my-tenant-app", testTenantId, nil, ClientTypeTenantApp)
	result := c.GetClientData(context.Background(), nil)

	require.NotNil(t, result)
	assert.Equal(t, "my-tenant-app", *result.ClientID)
	assert.Equal(t, testTenantId, *result.TenantId)
	assert.Equal(t, ClientTypeTenantApp, *result.Attributes.ClientType)
	assert.True(t, *result.Attributes.IsExternal)
	assert.False(t, *result.ClientConfig.PublicClient)
	assert.False(t, *result.ClientConfig.DirectAccessGrantsEnabled)
	assert.True(t, *result.ClientConfig.StandardFlowEnabled)
	assert.True(t, *result.ClientConfig.ServiceAccountsEnabled)
}

func TestRunaiApplicationClient_GetClientData_PreservesExistingName(t *testing.T) {
	c := NewRunaiServiceAccountClient("my-app", testTenantId, nil, ClientTypeUserApp)
	existing := &iamModels.Client{
		Name: ptr.To("custom-name"),
	}
	result := c.GetClientData(context.Background(), existing)

	require.NotNil(t, result)
	assert.Equal(t, "custom-name", *result.Name)
}

// McpClient

func TestMcpClient_GetClientID(t *testing.T) {
	c := NewMcpClient(testRealm, testTenantId)
	assert.Equal(t, utils.RUNAI_MCP_CLIENT_ID, c.GetClientID())
}

func TestMcpClient_GetClientData(t *testing.T) {
	c := NewMcpClient(testRealm, testTenantId)
	result := c.GetClientData(context.Background(), nil)

	require.NotNil(t, result)
	assert.Equal(t, utils.RUNAI_MCP_CLIENT_ID, *result.ClientID)
	assert.Equal(t, testTenantId, *result.TenantId)
	assert.True(t, *result.Enabled)
	assert.Equal(t, ClientTypeMCP, *result.Attributes.ClientType)
	assert.Equal(t, mcpRedirectURIs, *result.RedirectURIs)
	assert.Nil(t, result.WebOrigins)
	assert.False(t, *result.ClientConfig.PublicClient)
	assert.False(t, *result.ClientConfig.DirectAccessGrantsEnabled)
	assert.True(t, *result.ClientConfig.StandardFlowEnabled)
	assert.False(t, *result.ClientConfig.ServiceAccountsEnabled)
	assert.Equal(t, "S256", *result.Attributes.PkceCodeChallengeMethod)
}

// BuildClientData error path

func TestBuildClientData_TemplateError(t *testing.T) {
	result := BuildClientData(context.Background(), ClientTypeRunaiAdminUI, ClientParams{
		Realm:        testRealm,
		TenantId:     testTenantId,
		HostTemplate: "{{invalid",
	})
	assert.Nil(t, result)
}
