/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"k8s.io/utils/ptr"

	clientMock "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils/mocks"
	async_events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authMock "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/structs"
	internalCache "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/cache"
	cacheMock "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/cache/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/generated/pkg/api"
	iamError "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/errors"
	iamMock "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/mocks"
	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	authBusinessMocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/business/authorization/mocks"
	dalMock "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/business/mocks"
	idmModles "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/models"
	realm_client "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/realm/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
	cache "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	cache_testing "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/testing"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

func mockInstallerBootstrapLock(t *testing.T) *cache_testing.MockLocker {
	cache_testing.SetMock(t)
	t.Cleanup(func() { cache.SetMock(nil) })
	locker := cache_testing.SetLocksMock(t)
	locker.On("Lock", mock.Anything, mock.Anything).Return(nil)
	locker.On("Unlock", mock.Anything).Return(nil)
	return locker
}

const (
	RealmName          = "runai"
	TenantId           = "1001"
	CreatedBy          = "test@example.com"
	AppName            = "appName"
	ServiceAccountName = "serviceAccountName"
	ServiceAccountId   = "serviceAccountId"
)

func enrichContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, utils.AuthRealmKey, RealmName)
	ctx = context.WithValue(ctx, runaictx.CfTenantId, TenantId)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, CreatedBy)
	return ctx
}

func getAppModelForRequest(name, id, createdBy string) *iamModels.Client {
	client := &iamModels.Client{
		Name:     ptr.To(name),
		TenantId: ptr.To(id),
		Enabled:  ptr.To(true),
		ClientID: ptr.To(name),
		Attributes: &iamModels.ClientAttributes{
			CreatedBy:     ptr.To(createdBy),
			TokenLifespan: ptr.To(realm_client.DefaultServiceAccountTokenLifespan),
		},
	}
	realm_client.ApplyClientProfile(client, realm_client.ClientTypeTenantApp)
	return client
}

func getAppModelForResponse(name, id, createdBy string) *iamModels.Client {
	return &iamModels.Client{
		ID:       ptr.To("appId"),
		ClientID: ptr.To(name),
		Name:     ptr.To(name),
		TenantId: ptr.To(id),
		Secret:   ptr.To("secret"),
		Enabled:  ptr.To(true),
		Attributes: &iamModels.ClientAttributes{
			CreatedBy: ptr.To(createdBy),
		},
	}
}

func initServiceAccountTest(t *testing.T) (iamServicesContainerMock *iamMock.MockServicesContainer, authzClientMock *authMock.IAuthorizationClient, authManagerMock *authBusinessMocks.AuthManagerInterface, realmDalMock *dalMock.RealmDal, serviceAccountsCacheMock *cacheMock.ServiceAccountsCache, testSubject *ServiceAccountsApiService) {
	mockIamServicesContainer := iamMock.NewMockServicesContainer()
	authzClientMock = authMock.NewIAuthorizationClient(t)
	authManagerMock = authBusinessMocks.NewAuthManagerInterface(t)
	realmDalMock = dalMock.NewRealmDal(t)
	serviceAccountsCacheMock = cacheMock.NewServiceAccountsCache(t)

	testSubject = &ServiceAccountsApiService{
		authzClient:          authzClientMock,
		authManager:          authManagerMock,
		realmDal:             realmDalMock,
		iamClients:           mockIamServicesContainer.ToServicesContainer().ClientService,
		iamAuth:              mockIamServicesContainer.ToServicesContainer().AuthService,
		serviceAccountsCache: serviceAccountsCacheMock,
	}
	return mockIamServicesContainer, authzClientMock, authManagerMock, realmDalMock, serviceAccountsCacheMock, testSubject
}

func TestCreateServiceAccount_HappyFlow(t *testing.T) {
	iamMock, _, _, _, serviceAccountsCacheMock, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()

	appReq := getAppModelForRequest(ServiceAccountName, TenantId, CreatedBy)
	appResp := getAppModelForResponse(ServiceAccountName, TenantId, CreatedBy)
	iamMock.MockClientService.On("CreateClient", ctx, RealmName, appReq).Return(ServiceAccountId, nil)
	iamMock.MockClientService.On("SetEmailScopeOptional", ctx, RealmName, ServiceAccountId).Return(nil)
	iamMock.MockClientService.On("GetClient", ctx, RealmName, ServiceAccountId).Return(appResp, nil)
	serviceAccountsCacheMock.On("DeleteTenantServiceAccountsFromRealm", ctx, RealmName).Return(nil)
	response, err := testSubject.CreateServiceAccount(ctx, api.ServiceAccountCreationRequest{
		Name: ServiceAccountName,
	})
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusCreated, response.Code)
}

func TestCreateServiceAccount_CreateAppFail(t *testing.T) {
	iamMock, _, _, _, _, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()

	appReq := getAppModelForRequest(ServiceAccountName, TenantId, CreatedBy)
	iamMock.MockClientService.On("CreateClient", ctx, RealmName, appReq).Return("", fmt.Errorf("error"))

	response, err := testSubject.CreateServiceAccount(ctx, api.ServiceAccountCreationRequest{
		Name: ServiceAccountName,
	})
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestCreateServiceAccount_GetAppFail(t *testing.T) {
	iamMock, _, _, _, serviceAccountsCacheMock, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()

	appReq := getAppModelForRequest(ServiceAccountName, TenantId, CreatedBy)
	iamMock.MockClientService.On("CreateClient", ctx, RealmName, appReq).Return(ServiceAccountId, nil)
	iamMock.MockClientService.On("SetEmailScopeOptional", ctx, RealmName, ServiceAccountId).Return(nil)
	iamMock.MockClientService.On("GetClient", ctx, RealmName, ServiceAccountId).Return(nil, fmt.Errorf("error"))
	serviceAccountsCacheMock.On("DeleteTenantServiceAccountsFromRealm", ctx, RealmName).Return(nil)
	response, err := testSubject.CreateServiceAccount(ctx, api.ServiceAccountCreationRequest{
		Name: ServiceAccountName,
	})
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestDeleteServiceAccountById_HappyFlow(t *testing.T) {
	iamMock, authzClientMock, _, _, serviceAccountsCacheMock, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()
	appResp := getAppModelForResponse(ServiceAccountName, TenantId, CreatedBy)
	iamMock.MockClientService.On("GetClient", ctx, RealmName, ServiceAccountId).Return(appResp, nil)
	authzClientMock.On("DeleteSubjectAccessRules", mock.Anything).Return(authorization_api.ApiDeleteSubjectAccessRulesRequest{})
	authzClientMock.On("DeleteSubjectAccessRulesExecute", mock.Anything).Return(nil, nil)
	iamMock.MockClientService.On("DeleteClient", ctx, RealmName, ServiceAccountId).Return(nil)
	serviceAccountsCacheMock.On("DeleteTenantServiceAccountsFromRealm", ctx, RealmName).Return(nil)
	response, err := testSubject.DeleteServiceAccountById(ctx, ServiceAccountId)
	assert.Nil(t, err)
	assert.Nil(t, response.Body)
	assert.Equal(t, http.StatusNoContent, response.Code)
}

func TestDeleteServiceAccountById_GetAppError(t *testing.T) {
	iamMock, _, _, _, _, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()
	iamMock.MockClientService.On("GetClient", ctx, RealmName, ServiceAccountId).Return(nil, fmt.Errorf("error"))
	response, err := testSubject.DeleteServiceAccountById(ctx, ServiceAccountId)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestDeleteServiceAccountById_SubjectAccessRulesError(t *testing.T) {
	iamMock, authzClientMock, _, _, serviceAccountsCacheMock, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()
	appResp := getAppModelForResponse(ServiceAccountName, TenantId, CreatedBy)
	iamMock.MockClientService.On("GetClient", ctx, RealmName, ServiceAccountId).Return(appResp, nil)
	authzClientMock.On("DeleteSubjectAccessRules", mock.Anything).Return(authorization_api.ApiDeleteSubjectAccessRulesRequest{})
	authzClientMock.On("DeleteSubjectAccessRulesExecute", mock.Anything).Return(nil, fmt.Errorf("error"))
	serviceAccountsCacheMock.On("DeleteTenantServiceAccountsFromRealm", ctx, RealmName).Return(nil)
	response, err := testSubject.DeleteServiceAccountById(ctx, ServiceAccountId)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestDeleteServiceAccountById_DeleteAppError(t *testing.T) {
	iamMock, authzClientMock, _, _, serviceAccountsCacheMock, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()
	appResp := getAppModelForResponse(ServiceAccountName, TenantId, CreatedBy)
	iamMock.MockClientService.On("GetClient", ctx, RealmName, ServiceAccountId).Return(appResp, nil)
	authzClientMock.On("DeleteSubjectAccessRules", mock.Anything).Return(authorization_api.ApiDeleteSubjectAccessRulesRequest{})
	authzClientMock.On("DeleteSubjectAccessRulesExecute", mock.Anything).Return(nil, nil)
	iamMock.MockClientService.On("DeleteClient", ctx, RealmName, ServiceAccountId).Return(fmt.Errorf("error"))
	serviceAccountsCacheMock.On("DeleteTenantServiceAccountsFromRealm", ctx, RealmName).Return(nil)
	response, err := testSubject.DeleteServiceAccountById(ctx, ServiceAccountId)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestGetServiceAccountById_HappyFlow(t *testing.T) {
	iamMock, _, _, _, _, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()

	appResp := getAppModelForResponse(ServiceAccountName, TenantId, CreatedBy)
	iamMock.MockClientService.On("GetClient", ctx, RealmName, ServiceAccountId).Return(appResp, nil)

	response, err := testSubject.GetServiceAccountById(ctx, ServiceAccountId)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestGetServiceAccountById_GetAppError(t *testing.T) {
	iamMock, _, _, _, _, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()

	iamMock.MockClientService.On("GetClient", ctx, RealmName, ServiceAccountId).Return(nil, fmt.Errorf("error"))

	response, err := testSubject.GetServiceAccountById(ctx, ServiceAccountId)
	assert.NotNil(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestGetServiceAccounts_HappyFlow(t *testing.T) {
	_, _, _, _, serviceAccountsCacheMock, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()

	appArray := []iamModels.Client{
		*getAppModelForResponse(ServiceAccountName, TenantId, CreatedBy),
		*getAppModelForResponse(ServiceAccountName, TenantId, CreatedBy),
	}
	serviceAccountsCacheMock.On("GetTenantServiceAccountsForRealm", ctx, RealmName, mock.Anything).Return(&internalCache.TenantServiceAccounts{ServiceAccounts: appArray}, nil)
	response, err := testSubject.GetServiceAccounts(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, len(appArray), len(response.Body.([]api.ServiceAccount)))
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestGetServiceAccounts_GetAppsError(t *testing.T) {
	_, _, _, _, serviceAccountsCacheMock, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()
	serviceAccountsCacheMock.On("GetTenantServiceAccountsForRealm", ctx, RealmName, mock.Anything).Return(nil, fmt.Errorf("error"))

	response, err := testSubject.GetServiceAccounts(ctx)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestUpdateServiceAccountById_HappyFlow(t *testing.T) {
	iamMock, _, _, _, serviceAccountsCacheMock, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()

	iamMock.MockClientService.On("UpdateClientStatus", ctx, RealmName, ServiceAccountId, true).Return(nil)
	serviceAccountsCacheMock.On("DeleteTenantServiceAccountsFromRealm", ctx, RealmName).Return(nil)
	response, err := testSubject.UpdateServiceAccountById(ctx, ServiceAccountId, api.ServiceAccountPatchRequest{
		Enabled: true,
	})
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestUpdateServiceAccountById_UpdateStatusError(t *testing.T) {
	iamMock, _, _, _, _, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()

	iamMock.MockClientService.On("UpdateClientStatus", ctx, RealmName, ServiceAccountId, true).Return(fmt.Errorf("error"))

	response, err := testSubject.UpdateServiceAccountById(ctx, ServiceAccountId, api.ServiceAccountPatchRequest{
		Enabled: true,
	})
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestRegenerateServiceAccountSecret_HappyFlow(t *testing.T) {
	iamMock, authzClientMock, authManagerMock, _, serviceAccountsCacheMock, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()

	appResp := getAppModelForResponse(ServiceAccountName, TenantId, CreatedBy)
	iamMock.MockClientService.On("GetClient", ctx, RealmName, ServiceAccountId).Return(appResp, nil)
	serviceAccountsCacheMock.On("DeleteTenantServiceAccountsFromRealm", ctx, RealmName).Return(nil)
	authManagerMock.On("ValidatePrivilegeEscalation", mock.Anything, authzClientMock, ServiceAccountName).Return(nil)

	iamMock.MockClientService.On("RegenerateClientSecret", ctx, RealmName, ServiceAccountId).Return("secret", nil)

	response, err := testSubject.RegenerateServiceAccountSecret(ctx, ServiceAccountId)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestRegenerateServiceAccountSecret_RegenerateSecretError(t *testing.T) {
	iamMock, authzClientMock, authManagerMock, _, _, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()

	appResp := getAppModelForResponse(ServiceAccountName, TenantId, CreatedBy)
	iamMock.MockClientService.On("GetClient", ctx, RealmName, ServiceAccountId).Return(appResp, nil)

	authManagerMock.On("ValidatePrivilegeEscalation", mock.Anything, authzClientMock, ServiceAccountName).Return(nil)

	iamMock.MockClientService.On("RegenerateClientSecret", ctx, RealmName, ServiceAccountId).Return("", fmt.Errorf("error"))

	response, err := testSubject.RegenerateServiceAccountSecret(ctx, ServiceAccountId)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestRegenerateServiceAccountSecret_GetAppError(t *testing.T) {
	iamMock, _, _, _, _, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()

	iamMock.MockClientService.On("GetClient", ctx, RealmName, ServiceAccountId).Return(nil, fmt.Errorf("database connection error"))

	response, err := testSubject.RegenerateServiceAccountSecret(ctx, ServiceAccountId)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)

	utilsError := err.(utils.Error)
	assert.Equal(t, http.StatusInternalServerError, utilsError.Code)
	assert.Equal(t, "", utilsError.Message)
}

func TestRegenerateServiceAccountSecret_NilClientID(t *testing.T) {
	iamMock, _, _, _, _, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()

	serviceAccountWithNilClientID := &iamModels.Client{
		ID:       ptr.To(ServiceAccountId),
		Name:     ptr.To(ServiceAccountName),
		ClientID: nil,
		Enabled:  ptr.To(true),
	}
	iamMock.MockClientService.On("GetClient", ctx, RealmName, ServiceAccountId).Return(serviceAccountWithNilClientID, nil)

	response, err := testSubject.RegenerateServiceAccountSecret(ctx, ServiceAccountId)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	utilsError := err.(utils.Error)
	assert.Contains(t, utilsError.Message, "Service account client ID is required for privilege escalation check")
}

func TestRegenerateServiceAccountSecret_PrivilegeEscalationForbidden(t *testing.T) {
	iamMock, authzClientMock, authManagerMock, _, _, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()

	appResp := getAppModelForResponse(ServiceAccountName, TenantId, CreatedBy)
	iamMock.MockClientService.On("GetClient", ctx, RealmName, ServiceAccountId).Return(appResp, nil)

	forbiddenErr := fmt.Errorf("403 Forbidden")
	authManagerMock.On("ValidatePrivilegeEscalation", mock.Anything, authzClientMock, ServiceAccountName).Return(forbiddenErr)

	response, err := testSubject.RegenerateServiceAccountSecret(ctx, ServiceAccountId)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusForbidden, response.Code)

	utilsError := err.(utils.Error)
	assert.Contains(t, utilsError.Message, "You're not authorized to regenerate client secret for the selected service account")
}

func TestGetInstallerServiceAccount_HappyFlow_EmptyTenantId(t *testing.T) {
	iamMock, _, _, _, _, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()
	appResp := getAppModelForResponse(AppName, TenantId, CreatedBy)
	iamMock.MockClientService.On("GetClientByClientId", ctx, RealmName, utils.RunaiInstallerClientId).Return(appResp, nil)

	response, err := testSubject.GetInstallerServiceAccount(ctx, "")
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestGetInstallerServiceAccount_HappyFlow_MngRealm(t *testing.T) {
	iamMock, _, _, realmDalMock, _, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()
	ctxWithMng := context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")
	ctxWithMng = context.WithValue(ctxWithMng, structs.PermittedScopesKey, &authorization_api.PermittedScopes{System: ptr.To(true)})

	appResp := getAppModelForResponse(AppName, TenantId, CreatedBy)
	iamMock.MockClientService.On("GetClientByClientId", ctxWithMng, RealmName, utils.RunaiInstallerClientId).Return(appResp, nil)
	realmDalMock.On("GetRealmByTenantId", mock.Anything, mock.Anything).Return(&idmModles.Realm{Realm: RealmName, TenantId: TenantId}, nil)

	response, err := testSubject.GetInstallerServiceAccount(ctxWithMng, TenantId)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestGetInstallerServiceAccount_HappyFlow_FalseMngRealm(t *testing.T) {
	_, _, _, _, _, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()
	ctxWithMng := context.WithValue(ctx, runaictx.CfIsManagementTenant, "false")
	ctxWithMng = context.WithValue(ctxWithMng, structs.PermittedScopesKey, &authorization_api.PermittedScopes{System: ptr.To(true)})
	response, err := testSubject.GetInstallerServiceAccount(ctxWithMng, TenantId)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
}

func TestGetInstallerServiceAccount_HappyFlow_MngRealm_NoSystem(t *testing.T) {
	_, _, _, _, _, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()
	ctxWithMng := context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

	response, err := testSubject.GetInstallerServiceAccount(ctxWithMng, TenantId)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
}

func TestGetInstallerServiceAccount_HappyFlow_NilMngRealm(t *testing.T) {
	_, _, _, _, _, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()
	ctxWithMng := context.WithValue(ctx, runaictx.CfIsManagementTenant, nil)

	response, err := testSubject.GetInstallerServiceAccount(ctxWithMng, TenantId)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
}

func TestGetInstallerServiceAccount_HappyFlow_GetRealmError(t *testing.T) {
	iamMock, _, _, realmDalMock, _, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()
	ctxWithMng := context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")
	ctxWithMng = context.WithValue(ctxWithMng, structs.PermittedScopesKey, &authorization_api.PermittedScopes{System: ptr.To(true)})
	appResp := getAppModelForResponse(AppName, TenantId, CreatedBy)
	iamMock.MockClientService.On("GetClientByClientId", ctxWithMng, RealmName, utils.RunaiInstallerClientId).Return(appResp, nil)
	realmDalMock.On("GetRealmByTenantId", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error"))

	response, err := testSubject.GetInstallerServiceAccount(ctxWithMng, TenantId)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestGetInstallerServiceAccount_HappyFlow_GetAppError(t *testing.T) {
	iamMock, _, _, _, _, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()
	iamMock.MockClientService.On("GetClientByClientId", ctx, RealmName, utils.RunaiInstallerClientId).Return(nil, fmt.Errorf("error"))

	response, err := testSubject.GetInstallerServiceAccount(ctx, "")
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestGetInstallerServiceAccount_HappyFlow_GetAppError_nilError(t *testing.T) {
	iamMock, _, _, _, _, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()
	iamMock.MockClientService.On("GetClientByClientId", ctx, RealmName, utils.RunaiInstallerClientId).Return(nil, nil)

	response, _ := testSubject.GetInstallerServiceAccount(ctx, "")
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestClusterDeleteCallback(t *testing.T) {
	iamMock, _, _, dalMock, serviceAccountsCacheMock, testSubject := initServiceAccountTest(t)
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, clientMock.TENANT_ID)
	dalMock.On("GetRealmByTenantId", mock.Anything, clientMock.TENANT_ID).Return(&idmModles.Realm{TenantId: clientMock.TENANT_ID, Realm: clientMock.REALM_NAME}, nil)

	clients := []iamModels.Client{
		{
			ID:       ptr.To("client1"),
			Name:     ptr.To("workload-auth-client"),
			ClientID: ptr.To("workload-auth-client-id"),
			Attributes: &iamModels.ClientAttributes{
				ClientType: ptr.To(string(realm_client.ClientTypeWorkloadsAuth)),
			},
		},
		{
			ID:       ptr.To("client2"),
			Name:     ptr.To("admin-ui-client"),
			ClientID: ptr.To("admin-ui-client-id"),
			Attributes: &iamModels.ClientAttributes{
				ClientType: ptr.To(string(realm_client.ClientTypeRunaiAdminUI)),
			},
		},
		{
			ID:       ptr.To("client3"),
			Name:     ptr.To("agent-client"),
			ClientID: ptr.To("agent-client-id"),
			Attributes: &iamModels.ClientAttributes{
				ClientType: ptr.To(string(api.CLIENTTYPE_AGENT)),
			},
		},
	}

	iamMock.MockClientService.On("DeleteClient", ctx, clientMock.REALM_NAME, *clients[0].ID).Return(nil)
	iamMock.MockClientService.On("DeleteClient", ctx, clientMock.REALM_NAME, *clients[1].ID).Return(nil)
	iamMock.MockClientService.On("DeleteClient", ctx, clientMock.REALM_NAME, *clients[2].ID).Return(nil)
	iamMock.MockClientService.On("GetClients", ctx, clientMock.REALM_NAME, mock.Anything).Return(clients, nil)
	serviceAccountsCacheMock.On("DeleteTenantServiceAccountsFromRealm", ctx, clientMock.REALM_NAME).Return(nil)

	err := testSubject.ClusterDeleteCallback(ctx, &async_events.ClusterDelete{ClusterId: "cluster_id"})
	assert.Nil(t, err)
}

func TestMintInstallerBootstrapToken_RegenerateSecretError(t *testing.T) {
	iamMock, authzClientMock, _, _, _, testSubject := initServiceAccountTest(t)
	mockInstallerBootstrapLock(t)
	ctx := enrichContext()
	clusterId := "cluster-1"
	perClusterClientId := "runai-internal.cluster-1.installer"
	perClusterClient := &iamModels.Client{
		ID:       ptr.To("keycloak-id"),
		Enabled:  ptr.To(true),
		ClientID: ptr.To(perClusterClientId),
		Secret:   ptr.To("secret"),
		Attributes: &iamModels.ClientAttributes{
			ClusterId: ptr.To(clusterId),
		},
	}
	iamMock.MockClientService.On("GetClients", ctx, RealmName, &iamModels.ClientFilter{
		ClusterId:  &clusterId,
		ClientType: ptr.To(realm_client.ClientTypeInstaller),
	}).Return([]iamModels.Client{*perClusterClient}, nil)
	iamMock.MockClientService.On("GetClientByClientId", ctx, RealmName, perClusterClientId).Return(perClusterClient, nil)
	iamMock.MockClientService.On("RegenerateClientSecret", ctx, RealmName, *perClusterClient.ID).Return("", assert.AnError)
	authzClientMock.On("SetServiceAccountPermissions", ctx, TenantId, perClusterClientId,
		authorization_api.InternalCreateAccessRuleRequestScope{Id: clusterId, Type: authorization_api.SCOPETYPE_CLUSTER},
		authorization_api.ALLOWEDINTERNALCREATEROLES_INSTALLER_SERVICE_ACCOUNT).Return(nil)

	resp, err := testSubject.MintInstallerBootstrapToken(ctx, clusterId, "")
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestMintInstallerBootstrapToken_GetClientError(t *testing.T) {
	iamMock, authzClientMock, _, _, _, testSubject := initServiceAccountTest(t)
	mockInstallerBootstrapLock(t)
	ctx := enrichContext()
	clusterId := "cluster-1"
	perClusterClientId := "runai-internal.cluster-1.installer"
	perClusterClient := &iamModels.Client{
		ID:       ptr.To("keycloak-id"),
		Enabled:  ptr.To(true),
		ClientID: ptr.To(perClusterClientId),
		Secret:   ptr.To("secret"),
		Attributes: &iamModels.ClientAttributes{
			ClusterId: ptr.To(clusterId),
		},
	}
	iamMock.MockClientService.On("GetClients", ctx, RealmName, &iamModels.ClientFilter{
		ClusterId:  &clusterId,
		ClientType: ptr.To(realm_client.ClientTypeInstaller),
	}).Return([]iamModels.Client{*perClusterClient}, nil)
	iamMock.MockClientService.On("GetClientByClientId", ctx, RealmName, perClusterClientId).Return(nil, assert.AnError)
	authzClientMock.On("SetServiceAccountPermissions", ctx, TenantId, perClusterClientId,
		authorization_api.InternalCreateAccessRuleRequestScope{Id: clusterId, Type: authorization_api.SCOPETYPE_CLUSTER},
		authorization_api.ALLOWEDINTERNALCREATEROLES_INSTALLER_SERVICE_ACCOUNT).Return(nil)

	resp, err := testSubject.MintInstallerBootstrapToken(ctx, clusterId, "")
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestMintInstallerBootstrapToken_MissingClientID(t *testing.T) {
	iamMock, authzClientMock, _, _, _, testSubject := initServiceAccountTest(t)
	mockInstallerBootstrapLock(t)
	ctx := enrichContext()
	clusterId := "cluster-1"
	perClusterClientId := "runai-internal.cluster-1.installer"
	perClusterClient := &iamModels.Client{
		ID:       ptr.To("keycloak-id"),
		Enabled:  ptr.To(true),
		ClientID: ptr.To(perClusterClientId),
		Secret:   ptr.To("secret"),
		Attributes: &iamModels.ClientAttributes{
			ClusterId: ptr.To(clusterId),
		},
	}
	clientWithSecret := &iamModels.Client{Secret: ptr.To("secret")} // ClientID is nil
	iamMock.MockClientService.On("GetClients", ctx, RealmName, &iamModels.ClientFilter{
		ClusterId:  &clusterId,
		ClientType: ptr.To(realm_client.ClientTypeInstaller),
	}).Return([]iamModels.Client{*perClusterClient}, nil)
	iamMock.MockClientService.On("GetClientByClientId", ctx, RealmName, perClusterClientId).Return(clientWithSecret, nil)
	authzClientMock.On("SetServiceAccountPermissions", ctx, TenantId, perClusterClientId,
		authorization_api.InternalCreateAccessRuleRequestScope{Id: clusterId, Type: authorization_api.SCOPETYPE_CLUSTER},
		authorization_api.ALLOWEDINTERNALCREATEROLES_INSTALLER_SERVICE_ACCOUNT).Return(nil)

	resp, err := testSubject.MintInstallerBootstrapToken(ctx, clusterId, "")
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestMintInstallerBootstrapToken_AuthenticateError(t *testing.T) {
	iamMock, authzClientMock, _, _, _, testSubject := initServiceAccountTest(t)
	mockInstallerBootstrapLock(t)
	ctx := enrichContext()
	clusterId := "cluster-1"
	perClusterClientId := "runai-internal.cluster-1.installer"
	perClusterClient := &iamModels.Client{
		ID:       ptr.To("keycloak-id"),
		Enabled:  ptr.To(true),
		ClientID: ptr.To(perClusterClientId),
		Secret:   ptr.To("secret"),
		Attributes: &iamModels.ClientAttributes{
			ClusterId: ptr.To(clusterId),
		},
	}
	rotatedSecret := "rotated-secret"
	iamMock.MockClientService.On("GetClients", ctx, RealmName, &iamModels.ClientFilter{
		ClusterId:  &clusterId,
		ClientType: ptr.To(realm_client.ClientTypeInstaller),
	}).Return([]iamModels.Client{*perClusterClient}, nil)
	iamMock.MockClientService.On("GetClientByClientId", ctx, RealmName, perClusterClientId).Return(perClusterClient, nil)
	iamMock.MockClientService.On("RegenerateClientSecret", ctx, RealmName, *perClusterClient.ID).Return(rotatedSecret, nil)
	iamMock.MockAuthService.On("AuthenticateClient", ctx, RealmName, perClusterClientId, rotatedSecret).Return(nil, assert.AnError)
	authzClientMock.On("SetServiceAccountPermissions", ctx, TenantId, perClusterClientId,
		authorization_api.InternalCreateAccessRuleRequestScope{Id: clusterId, Type: authorization_api.SCOPETYPE_CLUSTER},
		authorization_api.ALLOWEDINTERNALCREATEROLES_INSTALLER_SERVICE_ACCOUNT).Return(nil)

	resp, err := testSubject.MintInstallerBootstrapToken(ctx, clusterId, "")
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestMintInstallerBootstrapToken_WrongTenant_NotFound(t *testing.T) {
	_, _, _, realmDalMock, _, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()
	ctxWithMng := context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")
	ctxWithMng = context.WithValue(ctxWithMng, structs.PermittedScopesKey, &authorization_api.PermittedScopes{System: ptr.To(true)})
	realmDalMock.On("GetRealmByTenantId", mock.Anything, "nonexistent-tenant").
		Return(nil, fmt.Errorf("realm not found for tenant nonexistent-tenant"))

	resp, err := testSubject.MintInstallerBootstrapToken(ctxWithMng, "cluster-1", "nonexistent-tenant")
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestMintInstallerBootstrapToken_TenantIdWithoutSystemPermissions_Unauthorized(t *testing.T) {
	_, _, _, _, _, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()

	resp, err := testSubject.MintInstallerBootstrapToken(ctx, "cluster-1", "some-other-tenant")
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestMintInstallerBootstrapToken_CacheNotInitialized(t *testing.T) {
	iamMock, authzClientMock, _, _, _, testSubject := initServiceAccountTest(t)
	cache.SetMock(nil)
	t.Cleanup(func() { cache.SetMock(nil) })
	ctx := enrichContext()
	clusterId := "cluster-1"
	perClusterClientId := "runai-internal.cluster-1.installer"
	perClusterClient := &iamModels.Client{
		ID:       ptr.To("keycloak-id"),
		Enabled:  ptr.To(true),
		ClientID: ptr.To(perClusterClientId),
		Secret:   ptr.To("secret"),
		Attributes: &iamModels.ClientAttributes{
			ClusterId: ptr.To(clusterId),
		},
	}
	existingSecret := "existing-secret"
	tokens := &iamModels.Tokens{AccessToken: "bootstrap-jwt"}
	iamMock.MockClientService.On("GetClients", ctx, RealmName, &iamModels.ClientFilter{
		ClusterId:  &clusterId,
		ClientType: ptr.To(realm_client.ClientTypeInstaller),
	}).Return([]iamModels.Client{*perClusterClient}, nil)
	iamMock.MockClientService.On("GetClientByClientId", ctx, RealmName, perClusterClientId).Return(perClusterClient, nil)
	iamMock.MockClientService.On("GetClientSecret", ctx, RealmName, *perClusterClient.ID).Return(existingSecret, nil)
	iamMock.MockAuthService.On("AuthenticateClient", ctx, RealmName, perClusterClientId, existingSecret).Return(tokens, nil)
	authzClientMock.On("SetServiceAccountPermissions", ctx, TenantId, perClusterClientId,
		authorization_api.InternalCreateAccessRuleRequestScope{Id: clusterId, Type: authorization_api.SCOPETYPE_CLUSTER},
		authorization_api.ALLOWEDINTERNALCREATEROLES_INSTALLER_SERVICE_ACCOUNT).Return(nil)

	resp, err := testSubject.MintInstallerBootstrapToken(ctx, clusterId, "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)

	// without a confirmed lock, rotation must be skipped so concurrent mints can't race
	iamMock.MockClientService.AssertNotCalled(t, "RegenerateClientSecret", mock.Anything, mock.Anything, mock.Anything)
}

func TestMintInstallerBootstrapToken_LockError(t *testing.T) {
	iamMock, _, _, _, _, testSubject := initServiceAccountTest(t)
	cache_testing.SetMock(t)
	t.Cleanup(func() { cache.SetMock(nil) })
	locker := cache_testing.SetLocksMock(t)
	locker.On("Lock", mock.Anything, mock.Anything).Return(assert.AnError)
	ctx := enrichContext()

	resp, err := testSubject.MintInstallerBootstrapToken(ctx, "cluster-1", "")
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	iamMock.MockClientService.AssertNotCalled(t, "GetClientByClientId", mock.Anything, mock.Anything, mock.Anything)
	iamMock.MockClientService.AssertNotCalled(t, "RegenerateClientSecret", mock.Anything, mock.Anything, mock.Anything)
	iamMock.MockAuthService.AssertNotCalled(t, "AuthenticateClient", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestMintInstallerBootstrapToken_UnlockUsesDetachedContext(t *testing.T) {
	iamMock, authzClientMock, _, _, _, testSubject := initServiceAccountTest(t)
	cache_testing.SetMock(t)
	t.Cleanup(func() { cache.SetMock(nil) })
	locker := cache_testing.SetLocksMock(t)
	locker.On("Lock", mock.Anything, mock.Anything).Return(nil)
	locker.On("Unlock", mock.MatchedBy(func(unlockCtx context.Context) bool {
		return unlockCtx.Err() == nil
	})).Return(nil)

	ctx, cancel := context.WithCancel(enrichContext())
	clusterId := "cluster-1"
	perClusterClientId := "runai-internal.cluster-1.installer"
	perClusterClient := &iamModels.Client{
		ID:       ptr.To("keycloak-id"),
		Enabled:  ptr.To(true),
		ClientID: ptr.To(perClusterClientId),
		Secret:   ptr.To("secret"),
		Attributes: &iamModels.ClientAttributes{
			ClusterId: ptr.To(clusterId),
		},
	}
	rotatedSecret := "rotated-secret"
	tokens := &iamModels.Tokens{AccessToken: "bootstrap-jwt"}
	iamMock.MockClientService.On("GetClients", ctx, RealmName, &iamModels.ClientFilter{
		ClusterId:  &clusterId,
		ClientType: ptr.To(realm_client.ClientTypeInstaller),
	}).Return([]iamModels.Client{*perClusterClient}, nil)
	iamMock.MockClientService.On("GetClientByClientId", ctx, RealmName, perClusterClientId).Return(perClusterClient, nil)
	iamMock.MockClientService.On("RegenerateClientSecret", ctx, RealmName, *perClusterClient.ID).Return(rotatedSecret, nil)
	iamMock.MockAuthService.On("AuthenticateClient", ctx, RealmName, perClusterClientId, rotatedSecret).
		Run(func(mock.Arguments) { cancel() }).
		Return(tokens, nil)
	authzClientMock.On("SetServiceAccountPermissions", ctx, TenantId, perClusterClientId,
		authorization_api.InternalCreateAccessRuleRequestScope{Id: clusterId, Type: authorization_api.SCOPETYPE_CLUSTER},
		authorization_api.ALLOWEDINTERNALCREATEROLES_INSTALLER_SERVICE_ACCOUNT).Return(nil)

	resp, err := testSubject.MintInstallerBootstrapToken(ctx, clusterId, "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestMintInstallerBootstrapToken_Unauthorized(t *testing.T) {
	iamMock, _, _, _, _, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()
	ctxWithMng := context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

	resp, err := testSubject.MintInstallerBootstrapToken(ctxWithMng, "cluster-1", TenantId)
	assert.NotNil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)

	iamMock.MockClientService.AssertNotCalled(t, "RegenerateClientSecret", mock.Anything, mock.Anything, mock.Anything)
	iamMock.MockAuthService.AssertNotCalled(t, "AuthenticateClient", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestMintInstallerBootstrapToken_GetRealmError(t *testing.T) {
	iamMock, _, _, realmDalMock, _, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()
	ctxWithMng := context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")
	ctxWithMng = context.WithValue(ctxWithMng, structs.PermittedScopesKey, &authorization_api.PermittedScopes{System: ptr.To(true)})
	realmDalMock.On("GetRealmByTenantId", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error"))

	resp, err := testSubject.MintInstallerBootstrapToken(ctxWithMng, "cluster-1", TenantId)
	assert.NotNil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	iamMock.MockClientService.AssertNotCalled(t, "RegenerateClientSecret", mock.Anything, mock.Anything, mock.Anything)
	iamMock.MockAuthService.AssertNotCalled(t, "AuthenticateClient", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestMintInstallerBootstrapToken_ClusterScoped_ExistingClient(t *testing.T) {
	iamMock, authzClientMock, _, _, _, testSubject := initServiceAccountTest(t)
	mockInstallerBootstrapLock(t)
	ctx := enrichContext()
	clusterId := "cluster-1"
	perClusterClientId := "runai-internal.cluster-1.installer"
	perClusterClient := &iamModels.Client{
		ID:       ptr.To("keycloak-id"),
		Enabled:  ptr.To(true),
		ClientID: ptr.To(perClusterClientId),
		Secret:   ptr.To("secret"),
		Attributes: &iamModels.ClientAttributes{
			ClusterId: ptr.To(clusterId),
		},
	}
	rotatedSecret := "rotated-secret"
	tokens := &iamModels.Tokens{AccessToken: "bootstrap-jwt"}

	iamMock.MockClientService.On("GetClients", ctx, RealmName, &iamModels.ClientFilter{
		ClusterId:  &clusterId,
		ClientType: ptr.To(realm_client.ClientTypeInstaller),
	}).Return([]iamModels.Client{*perClusterClient}, nil)
	iamMock.MockClientService.On("GetClientByClientId", ctx, RealmName, perClusterClientId).Return(perClusterClient, nil)
	iamMock.MockClientService.On("RegenerateClientSecret", ctx, RealmName, *perClusterClient.ID).Return(rotatedSecret, nil)
	iamMock.MockAuthService.On("AuthenticateClient", ctx, RealmName, perClusterClientId, rotatedSecret).Return(tokens, nil)
	authzClientMock.On("SetServiceAccountPermissions", ctx, TenantId, perClusterClientId,
		authorization_api.InternalCreateAccessRuleRequestScope{Id: clusterId, Type: authorization_api.SCOPETYPE_CLUSTER},
		authorization_api.ALLOWEDINTERNALCREATEROLES_INSTALLER_SERVICE_ACCOUNT).Return(nil)

	resp, err := testSubject.MintInstallerBootstrapToken(ctx, clusterId, "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "bootstrap-jwt", resp.Body.(api.InstallerBootstrapResponse).Token)

	iamMock.MockClientService.AssertNotCalled(t, "CreateClient", mock.Anything, mock.Anything, mock.Anything)
	authzClientMock.AssertCalled(t, "SetServiceAccountPermissions", ctx, TenantId, perClusterClientId,
		authorization_api.InternalCreateAccessRuleRequestScope{Id: clusterId, Type: authorization_api.SCOPETYPE_CLUSTER},
		authorization_api.ALLOWEDINTERNALCREATEROLES_INSTALLER_SERVICE_ACCOUNT)
}

func TestMintInstallerBootstrapToken_ClusterScoped_SelfHealsMissingClient(t *testing.T) {
	iamMock, authzClientMock, _, _, _, testSubject := initServiceAccountTest(t)
	mockInstallerBootstrapLock(t)
	ctx := enrichContext()
	clusterId := "cluster-1"
	perClusterClientId := "runai-internal.cluster-1.installer"
	perClusterClient := &iamModels.Client{
		ID:       ptr.To("keycloak-id"),
		Enabled:  ptr.To(true),
		ClientID: ptr.To(perClusterClientId),
		Secret:   ptr.To("secret"),
		Attributes: &iamModels.ClientAttributes{
			ClusterId: ptr.To(clusterId),
		},
	}
	rotatedSecret := "rotated-secret"
	tokens := &iamModels.Tokens{AccessToken: "bootstrap-jwt"}

	iamMock.MockClientService.On("GetClients", ctx, RealmName, &iamModels.ClientFilter{
		ClusterId:  &clusterId,
		ClientType: ptr.To(realm_client.ClientTypeInstaller),
	}).Return([]iamModels.Client{}, nil)
	iamMock.MockClientService.On("CreateClient", ctx, RealmName, mock.MatchedBy(func(c *iamModels.Client) bool {
		return c.ClientID != nil && *c.ClientID == perClusterClientId &&
			c.Name != nil && *c.Name == perClusterClientId &&
			c.Attributes != nil && c.Attributes.ClusterId != nil && *c.Attributes.ClusterId == clusterId
	})).Return("keycloak-id", nil)
	authzClientMock.On("SetServiceAccountPermissions", ctx, TenantId, perClusterClientId,
		authorization_api.InternalCreateAccessRuleRequestScope{Id: clusterId, Type: authorization_api.SCOPETYPE_CLUSTER},
		authorization_api.ALLOWEDINTERNALCREATEROLES_INSTALLER_SERVICE_ACCOUNT).Return(nil)
	iamMock.MockClientService.On("GetClientByClientId", ctx, RealmName, perClusterClientId).Return(perClusterClient, nil)
	iamMock.MockClientService.On("RegenerateClientSecret", ctx, RealmName, *perClusterClient.ID).Return(rotatedSecret, nil)
	iamMock.MockAuthService.On("AuthenticateClient", ctx, RealmName, perClusterClientId, rotatedSecret).Return(tokens, nil)

	resp, err := testSubject.MintInstallerBootstrapToken(ctx, clusterId, "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "bootstrap-jwt", resp.Body.(api.InstallerBootstrapResponse).Token)
}

func TestMintInstallerBootstrapToken_ClusterScoped_SetServiceAccountPermissionsError(t *testing.T) {
	iamMock, authzClientMock, _, _, _, testSubject := initServiceAccountTest(t)
	ctx := enrichContext()
	clusterId := "cluster-1"
	perClusterClientId := "runai-internal.cluster-1.installer"

	iamMock.MockClientService.On("GetClients", ctx, RealmName, &iamModels.ClientFilter{
		ClusterId:  &clusterId,
		ClientType: ptr.To(realm_client.ClientTypeInstaller),
	}).Return([]iamModels.Client{}, nil)
	iamMock.MockClientService.On("CreateClient", ctx, RealmName, mock.Anything).Return("keycloak-id", nil)
	authzClientMock.On("SetServiceAccountPermissions", ctx, TenantId, perClusterClientId, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))

	resp, err := testSubject.MintInstallerBootstrapToken(ctx, clusterId, "")
	assert.NotNil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	iamMock.MockClientService.AssertNotCalled(t, "GetClientByClientId", mock.Anything, mock.Anything, mock.Anything)
}

func TestMintInstallerBootstrapToken_ClusterScoped_RepairsGrantOnNextMint(t *testing.T) {
	iamMock, authzClientMock, _, _, _, testSubject := initServiceAccountTest(t)
	mockInstallerBootstrapLock(t)
	ctx := enrichContext()
	clusterId := "cluster-1"
	perClusterClientId := "runai-internal.cluster-1.installer"
	scope := authorization_api.InternalCreateAccessRuleRequestScope{Id: clusterId, Type: authorization_api.SCOPETYPE_CLUSTER}

	iamMock.MockClientService.On("GetClients", ctx, RealmName, &iamModels.ClientFilter{
		ClusterId:  &clusterId,
		ClientType: ptr.To(realm_client.ClientTypeInstaller),
	}).Return([]iamModels.Client{}, nil).Once()
	iamMock.MockClientService.On("CreateClient", ctx, RealmName, mock.Anything).Return("keycloak-id", nil).Once()
	authzClientMock.On("SetServiceAccountPermissions", ctx, TenantId, perClusterClientId, scope,
		authorization_api.ALLOWEDINTERNALCREATEROLES_INSTALLER_SERVICE_ACCOUNT).Return(fmt.Errorf("error")).Once()

	resp, err := testSubject.MintInstallerBootstrapToken(ctx, clusterId, "")
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	perClusterClient := &iamModels.Client{
		ID:       ptr.To("keycloak-id"),
		Enabled:  ptr.To(true),
		ClientID: ptr.To(perClusterClientId),
		Secret:   ptr.To("secret"),
		Attributes: &iamModels.ClientAttributes{
			ClusterId: ptr.To(clusterId),
		},
	}
	rotatedSecret := "rotated-secret"
	tokens := &iamModels.Tokens{AccessToken: "bootstrap-jwt"}

	iamMock.MockClientService.On("GetClients", ctx, RealmName, &iamModels.ClientFilter{
		ClusterId:  &clusterId,
		ClientType: ptr.To(realm_client.ClientTypeInstaller),
	}).Return([]iamModels.Client{*perClusterClient}, nil).Once()
	authzClientMock.On("SetServiceAccountPermissions", ctx, TenantId, perClusterClientId, scope,
		authorization_api.ALLOWEDINTERNALCREATEROLES_INSTALLER_SERVICE_ACCOUNT).Return(nil).Once()
	iamMock.MockClientService.On("GetClientByClientId", ctx, RealmName, perClusterClientId).Return(perClusterClient, nil)
	iamMock.MockClientService.On("RegenerateClientSecret", ctx, RealmName, *perClusterClient.ID).Return(rotatedSecret, nil)
	iamMock.MockAuthService.On("AuthenticateClient", ctx, RealmName, perClusterClientId, rotatedSecret).Return(tokens, nil)

	resp, err = testSubject.MintInstallerBootstrapToken(ctx, clusterId, "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "bootstrap-jwt", resp.Body.(api.InstallerBootstrapResponse).Token)

	iamMock.MockClientService.AssertNumberOfCalls(t, "CreateClient", 1)
	authzClientMock.AssertNumberOfCalls(t, "SetServiceAccountPermissions", 2)
}

func TestMintInstallerBootstrapToken_ClusterScoped_CreateConflict_FallsBackToSetServiceAccountPermissions(t *testing.T) {
	iamMock, authzClientMock, _, _, _, testSubject := initServiceAccountTest(t)
	mockInstallerBootstrapLock(t)
	ctx := enrichContext()
	clusterId := "cluster-1"
	perClusterClientId := "runai-internal.cluster-1.installer"
	perClusterClient := &iamModels.Client{
		ID:       ptr.To("keycloak-id"),
		Enabled:  ptr.To(true),
		ClientID: ptr.To(perClusterClientId),
		Secret:   ptr.To("secret"),
		Attributes: &iamModels.ClientAttributes{
			ClusterId: ptr.To(clusterId),
		},
	}
	rotatedSecret := "rotated-secret"
	tokens := &iamModels.Tokens{AccessToken: "bootstrap-jwt"}

	iamMock.MockClientService.On("GetClients", ctx, RealmName, &iamModels.ClientFilter{
		ClusterId:  &clusterId,
		ClientType: ptr.To(realm_client.ClientTypeInstaller),
	}).Return([]iamModels.Client{}, nil)
	iamMock.MockClientService.On("CreateClient", ctx, RealmName, mock.Anything).
		Return("", fmt.Errorf("keycloak error: conflict: %w", iamError.ErrConflict))
	authzClientMock.On("SetServiceAccountPermissions", ctx, TenantId, perClusterClientId,
		authorization_api.InternalCreateAccessRuleRequestScope{Id: clusterId, Type: authorization_api.SCOPETYPE_CLUSTER},
		authorization_api.ALLOWEDINTERNALCREATEROLES_INSTALLER_SERVICE_ACCOUNT).Return(nil)
	iamMock.MockClientService.On("GetClientByClientId", ctx, RealmName, perClusterClientId).Return(perClusterClient, nil)
	iamMock.MockClientService.On("RegenerateClientSecret", ctx, RealmName, *perClusterClient.ID).Return(rotatedSecret, nil)
	iamMock.MockAuthService.On("AuthenticateClient", ctx, RealmName, perClusterClientId, rotatedSecret).Return(tokens, nil)

	resp, err := testSubject.MintInstallerBootstrapToken(ctx, clusterId, "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "bootstrap-jwt", resp.Body.(api.InstallerBootstrapResponse).Token)
}
