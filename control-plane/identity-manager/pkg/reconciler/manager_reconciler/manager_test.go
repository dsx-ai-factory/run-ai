/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manager_reconciler

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	iamMock "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/mocks"
	businessMocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/business/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/models"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client/generated"
	clientMocks "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client/mocks"
)

type Mocks struct {
	iam                  *iamMock.MockServicesContainer
	realmDal             *businessMocks.RealmDal
	tenantsManagerClient *clientMocks.ITenantsManager
}

func initManagerTest(t *testing.T) (Mocks, *IamReconcileManager) {
	mockIamServicesContainer := iamMock.NewMockServicesContainer()
	realmDalMock := businessMocks.NewRealmDal(t)
	tenantsManagerClient := clientMocks.NewITenantsManager(t)

	mock := Mocks{iam: mockIamServicesContainer, realmDal: realmDalMock, tenantsManagerClient: tenantsManagerClient}
	testSubject := &IamReconcileManager{
		iam:                  *mockIamServicesContainer.ToServicesContainer(),
		realmDal:             mock.realmDal,
		masterReconciler:     nil,
		tenantsManagerClient: tenantsManagerClient,
	}
	return mock, testSubject
}

func TestManager_UpdateTenants(t *testing.T) {
	mocks, testSubject := initManagerTest(t)
	a := make([]string, len(os.Args)+2)
	a[0] = os.Args[0]
	a[1] = "reconcile"
	a[2] = "updateTenantRealms"
	copy(a[3:], os.Args[1:])

	os.Args = a
	mocks.realmDal.On("GetRealms", mock.Anything).Return(&[]models.Realm{}, nil)

	res := testSubject.ShouldUpdateTenants(context.Background())
	assert.Equal(t, "updateTenantRealms", os.Args[2])
	assert.True(t, res)
}

func TestManager_UpdateTenantsOnly(t *testing.T) {
	mocks, testSubject := initManagerTest(t)
	a := make([]string, len(os.Args)+2)
	a[0] = os.Args[0]
	a[1] = "updateTenantRealms"
	copy(a[3:], os.Args[1:])

	os.Args = a
	mocks.realmDal.On("GetRealms", mock.Anything).Return(&[]models.Realm{}, nil)

	res := testSubject.ShouldUpdateTenants(context.Background())
	assert.Equal(t, "updateTenantRealms", os.Args[1])
	assert.True(t, res)
}

func TestManager_Reconcile(t *testing.T) {
	_, testSubject := initManagerTest(t)
	a := make([]string, len(os.Args)+2)
	a[0] = os.Args[0]
	a[1] = "reconcile"
	a[2] = "updateTenantRealms"
	copy(a[3:], os.Args[1:])

	os.Args = a

	res := testSubject.ShouldReconcile()
	assert.Equal(t, "reconcile", os.Args[1])
	assert.True(t, res)
}

func TestManager_SyncTenantsRealmError(t *testing.T) {
	mocks, testSubject := initManagerTest(t)
	a := make([]string, len(os.Args)+2)
	a[0] = os.Args[0]
	a[1] = "updateTenantRealms"
	copy(a[3:], os.Args[1:])

	os.Args = a

	mocks.realmDal.On("GetRealms", mock.Anything).Return(nil, errors.New("error"))

	res := testSubject.ShouldUpdateTenants(context.Background())
	assert.Equal(t, "updateTenantRealms", os.Args[1])
	assert.True(t, res)
}

func TestManager_SyncTenantsRealmNoRealms(t *testing.T) {
	mocks, testSubject := initManagerTest(t)
	a := make([]string, len(os.Args)+2)
	a[0] = os.Args[0]
	a[1] = "updateTenantRealms"
	copy(a[3:], os.Args[1:])

	os.Args = a

	mocks.realmDal.On("GetRealms", mock.Anything).Return(nil, nil)

	res := testSubject.ShouldUpdateTenants(context.Background())
	assert.Equal(t, "updateTenantRealms", os.Args[1])
	assert.True(t, res)
}

func TestManager_SyncTenantsRealmEmptyRealms(t *testing.T) {
	mocks, testSubject := initManagerTest(t)
	a := make([]string, len(os.Args)+2)
	a[0] = os.Args[0]
	a[1] = "updateTenantRealms"
	copy(a[3:], os.Args[1:])

	os.Args = a

	mocks.realmDal.On("GetRealms", mock.Anything).Return(&[]models.Realm{}, nil)

	res := testSubject.ShouldUpdateTenants(context.Background())
	assert.Equal(t, "updateTenantRealms", os.Args[1])
	assert.True(t, res)
}

func TestManager_SyncTenantsRealmWithRealms(t *testing.T) {
	mocks, testSubject := initManagerTest(t)
	a := make([]string, len(os.Args)+2)
	a[0] = os.Args[0]
	a[1] = "updateTenantRealms"
	copy(a[3:], os.Args[1:])

	os.Args = a

	mocks.realmDal.On("GetRealms", mock.Anything).Return(&[]models.Realm{{
		TenantId: "1",
		Realm:    "test",
	}}, nil)

	res := testSubject.ShouldUpdateTenants(context.Background())
	assert.Equal(t, "updateTenantRealms", os.Args[1])
	assert.False(t, res)
}

func TestManager_listAllTenants(t *testing.T) {
	mocks, testSubject := initManagerTest(t)

	tenant := openapi.Tenant{
		Id:     123,
		Status: tenantStatusReady,
	}
	mocks.tenantsManagerClient.On("ListAllTenantsInternal", mock.Anything).Return([]openapi.Tenant{tenant}, &http.Response{
		StatusCode: http.StatusOK,
	}, nil)

	res, err := testSubject.listAllTenants(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "123", res[0].TenantId)
}
