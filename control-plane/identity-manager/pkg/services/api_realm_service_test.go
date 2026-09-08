/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"

	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	"github.com/google/uuid"

	iamMock "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/mocks"
	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/generated/pkg/api"
	iamError "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/errors"
	businessMocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/business/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/models"
	reconcilerMocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/reconciler/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
	mockUtils "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils/mocks"
)

func initRealmServiceTests(t *testing.T) (reconcilerMock *reconcilerMocks.MasterReconcilerI, idpService api.IdpsAPIServicer, dalMock *businessMocks.RealmDal, iamServicesContainerMock *iamMock.MockServicesContainer, testSubject *RealmApiService, ctx context.Context) {
	config := setConfig()
	dalMock = businessMocks.NewRealmDal(t)
	reconcilerMock = reconcilerMocks.NewMasterReconcilerI(t)
	mockIamServicesContainer := iamMock.NewMockServicesContainer()

	testSubject = &RealmApiService{
		config:           &config,
		dal:              dalMock,
		iamServices:      *mockIamServicesContainer.ToServicesContainer(),
		masterReconciler: reconcilerMock,
	}
	ctx = context.WithValue(context.Background(), utils.AuthRealmKey, "realm")
	ctx = context.WithValue(ctx, runaictx.CfTenantId, "tenant-id")

	return reconcilerMock, idpService, dalMock, mockIamServicesContainer, testSubject, ctx
}

func setConfig() config.TenantRealm {
	conf := config.TenantRealm{
		DefaultClientId:      "default",
		TenantDomainTemplate: "https://{{.}}.mock-q7z3k.example.com",
		AppHost:              "https://mock-q7z3k.example.com",
	}
	runai_config.Set(&conf)
	return conf
}

func TestCreateRealm_NotOcpHappyFlow(t *testing.T) {
	assert.Nil(t, nil)
	reconcilerMock, _, dalMock, iamServicesMocks, testSubject, ctx := initRealmServiceTests(t)
	client := mockUtils.GetBasicClientArray()[0]

	dalMock.On("CreateRealm", mock.Anything, mock.Anything).Return(nil, nil)
	reconcilerMock.On("Reconcile", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	iamServicesMocks.MockRealmService.On("CreateRealm", mock.Anything, mock.Anything).Return("", nil)
	iamServicesMocks.MockClientService.On("GetClientByClientId", mock.Anything, mock.Anything, mock.Anything).Return(client, nil)

	request := api.RealmCreationRequest{
		RealmName:           "realm",
		TenantId:            "123",
		ExternalRedirectUri: []string{"http://localhost:8080"},
	}
	response, err := testSubject.CreateRealm(ctx, request)
	assert.Nil(t, err)
	reconcilerMock.AssertNumberOfCalls(t, "Reconcile", 1)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, request.RealmName, response.Body.(api.RealmCreationResponse).RealmName)
	assert.Equal(t, fmt.Sprintf("https://%v.mock-q7z3k.example.com", request.RealmName), response.Body.(api.RealmCreationResponse).Domain)
}

func TestCreateRealm_StatusConflictFlow(t *testing.T) {
	assert.Nil(t, nil)
	_, _, dalMock, iamServicesMocks, testSubject, ctx := initRealmServiceTests(t)

	dalMock.On("CreateRealm", mock.Anything, mock.Anything).Return(nil, nil)
	iamServicesMocks.MockRealmService.On("CreateRealm", mock.Anything, mock.Anything).Return("", iamError.ErrConflict)

	request := api.RealmCreationRequest{
		RealmName:           "realm",
		TenantId:            "123",
		ExternalRedirectUri: []string{"http://localhost:8080"},
	}
	response, err := testSubject.CreateRealm(ctx, request)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusConflict, response.Code)
}

func TestCreateRealm_OcpHappyFlow(t *testing.T) {
	assert.Nil(t, nil)
	reconcilerMock, _, dalMock, iamServicesMocks, testSubject, ctx := initRealmServiceTests(t)
	client := mockUtils.GetBasicClientArray()[0]
	wg := sync.WaitGroup{}
	wg.Add(1)

	dalMock.On("CreateRealm", mock.Anything, mock.Anything).Return(nil, nil)
	reconcilerMock.On("Reconcile", mock.Anything, mock.Anything, mock.Anything).Run(func(arg mock.Arguments) {
		defer wg.Done()
	}).Return(nil)
	iamServicesMocks.MockRealmService.On("CreateRealm", mock.Anything, mock.Anything).Return("", nil)
	iamServicesMocks.MockClientService.On("GetClientByClientId", mock.Anything, mock.Anything, mock.Anything).Return(client, nil)

	request := api.RealmCreationRequest{
		RealmName:           "realm",
		TenantId:            "123",
		ExternalRedirectUri: []string{"http://localhost:8080"},
	}
	response, err := testSubject.CreateRealm(ctx, request)
	wg.Wait()

	reconcilerMock.AssertNumberOfCalls(t, "Reconcile", 1)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, request.RealmName, response.Body.(api.RealmCreationResponse).RealmName)
	assert.Equal(t, fmt.Sprintf("https://%v.mock-q7z3k.example.com", request.RealmName), response.Body.(api.RealmCreationResponse).Domain)
}

func TestCreateRealm_OcpNotNotFoundFlow(t *testing.T) {
	assert.Nil(t, nil)
	reconcilerMock, _, dalMock, iamServicesMocks, testSubject, ctx := initRealmServiceTests(t)
	client := mockUtils.GetBasicClientArray()[0]

	dalMock.On("CreateRealm", mock.Anything, mock.Anything).Return(nil, nil)
	reconcilerMock.On("Reconcile", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	iamServicesMocks.MockRealmService.On("CreateRealm", mock.Anything, mock.Anything).Return("", nil)
	iamServicesMocks.MockClientService.On("GetClientByClientId", mock.Anything, mock.Anything, mock.Anything).Return(client, nil)

	request := api.RealmCreationRequest{
		RealmName:           "realm",
		TenantId:            "123",
		ExternalRedirectUri: []string{"http://localhost:8080"},
	}
	response, err := testSubject.CreateRealm(ctx, request)
	reconcilerMock.AssertNumberOfCalls(t, "Reconcile", 1)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, request.RealmName, response.Body.(api.RealmCreationResponse).RealmName)
	assert.Equal(t, fmt.Sprintf("https://%v.mock-q7z3k.example.com", request.RealmName), response.Body.(api.RealmCreationResponse).Domain)
}

func TestCreateRealm_OcpUpgradeFlow(t *testing.T) {
	assert.Nil(t, nil)
	reconcilerMock, _, dalMock, iamServicesMocks, testSubject, ctx := initRealmServiceTests(t)
	client := mockUtils.GetBasicClientArray()[0]

	dalMock.On("CreateRealm", mock.Anything, mock.Anything).Return(nil, nil)
	reconcilerMock.On("Reconcile", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	iamServicesMocks.MockRealmService.On("CreateRealm", mock.Anything, mock.Anything).Return("", nil)
	iamServicesMocks.MockClientService.On("GetClientByClientId", mock.Anything, mock.Anything, mock.Anything).Return(client, nil)

	request := api.RealmCreationRequest{
		RealmName:           "realm",
		TenantId:            "123",
		ExternalRedirectUri: []string{"http://localhost:8080"},
	}
	response, err := testSubject.CreateRealm(ctx, request)
	reconcilerMock.AssertNumberOfCalls(t, "Reconcile", 1)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, request.RealmName, response.Body.(api.RealmCreationResponse).RealmName)
	assert.Equal(t, fmt.Sprintf("https://%v.mock-q7z3k.example.com", request.RealmName), response.Body.(api.RealmCreationResponse).Domain)
}

func TestCreateRealm_StatusConflictDB(t *testing.T) {
	assert.Nil(t, nil)
	_, _, dalMock, _, testSubject, ctx := initRealmServiceTests(t)

	dalMock.On("CreateRealm", mock.Anything, mock.Anything).Return(nil, iamError.ErrConflict)

	request := api.RealmCreationRequest{
		RealmName:           "realm",
		TenantId:            "123",
		ExternalRedirectUri: []string{"http://localhost:8080"},
	}
	response, err := testSubject.CreateRealm(ctx, request)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusConflict, response.Code)
}

func TestDisableRealm_MissingInDb(t *testing.T) {
	assert.Nil(t, nil)
	_, _, dalMock, iamServicesMocks, testSubject, ctx := initRealmServiceTests(t)

	dalMock.On("DisableRealm", mock.Anything, mock.Anything).Return(nil, iamError.ErrNotFound)
	iamServicesMocks.MockRealmService.On("DisableRealm", mock.Anything, mock.Anything).Return(nil)

	response, err := testSubject.DeleteRealm(ctx, "test", false)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestDisableRealm_FailedDeleteFromDb(t *testing.T) {
	assert.Nil(t, nil)
	_, _, dalMock, iamServicesMocks, testSubject, ctx := initRealmServiceTests(t)

	dalMock.On("DisableRealm", mock.Anything, mock.Anything).Return(nil, iamError.ErrInternal)
	iamServicesMocks.MockRealmService.On("DisableRealm", mock.Anything, mock.Anything).Return(iamError.ErrInternal)

	response, err := testSubject.DeleteRealm(ctx, "test", false)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestDisableRealm_FailedGetFromKC(t *testing.T) {
	assert.Nil(t, nil)
	_, _, dalMock, iamServicesMocks, testSubject, ctx := initRealmServiceTests(t)
	uuidDal, _ := uuid.NewUUID()
	dalMock.On("DisableRealm", mock.Anything, mock.Anything).Return(&uuidDal, nil)
	iamServicesMocks.MockRealmService.On("DisableRealm", mock.Anything, mock.Anything).Return(iamError.ErrInternal)

	response, err := testSubject.DeleteRealm(ctx, "test", false)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestDisableRealm_Success(t *testing.T) {
	assert.Nil(t, nil)
	_, _, dalMock, iamServicesMocks, testSubject, ctx := initRealmServiceTests(t)
	uuidDal, _ := uuid.NewUUID()
	dalMock.On("DisableRealm", mock.Anything, mock.Anything).Return(&uuidDal, nil)
	iamServicesMocks.MockRealmService.On("DisableRealm", mock.Anything, mock.Anything).Return(nil)

	response, err := testSubject.DeleteRealm(ctx, "test", false)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, response.Body, nil)
}

func TestGetOrCreateRealm_Created(t *testing.T) {
	reconcilerMock, _, dalMock, iamServicesMocks, testSubject, ctx := initRealmServiceTests(t)
	client := mockUtils.GetBasicClientArray()[0]

	dalMock.On("CreateRealm", mock.Anything, mock.Anything).Return(nil, nil)
	reconcilerMock.On("Reconcile", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	iamServicesMocks.MockRealmService.On("CreateRealm", mock.Anything, mock.Anything).Return("", nil)
	iamServicesMocks.MockClientService.On("GetClientByClientId", mock.Anything, mock.Anything, mock.Anything).Return(client, nil)

	request := api.RealmCreationRequest{
		RealmName: "realm",
		TenantId:  "123",
	}
	response, err := testSubject.GetOrCreateRealm(ctx, request)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, request.RealmName, response.Body.(api.RealmCreationResponse).RealmName)
	assert.Equal(t, fmt.Sprintf("https://%v.mock-q7z3k.example.com", request.RealmName), response.Body.(api.RealmCreationResponse).Domain)
}

func TestGetOrCreateRealm_Exists(t *testing.T) {
	reconcilerMock, _, dalMock, iamServicesMocks, testSubject, ctx := initRealmServiceTests(t)
	client := mockUtils.GetBasicClientArray()[0]
	existingRealm := &models.Realm{TenantId: "123", Realm: "realm"}

	dalMock.On("CreateRealm", mock.Anything, mock.Anything).Return(nil, iamError.ErrConflict)
	dalMock.On("GetRealmByName", mock.Anything, mock.Anything).Return(existingRealm, nil)
	reconcilerMock.On("Reconcile", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	iamServicesMocks.MockClientService.On("GetClientByClientId", mock.Anything, mock.Anything, mock.Anything).Return(client, nil)

	request := api.RealmCreationRequest{
		RealmName: "realm",
		TenantId:  "123",
	}
	response, err := testSubject.GetOrCreateRealm(ctx, request)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, request.RealmName, response.Body.(api.RealmCreationResponse).RealmName)
	assert.Equal(t, fmt.Sprintf("https://%v.mock-q7z3k.example.com", request.RealmName), response.Body.(api.RealmCreationResponse).Domain)
}

func TestGetOrCreateRealm_TenantIdMismatch(t *testing.T) {
	_, _, dalMock, _, testSubject, ctx := initRealmServiceTests(t)
	existingRealm := &models.Realm{TenantId: "other-tenant", Realm: "realm"}

	dalMock.On("CreateRealm", mock.Anything, mock.Anything).Return(nil, iamError.ErrConflict)
	dalMock.On("GetRealmByName", mock.Anything, mock.Anything).Return(existingRealm, nil)

	request := api.RealmCreationRequest{
		RealmName: "realm",
		TenantId:  "123",
	}
	response, err := testSubject.GetOrCreateRealm(ctx, request)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusConflict, response.Code)
}

func TestGetOrCreateRealm_GetRealmError(t *testing.T) {
	_, _, dalMock, _, testSubject, ctx := initRealmServiceTests(t)

	dalMock.On("CreateRealm", mock.Anything, mock.Anything).Return(nil, iamError.ErrConflict)
	dalMock.On("GetRealmByName", mock.Anything, mock.Anything).Return(nil, iamError.ErrInternal)

	request := api.RealmCreationRequest{
		RealmName: "realm",
		TenantId:  "123",
	}
	response, err := testSubject.GetOrCreateRealm(ctx, request)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusConflict, response.Code)
}

func TestGetOrCreateRealm_ReconcileError(t *testing.T) {
	reconcilerMock, _, dalMock, _, testSubject, ctx := initRealmServiceTests(t)
	existingRealm := &models.Realm{TenantId: "123", Realm: "realm"}

	dalMock.On("CreateRealm", mock.Anything, mock.Anything).Return(nil, iamError.ErrConflict)
	dalMock.On("GetRealmByName", mock.Anything, mock.Anything).Return(existingRealm, nil)
	reconcilerMock.On("Reconcile", mock.Anything, mock.Anything, mock.Anything).Return(iamError.ErrInternal)

	request := api.RealmCreationRequest{
		RealmName: "realm",
		TenantId:  "123",
	}
	response, err := testSubject.GetOrCreateRealm(ctx, request)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, response.Code)
}

func TestDeleteRealm_Success(t *testing.T) {
	assert.Nil(t, nil)
	_, _, dalMock, iamServicesMocks, testSubject, ctx := initRealmServiceTests(t)
	uuidDal, _ := uuid.NewUUID()
	dalMock.On("DeleteRealm", mock.Anything, mock.Anything).Return(&uuidDal, nil)
	iamServicesMocks.MockRealmService.On("DeleteRealm", mock.Anything, mock.Anything).Return(nil)

	response, err := testSubject.DeleteRealm(ctx, "test", true)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, response.Body, nil)
}

func TestGetClientsResponse_MissingRequiredFields(t *testing.T) {
	_, _, _, iamServicesMocks, testSubject, ctx := initRealmServiceTests(t)

	iamServicesMocks.MockClientService.On(
		"GetClientByClientId", mock.Anything, "realm", testSubject.config.DefaultClientId,
	).Return(&iamModels.Client{ClientID: ptr.To("default")}, nil)

	_, err := testSubject.getClientsResponse(ctx, "realm")
	assert.ErrorContains(t, err, "client default in realm realm is missing required fields")
}
