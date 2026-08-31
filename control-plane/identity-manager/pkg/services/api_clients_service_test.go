/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/generated/pkg/api"
	iamMock "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/mocks"
	iamModels "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	businessMocks "github.com/run-ai/runai/control-plane/identity-manager/pkg/business/mocks"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/models"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	clientMock "github.com/run-ai/runai/control-plane/identity-manager/pkg/utils/mocks"
)

func initClientServiceTest(t *testing.T) (iamServicesContainerMock *iamMock.MockServicesContainer, realmDalMock *businessMocks.RealmDal, testSubject *ClientsApiService) {
	mockIamServicesContainer := iamMock.NewMockServicesContainer()
	realmDalMock = businessMocks.NewRealmDal(t)

	testSubject = &ClientsApiService{
		iamClients: mockIamServicesContainer.ToServicesContainer().ClientService,
		realmDal:   realmDalMock,
	}
	return mockIamServicesContainer, realmDalMock, testSubject
}

func TestGetAllClients_HappyFlow(t *testing.T) {
	iamMock, _, testSubject := initClientServiceTest(t)
	clients := getMockClients()
	ctx := context.Background()
	ctx = context.WithValue(ctx, utils.AuthRealmKey, clientMock.REALM_NAME)
	filters := &iamModels.ClientFilter{}
	iamMock.MockClientService.On("GetClients", ctx, clientMock.REALM_NAME, filters).Return(clients, nil)

	response, err := testSubject.GetAllClients(ctx, "")
	assert.Nil(t, err)
	assert.Equal(t, len(clients), len(response.Body.([]iamModels.Client)))
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestGetAllClients_IamError(t *testing.T) {
	iamMock, _, testSubject := initClientServiceTest(t)
	ctx := context.Background()
	ctx = context.WithValue(ctx, utils.AuthRealmKey, clientMock.REALM_NAME)
	filters := &iamModels.ClientFilter{}
	iamMock.MockClientService.On("GetClients", ctx, clientMock.REALM_NAME, filters).Return(nil, assert.AnError)

	response, err := testSubject.GetAllClients(ctx, "")
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestCreateInternalClients_HappyFlow(t *testing.T) {
	iamMock, _, testSubject := initClientServiceTest(t)
	ctx := context.Background()
	ctx = context.WithValue(ctx, utils.AuthRealmKey, clientMock.REALM_NAME)
	iamMock.MockClientService.On("CreateClient", ctx, clientMock.REALM_NAME, mock.Anything).Return(clientMock.ID_OF_CLIENT, nil)

	response, err := testSubject.CreateInternalClient(ctx, api.ClientCreationRequest{
		Type:     "Agent",
		ClientId: clientMock.CLIENT_ID,
		TenantId: clientMock.TENANT_ID,
	})
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, response.Code)
}

func TestCreateInternalClients_NonSupportedType_InternalError(t *testing.T) {
	_, _, testSubject := initClientServiceTest(t)
	ctx := context.Background()
	ctx = context.WithValue(ctx, utils.AuthRealmKey, clientMock.REALM_NAME)
	response, err := testSubject.CreateInternalClient(ctx, api.ClientCreationRequest{
		Type:     "newType",
		ClientId: clientMock.CLIENT_ID,
		TenantId: clientMock.TENANT_ID,
	})
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestCreateInternalClients_InternalError(t *testing.T) {
	apiError := utils.Error{
		Code: http.StatusInternalServerError,
	}
	iamMock, dalMock, testSubject := initClientServiceTest(t)
	dalMock.On("GetRealmByTenantId", mock.Anything, clientMock.TENANT_ID).Return(&models.Realm{TenantId: clientMock.TENANT_ID, Realm: clientMock.REALM_NAME}, nil)
	iamMock.MockClientService.On("CreateClient", mock.Anything, clientMock.REALM_NAME, mock.Anything).Return("", apiError)
	response, err := testSubject.CreateInternalClient(context.Background(), api.ClientCreationRequest{
		Type:     "Agent",
		ClientId: clientMock.CLIENT_ID,
		TenantId: clientMock.TENANT_ID,
	})
	assert.NotNil(t, err)
	assert.Nil(t, response.Body)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestDeleteClient_HappyFlow(t *testing.T) {
	iamMock, _, testSubject := initClientServiceTest(t)
	ctx := context.Background()
	ctx = context.WithValue(ctx, utils.AuthRealmKey, clientMock.REALM_NAME)
	iamMock.MockClientService.On("DeleteClient", ctx, clientMock.REALM_NAME, clientMock.ID_OF_CLIENT).Return(nil)
	response, err := testSubject.DeleteClient(ctx, clientMock.ID_OF_CLIENT)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, response.Code)
}

func TestDeleteClient_InternalError(t *testing.T) {
	apiError := utils.Error{
		Code: http.StatusInternalServerError,
	}
	iamMock, _, testSubject := initClientServiceTest(t)
	ctx := context.Background()
	ctx = context.WithValue(ctx, utils.AuthRealmKey, clientMock.REALM_NAME)
	iamMock.MockClientService.On("DeleteClient", ctx, clientMock.REALM_NAME, clientMock.ID_OF_CLIENT).Return(apiError)
	response, err := testSubject.DeleteClient(ctx, clientMock.ID_OF_CLIENT)
	assert.NotNil(t, err)
	assert.Nil(t, response.Body)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestGetClientSecret_HappyFlow(t *testing.T) {
	iamMock, _, testSubject := initClientServiceTest(t)
	ctx := context.Background()
	ctx = context.WithValue(ctx, utils.AuthRealmKey, clientMock.REALM_NAME)
	iamMock.MockClientService.On("GetClientSecret", ctx, clientMock.REALM_NAME, clientMock.ID_OF_CLIENT).Return(clientMock.CLIENT_SECRET, nil)
	response, err := testSubject.GetClientSecret(ctx, clientMock.ID_OF_CLIENT)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, clientMock.CLIENT_SECRET, response.Body.(api.GetClientSecret200Response).Secret)
}

func TestGetRealm_RealmFromCtx(t *testing.T) {
	_, _, testSubject := initClientServiceTest(t)
	ctx := context.Background()
	ctx = context.WithValue(ctx, utils.AuthRealmKey, "ctxRealm")
	realmName := testSubject.getRealmName(ctx, clientMock.TENANT_ID)
	assert.Equal(t, "ctxRealm", realmName)
}

func TestGetRealm_RealmFromTenantId(t *testing.T) {
	_, dalMock, testSubject := initClientServiceTest(t)
	dalMock.On("GetRealmByTenantId", mock.Anything, clientMock.TENANT_ID).Return(&models.Realm{TenantId: clientMock.TENANT_ID, Realm: clientMock.REALM_NAME}, nil)
	realmName := testSubject.getRealmName(context.Background(), clientMock.TENANT_ID)
	assert.Equal(t, clientMock.REALM_NAME, realmName)
}

func TestGetRealm_EmptyRealm(t *testing.T) {
	_, dalMock, testSubject := initClientServiceTest(t)
	dalMock.On("GetRealmByTenantId", mock.Anything, clientMock.TENANT_ID).Return(nil, nil)
	realmName := testSubject.getRealmName(context.Background(), clientMock.TENANT_ID)
	assert.Equal(t, "", realmName)
}

func TestGetClientFilters_RealmFromCtx(t *testing.T) {
	_, _, testSubject := initClientServiceTest(t)
	ctx := context.Background()
	ctx = context.WithValue(ctx, utils.AuthRealmKey, "ctxRealm")
	realmName := testSubject.getRealmName(ctx, clientMock.TENANT_ID)
	assert.Equal(t, "ctxRealm", realmName)
}

func TestGetFilters(t *testing.T) {
	service := ClientsApiService{}

	tests := []struct {
		input    string
		expected *iamModels.ClientFilter
	}{
		{
			input: "runai.client_type:workloads_auth runai.cluster_id:111",
			expected: &iamModels.ClientFilter{
				ClientType: &[]string{"workloads_auth"}[0],
				ClusterId:  &[]string{"111"}[0],
			},
		},
		{
			input: "runai.is_external:true",
			expected: &iamModels.ClientFilter{
				IsExternal: ptr.To(true),
			},
		},
		{
			input:    "",
			expected: &iamModels.ClientFilter{},
		},
		{
			input:    "invalid_key:value",
			expected: &iamModels.ClientFilter{},
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := service.getFilters(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func getMockClients() []iamModels.Client {
	return []iamModels.Client{
		{
			ID:       ptr.To(clientMock.CLIENT_ID + "1"),
			Name:     ptr.To("name1"),
			ClientID: ptr.To("clientIdName1"),
		},
		{
			ID:       ptr.To(clientMock.CLIENT_ID + "2"),
			Name:     ptr.To("name2"),
			ClientID: ptr.To("clientIdName2"),
		},
	}
}
