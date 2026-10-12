/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client_reconciler

import (
	"context"
	"testing"

	iamMock "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/mocks"
	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	reconciler_mocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/reconciler/client_reconciler/mocks"
	tenant_mocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/reconciler/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
	test_utils "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils/mocks"

	"errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const TenantId = "1001"

func initClientReconcilerTest(t *testing.T) (iamServicesContainerMock *iamMock.MockServicesContainer, tenantResolverMock *tenant_mocks.TenantResolver, reconcilerHelperMock *reconciler_mocks.ClientReconcilerHelperInterface, testSubject *ClientsReconciler) {
	mockIamServicesContainer := iamMock.NewMockServicesContainer()
	tenantResolverMock = tenant_mocks.NewTenantResolver(t)
	reconcilerHelperMock = reconciler_mocks.NewClientReconcilerHelperInterface(t)

	testSubject = &ClientsReconciler{
		iamServices: *mockIamServicesContainer.ToServicesContainer(),
		config: ClientsReconcilerConfig{
			hostTemplate: "example-host.com",
			appHost:      "example-app-host.com",
		},
		tenantResolver:         tenantResolverMock,
		clientReconcilerHelper: reconcilerHelperMock,
	}
	return mockIamServicesContainer, tenantResolverMock, reconcilerHelperMock, testSubject
}

func defaultIamMockSetup(iamMock *iamMock.MockServicesContainer, clients []*iamModels.Client, realmName string, isClientsExist bool) {
	iamMock.MockIdpService.On("GetIdps", mock.Anything, mock.Anything).Return([]*iamModels.Idp{}, nil)

	for _, client := range clients {
		var clientRes *iamModels.Client
		if isClientsExist {
			clientRes = client
		}
		iamMock.MockClientService.On("GetClientByClientId", context.Background(), realmName, *client.ClientID).Return(clientRes, nil)
		iamMock.MockClientService.On("UpdateClient", context.Background(), realmName, mock.Anything).Return(nil)
		iamMock.MockClientService.On("CreateClient", context.Background(), realmName, mock.Anything).Return("", nil)
	}

}

func defaultTenantResolverMockSetup(tenantResolverMock *tenant_mocks.TenantResolver, tenantId string) {
	tenantResolverMock.On("GetRealmTenantId", context.Background(), mock.Anything).Return(tenantId, nil)
}

func TestNonSSOMappersWithExistsClients(t *testing.T) {
	iamMock, tenantResolverMock, reconcilerHelperMock, testSubject := initClientReconcilerTest(t)
	clients := test_utils.GetBasicReconcilerClientWithTenantIdMapperArray()
	realmName := "test-realm"
	defaultIamMockSetup(iamMock, clients, realmName, true)
	defaultTenantResolverMockSetup(tenantResolverMock, TenantId)
	reconcilerHelperMock.On("ReconcileClient", context.Background(), mock.Anything, mock.Anything).Return(nil)
	reconcilerHelperMock.On("DeleteClientIfExists", context.Background(), realmName, utils.RUNAI_CLIENT_ID).Return(nil)

	err := testSubject.Reconcile(context.Background(), realmName, true)
	assert.Nil(t, err)
	tenantResolverMock.AssertNumberOfCalls(t, "GetRealmTenantId", 1)
	reconcilerHelperMock.AssertNumberOfCalls(t, "ReconcileClient", len(clients))
	reconcilerHelperMock.AssertNumberOfCalls(t, "DeleteClientIfExists", 1)
}

func TestNonSSOMappersWithoutPreCreatedTenantIdMapper(t *testing.T) {
	iamMock, tenantResolverMock, _, testSubject := initClientReconcilerTest(t)
	defaultTenantResolverMockSetup(tenantResolverMock, "")
	defaultIamMockSetup(iamMock, nil, "test-realm", true)
	err := testSubject.Reconcile(context.Background(), "test-realm", true)
	assert.NotNil(t, err)
}

func TestSSOMappersWithExistsClients(t *testing.T) {
	iamMock, tenantResolverMock, reconcilerHelperMock, testSubject := initClientReconcilerTest(t)
	clients := test_utils.GetBasicReconcilerClientWithTenantIdMapperArray()
	realmName := "test-realm"
	defaultIamMockSetup(iamMock, clients, realmName, true)
	defaultTenantResolverMockSetup(tenantResolverMock, TenantId)
	reconcilerHelperMock.On("ReconcileClient", context.Background(), mock.Anything, mock.Anything).Return(nil)
	reconcilerHelperMock.On("DeleteClientIfExists", context.Background(), realmName, utils.RUNAI_CLIENT_ID).Return(nil)

	err := testSubject.Reconcile(context.Background(), realmName, true)
	assert.Nil(t, err)

	tenantResolverMock.AssertNumberOfCalls(t, "GetRealmTenantId", 1)
	reconcilerHelperMock.AssertNumberOfCalls(t, "ReconcileClient", len(clients))
	reconcilerHelperMock.AssertNumberOfCalls(t, "DeleteClientIfExists", 1)
}

func TestReconcileDefaultClients_DeleteClientIfExistsError(t *testing.T) {
	iamMock, tenantResolverMock, reconcilerHelperMock, testSubject := initClientReconcilerTest(t)
	clients := test_utils.GetBasicReconcilerClientWithTenantIdMapperArray()
	realmName := "test-realm"
	defaultIamMockSetup(iamMock, clients, realmName, true)
	defaultTenantResolverMockSetup(tenantResolverMock, TenantId)
	reconcilerHelperMock.On("ReconcileClient", context.Background(), mock.Anything, mock.Anything).Return(nil)
	reconcilerHelperMock.On("DeleteClientIfExists", context.Background(), realmName, utils.RUNAI_CLIENT_ID).Return(errors.New("delete failed"))

	err := testSubject.Reconcile(context.Background(), realmName, true)
	assert.ErrorContains(t, err, "delete failed")
}
