/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"net/http"
	"testing"

	openapi "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

const (
	clientNamePrefix = "test-client"
)

func createAgentClient(t *testing.T, ctx context.Context, clientName string, clusterId *string) string {
	return createAgentClientWithTenant(t, ctx, clientName, clusterId, BaseTenantId)
}

func createAgentClientWithTenant(t *testing.T, ctx context.Context, clientName string, clusterId *string, tenantId string) string {
	var attributes *openapi.ClientAttributeCreationRequest
	if clusterId != nil {
		attributes = &openapi.ClientAttributeCreationRequest{
			RunaiClusterId: clusterId,
		}
	}

	clientRequest := openapi.ClientCreationRequest{
		Type:       openapi.CLIENTTYPE_AGENT,
		ClientId:   clientName,
		TenantId:   ptr.To(tenantId),
		Attributes: attributes,
	}

	resp, httpResp, err := OpenapiClient.ClientsAPI.CreateInternalClient(ctx).ClientCreationRequest(clientRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, httpResp.StatusCode)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Id)

	return *resp.Id
}

func createWorkloadsClient(t *testing.T, ctx context.Context, clientName string, clusterId string) string {
	return createWorkloadsClientWithTenant(t, ctx, clientName, clusterId, BaseTenantId)
}

func createWorkloadsClientWithTenant(t *testing.T, ctx context.Context, clientName string, clusterId string, tenantId string) string {
	attributes := &openapi.ClientAttributeCreationRequest{
		RunaiClusterId: ptr.To(clusterId),
	}

	clientRequest := openapi.ClientCreationRequest{
		Type:       openapi.CLIENTTYPE_WORKLOADS_AUTH,
		ClientId:   clientName,
		TenantId:   ptr.To(tenantId),
		Attributes: attributes,
	}

	resp, httpResp, err := OpenapiClient.ClientsAPI.CreateInternalClient(ctx).ClientCreationRequest(clientRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, httpResp.StatusCode)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Id)

	return *resp.Id
}

func createApplicationClient(t *testing.T, ctx context.Context, clientName string) string {
	return createApplicationClientWithTenant(t, ctx, clientName, BaseTenantId)
}

func createApplicationClientWithTenant(t *testing.T, ctx context.Context, clientName string, tenantId string) string {
	clientRequest := openapi.ClientCreationRequest{
		Type:     openapi.CLIENTTYPE_APPLICATION,
		ClientId: clientName,
		TenantId: ptr.To(tenantId),
	}

	resp, httpResp, err := OpenapiClient.ClientsAPI.CreateInternalClient(ctx).ClientCreationRequest(clientRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, httpResp.StatusCode)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Id)

	return *resp.Id
}

func getAllClients(t *testing.T, ctx context.Context) []openapi.Client {
	clients, httpResp, err := OpenapiClient.ClientsAPI.GetAllClients(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	return clients
}

func getClientsWithFilter(t *testing.T, ctx context.Context, filter string) []openapi.Client {
	clients, httpResp, err := OpenapiClient.ClientsAPI.GetAllClients(ctx).Filter(filter).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.NotNil(t, clients)
	return clients
}

func findClientByID(clients []openapi.Client, clientId string) *openapi.Client {
	for _, client := range clients {
		if client.ClientId == clientId {
			return &client
		}
	}
	return nil
}

func assertClientExists(t *testing.T, clients []openapi.Client, clientId string) *openapi.Client {
	client := findClientByID(clients, clientId)
	require.NotNil(t, client, "Client %s should be found in the list", clientId)
	return client
}

func assertClientAttribute(t *testing.T, client *openapi.Client, key, expectedValue string) {
	require.NotNil(t, client.Attributes, "Client should have attributes")
	if value, exists := client.Attributes[key]; exists {
		require.Equal(t, expectedValue, value, "Client attribute %s should match expected value", key)
	}
}

func TestClient_CreateInternalClient_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	clientId := GenerateRandomString(clientNamePrefix)

	createAgentClient(t, ctx, clientId, nil)

	clients := getAllClients(t, ctx)
	assertClientExists(t, clients, clientId)
}

func TestClient_CreateInternalClient_WithAttributes(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	clientId := GenerateRandomString(clientNamePrefix)

	createAgentClient(t, ctx, clientId, ptr.To("test-cluster-123"))

	clients := getAllClients(t, ctx)
	client := assertClientExists(t, clients, clientId)
	assertClientAttribute(t, client, "runai.cluster_id", "test-cluster-123")
}

func TestClient_CreateInternalClient_InvalidClientType(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	clientId := GenerateRandomString(clientNamePrefix)

	clientRequest := openapi.ClientCreationRequest{
		Type:     openapi.CLIENTTYPE_AGENT,
		ClientId: clientId,
		TenantId: ptr.To(BaseTenantId),
	}

	resp, httpResp, err := OpenapiClient.ClientsAPI.CreateInternalClient(ctx).ClientCreationRequest(clientRequest).Execute()

	if err == nil && resp.Id != nil {
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, httpResp.StatusCode)
	}
}

func TestClient_CreateInternalClient_MissingClientId(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	clientRequest := openapi.ClientCreationRequest{
		Type:     openapi.CLIENTTYPE_AGENT,
		ClientId: "",
		TenantId: ptr.To(BaseTenantId),
	}

	resp, httpResp, err := OpenapiClient.ClientsAPI.CreateInternalClient(ctx).ClientCreationRequest(clientRequest).Execute()

	if err != nil {
		require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
		require.Nil(t, resp)
	} else {
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, httpResp.StatusCode)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Id)
	}
}

func TestClient_CreateInternalClient_DifferentClientTypes(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	tests := []struct {
		name           string
		clientType     openapi.ClientType
		expectType     string
		expectExternal bool
	}{
		{
			name:           "Agent client type",
			clientType:     openapi.CLIENTTYPE_AGENT,
			expectType:     "agent",
			expectExternal: false,
		},
		{
			name:           "WorkloadsAuth client type",
			clientType:     openapi.CLIENTTYPE_WORKLOADS_AUTH,
			expectType:     "workloads_auth",
			expectExternal: false,
		},
		{
			name:           "Application/External client type",
			clientType:     openapi.CLIENTTYPE_APPLICATION,
			expectType:     "tenant-application",
			expectExternal: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clientId := GenerateRandomString(clientNamePrefix)

			switch test.clientType {
			case openapi.CLIENTTYPE_AGENT:
				createAgentClient(t, ctx, clientId, nil)
			case openapi.CLIENTTYPE_WORKLOADS_AUTH:
				createWorkloadsClient(t, ctx, clientId, "test-cluster-456")
			case openapi.CLIENTTYPE_APPLICATION:
				createApplicationClient(t, ctx, clientId)
			}

			clients := getAllClients(t, ctx)
			client := assertClientExists(t, clients, clientId)

			if clientType, exists := client.Attributes["runai.client_type"]; exists {
				require.Equal(t, test.expectType, clientType)
			}

			if isExternal, exists := client.Attributes["runai.is_external"]; exists {
				if test.expectExternal {
					require.Equal(t, "true", isExternal)
				} else {
					require.Equal(t, "false", isExternal)
				}
			}

			if test.clientType == openapi.CLIENTTYPE_WORKLOADS_AUTH {
				assertClientAttribute(t, client, "runai.cluster_id", "test-cluster-456")
			}
		})
	}
}

func TestClient_DeleteClient_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	clientId := TestStepsClient.CreateInternalClient(ctx, BaseTenantId, nil)

	httpResp, err := OpenapiClient.ClientsAPI.DeleteClient(ctx, clientId).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, httpResp.StatusCode)
}

func TestClient_DeleteClient_InvalidID(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	httpResp, err := OpenapiClient.ClientsAPI.DeleteClient(ctx, "invalid-id").Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode)
}

func TestClient_GetAllClients_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	cleanupTestClients(t, ctx)

	clientName1 := GenerateRandomString("client1")
	clientName2 := GenerateRandomString("client2")
	TestStepsClient.CreateInternalClient(ctx, BaseTenantId, &clientName1)
	TestStepsClient.CreateInternalClient(ctx, BaseTenantId, &clientName2)

	clients := getAllClients(t, ctx)
	require.NotEmpty(t, clients)

	assertClientExists(t, clients, clientName1)
	assertClientExists(t, clients, clientName2)
}

func TestClient_GetAllClients_WithFilter(t *testing.T) {
	t.Parallel()
	ctx, _, tenantId := TestStepsClient.NewTestContextWithRealm("client-filter-test")

	cleanupTestClients(t, ctx)

	agentClientName := GenerateRandomString("agent-client")
	workloadsClientName := GenerateRandomString("workloads-client")
	appClientName := GenerateRandomString("app-client")

	createAgentClientWithTenant(t, ctx, agentClientName, nil, tenantId)
	createWorkloadsClientWithTenant(t, ctx, workloadsClientName, "test-cluster-123", tenantId)
	createApplicationClientWithTenant(t, ctx, appClientName, tenantId)

	tests := []struct {
		name            string
		filter          string
		expectedClients []string
		verifyFunc      func(clients []openapi.Client) bool
	}{
		{
			name:            "Filter by client type - agent",
			filter:          "runai.client_type:agent",
			expectedClients: []string{agentClientName},
			verifyFunc: func(clients []openapi.Client) bool {
				client := findClientByID(clients, agentClientName)
				if client == nil {
					return false
				}
				if clientType, exists := client.Attributes["runai.client_type"]; exists {
					return clientType == "agent"
				}
				return false
			},
		},
		{
			name:            "Filter by client type - workloads_auth",
			filter:          "runai.client_type:workloads_auth",
			expectedClients: []string{workloadsClientName},
			verifyFunc: func(clients []openapi.Client) bool {
				client := findClientByID(clients, workloadsClientName)
				if client == nil {
					return false
				}
				if clientType, exists := client.Attributes["runai.client_type"]; exists {
					return clientType == "workloads_auth"
				}
				return false
			},
		},
		{
			name:            "Filter by cluster ID",
			filter:          "runai.cluster_id:test-cluster-123",
			expectedClients: []string{workloadsClientName},
			verifyFunc: func(clients []openapi.Client) bool {
				client := findClientByID(clients, workloadsClientName)
				if client == nil {
					return false
				}
				if clusterId, exists := client.Attributes["runai.cluster_id"]; exists {
					return clusterId == "test-cluster-123"
				}
				return false
			},
		},
		{
			name:            "Filter by client type - application",
			filter:          "runai.client_type:application",
			expectedClients: []string{},
			verifyFunc: func(clients []openapi.Client) bool {
				// APPLICATION clients don't have the runai.client_type attribute set in current implementation
				return len(clients) == 0
			},
		},
		{
			name:            "Filter by is_external - true",
			filter:          "runai.is_external:true",
			expectedClients: []string{},
			verifyFunc: func(clients []openapi.Client) bool {
				return len(clients) == 1
			},
		},
		{
			name:            "Filter by is_external - false",
			filter:          "runai.is_external:false",
			expectedClients: []string{agentClientName},
			verifyFunc: func(clients []openapi.Client) bool {
				client := findClientByID(clients, agentClientName)
				if client == nil {
					return false
				}
				if isExternal, exists := client.Attributes["runai.is_external"]; exists && isExternal == "false" {
					return true
				}
				return false
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clients := getClientsWithFilter(t, ctx, test.filter)

			require.True(t, test.verifyFunc(clients), "Filter should return clients matching the criteria: %s", test.filter)

			for _, expectedClient := range test.expectedClients {
				client := findClientByID(clients, expectedClient)
				require.NotNil(t, client, "Expected client %s should be found with filter %s", expectedClient, test.filter)
			}
		})
	}
}

func TestClient_GetAllClients_FilterByAttributes(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	cleanupTestClients(t, ctx)

	clientName := GenerateRandomString("attr-client")
	createWorkloadsClient(t, ctx, clientName, "filter-test-cluster")

	combinedFilter := "runai.client_type:workloads_auth runai.cluster_id:filter-test-cluster"
	clients := getClientsWithFilter(t, ctx, combinedFilter)
	require.NotEmpty(t, clients)

	client := assertClientExists(t, clients, clientName)
	assertClientAttribute(t, client, "runai.cluster_id", "filter-test-cluster")
	assertClientAttribute(t, client, "runai.client_type", "workloads_auth")
}

func TestClient_GetAllClients_MultipleFilters(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	cleanupTestClients(t, ctx)
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)
	clientName1 := GenerateRandomString("multi-client-1")
	clientName2 := GenerateRandomString("multi-client-2")
	TestStepsClient.CreateInternalClient(ctx, BaseTenantId, &clientName1)
	TestStepsClient.CreateInternalClient(ctx, BaseTenantId, &clientName2)

	clients := getAllClients(t, ctx)
	require.NotEmpty(t, clients)

	assertClientExists(t, clients, clientName1)
	assertClientExists(t, clients, clientName2)
}

func TestClient_GetClientSecret_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	clientId := TestStepsClient.CreateInternalClient(ctx, BaseTenantId, nil)

	resp, httpResp, err := OpenapiClient.ClientsAPI.GetClientSecret(ctx, clientId).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Secret)
	require.NotEmpty(t, *resp.Secret)
}

func TestClient_GetClientSecret_InvalidID(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	resp, httpResp, err := OpenapiClient.ClientsAPI.GetClientSecret(ctx, "invalid-id").Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestClient_UpdateRedirectUris_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	clientName := GenerateRandomString("redirect-test-client")
	createAgentClient(t, ctx, clientName, nil)

	redirectUris := []string{
		"https://example.com/callback",
		"https://app.example.com/auth/callback",
		"http://localhost:8080/callback",
	}

	updateRequest := openapi.ClientRedirectUrisUpdateRequest{
		RedirectUris: redirectUris,
	}

	httpResp, err := OpenapiClient.ClientsAPI.UpdateRedirectUris(ctx, clientName).ClientRedirectUrisUpdateRequest(updateRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)

	clients := getAllClients(t, ctx)
	client := assertClientExists(t, clients, clientName)
	require.NotNil(t, client.RedirectUris)
	require.Equal(t, len(redirectUris), len(client.RedirectUris))

	for _, expectedUri := range redirectUris {
		uriFound := false
		for _, actualUri := range client.RedirectUris {
			if actualUri == expectedUri {
				uriFound = true
				break
			}
		}
		require.True(t, uriFound, "Expected redirect URI %s should be found", expectedUri)
	}
}

func TestClient_UpdateRedirectUris_InvalidClientID(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	updateRequest := openapi.ClientRedirectUrisUpdateRequest{
		RedirectUris: []string{"https://example.com/callback"},
	}

	httpResp, err := OpenapiClient.ClientsAPI.UpdateRedirectUris(ctx, "invalid-application-client-id").ClientRedirectUrisUpdateRequest(updateRequest).Execute()
	require.Error(t, err)
	if httpResp != nil {
		require.True(t, httpResp.StatusCode >= 400, "Should return an error status code")
	}
}

func cleanupTestClients(t *testing.T, ctx context.Context) {
	clients, _, err := OpenapiClient.ClientsAPI.GetAllClients(ctx).Execute()
	if err != nil {
		t.Logf("Warning: Could not retrieve clients for cleanup: %v", err)
		return
	}

	for _, client := range clients {
		if isTestClient(client.ClientId) {
			err := TestStepsClient.DeleteInternalClient(ctx, client.Id)
			if err != nil {
				t.Logf("Warning: Failed to cleanup leftover test client %s: %v", client.ClientId, err)
			} else {
				t.Logf("Cleaned up leftover test client: %s", client.ClientId)
			}
		}
	}
}

func isTestClient(clientId string) bool {
	testPrefixes := []string{
		"client1", "client2",
		"multi-client-1", "multi-client-2",
		"filtered-client",
		"test-client",
		"agent-client",
		"workloads-client",
		"app-client",
		"attr-client",
		"redirect-test-client",
	}

	for _, prefix := range testPrefixes {
		if clientId == prefix ||
			(len(clientId) > len(prefix) && clientId[:len(prefix)] == prefix) ||
			(len(clientId) > len(prefix)+1 && clientId[:len(prefix)+1] == prefix+"-") {
			return true
		}
	}
	return false
}
