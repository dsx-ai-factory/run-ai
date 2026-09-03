/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package app_reconciler

import (
	"context"
	"errors"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/models"

	iamMock "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/mocks"
	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	dal_mocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/business/mocks"
	reconciler_mocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/reconciler/mocks"
	test_utils "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils/mocks"

	"k8s.io/utils/ptr"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func initExternalAppReconcilerTest(t *testing.T) (iamServicesContainerMock *iamMock.MockServicesContainer, tenantResolverMock *reconciler_mocks.TenantResolver, reconcilerVersionsDalMock *dal_mocks.ReconcilerVersionsDal, testSubject *AppsReconciler) {
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
		isExternal:     ptr.To(true),
		version:        "v1",
	}
	return mockIamServicesContainer, tenantResolverMock, reconcilerVersionsDalMock, testSubject
}

func TestSameVersionExternal_NoNeedToReconcile(t *testing.T) {
	iamMock, tenantResolverMock, versionsDaoMock, testSubject := initExternalAppReconcilerTest(t)
	realmName := "test-realm"
	versionRes := models.ReconcilerVersions{
		TenantId:        TenantId,
		ServiceAccounts: testSubject.version,
	}
	tenantResolverMock.On("GetRealmTenantId", context.Background(), mock.Anything).Return(TenantId, nil)
	versionsDaoMock.On("GetReconcilerVersionsByTenantId", mock.Anything, TenantId).Return(&versionRes, nil)

	err := testSubject.Reconcile(context.Background(), realmName, false)
	assert.Nil(t, err)
	tenantResolverMock.AssertNumberOfCalls(t, "CreateApp", 0)
	tenantResolverMock.AssertNumberOfCalls(t, "GetRealmTenantId", 1)
	iamMock.MockClientService.AssertNumberOfCalls(t, "GetClients", 0)
}

func TestSameVersionExternal_NeedToReconcile(t *testing.T) {
	iamMock, tenantResolverMock, versionsDaoMock, testSubject := initExternalAppReconcilerTest(t)
	realmName := "test-realm"
	versionRes := models.ReconcilerVersions{
		TenantId:                TenantId,
		InternalServiceAccounts: "testSubject.version",
	}
	tenantResolverMock.On("GetRealmTenantId", context.Background(), mock.Anything).Return(TenantId, nil)

	versionsDaoMock.On("GetReconcilerVersionsByTenantId", mock.Anything, TenantId).Return(&versionRes, nil)
	versionsDaoMock.On("UpdateServiceAccountsReconcilerVersions", mock.Anything, TenantId, testSubject.version).Return(nil)

	externalApps := test_utils.GetServiceAccountsArray()
	defaultIamMockSetup(iamMock, externalApps, realmName, true)
	iamMock.MockClientService.On("GetClients", mock.Anything, mock.Anything, &iamModels.ClientFilter{IsExternal: ptr.To(true)}).Return(externalApps, nil)

	err := testSubject.Reconcile(context.Background(), realmName, false)
	assert.Nil(t, err)
	tenantResolverMock.AssertNumberOfCalls(t, "CreateApp", 0)
	tenantResolverMock.AssertNumberOfCalls(t, "GetRealmTenantId", 1)
	iamMock.MockClientService.AssertNumberOfCalls(t, "GetClients", 1)
	iamMock.MockClientService.AssertNumberOfCalls(t, "UpdateClient", len(externalApps))
}

func TestMissingVersionExternal_NeedToReconcile(t *testing.T) {
	iamMock, tenantResolverMock, versionsDaoMock, testSubject := initExternalAppReconcilerTest(t)
	realmName := "test-realm"
	versionRes := models.ReconcilerVersions{
		TenantId: TenantId,
	}
	tenantResolverMock.On("GetRealmTenantId", context.Background(), mock.Anything).Return(TenantId, nil)

	versionsDaoMock.On("GetReconcilerVersionsByTenantId", mock.Anything, TenantId).Return(&versionRes, nil)
	versionsDaoMock.On("UpdateServiceAccountsReconcilerVersions", mock.Anything, TenantId, testSubject.version).Return(nil)

	externalApps := test_utils.GetServiceAccountsArray()
	defaultIamMockSetup(iamMock, externalApps, realmName, true)
	iamMock.MockClientService.On("GetClients", mock.Anything, mock.Anything, &iamModels.ClientFilter{IsExternal: ptr.To(true)}).Return(externalApps, nil)

	err := testSubject.Reconcile(context.Background(), realmName, false)
	assert.Nil(t, err)
	tenantResolverMock.AssertNumberOfCalls(t, "CreateApp", 0)
	tenantResolverMock.AssertNumberOfCalls(t, "GetRealmTenantId", 1)
	iamMock.MockClientService.AssertNumberOfCalls(t, "GetClients", 1)
	iamMock.MockClientService.AssertNumberOfCalls(t, "UpdateClient", len(externalApps))
}

func TestErrorVersionExternal_NeedToReconcile(t *testing.T) {
	iamMock, tenantResolverMock, versionsDaoMock, testSubject := initExternalAppReconcilerTest(t)
	realmName := "test-realm"

	tenantResolverMock.On("GetRealmTenantId", context.Background(), mock.Anything).Return(TenantId, nil)

	versionsDaoMock.On("GetReconcilerVersionsByTenantId", mock.Anything, TenantId).Return(nil, errors.New("error"))
	versionsDaoMock.On("UpdateServiceAccountsReconcilerVersions", mock.Anything, TenantId, testSubject.version).Return(nil)

	externalApps := test_utils.GetServiceAccountsArray()
	defaultIamMockSetup(iamMock, externalApps, realmName, true)
	iamMock.MockClientService.On("GetClients", mock.Anything, mock.Anything, &iamModels.ClientFilter{IsExternal: ptr.To(true)}).Return(externalApps, nil)

	err := testSubject.Reconcile(context.Background(), realmName, false)
	assert.Nil(t, err)
	tenantResolverMock.AssertNumberOfCalls(t, "CreateApp", 0)
	tenantResolverMock.AssertNumberOfCalls(t, "GetRealmTenantId", 1)
	iamMock.MockClientService.AssertNumberOfCalls(t, "GetClients", 1)
	iamMock.MockClientService.AssertNumberOfCalls(t, "UpdateClient", len(externalApps))
}
