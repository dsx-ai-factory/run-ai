/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/generated/pkg/api"
	iamError "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/errors"
	iamMock "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/mocks"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	realm_client "github.com/run-ai/runai/control-plane/identity-manager/pkg/realm/client"
)

const (
	AccessKeyId   = "accessKeyId"
	AccessKeyName = "accessKeyName"
	createdBy     = "test@run.ai"
	clientId      = "clientId"
)

func initAccessKeysTest(t *testing.T) (iamServicesContainerMock *iamMock.MockServicesContainer, testSubject *AccessKeysAPIService) {
	mockIamServicesContainer := iamMock.NewMockServicesContainer()

	testSubject = &AccessKeysAPIService{
		iamClients: mockIamServicesContainer.ToServicesContainer().ClientService,
	}
	return mockIamServicesContainer, testSubject
}

func getClientModelForResponse() *models.Client {
	return &models.Client{
		ID:       ptr.To(AccessKeyId),
		ClientID: ptr.To("clientId"),
		Name:     ptr.To(AccessKeyName),
		Secret:   ptr.To("secret"),
		Attributes: &models.ClientAttributes{
			CreatedBy:  ptr.To(CreatedBy),
			ClientType: ptr.To(realm_client.ClientTypeUserApp),
		},
	}
}

func TestCreateAccessKey_HappyFlow(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()
	iamMock.MockClientService.On("GetClientsCount", ctx, RealmName, &models.ClientFilter{ClientType: ptr.To(realm_client.ClientTypeUserApp), IsTenantAdmin: false, CreatedBy: ptr.To(createdBy)}).Return(2, nil)
	iamMock.MockClientService.On("CreateClient", ctx, RealmName, mock.MatchedBy(func(c *models.Client) bool {
		return c != nil &&
			c.Name != nil && *c.Name == AccessKeyName &&
			c.ClientConfig != nil && c.ClientConfig.ServiceAccountsEnabled != nil && *c.ClientConfig.ServiceAccountsEnabled &&
			c.Attributes != nil && c.Attributes.ClientType != nil && *c.Attributes.ClientType == realm_client.ClientTypeUserApp &&
			c.Attributes.CreatedBy != nil && *c.Attributes.CreatedBy == createdBy &&
			c.Attributes.TokenLifespan != nil && *c.Attributes.TokenLifespan == DefaultServiceAccountTokenLifespan
	})).Return(AccessKeyId, nil)
	iamMock.MockClientService.On("SetEmailScopeOptional", ctx, RealmName, AccessKeyId).Return(nil)
	iamMock.MockClientService.On("GetClient", ctx, RealmName, AccessKeyId).Return(getClientModelForResponse(), nil)
	response, err := testSubject.CreateAccessKey(ctx, api.AccessKeyCreationRequest{
		Name: AccessKeyName,
	})
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusCreated, response.Code)

	accessKeyResp, ok := response.Body.(api.AccessKeyPostResponse)
	assert.True(t, ok)
	assert.Equal(t, AccessKeyName, accessKeyResp.Name)
}

func TestCreateAccessKey_GetClientsCount_Error(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()
	iamMock.MockClientService.On("GetClientsCount", ctx, RealmName, &models.ClientFilter{ClientType: ptr.To(realm_client.ClientTypeUserApp), IsTenantAdmin: false, CreatedBy: ptr.To(createdBy)}).Return(0, fmt.Errorf("error"))
	response, err := testSubject.CreateAccessKey(ctx, api.AccessKeyCreationRequest{
		Name: AccessKeyName,
	})
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestCreateAccessKey_CreateClient_Error(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()
	iamMock.MockClientService.On("GetClientsCount", ctx, RealmName, &models.ClientFilter{ClientType: ptr.To(realm_client.ClientTypeUserApp), IsTenantAdmin: false, CreatedBy: ptr.To(createdBy)}).Return(2, nil)
	iamMock.MockClientService.On("CreateClient", ctx, RealmName, mock.MatchedBy(func(c *models.Client) bool {
		return c != nil &&
			c.Name != nil && *c.Name == AccessKeyName &&
			c.ClientConfig != nil && c.ClientConfig.ServiceAccountsEnabled != nil && *c.ClientConfig.ServiceAccountsEnabled &&
			c.Attributes != nil && c.Attributes.ClientType != nil && *c.Attributes.ClientType == realm_client.ClientTypeUserApp &&
			c.Attributes.CreatedBy != nil && *c.Attributes.CreatedBy == createdBy
	})).Return("", fmt.Errorf("error"))
	response, err := testSubject.CreateAccessKey(ctx, api.AccessKeyCreationRequest{
		Name: AccessKeyName,
	})
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestCreateAccessKey_ExceedsLimit(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()
	iamMock.MockClientService.On("GetClientsCount", ctx, RealmName, &models.ClientFilter{ClientType: ptr.To(realm_client.ClientTypeUserApp), IsTenantAdmin: false, CreatedBy: ptr.To(createdBy)}).Return(21, nil)
	response, err := testSubject.CreateAccessKey(ctx, api.AccessKeyCreationRequest{
		Name: AccessKeyName,
	})
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestDeleteAccessKeyById_HappyFlow(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()
	iamMock.MockClientService.On("GetAccessKeyClientById", ctx, RealmName, createdBy, AccessKeyId).Return(getClientModelForResponse(), nil)
	iamMock.MockClientService.On("DeleteClient", ctx, RealmName, AccessKeyId).Return(nil)
	response, err := testSubject.DeleteAccessKeyById(ctx, AccessKeyId)
	assert.Nil(t, err)
	assert.Nil(t, response.Body)
	assert.Equal(t, http.StatusNoContent, response.Code)
}

func TestDeleteAccessKeyById_GetAppError(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()
	iamMock.MockClientService.On("GetAccessKeyClientById", ctx, RealmName, createdBy, AccessKeyId).Return(nil, fmt.Errorf("error"))
	response, err := testSubject.DeleteAccessKeyById(ctx, AccessKeyId)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestDeleteAccessKeyById_DeleteAppError(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()
	iamMock.MockClientService.On("GetAccessKeyClientById", ctx, RealmName, createdBy, AccessKeyId).Return(getClientModelForResponse(), nil)
	iamMock.MockClientService.On("DeleteClient", ctx, RealmName, AccessKeyId).Return(fmt.Errorf("error"))
	response, err := testSubject.DeleteAccessKeyById(ctx, AccessKeyId)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestGetAccessKeyById_HappyFlow(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()

	iamMock.MockClientService.On("GetAccessKeyClientById", ctx, RealmName, createdBy, AccessKeyId).Return(getClientModelForResponse(), nil)

	response, err := testSubject.GetAccessKeyById(ctx, AccessKeyId)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)

	accessKey, ok := response.Body.(api.AccessKey)
	assert.True(t, ok)
	assert.Equal(t, AccessKeyName, accessKey.Name)
}

func TestGetAccessKeyById_GetAppError(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()
	iamMock.MockClientService.On("GetAccessKeyClientById", ctx, RealmName, createdBy, AccessKeyId).Return(nil, fmt.Errorf("error"))
	response, err := testSubject.GetAccessKeyById(ctx, AccessKeyId)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestRegenerateAccessKeySecret_HappyFlow(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()
	iamMock.MockClientService.On("GetAccessKeyClientById", ctx, RealmName, createdBy, AccessKeyId).Return(getClientModelForResponse(), nil)
	iamMock.MockClientService.On("RegenerateClientSecret", ctx, RealmName, AccessKeyId).Return("secret", nil)

	response, err := testSubject.RegenerateAccessKeySecret(ctx, AccessKeyId)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestRegenerateAccessKeySecret_GetAppError(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()
	iamMock.MockClientService.On("GetAccessKeyClientById", ctx, RealmName, createdBy, AccessKeyId).Return(nil, fmt.Errorf("error"))
	response, err := testSubject.RegenerateAccessKeySecret(ctx, AccessKeyId)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestRegenerateAccessKeySecret_RegenerateSecretError(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()
	iamMock.MockClientService.On("GetAccessKeyClientById", ctx, RealmName, createdBy, AccessKeyId).Return(getClientModelForResponse(), nil)
	iamMock.MockClientService.On("RegenerateClientSecret", ctx, RealmName, AccessKeyId).Return("", fmt.Errorf("error"))

	response, err := testSubject.RegenerateAccessKeySecret(ctx, AccessKeyId)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestGetAccessKeys_HappyFlow(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()

	clientsArray := []models.Client{
		*getClientModelForResponse(),
		*getClientModelForResponse(),
	}
	iamMock.MockClientService.On("GetClients", ctx, RealmName, mock.Anything).Return(clientsArray, nil)

	response, err := testSubject.GetAccessKeys(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)

	accessKeysResp, ok := response.Body.(api.AccessKeys)
	assert.True(t, ok)
	assert.Equal(t, len(clientsArray), len(accessKeysResp.AccessKeys))
}

func TestGetAccessKeys_GetAppsError(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()
	iamMock.MockClientService.On("GetClients", ctx, RealmName, mock.Anything).Return(nil, fmt.Errorf("error"))

	response, err := testSubject.GetAccessKeys(ctx)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestGetAccessKeysAdministration_HappyFlow_CreatedBy(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()

	clientsArray := []models.Client{
		*getClientModelForResponse(),
		*getClientModelForResponse(),
	}
	iamMock.MockClientService.On("GetClients", ctx, RealmName, mock.Anything).Return(clientsArray, nil)

	response, err := testSubject.GetAccessKeysAdministration(ctx, "", createdBy)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)

	accessKeysResp, ok := response.Body.(api.AccessKeysAdministration)
	assert.True(t, ok)
	assert.Equal(t, len(clientsArray), len(accessKeysResp.AccessKeys))
}

func TestGetAccessKeysAdministration_HappyFlow_ClientId(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()

	iamMock.MockClientService.On("GetClientByClientId", ctx, RealmName, clientId).Return(getClientModelForResponse(), nil)

	response, err := testSubject.GetAccessKeysAdministration(ctx, clientId, "")
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)

	accessKeysResp, ok := response.Body.(api.AccessKeysAdministration)
	assert.True(t, ok)
	assert.Equal(t, 1, len(accessKeysResp.AccessKeys))
}

func TestGetAccessKeysAdministration_FiltersError(t *testing.T) {
	_, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()

	response, err := testSubject.GetAccessKeysAdministration(ctx, clientId, createdBy)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestGetAccessKeysAdministration_GetAppsError(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()

	iamMock.MockClientService.On("GetClients", ctx, RealmName, mock.Anything).Return(nil, fmt.Errorf("error"))

	response, err := testSubject.GetAccessKeysAdministration(ctx, "", createdBy)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestGetAccessKeysAdministration_GetAppByClientIdError(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()

	iamMock.MockClientService.On("GetClientByClientId", ctx, RealmName, clientId).Return(nil, fmt.Errorf("error"))
	response, err := testSubject.GetAccessKeysAdministration(ctx, clientId, "")

	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestGetAccessKeysAdministration_GetAppByClientIdError_NotFound(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()

	iamMock.MockClientService.On("GetClientByClientId", ctx, RealmName, clientId).Return(nil, iamError.ErrNotFound)
	response, err := testSubject.GetAccessKeysAdministration(ctx, clientId, "")

	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestDeleteAccessKeyByIdAdministration_HappyFlow(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()
	iamMock.MockClientService.On("GetClient", ctx, RealmName, AccessKeyId).Return(getClientModelForResponse(), nil)
	iamMock.MockClientService.On("DeleteClient", ctx, RealmName, AccessKeyId).Return(nil)
	response, err := testSubject.DeleteAccessKeyByIdAdministration(ctx, AccessKeyId)
	assert.Nil(t, err)
	assert.Nil(t, response.Body)
	assert.Equal(t, http.StatusNoContent, response.Code)
}

func TestDeleteAccessKeyByIdAdministration_InvalidUserApp(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()
	tenantAppClient := &models.Client{
		ID:       ptr.To(AccessKeyId),
		ClientID: ptr.To("clientId"),
		Name:     ptr.To(AccessKeyName),
		Secret:   ptr.To("secret"),
		Attributes: &models.ClientAttributes{
			CreatedBy:  ptr.To(CreatedBy),
			ClientType: ptr.To(realm_client.ClientTypeTenantApp),
		},
	}
	iamMock.MockClientService.On("GetClient", ctx, RealmName, AccessKeyId).Return(tenantAppClient, nil)
	response, err := testSubject.DeleteAccessKeyByIdAdministration(ctx, AccessKeyId)
	assert.NotNil(t, err)
	assert.Nil(t, response.Body)
	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestDeleteAccessKeyByIdAdministration_GetAppError(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()
	iamMock.MockClientService.On("GetClient", ctx, RealmName, AccessKeyId).Return(nil, fmt.Errorf("error"))
	response, err := testSubject.DeleteAccessKeyByIdAdministration(ctx, AccessKeyId)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestDeleteAccessKeyByIdAdministration_DeleteAppError(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()
	iamMock.MockClientService.On("GetClient", ctx, RealmName, AccessKeyId).Return(getClientModelForResponse(), nil)
	iamMock.MockClientService.On("DeleteClient", ctx, RealmName, AccessKeyId).Return(fmt.Errorf("error"))
	response, err := testSubject.DeleteAccessKeyByIdAdministration(ctx, AccessKeyId)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestCreateAccessKey_GetClient_Error(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()
	iamMock.MockClientService.On("GetClientsCount", ctx, RealmName, &models.ClientFilter{ClientType: ptr.To(realm_client.ClientTypeUserApp), IsTenantAdmin: false, CreatedBy: ptr.To(createdBy)}).Return(2, nil)
	iamMock.MockClientService.On("CreateClient", ctx, RealmName, mock.MatchedBy(func(c *models.Client) bool {
		return c != nil && c.Name != nil && *c.Name == AccessKeyName
	})).Return(AccessKeyId, nil)
	iamMock.MockClientService.On("SetEmailScopeOptional", ctx, RealmName, AccessKeyId).Return(nil)
	iamMock.MockClientService.On("GetClient", ctx, RealmName, AccessKeyId).Return(nil, fmt.Errorf("error"))
	response, err := testSubject.CreateAccessKey(ctx, api.AccessKeyCreationRequest{
		Name: AccessKeyName,
	})
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestGetAccessKeysAdministration_HappyFlow_NoFilter(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()

	clientsArray := []models.Client{
		*getClientModelForResponse(),
		*getClientModelForResponse(),
	}
	emptyCreatedBy := ""
	iamMock.MockClientService.On("GetClients", ctx, RealmName, &models.ClientFilter{ClientType: ptr.To(realm_client.ClientTypeUserApp), IsTenantAdmin: true, CreatedBy: &emptyCreatedBy}).Return(clientsArray, nil)

	response, err := testSubject.GetAccessKeysAdministration(ctx, "", "")
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)

	accessKeysResp, ok := response.Body.(api.AccessKeysAdministration)
	assert.True(t, ok)
	assert.Equal(t, len(clientsArray), len(accessKeysResp.AccessKeys))
}

func TestGetAccessKeysAdministration_ClientId_NotFound(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()

	iamMock.MockClientService.On("GetClientByClientId", ctx, RealmName, clientId).Return(nil, nil)

	response, err := testSubject.GetAccessKeysAdministration(ctx, clientId, "")
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestGetAccessKeysAdministration_ClientId_NonAccessKey(t *testing.T) {
	iamMock, testSubject := initAccessKeysTest(t)
	ctx := enrichContext()

	tenantAppClient := &models.Client{
		ID:       ptr.To(AccessKeyId),
		ClientID: ptr.To(clientId),
		Name:     ptr.To(AccessKeyName),
		Attributes: &models.ClientAttributes{
			ClientType: ptr.To(realm_client.ClientTypeTenantApp),
		},
	}
	iamMock.MockClientService.On("GetClientByClientId", ctx, RealmName, clientId).Return(tenantAppClient, nil)

	response, err := testSubject.GetAccessKeysAdministration(ctx, clientId, "")
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusNotFound, response.Code)
}
