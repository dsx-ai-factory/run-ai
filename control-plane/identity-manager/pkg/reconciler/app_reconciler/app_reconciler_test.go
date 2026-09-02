/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package app_reconciler

import (
	"context"
	"testing"

	dal_mocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/business/mocks"

	iamMock "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/mocks"
	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	reconciler_mocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/reconciler/mocks"
	test_utils "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils/mocks"

	"k8s.io/utils/ptr"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	TenantId = "1001"
	version  = "v1"
)

func initAppReconcilerTest(t *testing.T) (iamServicesContainerMock *iamMock.MockServicesContainer, tenantResolverMock *reconciler_mocks.TenantResolver, reconcilerVersionsDalMock *dal_mocks.ReconcilerVersionsDal, testSubject *AppsReconciler) {
	mockIamServicesContainer := iamMock.NewMockServicesContainer()
	tenantResolverMock = reconciler_mocks.NewTenantResolver(t)
	reconcilerVersionsDalMock = dal_mocks.NewReconcilerVersionsDal(t)

	testSubject = &AppsReconciler{
		iamServices:           *mockIamServicesContainer.ToServicesContainer(),
		reconcilerVersionsDal: reconcilerVersionsDalMock,
		config: AppsReconcilerConfig{
			hostTemplate: "example-host.com",
			appHost:      "example-app-host.com",
		},
		tenantResolver: tenantResolverMock,
		version:        version,
	}
	return mockIamServicesContainer, tenantResolverMock, reconcilerVersionsDalMock, testSubject
}

func defaultIamMockSetup(iamMock *iamMock.MockServicesContainer, clients []iamModels.Client, realmName string, isClientsExist bool) {
	for _, client := range clients {
		var clientRes *iamModels.Client
		if isClientsExist {
			c := client
			clientRes = &c
		}
		iamMock.MockClientService.On("GetClientByClientId", context.Background(), realmName, *client.ClientID).Return(clientRes, nil)
		if isClientsExist {
			iamMock.MockClientService.On("UpdateClient", context.Background(), realmName, mock.Anything).Return(nil)
		} else {
			iamMock.MockClientService.On("CreateClient", context.Background(), realmName, mock.Anything).Return("", nil)
		}
	}
}

func TestNonSSOMappersWithExistsClients(t *testing.T) {
	iamMock, tenantResolverMock, reconcilerVersionsDalMock, testSubject := initAppReconcilerTest(t)
	realmName := "test-realm"

	tenantResolverMock.On("GetRealmTenantId", context.Background(), mock.Anything).Return(TenantId, nil)
	internalApps := test_utils.GetInternalApplicationsArray()

	defaultIamMockSetup(iamMock, internalApps, realmName, true)
	defaultReconcilerVersionMockSetup(reconcilerVersionsDalMock)

	iamMock.MockClientService.On("GetClients", mock.Anything, mock.Anything, &iamModels.ClientFilter{IsExternal: ptr.To(false)}).Return(internalApps, nil)
	externalApps := test_utils.GetServiceAccountsArray()
	defaultIamMockSetup(iamMock, externalApps, realmName, true)
	iamMock.MockClientService.On("GetClients", mock.Anything, mock.Anything, &iamModels.ClientFilter{IsExternal: ptr.To(true)}).Return(externalApps, nil)

	err := testSubject.Reconcile(context.Background(), realmName, true)
	assert.Nil(t, err)
	tenantResolverMock.AssertNumberOfCalls(t, "CreateApp", 0)
	iamMock.MockClientService.AssertNumberOfCalls(t, "UpdateClient", len(internalApps)+len(externalApps))
	tenantResolverMock.AssertNumberOfCalls(t, "GetRealmTenantId", 1)
	iamMock.MockClientService.AssertNumberOfCalls(t, "GetClients", 2)
}

func defaultReconcilerVersionMockSetup(dalMock *dal_mocks.ReconcilerVersionsDal) {
	dalMock.On("UpdateServiceAccountsReconcilerVersions", mock.Anything, TenantId, version).Return(nil)
	dalMock.On("UpdateInternalServiceAccountsReconcilerVersions", mock.Anything, TenantId, version).Return(nil)
}

func TestNonSSOMappersWithoutPreCreatedTenantIdMapper(t *testing.T) {
	iamMock, tenantResolverMock, _, testSubject := initAppReconcilerTest(t)

	tenantResolverMock.On("GetRealmTenantId", context.Background(), mock.Anything).Return("", nil)
	defaultIamMockSetup(iamMock, nil, "test-realm", true)

	err := testSubject.Reconcile(context.Background(), "test-realm", true)
	assert.NotNil(t, err)
}

func TestNonSSOMappersWithoutApps(t *testing.T) {
	iamMock, tenantResolverMock, reconcilerVersionsDalMock, testSubject := initAppReconcilerTest(t)
	realmName := "test-realm"

	defaultReconcilerVersionMockSetup(reconcilerVersionsDalMock)

	tenantResolverMock.On("GetRealmTenantId", context.Background(), mock.Anything).Return(TenantId, nil)
	internalApps := test_utils.GetInternalApplicationsArray()
	defaultIamMockSetup(iamMock, internalApps, realmName, false)
	iamMock.MockClientService.On("GetClients", mock.Anything, mock.Anything, &iamModels.ClientFilter{IsExternal: ptr.To(false)}).Return(internalApps, nil)
	externalApps := test_utils.GetServiceAccountsArray()
	defaultIamMockSetup(iamMock, externalApps, realmName, false)
	iamMock.MockClientService.On("GetClients", mock.Anything, mock.Anything, &iamModels.ClientFilter{IsExternal: ptr.To(true)}).Return(externalApps, nil)

	err := testSubject.Reconcile(context.Background(), realmName, true)
	assert.Nil(t, err)
	iamMock.MockClientService.AssertNumberOfCalls(t, "CreateClient", len(internalApps)+len(externalApps))
	tenantResolverMock.AssertNumberOfCalls(t, "GetRealmTenantId", 1)
	iamMock.MockClientService.AssertNumberOfCalls(t, "GetClients", 2)
}

func TestSSOMappersWithExistsClients(t *testing.T) {
	iamMock, tenantResolverMock, reconcilerVersionsDalMock, testSubject := initAppReconcilerTest(t)
	realmName := "test-realm"

	defaultReconcilerVersionMockSetup(reconcilerVersionsDalMock)

	tenantResolverMock.On("GetRealmTenantId", context.Background(), mock.Anything).Return(TenantId, nil)

	internalApps := test_utils.GetInternalApplicationsArray()
	defaultIamMockSetup(iamMock, internalApps, realmName, true)
	iamMock.MockClientService.On("GetClients", mock.Anything, mock.Anything, &iamModels.ClientFilter{IsExternal: ptr.To(false)}).Return(internalApps, nil)
	externalApps := test_utils.GetServiceAccountsArray()
	defaultIamMockSetup(iamMock, externalApps, realmName, true)
	iamMock.MockClientService.On("GetClients", mock.Anything, mock.Anything, &iamModels.ClientFilter{IsExternal: ptr.To(true)}).Return(externalApps, nil)

	err := testSubject.Reconcile(context.Background(), realmName, true)
	assert.Nil(t, err)

	tenantResolverMock.AssertNumberOfCalls(t, "CreateApp", 0)
	iamMock.MockClientService.AssertNumberOfCalls(t, "UpdateClient", len(internalApps)+len(externalApps))
	tenantResolverMock.AssertNumberOfCalls(t, "GetRealmTenantId", 1)
	iamMock.MockClientService.AssertNumberOfCalls(t, "GetClients", 2)
}
