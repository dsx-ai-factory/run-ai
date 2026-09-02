/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tenant

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
	runaipubsub_mocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/mocks"

	"github.com/google/uuid"
	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	iamMock "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/builders"
	buildersMock "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/builders/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/database/dao"
	dm "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/database/dao/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func newPubsuberMock(t *testing.T) *runaipubsub_mocks.PubSuber {
	return runaipubsub_mocks.NewPubSuber(t)
}

type tenantProviderMockManager struct {
	tenantDaoMock *dm.Tenant
	idpClient     *iamMock.IIdm
	buildersMock  *buildersMock.Interface
}

// assertAdditionalDataContent verifies AdditionalData has tenant_domain and customer email password.
func assertAdditionalDataContent(t *testing.T, additionalData *map[string]interface{}, tenantName, customerEmail string) {
	t.Helper()
	assert.NotNil(t, additionalData)
	ad := *additionalData
	assert.Contains(t, ad, "tenant_domain", "AdditionalData should contain tenant_domain")
	tenantDomain, ok := ad["tenant_domain"].(string)
	assert.True(t, ok, "tenant_domain should be a string")
	assert.NotEmpty(t, tenantDomain)
	assert.Contains(t, ad, customerEmail, "AdditionalData should contain customer email key")
	customerPass, ok := ad[customerEmail].(string)
	assert.True(t, ok, "customer password should be a string")
	assert.NotEmpty(t, customerPass)
}

func initTenantProviderTest(t *testing.T) (mm *tenantProviderMockManager, serviceTestSubject *TenantProviderImp) {
	mm = &tenantProviderMockManager{
		tenantDaoMock: dm.NewTenant(t),
		idpClient:     iamMock.NewIIdm(t),
		buildersMock:  buildersMock.NewInterface(t),
	}

	serviceTestSubject = &TenantProviderImp{
		tenantDao:      mm.tenantDaoMock,
		idpClient:      mm.idpClient,
		builder:        mm.buildersMock,
		serviceUtils:   utils.NewErrorUtils(),
		domainTemplate: "https://{{.}}.run.ai",
	}
	return
}

func TestGetFirstUserRole_NoRole(t *testing.T) {
	_, serviceTestSubject := initTenantProviderTest(t)
	role, err := serviceTestSubject.GetFirstUserRole(context.Background(), nil)
	assert.Equal(t, authorization_api.ALLOWEDINTERNALCREATEROLES_SYSTEM_ADMINISTRATOR.Ptr(), role)
	assert.Nil(t, err)
}

func TestGetFirstUserRole_SystemAdmin(t *testing.T) {
	_, serviceTestSubject := initTenantProviderTest(t)
	systemAdminRole := api.ROLES_SYSTEM_ADMINISTRATOR
	role, err := serviceTestSubject.GetFirstUserRole(context.Background(), &systemAdminRole)
	assert.Equal(t, authorization_api.ALLOWEDINTERNALCREATEROLES_SYSTEM_ADMINISTRATOR.Ptr(), role)
	assert.Nil(t, err)
}

func TestGetFirstUserRole_CloudOperator(t *testing.T) {
	_, serviceTestSubject := initTenantProviderTest(t)
	cloudOperatorRole := api.ROLES_CLOUD_OPERATOR
	role, err := serviceTestSubject.GetFirstUserRole(context.Background(), &cloudOperatorRole)
	assert.Equal(t, authorization_api.ALLOWEDINTERNALCREATEROLES_CLOUD_OPERATOR.Ptr(), role)
	assert.Nil(t, err)
}

func TestGetContractType_HappyFlow(t *testing.T) {
	_, serviceTestSubject := initTenantProviderTest(t)
	contractType := serviceTestSubject.GetContractType("type")
	assert.Equal(t, "type", contractType)
}

func TestGetContractType_EmptyType(t *testing.T) {
	_, serviceTestSubject := initTenantProviderTest(t)
	contractType := serviceTestSubject.GetContractType("")
	assert.Equal(t, models.ContractTypeNormal, contractType)
}

func TestGetTenantByIdInternal_HappyFlow(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	mm.tenantDaoMock.On("GetById", mock.Anything, mock.Anything).Return(getTenantModel(models.ContractTypeNormal), nil)
	tenant, err := serviceTestSubject.GetTenantById(context.Background(), 1)
	assert.NotNil(t, tenant)
	assert.Nil(t, err)
}

func TestGetTenantByIdInternal_DbError(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	mm.tenantDaoMock.On("GetById", mock.Anything, mock.Anything).Return(nil, getDaoError())
	tenant, err := serviceTestSubject.GetTenantById(context.Background(), 1)
	assert.NotNil(t, err)
	assert.Nil(t, tenant)
}

func TestDeleteTenantByIdInternal_SoftDelete_HappyFlow(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	mm.tenantDaoMock.On("GetByIdIncludeDeleted", mock.Anything, mock.Anything, false).Return(getTenantModel(models.ContractTypeNormal), nil)
	mm.tenantDaoMock.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything).Return(getTenantModel(models.ContractTypeNormal), nil)
	mm.idpClient.On("DeleteRealm", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	mm.tenantDaoMock.On("Delete", mock.Anything, mock.Anything).Return(nil)
	tenantUid, err := serviceTestSubject.DeleteTenantById(context.Background(), 1, false)
	assert.NotNil(t, tenantUid)
	assert.Nil(t, err)
}

func TestDeleteTenantByIdInternal_HardDelete_HappyFlow(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	mm.tenantDaoMock.On("GetByIdIncludeDeleted", mock.Anything, mock.Anything, true).Return(getTenantModel(models.ContractTypeNormal), nil)
	mm.tenantDaoMock.On("UpdateStatus", mock.Anything, mock.Anything, models.TenantDeleting).Return(getTenantModel(models.ContractTypeNormal), nil)
	pubsuberMock := newPubsuberMock(t)
	pubsuberMock.On("Publish", mock.Anything, mock.Anything).Return(nil)
	runaipubsub.SetMock(pubsuberMock)
	defer runaipubsub.SetMock(nil)
	// Polling: tenant is already gone on first check.
	mm.tenantDaoMock.On("GetById", mock.Anything, uint(1)).Return(nil, &dao.DaoError{Reason: dao.NotFound})
	tenantUid, err := serviceTestSubject.DeleteTenantById(context.Background(), 1, true)
	assert.NotNil(t, tenantUid)
	assert.Nil(t, err)
}

func TestDeleteTenantByIdInternal_GetById_Error(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	mm.tenantDaoMock.On("GetByIdIncludeDeleted", mock.Anything, mock.Anything, false).Return(nil, getDaoError())
	tenantUid, err := serviceTestSubject.DeleteTenantById(context.Background(), 1, false)
	assert.NotNil(t, err)
	assert.Nil(t, tenantUid)
}

func TestDeleteTenantByIdInternal_UpdateStatus_Error(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	mm.tenantDaoMock.On("GetByIdIncludeDeleted", mock.Anything, mock.Anything, false).Return(getTenantModel(models.ContractTypeNormal), nil)
	mm.tenantDaoMock.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything).Return(nil, getDaoError())
	tenantUid, err := serviceTestSubject.DeleteTenantById(context.Background(), 1, false)
	assert.NotNil(t, err)
	assert.Nil(t, tenantUid)
}

func TestDeleteTenantByIdInternal_DisableAuthRealm_Error(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	mm.tenantDaoMock.On("GetByIdIncludeDeleted", mock.Anything, mock.Anything, false).Return(getTenantModel(models.ContractTypeNormal), nil)
	mm.tenantDaoMock.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything).Return(getTenantModel(models.ContractTypeNormal), nil)
	mm.idpClient.On("DeleteRealm", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	mm.tenantDaoMock.On("Delete", mock.Anything, mock.Anything).Return(getDaoError())
	tenantUid, err := serviceTestSubject.DeleteTenantById(context.Background(), 1, false)
	assert.NotNil(t, err)
	assert.Nil(t, tenantUid)
}

func TestDeleteTenantByIdInternal_Delete_Error(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	mm.tenantDaoMock.On("GetByIdIncludeDeleted", mock.Anything, mock.Anything, false).Return(getTenantModel(models.ContractTypeNormal), nil)
	mm.tenantDaoMock.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything).Return(getTenantModel(models.ContractTypeNormal), nil)
	mm.idpClient.On("DeleteRealm", mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error"))
	tenantUid, err := serviceTestSubject.DeleteTenantById(context.Background(), 1, false)
	assert.NotNil(t, err)
	assert.Nil(t, tenantUid)
}

func TestCreateTenantInternal_NoramlContract_HappyFlow(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	mm.tenantDaoMock.On("GetByNameOrAlias", mock.Anything, mock.Anything).Return(nil, &dao.DaoError{Reason: dao.NotFound})
	mm.tenantDaoMock.On("Create", mock.Anything, mock.Anything).Return(getTenantModel(models.ContractTypeNormal), nil)
	pubsuberMock := newPubsuberMock(t)
	pubsuberMock.On("Publish", mock.Anything, mock.Anything).Return(nil)
	runaipubsub.SetMock(pubsuberMock)
	defer runaipubsub.SetMock(nil)
	response, err := serviceTestSubject.CreateTenant(context.Background(), getTenantCreationRequest(models.ContractTypeNormal))
	assert.NotNil(t, response)
	assert.Equal(t, models.ContractTypeNormal, response.Tenant.ContractType)
	assert.Nil(t, err)
	assertAdditionalDataContent(t, response.AdditionalData, "name", "email")
	pubsuberMock.AssertCalled(t, "Publish", mock.Anything, mock.Anything)
}

func TestCreateTenantInternal_Error_409(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	mm.tenantDaoMock.On("GetByNameOrAlias", mock.Anything, mock.Anything).Return(nil, &dao.DaoError{Reason: dao.NotFound})
	mm.tenantDaoMock.On("Create", mock.Anything, mock.Anything).Return(nil, &dao.DaoError{
		Reason: dao.Conflict,
		Err:    errors.New("error"),
	})

	response, err := serviceTestSubject.CreateTenant(context.Background(), getTenantCreationRequest(models.ContractTypeNormal))
	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusConflict, int(err.Code))
}

func TestCreateTenantInternal_CorrectBuilderInfoPassed(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	tenant := getTenantModel(models.ContractTypeNormal)
	createRequest := getTenantCreationRequest(models.ContractTypeNormal)
	mm.tenantDaoMock.On("GetByNameOrAlias", mock.Anything, mock.Anything).Return(nil, &dao.DaoError{Reason: dao.NotFound})
	mm.tenantDaoMock.On("Create", mock.Anything, mock.Anything).Return(tenant, nil)
	pubsuberMock := newPubsuberMock(t)
	pubsuberMock.On("Publish", mock.Anything, mock.Anything).Return(nil)
	runaipubsub.SetMock(pubsuberMock)
	defer runaipubsub.SetMock(nil)
	response, err := serviceTestSubject.CreateTenant(context.Background(), createRequest)
	assert.NotNil(t, response)
	assert.Nil(t, err)
	assertAdditionalDataContent(t, response.AdditionalData, "name", "email")
	pubsuberMock.AssertCalled(t, "Publish", mock.Anything, mock.Anything)
}

func TestCreateTenant_WithMetricStore_HappyFlow(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	mm.tenantDaoMock.On("GetByNameOrAlias", mock.Anything, mock.Anything).Return(nil, &dao.DaoError{Reason: dao.NotFound})
	mm.tenantDaoMock.On("Create", mock.Anything, mock.Anything).Return(getTenantModel(models.ContractTypeNormal), nil)
	pubsuberMock := newPubsuberMock(t)
	pubsuberMock.On("Publish", mock.Anything, mock.Anything).Return(nil)
	runaipubsub.SetMock(pubsuberMock)
	defer runaipubsub.SetMock(nil)
	response, err := serviceTestSubject.CreateTenant(context.Background(), getTenantCreationRequestWithMetricsStore())
	assert.NotNil(t, response)
	assert.Equal(t, models.ContractTypeNormal, response.Tenant.ContractType)
	assert.Nil(t, err)
	assertAdditionalDataContent(t, response.AdditionalData, "name", "email")
	pubsuberMock.AssertCalled(t, "Publish", mock.Anything, mock.Anything)
}

func TestCreateTenantInternal_CreateError(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	mm.tenantDaoMock.On("GetByNameOrAlias", mock.Anything, mock.Anything).Return(nil, &dao.DaoError{Reason: dao.NotFound})
	mm.tenantDaoMock.On("Create", mock.Anything, mock.Anything).Return(nil, getDaoError())
	response, err := serviceTestSubject.CreateTenant(context.Background(), getTenantCreationRequest(models.ContractTypeNormal))
	assert.Nil(t, response)
	assert.NotNil(t, err)
}

func TestCreateTenant_PublishError_RollsBackDB(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	mm.tenantDaoMock.On("GetByNameOrAlias", mock.Anything, mock.Anything).Return(nil, &dao.DaoError{Reason: dao.NotFound})
	mm.tenantDaoMock.On("Create", mock.Anything, mock.Anything).Return(getTenantModel(models.ContractTypeNormal), nil)
	mm.tenantDaoMock.On("HardDeleteById", mock.Anything, mock.Anything).Return(nil)
	pubsuberMock := newPubsuberMock(t)
	pubsuberMock.On("Publish", mock.Anything, mock.Anything).Return(fmt.Errorf("pubsub error"))
	runaipubsub.SetMock(pubsuberMock)
	defer runaipubsub.SetMock(nil)
	response, err := serviceTestSubject.CreateTenant(context.Background(), getTenantCreationRequest(models.ContractTypeNormal))
	assert.Nil(t, response)
	assert.NotNil(t, err)
	mm.tenantDaoMock.AssertCalled(t, "HardDeleteById", mock.Anything, mock.Anything)
}

func TestCreateTenant_NameConflictsWithExistingNameOrAlias(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	mm.tenantDaoMock.On("GetByNameOrAlias", mock.Anything, "name").Return(getTenantModel(models.ContractTypeNormal), nil)
	response, err := serviceTestSubject.CreateTenant(context.Background(), getTenantCreationRequest(models.ContractTypeNormal))
	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusConflict, int(err.Code))
}

func TestCreateTenant_AliasConflictsWithExistingNameOrAlias(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	alias := "existing-alias"
	mm.tenantDaoMock.On("GetByNameOrAlias", mock.Anything, "name").Return(nil, &dao.DaoError{Reason: dao.NotFound})
	mm.tenantDaoMock.On("GetByNameOrAlias", mock.Anything, alias).Return(getTenantModel(models.ContractTypeNormal), nil)

	request := getTenantCreationRequest(models.ContractTypeNormal)
	request.Alias = &alias
	response, err := serviceTestSubject.CreateTenant(context.Background(), request)
	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusConflict, int(err.Code))
}

func TestCreateTenant_WithAlias_HappyFlow(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	alias := "tenant-alias"
	mm.tenantDaoMock.On("GetByNameOrAlias", mock.Anything, mock.Anything).Return(nil, &dao.DaoError{Reason: dao.NotFound}).Twice()
	mm.tenantDaoMock.On("Create", mock.Anything, mock.Anything).Return(getTenantModel(models.ContractTypeNormal), nil)
	pubsuberMock := newPubsuberMock(t)
	pubsuberMock.On("Publish", mock.Anything, mock.Anything).Return(nil)
	runaipubsub.SetMock(pubsuberMock)
	defer runaipubsub.SetMock(nil)

	request := getTenantCreationRequest(models.ContractTypeNormal)
	request.Alias = &alias
	response, err := serviceTestSubject.CreateTenant(context.Background(), request)
	assert.NotNil(t, response)
	assert.Nil(t, err)
	assertAdditionalDataContent(t, response.AdditionalData, "name", "email")
	pubsuberMock.AssertCalled(t, "Publish", mock.Anything, mock.Anything)
}

func TestCreateTenant_NameEqualsAlias_ReturnsBadRequest(t *testing.T) {
	_, serviceTestSubject := initTenantProviderTest(t)
	alias := "name"
	request := getTenantCreationRequest(models.ContractTypeNormal)
	request.Alias = &alias
	response, err := serviceTestSubject.CreateTenant(context.Background(), request)
	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, int(err.Code))
	assert.Equal(t, "tenant name and alias must not be the same", err.Message)
}

func TestCreateTenant_NameEqualsAlias_ValidateCaseInsensitive_ReturnsBadRequest(t *testing.T) {
	_, serviceTestSubject := initTenantProviderTest(t)
	alias := "Name"
	request := getTenantCreationRequest(models.ContractTypeNormal)
	request.Alias = &alias
	response, err := serviceTestSubject.CreateTenant(context.Background(), request)
	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, int(err.Code))
	assert.Equal(t, "tenant name and alias must not be the same", err.Message)
}

func TestCreateTenant_ValidateAliasIsStoredInLowercase(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	alias := "My-Alias"
	mm.tenantDaoMock.On("GetByNameOrAlias", mock.Anything, mock.Anything).Return(nil, &dao.DaoError{Reason: dao.NotFound})
	mm.tenantDaoMock.On("Create", mock.Anything, mock.MatchedBy(func(tenant models.Tenant) bool {
		return tenant.Alias != nil && *tenant.Alias == "my-alias"
	})).Return(getTenantModel(models.ContractTypeNormal), nil)
	pubsuberMock := newPubsuberMock(t)
	pubsuberMock.On("Publish", mock.Anything, mock.Anything).Return(nil)
	runaipubsub.SetMock(pubsuberMock)
	defer runaipubsub.SetMock(nil)

	request := getTenantCreationRequest(models.ContractTypeNormal)
	request.Alias = &alias
	response, err := serviceTestSubject.CreateTenant(context.Background(), request)
	assert.NotNil(t, response)
	assert.Nil(t, err)
	assertAdditionalDataContent(t, response.AdditionalData, "name", "email")
	pubsuberMock.AssertCalled(t, "Publish", mock.Anything, mock.Anything)
}

func TestCreateTenant_NameEqualsAlias_NilAlias_Succeeds(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	mm.tenantDaoMock.On("GetByNameOrAlias", mock.Anything, mock.Anything).Return(nil, &dao.DaoError{Reason: dao.NotFound})
	mm.tenantDaoMock.On("Create", mock.Anything, mock.Anything).Return(getTenantModel(models.ContractTypeNormal), nil)
	pubsuberMock := newPubsuberMock(t)
	pubsuberMock.On("Publish", mock.Anything, mock.Anything).Return(nil)
	runaipubsub.SetMock(pubsuberMock)
	defer runaipubsub.SetMock(nil)

	request := getTenantCreationRequest(models.ContractTypeNormal)
	request.Alias = nil
	response, err := serviceTestSubject.CreateTenant(context.Background(), request)
	assert.NotNil(t, response)
	assert.Nil(t, err)
	assertAdditionalDataContent(t, response.AdditionalData, "name", "email")
	pubsuberMock.AssertCalled(t, "Publish", mock.Anything, mock.Anything)
}

func TestCreateTenant_NameEqualsAlias_EmptyAlias_Succeeds(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	mm.tenantDaoMock.On("GetByNameOrAlias", mock.Anything, mock.Anything).Return(nil, &dao.DaoError{Reason: dao.NotFound})
	mm.tenantDaoMock.On("Create", mock.Anything, mock.Anything).Return(getTenantModel(models.ContractTypeNormal), nil)
	pubsuberMock := newPubsuberMock(t)
	pubsuberMock.On("Publish", mock.Anything, mock.Anything).Return(nil)
	runaipubsub.SetMock(pubsuberMock)
	defer runaipubsub.SetMock(nil)

	emptyAlias := ""
	request := getTenantCreationRequest(models.ContractTypeNormal)
	request.Alias = &emptyAlias
	response, err := serviceTestSubject.CreateTenant(context.Background(), request)
	assert.NotNil(t, response)
	assert.Nil(t, err)
	assertAdditionalDataContent(t, response.AdditionalData, "name", "email")
	pubsuberMock.AssertCalled(t, "Publish", mock.Anything, mock.Anything)
}

func TestGetAllTenants_HappyFlow(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	mm.tenantDaoMock.On("GetAll", mock.Anything, mock.Anything, mock.Anything).Return(getTenantArrayModel(), nil)
	response, err := serviceTestSubject.GetAllTenants(context.Background(), nil, nil)
	assert.NotNil(t, response)
	assert.Equal(t, 2, len(*response))
	assert.Nil(t, err)
}

func TestGetAllTenants_DbError(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	mm.tenantDaoMock.On("GetAll", mock.Anything, mock.Anything, mock.Anything).Return(nil, getDaoError())
	response, err := serviceTestSubject.GetAllTenants(context.Background(), nil, nil)
	assert.Nil(t, response)
	assert.NotNil(t, err)
}

func TestGetAllTenants_WithNameFilter(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	filters := &dao.TenantFilters{Name: "test-tenant"}
	mm.tenantDaoMock.On("GetAll", mock.Anything, mock.Anything, filters).Return(getTenantArrayModel(), nil)
	response, err := serviceTestSubject.GetAllTenants(context.Background(), nil, filters)
	assert.NotNil(t, response)
	assert.Nil(t, err)
	mm.tenantDaoMock.AssertCalled(t, "GetAll", mock.Anything, mock.Anything, filters)
}

func TestGetAllTenants_WithAliasFilter(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	filters := &dao.TenantFilters{Alias: "test-alias"}
	mm.tenantDaoMock.On("GetAll", mock.Anything, mock.Anything, filters).Return(getTenantArrayModel(), nil)
	response, err := serviceTestSubject.GetAllTenants(context.Background(), nil, filters)
	assert.NotNil(t, response)
	assert.Nil(t, err)
	mm.tenantDaoMock.AssertCalled(t, "GetAll", mock.Anything, mock.Anything, filters)
}

func TestGetAllTenants_WithBothFilters(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	filters := &dao.TenantFilters{Name: "test-tenant", Alias: "test-alias"}
	mm.tenantDaoMock.On("GetAll", mock.Anything, mock.Anything, filters).Return(getTenantArrayModel(), nil)
	response, err := serviceTestSubject.GetAllTenants(context.Background(), nil, filters)
	assert.NotNil(t, response)
	assert.Nil(t, err)
	mm.tenantDaoMock.AssertCalled(t, "GetAll", mock.Anything, mock.Anything, filters)
}

func TestGetTenantByName_HappyFlow(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	mm.tenantDaoMock.On("GetByName", mock.Anything, mock.Anything).Return(getTenantModel(models.ContractTypeNormal), nil)
	tenant, err := serviceTestSubject.GetTenantByName(context.Background(), "name")
	assert.NotNil(t, tenant)
	assert.Nil(t, err)
}

func TestGetTenantByName_DbError(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	mm.tenantDaoMock.On("GetByName", mock.Anything, mock.Anything).Return(nil, getDaoError())
	tenant, err := serviceTestSubject.GetTenantByName(context.Background(), "name")
	assert.NotNil(t, err)
	assert.Nil(t, tenant)
}

func TestHardDeleteTenant_HappyFlow(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	tenantModel := getTenantModel(models.ContractTypeNormal)
	mm.tenantDaoMock.On("UpdateStatus", mock.Anything, mock.Anything, models.TenantDeleting).Return(tenantModel, nil)
	pubsuberMock := newPubsuberMock(t)
	pubsuberMock.On("Publish", mock.Anything, mock.Anything).Return(nil)
	runaipubsub.SetMock(pubsuberMock)
	defer runaipubsub.SetMock(nil)
	// Polling: tenant is already gone on first check.
	mm.tenantDaoMock.On("GetById", mock.Anything, uint(1)).Return(nil, &dao.DaoError{Reason: dao.NotFound})
	err := serviceTestSubject.hardDeleteTenant(context.Background(), 1, tenantModel)
	assert.Nil(t, err)
}

func TestHardDeleteTenant_UpdateStatus_Error(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	tenantModel := getTenantModel(models.ContractTypeNormal)
	mm.tenantDaoMock.On("UpdateStatus", mock.Anything, mock.Anything, models.TenantDeleting).Return(nil, getDaoError())
	err := serviceTestSubject.hardDeleteTenant(context.Background(), 1, tenantModel)
	assert.NotNil(t, err)
}

func TestHardDeleteTenant_Publish_Error(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	tenantModel := getTenantModel(models.ContractTypeNormal)
	mm.tenantDaoMock.On("UpdateStatus", mock.Anything, mock.Anything, models.TenantDeleting).Return(tenantModel, nil)
	pubsuberMock := newPubsuberMock(t)
	pubsuberMock.On("Publish", mock.Anything, mock.Anything).Return(fmt.Errorf("pubsub error"))
	runaipubsub.SetMock(pubsuberMock)
	defer runaipubsub.SetMock(nil)
	err := serviceTestSubject.hardDeleteTenant(context.Background(), 1, tenantModel)
	assert.NotNil(t, err)
}

func TestSoftDeleteTenant_HappyFlow(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	tenantModel := getTenantModel(models.ContractTypeNormal)
	mm.tenantDaoMock.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything).Return(getTenantModel(models.ContractTypeNormal), nil)
	mm.idpClient.On("DeleteRealm", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	mm.tenantDaoMock.On("Delete", mock.Anything, mock.Anything).Return(nil)
	err := serviceTestSubject.softDeleteTenant(context.Background(), 1, tenantModel)
	assert.Nil(t, err)
}

func TestSoftDeleteTenant_UpdateStatus_Error(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	tenantModel := getTenantModel(models.ContractTypeNormal)
	mm.tenantDaoMock.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything).Return(nil, getDaoError())
	err := serviceTestSubject.softDeleteTenant(context.Background(), 1, tenantModel)
	assert.NotNil(t, err)
}

func TestSoftDeleteTenant_DeleteRealm_Error(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	tenantModel := getTenantModel(models.ContractTypeNormal)
	mm.tenantDaoMock.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything).Return(getTenantModel(models.ContractTypeNormal), nil)
	mm.idpClient.On("DeleteRealm", mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error"))
	err := serviceTestSubject.softDeleteTenant(context.Background(), 1, tenantModel)
	assert.NotNil(t, err)
}

func TestSoftDeleteTenant_DeleteRealm_404_Succeeds(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	tenantModel := getTenantModel(models.ContractTypeNormal)
	mm.tenantDaoMock.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything).Return(getTenantModel(models.ContractTypeNormal), nil)
	mm.idpClient.On("DeleteRealm", mock.Anything, mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusNotFound}, fmt.Errorf("realm not found"))
	mm.tenantDaoMock.On("Delete", mock.Anything, mock.Anything).Return(nil)
	err := serviceTestSubject.softDeleteTenant(context.Background(), 1, tenantModel)
	assert.Nil(t, err)
}

func TestSoftDeleteTenant_db_error(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	tenantModel := getTenantModel(models.ContractTypeNormal)
	mm.tenantDaoMock.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything).Return(getTenantModel(models.ContractTypeNormal), nil)
	mm.idpClient.On("DeleteRealm", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	mm.tenantDaoMock.On("Delete", mock.Anything, mock.Anything).Return(getDaoError())
	err := serviceTestSubject.softDeleteTenant(context.Background(), 1, tenantModel)
	assert.NotNil(t, err)
}

func TestGetTenantFromRequest_AliasIsLowercased(t *testing.T) {
	_, serviceTestSubject := initTenantProviderTest(t)
	alias := "My-Alias"
	request := getTenantCreationRequest(models.ContractTypeNormal)
	request.Alias = &alias
	tenantModel := serviceTestSubject.getTenantModelFromRequest(request)
	assert.NotNil(t, tenantModel.Alias)
	assert.Equal(t, "my-alias", *tenantModel.Alias)
}

func TestUpdateTenantAlias_HappyFlow(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	existingTenant := getTenantModel(models.ContractTypeNormal)
	alias := "new-alias"
	mm.tenantDaoMock.On("GetById", mock.Anything, uint(1)).Return(existingTenant, nil)
	mm.tenantDaoMock.On("GetByNameOrAlias", mock.Anything, alias).Return(nil, &dao.DaoError{Reason: dao.NotFound})
	mm.tenantDaoMock.On("UpdateByField", mock.Anything, uint(1), "Alias", &alias).Return(existingTenant, nil)
	tenant, err := serviceTestSubject.UpdateTenantAlias(context.Background(), 1, &alias)
	assert.NotNil(t, tenant)
	assert.Nil(t, err)
}

func TestUpdateTenantAlias_NilAlias(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	existingTenant := getTenantModel(models.ContractTypeNormal)
	mm.tenantDaoMock.On("GetById", mock.Anything, uint(1)).Return(existingTenant, nil)
	mm.tenantDaoMock.On("UpdateByField", mock.Anything, uint(1), "Alias", (*string)(nil)).Return(existingTenant, nil)
	tenant, err := serviceTestSubject.UpdateTenantAlias(context.Background(), 1, nil)
	assert.NotNil(t, tenant)
	assert.Nil(t, err)
}

func TestUpdateTenantAlias_EmptyAlias(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	existingTenant := getTenantModel(models.ContractTypeNormal)
	emptyAlias := ""
	mm.tenantDaoMock.On("GetById", mock.Anything, uint(1)).Return(existingTenant, nil)
	mm.tenantDaoMock.On("UpdateByField", mock.Anything, uint(1), "Alias", &emptyAlias).Return(existingTenant, nil)
	tenant, err := serviceTestSubject.UpdateTenantAlias(context.Background(), 1, &emptyAlias)
	assert.NotNil(t, tenant)
	assert.Nil(t, err)
}

func TestUpdateTenantAlias_AliasIsLowercased(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	existingTenant := getTenantModel(models.ContractTypeNormal)
	alias := "New-Alias"
	expectedLowercased := "new-alias"
	mm.tenantDaoMock.On("GetById", mock.Anything, uint(1)).Return(existingTenant, nil)
	mm.tenantDaoMock.On("GetByNameOrAlias", mock.Anything, expectedLowercased).Return(nil, &dao.DaoError{Reason: dao.NotFound})
	mm.tenantDaoMock.On("UpdateByField", mock.Anything, uint(1), "Alias", &expectedLowercased).Return(existingTenant, nil)
	tenant, err := serviceTestSubject.UpdateTenantAlias(context.Background(), 1, &alias)
	assert.NotNil(t, tenant)
	assert.Nil(t, err)
}

func TestUpdateTenantAlias_AliasEqualsName_ReturnsBadRequest(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	existingTenant := getTenantModel(models.ContractTypeNormal)
	alias := existingTenant.Name
	mm.tenantDaoMock.On("GetById", mock.Anything, uint(1)).Return(existingTenant, nil)
	tenant, err := serviceTestSubject.UpdateTenantAlias(context.Background(), 1, &alias)
	assert.Nil(t, tenant)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, int(err.Code))
	assert.Equal(t, "tenant name and alias must not be the same", err.Message)
}

func TestUpdateTenantAlias_AliasEqualsName_DifferentCase_ReturnsBadRequest(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	existingTenant := getTenantModel(models.ContractTypeNormal)
	alias := "Name"
	mm.tenantDaoMock.On("GetById", mock.Anything, uint(1)).Return(existingTenant, nil)
	tenant, err := serviceTestSubject.UpdateTenantAlias(context.Background(), 1, &alias)
	assert.Nil(t, tenant)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, int(err.Code))
	assert.Equal(t, "tenant name and alias must not be the same", err.Message)
}

func TestUpdateTenantAlias_TenantNotFound(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	alias := "new-alias"
	mm.tenantDaoMock.On("GetById", mock.Anything, uint(1)).Return(nil, &dao.DaoError{Reason: dao.NotFound, Err: errors.New("not found")})
	tenant, err := serviceTestSubject.UpdateTenantAlias(context.Background(), 1, &alias)
	assert.Nil(t, tenant)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusNotFound, int(err.Code))
}

func TestUpdateTenantAlias_ConflictWithExistingTenant(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	existingTenant := getTenantModel(models.ContractTypeNormal)
	existingTenant.ID = 1
	conflictingTenant := getTenantModel(models.ContractTypeNormal)
	conflictingTenant.ID = 2
	alias := "existing-alias"
	mm.tenantDaoMock.On("GetById", mock.Anything, uint(1)).Return(existingTenant, nil)
	mm.tenantDaoMock.On("GetByNameOrAlias", mock.Anything, alias).Return(conflictingTenant, nil)
	tenant, err := serviceTestSubject.UpdateTenantAlias(context.Background(), 1, &alias)
	assert.Nil(t, tenant)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusConflict, int(err.Code))
}

func TestUpdateTenantAlias_SameTenantAlias(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	existingTenant := getTenantModel(models.ContractTypeNormal)
	existingTenant.ID = 1
	alias := "same-alias"
	mm.tenantDaoMock.On("GetById", mock.Anything, uint(1)).Return(existingTenant, nil)
	mm.tenantDaoMock.On("GetByNameOrAlias", mock.Anything, alias).Return(existingTenant, nil)
	mm.tenantDaoMock.On("UpdateByField", mock.Anything, uint(1), "Alias", &alias).Return(existingTenant, nil)
	tenant, err := serviceTestSubject.UpdateTenantAlias(context.Background(), 1, &alias)
	assert.NotNil(t, tenant)
	assert.Nil(t, err)
}

func TestUpdateTenantAlias_GetByNameOrAliasDbError(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	existingTenant := getTenantModel(models.ContractTypeNormal)
	alias := "new-alias"
	mm.tenantDaoMock.On("GetById", mock.Anything, uint(1)).Return(existingTenant, nil)
	mm.tenantDaoMock.On("GetByNameOrAlias", mock.Anything, alias).Return(nil, getDaoError())
	tenant, err := serviceTestSubject.UpdateTenantAlias(context.Background(), 1, &alias)
	assert.Nil(t, tenant)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, int(err.Code))
}

func TestUpdateTenantAlias_UpdateByFieldDbError(t *testing.T) {
	mm, serviceTestSubject := initTenantProviderTest(t)
	existingTenant := getTenantModel(models.ContractTypeNormal)
	alias := "new-alias"
	mm.tenantDaoMock.On("GetById", mock.Anything, uint(1)).Return(existingTenant, nil)
	mm.tenantDaoMock.On("GetByNameOrAlias", mock.Anything, alias).Return(nil, &dao.DaoError{Reason: dao.NotFound})
	mm.tenantDaoMock.On("UpdateByField", mock.Anything, uint(1), "Alias", &alias).Return(nil, getDaoError())
	tenant, err := serviceTestSubject.UpdateTenantAlias(context.Background(), 1, &alias)
	assert.Nil(t, tenant)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, int(err.Code))
}

func getTenantInfo() *builders.TenantInfo {
	return &builders.TenantInfo{
		Name: "name",
		Id:   1,
	}
}

func getTenantModel(contractType string) *models.Tenant {
	alias := "alias"
	return &models.Tenant{
		Name:         "name",
		DisplayName:  "display-name",
		Uuid:         uuid.New().String(),
		Status:       models.TenantCreating,
		ContractType: contractType,
		MimirId:      uuid.New().String(),
		Alias:        &alias,
	}
}

func getTenantArrayModel() *[]models.Tenant {
	alias1 := "alias1"
	alias2 := "alias2"
	return &[]models.Tenant{
		{
			Name:        "name1",
			DisplayName: "display-name1",
			Uuid:        uuid.New().String(),
			Alias:       &alias1,
		},
		{
			Name:        "name2",
			DisplayName: "display-name2",
			Uuid:        uuid.New().String(),
			Alias:       &alias2,
		},
	}
}

func getTenantCreationRequest(contractType string) api.TenantCreationRequest {
	return api.TenantCreationRequest{
		Name:         "name",
		Email:        "email",
		ContractType: contractType,
	}
}

func getTenantCreationRequestWithMetricsStore() api.TenantCreationRequest {
	return api.TenantCreationRequest{
		Name:         "name",
		Email:        "email",
		ContractType: models.ContractTypeNormal,
		MetricStore: &api.MetricStore{
			Read: api.MetricStoreEndpoint{
				Url: "url",
				Auth: &api.MetricStoreEndpointAuth{
					Basic: &api.MetricStoreEndpointAuthBasic{
						Username: "username",
						Password: "password",
					},
				},
				Headers: &map[string]string{
					"key": "value",
				},
			},
			Write: api.MetricStoreEndpoint{
				Url: "url",
				Auth: &api.MetricStoreEndpointAuth{
					Basic: &api.MetricStoreEndpointAuthBasic{
						Username: "username",
						Password: "password",
					},
				},
				Headers: &map[string]string{
					"key": "value",
				},
			},
		},
	}
}

func getDaoError() *dao.DaoError {
	return &dao.DaoError{
		Reason: dao.General,
		Err:    errors.Join(dao.ErrNotFound),
	}
}

// ---------------------------------------------------------------------------
// RequeueProvision
// ---------------------------------------------------------------------------

func existingFailedTenant() *models.Tenant {
	return &models.Tenant{
		Model:        gorm.Model{ID: 99},
		Name:         "failed-tenant",
		MimirId:      "mimir-99",
		ContractType: "normal",
		Status:       models.TenantFailed,
	}
}

func TestRequeueProvision_HappyFlow(t *testing.T) {
	mm, svc := initTenantProviderTest(t)
	tenant := existingFailedTenant()

	mm.tenantDaoMock.On("UpdateStatus", mock.Anything, uint(99), models.TenantCreating).Return(tenant, nil)

	pubsuberMock := newPubsuberMock(t)
	pubsuberMock.On("Publish", mock.Anything, mock.Anything).Return(nil)
	runaipubsub.SetMock(pubsuberMock)
	defer runaipubsub.SetMock(nil)

	apiErr := svc.RequeueProvision(context.Background(), tenant, ProvisionConfig{Email: "admin@test.ai", Role: authorization_api.ALLOWEDINTERNALCREATEROLES_SYSTEM_ADMINISTRATOR})
	assert.Nil(t, apiErr)
	mm.tenantDaoMock.AssertCalled(t, "UpdateStatus", mock.Anything, uint(99), models.TenantCreating)
}

func TestRequeueProvision_PublishFails_RestoresFailed(t *testing.T) {
	mm, svc := initTenantProviderTest(t)
	tenant := existingFailedTenant()

	mm.tenantDaoMock.On("UpdateStatus", mock.Anything, uint(99), models.TenantCreating).Return(tenant, nil)

	pubsuberMock := newPubsuberMock(t)
	pubsuberMock.On("Publish", mock.Anything, mock.Anything).Return(errors.New("nats down"))
	runaipubsub.SetMock(pubsuberMock)
	defer runaipubsub.SetMock(nil)

	mm.tenantDaoMock.On("UpdateStatus", mock.Anything, uint(99), models.TenantFailed).Return(tenant, nil)

	apiErr := svc.RequeueProvision(context.Background(), tenant, ProvisionConfig{Email: "admin@test.ai", Role: authorization_api.ALLOWEDINTERNALCREATEROLES_SYSTEM_ADMINISTRATOR})
	assert.NotNil(t, apiErr)
	assert.Equal(t, int32(http.StatusInternalServerError), apiErr.Code)
	mm.tenantDaoMock.AssertCalled(t, "UpdateStatus", mock.Anything, uint(99), models.TenantFailed)
}

func TestRequeueProvision_ResetFails_ReturnsError(t *testing.T) {
	mm, svc := initTenantProviderTest(t)
	tenant := existingFailedTenant()

	mm.tenantDaoMock.On("UpdateStatus", mock.Anything, uint(99), models.TenantCreating).
		Return(nil, &dao.DaoError{Reason: dao.General, Err: errors.New("db error")})

	apiErr := svc.RequeueProvision(context.Background(), tenant, ProvisionConfig{Email: "admin@test.ai", Role: authorization_api.ALLOWEDINTERNALCREATEROLES_SYSTEM_ADMINISTRATOR})
	assert.NotNil(t, apiErr)
	assert.Equal(t, int32(http.StatusInternalServerError), apiErr.Code)
}
