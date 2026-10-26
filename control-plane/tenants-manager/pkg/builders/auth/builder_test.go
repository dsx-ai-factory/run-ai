/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authMock "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/mocks"
	idmopenapi "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	idmMock "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/builders"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/config"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getExistingClients() []idmopenapi.Client {
	return []idmopenapi.Client{
		*idmopenapi.NewClient(AgentClientName+NAME, "agentClientName", "agentSecret"),
		*idmopenapi.NewClient(AgentClientName+NAME+"-2", "agentClientName2", "agentSecret2"),
	}
}

const (
	RANDOM_STRING = "123"
	TOKEN         = "Bearer fakeToken"
)

// user data
const (
	NAME     = "tenantName"
	ID       = 12345
	EMAIL    = "adminuser@run.ai"
	PASSWORD = "password"
)

func initTest(t *testing.T, identityServiceUrl string) (testSubject *Builder, authClientMock *authMock.IAuthorizationClient, idmClient *idmMock.IIdm) {
	authClientMock = authMock.NewIAuthorizationClient(t)
	idmClient = idmMock.NewIIdm(t)
	testSubject = &Builder{
		Auth: &config.Auth{
			ManagementRealm:    RANDOM_STRING,
			ManagementGroups:   nil,
			IdentityServiceUrl: identityServiceUrl,
			ServiceUrl:         RANDOM_STRING,
		},
		httpClient: resty.New(),
		authClient: authClientMock,
		idmClient:  idmClient,
	}
	return
}

func TestConstructWithAuthClient_HappyFlow(t *testing.T) {
	fakeIdentityServer := getFakeIdentityServer()
	defer fakeIdentityServer.Close()

	testSubject, authClientMock, idmClient := initTest(t, fakeIdentityServer.URL)

	builderResponse := make(map[string]interface{})
	context := createFakeAuthCtx()

	httpOK := getHttpRespone(http.StatusOK)
	realmResp := idmopenapi.NewRealmCreationResponse(NAME, "Domain", []idmopenapi.Client{
		*idmopenapi.NewClient("clientId", "clientName", "clientSecret"),
	})
	idmClient.On("GetOrCreateRealm", mock.Anything, NAME, fmt.Sprintf("%d", ID)).Return(realmResp, httpOK, nil)
	idmClient.On("DeleteInternalUserByEmail", mock.Anything, NAME, EMAIL).Return(httpOK, nil)
	idmClient.On("CreateInternalUser", mock.Anything, mock.Anything).Return(
		idmopenapi.NewUserPostResponse(fmt.Sprintf("%d", ID), EMAIL, PASSWORD), httpOK, nil)

	authClientMock.On("CreateInternalAccessRule", mock.Anything, mock.Anything, mock.Anything).Return(getAdminUserRequest())
	authClientMock.On("CreateInternalAccessRuleExecute", mock.Anything).Return(nil, httpOK, nil)

	err := testSubject.Construct(context, getTenantInfo(), builderResponse)

	assert.Nil(t, err)
	assert.Len(t, builderResponse, 2)
	assert.Equal(t, "Domain", builderResponse["tenant_domain"])
	assert.Equal(t, PASSWORD, builderResponse[EMAIL])
}

func TestConstructWithAuthClient_AuthError(t *testing.T) {
	fakeIdentityServer := getFakeIdentityServer()
	defer fakeIdentityServer.Close()

	testSubject, authClientMock, idmClient := initTest(t, fakeIdentityServer.URL)

	builderResponse := make(map[string]interface{})
	context := createFakeAuthCtx()

	httpOK := getHttpRespone(http.StatusOK)
	realmResp := idmopenapi.NewRealmCreationResponse(NAME, "Domain", []idmopenapi.Client{
		*idmopenapi.NewClient("clientId", "clientName", "clientSecret"),
	})
	idmClient.On("GetOrCreateRealm", mock.Anything, NAME, fmt.Sprintf("%d", ID)).Return(realmResp, httpOK, nil)
	idmClient.On("DeleteInternalUserByEmail", mock.Anything, NAME, EMAIL).Return(httpOK, nil)
	idmClient.On("CreateInternalUser", mock.Anything, mock.Anything).Return(
		idmopenapi.NewUserPostResponse(fmt.Sprintf("%d", ID), EMAIL, PASSWORD), httpOK, nil)

	authClientMock.On("CreateInternalAccessRule", mock.Anything, mock.Anything, mock.Anything).Return(getAdminUserRequest())
	authClientMock.On("CreateInternalAccessRuleExecute", mock.Anything).Return(nil, getHttpRespone(http.StatusConflict), errors.New("conflict"))

	err := testSubject.Construct(context, getTenantInfo(), builderResponse)
	assert.Nil(t, err)
}

func TestConstructWithAuthClient_AuthNetworkError(t *testing.T) {
	fakeIdentityServer := getFakeIdentityServer()
	defer fakeIdentityServer.Close()

	testSubject, authClientMock, idmClient := initTest(t, fakeIdentityServer.URL)

	builderResponse := make(map[string]interface{})
	context := createFakeAuthCtx()

	httpOK := getHttpRespone(http.StatusOK)
	realmResp := idmopenapi.NewRealmCreationResponse(NAME, "Domain", []idmopenapi.Client{
		*idmopenapi.NewClient("clientId", "clientName", "clientSecret"),
	})
	idmClient.On("GetOrCreateRealm", mock.Anything, NAME, fmt.Sprintf("%d", ID)).Return(realmResp, httpOK, nil)
	idmClient.On("DeleteInternalUserByEmail", mock.Anything, NAME, EMAIL).Return(httpOK, nil)
	idmClient.On("CreateInternalUser", mock.Anything, mock.Anything).Return(
		idmopenapi.NewUserPostResponse(fmt.Sprintf("%d", ID), EMAIL, PASSWORD), httpOK, nil)

	authClientMock.On("CreateInternalAccessRule", mock.Anything, mock.Anything, mock.Anything).Return(getAdminUserRequest())
	authClientMock.On("CreateInternalAccessRuleExecute", mock.Anything).Return(nil, nil, errors.New("connection refused"))

	err := testSubject.Construct(context, getTenantInfo(), builderResponse)
	assert.NotNil(t, err)
}

func TestGetFirstUserRoleAndScope_SystemAdmin(t *testing.T) {
	scope := getScopeByRole(authorization_api.ALLOWEDINTERNALCREATEROLES_SYSTEM_ADMINISTRATOR)
	assert.Equal(t, authorization_api.SCOPETYPE_TENANT, scope)
}

func TestGetFirstUserRoleAndScope_CloudOperator(t *testing.T) {
	scope := getScopeByRole(authorization_api.ALLOWEDINTERNALCREATEROLES_CLOUD_OPERATOR)
	assert.Equal(t, authorization_api.SCOPETYPE_SYSTEM, scope)
}

func TestGetFirstUserRoleAndScope_UnspportedRole(t *testing.T) {
	scope := getScopeByRole(authorization_api.ALLOWEDINTERNALCREATEROLES_INSTALLER_SERVICE_ACCOUNT)
	assert.Equal(t, authorization_api.SCOPETYPE_TENANT, scope)
}

func TestDestruct_HappyFlow(t *testing.T) {
	fakeIdentityServer := getFakeIdentityServer()
	defer fakeIdentityServer.Close()
	context := createFakeAuthCtx()

	testSubject, authClientMock, idmMock := initTest(t, fakeIdentityServer.URL)
	idmMock.On("DeleteRealm", mock.Anything, NAME, mock.Anything).Return(nil, nil)
	authClientMock.On("DeleteSubjectAccessRules", mock.Anything).Return(authorization_api.ApiDeleteSubjectAccessRulesRequest{})
	authClientMock.On("DeleteSubjectAccessRulesExecute", mock.Anything).Return(nil, nil)

	err := testSubject.Destruct(context, getTenantInfo())
	assert.Nil(t, err)
}

func TestDestruct_AuthRealm_404Error(t *testing.T) {
	fakeIdentityServer := getFakeIdentityServer()
	defer fakeIdentityServer.Close()
	context := createFakeAuthCtx()

	testSubject, authClientMock, idmMock := initTest(t, fakeIdentityServer.URL)
	idmMock.On("DeleteRealm", mock.Anything, NAME, mock.Anything).Return(getHttpRespone(404), fmt.Errorf("error"))
	authClientMock.On("DeleteSubjectAccessRules", mock.Anything).Return(authorization_api.ApiDeleteSubjectAccessRulesRequest{})
	authClientMock.On("DeleteSubjectAccessRulesExecute", mock.Anything).Return(nil, nil)

	err := testSubject.Destruct(context, getTenantInfo())
	assert.Nil(t, err)
}

func TestDestruct_AuthRealm_Error(t *testing.T) {
	fakeIdentityServer := getFakeIdentityServer()
	defer fakeIdentityServer.Close()
	context := createFakeAuthCtx()

	testSubject, authClientMock, idmMock := initTest(t, fakeIdentityServer.URL)
	idmMock.On("DeleteRealm", mock.Anything, NAME, mock.Anything).Return(nil, fmt.Errorf("error"))
	authClientMock.On("DeleteSubjectAccessRules", mock.Anything).Return(authorization_api.ApiDeleteSubjectAccessRulesRequest{})
	authClientMock.On("DeleteSubjectAccessRulesExecute", mock.Anything).Return(nil, nil)

	err := testSubject.Destruct(context, getTenantInfo())
	assert.NotNil(t, err)
}

func TestDestruct_Errors(t *testing.T) {
	fakeIdentityServer := getFakeIdentityServer()
	defer fakeIdentityServer.Close()
	context := createFakeAuthCtx()

	testSubject, authClientMock, idmMock := initTest(t, fakeIdentityServer.URL)
	idmMock.On("DeleteRealm", mock.Anything, NAME, mock.Anything).Return(getHttpRespone(500), fmt.Errorf("error"))
	authClientMock.On("DeleteSubjectAccessRules", mock.Anything).Return(authorization_api.ApiDeleteSubjectAccessRulesRequest{})
	authClientMock.On("DeleteSubjectAccessRulesExecute", mock.Anything).Return(getHttpRespone(500), fmt.Errorf("error"))

	err := testSubject.Destruct(context, getTenantInfo())
	assert.NotNil(t, err)
}

func TestDestruct_DeleteAccessRules_404Error(t *testing.T) {
	fakeIdentityServer := getFakeIdentityServer()
	defer fakeIdentityServer.Close()
	context := createFakeAuthCtx()

	testSubject, authClientMock, idmMock := initTest(t, fakeIdentityServer.URL)
	idmMock.On("DeleteRealm", mock.Anything, NAME, mock.Anything).Return(nil, nil)
	authClientMock.On("DeleteSubjectAccessRules", mock.Anything).Return(authorization_api.ApiDeleteSubjectAccessRulesRequest{})
	authClientMock.On("DeleteSubjectAccessRulesExecute", mock.Anything).Return(getHttpRespone(404), fmt.Errorf("error"))

	err := testSubject.Destruct(context, getTenantInfo())
	assert.Nil(t, err)
}

func TestDestruct_EmptyEmail(t *testing.T) {
	fakeIdentityServer := getFakeIdentityServer()
	defer fakeIdentityServer.Close()
	context := createFakeAuthCtx()

	testSubject, _, idmMock := initTest(t, fakeIdentityServer.URL)
	idmMock.On("DeleteRealm", mock.Anything, NAME, mock.Anything).Return(nil, nil)

	err := testSubject.Destruct(context, &builders.TenantInfo{
		Name: NAME,
		Id:   ID,
	})
	assert.Nil(t, err)
}

func TestDestruct_DeleteAccessRules_Error(t *testing.T) {
	fakeIdentityServer := getFakeIdentityServer()
	defer fakeIdentityServer.Close()
	context := createFakeAuthCtx()

	testSubject, authClientMock, idmMock := initTest(t, fakeIdentityServer.URL)
	idmMock.On("DeleteRealm", mock.Anything, NAME, mock.Anything).Return(nil, nil)
	authClientMock.On("DeleteSubjectAccessRules", mock.Anything).Return(authorization_api.ApiDeleteSubjectAccessRulesRequest{})
	authClientMock.On("DeleteSubjectAccessRulesExecute", mock.Anything).Return(getHttpRespone(500), fmt.Errorf("error"))

	err := testSubject.Destruct(context, getTenantInfo())
	assert.NotNil(t, err)
}

func TestConstruct_CreateRealm_AlreadyExists_IsIdempotent(t *testing.T) {
	testSubject, authClientMock, idmClient := initTest(t, "")

	builderResponse := make(map[string]interface{})
	ctx := createFakeAuthCtx()

	httpOK := getHttpRespone(http.StatusOK)
	realmResp := idmopenapi.NewRealmCreationResponse(NAME, "", getExistingClients())
	idmClient.On("GetOrCreateRealm", mock.Anything, NAME, fmt.Sprintf("%d", ID)).Return(realmResp, httpOK, nil)
	idmClient.On("DeleteInternalUserByEmail", mock.Anything, NAME, EMAIL).Return(httpOK, nil)
	idmClient.On("CreateInternalUser", mock.Anything, mock.Anything).Return(
		idmopenapi.NewUserPostResponse(fmt.Sprintf("%d", ID), EMAIL, PASSWORD), httpOK, nil)
	authClientMock.On("CreateInternalAccessRule", mock.Anything, mock.Anything, mock.Anything).Return(getAdminUserRequest())
	authClientMock.On("CreateInternalAccessRuleExecute", mock.Anything).Return(nil, httpOK, nil)

	err := testSubject.Construct(ctx, getTenantInfo(), builderResponse)
	assert.Nil(t, err)
	assert.NotContains(t, builderResponse, "tenant_domain")
}

func TestConstruct_DeleteFirstUser_NotFound_IsIdempotent(t *testing.T) {
	testSubject, authClientMock, idmClient := initTest(t, "")

	builderResponse := make(map[string]interface{})
	ctx := createFakeAuthCtx()

	httpOK := getHttpRespone(http.StatusOK)
	realmResp := idmopenapi.NewRealmCreationResponse(NAME, "Domain", []idmopenapi.Client{
		*idmopenapi.NewClient(AgentClientName+NAME, "agentClientName", "agentSecret"),
	})
	idmClient.On("GetOrCreateRealm", mock.Anything, NAME, fmt.Sprintf("%d", ID)).Return(realmResp, httpOK, nil)
	idmClient.On("DeleteInternalUserByEmail", mock.Anything, NAME, EMAIL).Return(getHttpRespone(http.StatusNotFound), errors.New("not found"))
	idmClient.On("CreateInternalUser", mock.Anything, mock.Anything).Return(
		idmopenapi.NewUserPostResponse(fmt.Sprintf("%d", ID), EMAIL, PASSWORD), httpOK, nil)
	authClientMock.On("CreateInternalAccessRule", mock.Anything, mock.Anything, mock.Anything).Return(getAdminUserRequest())
	authClientMock.On("CreateInternalAccessRuleExecute", mock.Anything).Return(nil, httpOK, nil)

	err := testSubject.Construct(ctx, getTenantInfo(), builderResponse)
	assert.Nil(t, err)
}

func TestConstruct_DeleteFirstUser_Error(t *testing.T) {
	testSubject, _, idmClient := initTest(t, "")

	builderResponse := make(map[string]interface{})
	ctx := createFakeAuthCtx()

	httpOK := getHttpRespone(http.StatusOK)
	realmResp := idmopenapi.NewRealmCreationResponse(NAME, "Domain", []idmopenapi.Client{
		*idmopenapi.NewClient(AgentClientName+NAME, "agentClientName", "agentSecret"),
	})
	idmClient.On("GetOrCreateRealm", mock.Anything, NAME, fmt.Sprintf("%d", ID)).Return(realmResp, httpOK, nil)
	idmClient.On("DeleteInternalUserByEmail", mock.Anything, NAME, EMAIL).Return(getHttpRespone(http.StatusInternalServerError), errors.New("internal error"))

	err := testSubject.Construct(ctx, getTenantInfo(), builderResponse)
	assert.NotNil(t, err)
}

func TestConstruct_AddAccessRules_Conflict_IsIdempotent(t *testing.T) {
	testSubject, authClientMock, idmClient := initTest(t, "")

	builderResponse := make(map[string]interface{})
	ctx := createFakeAuthCtx()

	httpOK := getHttpRespone(http.StatusOK)
	realmResp := idmopenapi.NewRealmCreationResponse(NAME, "Domain", []idmopenapi.Client{
		*idmopenapi.NewClient(AgentClientName+NAME, "agentClientName", "agentSecret"),
	})
	idmClient.On("GetOrCreateRealm", mock.Anything, NAME, fmt.Sprintf("%d", ID)).Return(realmResp, httpOK, nil)
	idmClient.On("DeleteInternalUserByEmail", mock.Anything, NAME, EMAIL).Return(httpOK, nil)
	idmClient.On("CreateInternalUser", mock.Anything, mock.Anything).Return(
		idmopenapi.NewUserPostResponse(fmt.Sprintf("%d", ID), EMAIL, PASSWORD), httpOK, nil)
	authClientMock.On("CreateInternalAccessRule", mock.Anything, mock.Anything, mock.Anything).Return(getAdminUserRequest())
	authClientMock.On("CreateInternalAccessRuleExecute", mock.Anything).Return(nil, getHttpRespone(http.StatusConflict), errors.New("conflict"))

	err := testSubject.Construct(ctx, getTenantInfo(), builderResponse)
	assert.Nil(t, err)
}

func createFakeAuthCtx() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "Bear ", TOKEN)
	return ctx
}

func getFakeIdentityServer() *httptest.Server {
	clients := getRealm()
	clientsJson, _ := json.Marshal(clients)

	user := getUser()
	userJson, _ := json.Marshal(user)

	return httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.RequestURI == "/internal/realms" {
			res.WriteHeader(http.StatusOK)
			res.Write(clientsJson)
		}
		if req.RequestURI == "/internal/users" {
			res.WriteHeader(http.StatusCreated)
			res.Write(userJson)
		}
	}))
}

func getTenantInfo() *builders.TenantInfo {
	return &builders.TenantInfo{
		Name:  NAME,
		Id:    ID,
		Email: EMAIL,
	}
}

func getRealm() *config.Realm {
	return &config.Realm{
		Id:     "realmId",
		Name:   "realmName",
		Domain: "Domain",
		Clients: []config.Client{{
			ClientId: "clientName",
			Id:       "clientId",
			Secret:   "clientSecret",
		}},
	}
}

func getUser() map[string]interface{} {
	return map[string]interface{}{
		"Id":           ID,
		"Username":     EMAIL,
		"TempPassword": PASSWORD,
	}
}

func getAdminUserRequest() authorization_api.ApiCreateInternalAccessRuleRequest {
	return authorization_api.ApiCreateInternalAccessRuleRequest{
		ApiService: nil,
	}
}

func getHttpRespone(status int) *http.Response {
	return &http.Response{
		StatusCode: status,
	}
}
