/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tenant_lifecycle

import (
	"context"
	"errors"
	"testing"

	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	builder_mocks "github.com/run-ai/runai/control-plane/tenants-manager/pkg/builders/mocks"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/database/dao"
	dao_mocks "github.com/run-ai/runai/control-plane/tenants-manager/pkg/database/dao/mocks"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/domain/models"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
	runaipubsub_mocks "github.com/run-ai/runai/runai-common-packages/runaipubsub/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newSubscriber(t *testing.T) *TenantLifecycleSubscriber {
	return NewTenantLifecycleSubscriber(
		dao_mocks.NewTenant(t),
		dao_mocks.NewProvisioningRun(t),
		builder_mocks.NewInterface(t),
	)
}

func sampleProvision() *TenantProvision {
	return &TenantProvision{
		TenantId:   42,
		TenantName: "test-tenant",
		MimirId:    "mimir-42",
		Email:      "admin@test.ai",
		Role:       authorization_api.ALLOWEDINTERNALCREATEROLES_SYSTEM_ADMINISTRATOR,
	}
}

func failedTenant() *models.Tenant {
	return &models.Tenant{Model: gorm.Model{ID: 42}, Name: "test-tenant", Status: models.TenantFailed}
}

func creatingTenant() *models.Tenant {
	return &models.Tenant{Model: gorm.Model{ID: 42}, Name: "test-tenant", Status: models.TenantCreating}
}

func sampleRun() *models.ProvisioningRun {
	return &models.ProvisioningRun{ID: 1, TenantId: 42, Status: models.ProvisioningRunRunning}
}

// ---------------------------------------------------------------------------
// Reset (Failed → Creating) path
// ---------------------------------------------------------------------------

func TestHandleTenantProvision_FailedTenant_ResetsBeforeConstruct(t *testing.T) {
	tenantDao := dao_mocks.NewTenant(t)
	runDao := dao_mocks.NewProvisioningRun(t)
	builder := builder_mocks.NewInterface(t)
	s := NewTenantLifecycleSubscriber(tenantDao, runDao, builder)

	tenantDao.On("GetById", mock.Anything, uint(42)).Return(failedTenant(), nil)

	constructCalled := false
	tenantDao.On("UpdateStatus", mock.Anything, uint(42), models.TenantCreating).
		Run(func(args mock.Arguments) {
			assert.False(t, constructCalled, "UpdateStatus must be called before Construct")
		}).
		Return(creatingTenant(), nil)
	runDao.On("Create", mock.Anything, uint(42)).Return(sampleRun(), nil)
	builder.On("Construct", mock.Anything, mock.Anything, mock.Anything).
		Run(func(_ mock.Arguments) { constructCalled = true }).
		Return(nil)
	runDao.On("Complete", mock.Anything, uint(1), models.ProvisioningRunSuccess, "").Return(nil)
	tenantDao.On("UpdateStatus", mock.Anything, uint(42), models.TenantReady).Return(creatingTenant(), nil)

	err := s.handleTenantProvision(context.Background(), sampleProvision())
	assert.NoError(t, err)
	assert.True(t, constructCalled)
}

func TestHandleTenantProvision_FailedTenant_ResetFails_NacksWithoutConstruct(t *testing.T) {
	tenantDao := dao_mocks.NewTenant(t)
	runDao := dao_mocks.NewProvisioningRun(t)
	builder := builder_mocks.NewInterface(t)
	s := NewTenantLifecycleSubscriber(tenantDao, runDao, builder)

	tenantDao.On("GetById", mock.Anything, uint(42)).Return(failedTenant(), nil)
	tenantDao.On("UpdateStatus", mock.Anything, uint(42), models.TenantCreating).
		Return(nil, &dao.DaoError{Reason: dao.General, Err: errors.New("db error")})

	err := s.handleTenantProvision(context.Background(), sampleProvision())
	assert.Error(t, err)
	builder.AssertNotCalled(t, "Construct")
}

// ---------------------------------------------------------------------------
// Builder failure path
// ---------------------------------------------------------------------------

func TestHandleTenantProvision_BuilderError_PersistsFailedWithDiagnostics(t *testing.T) {
	tenantDao := dao_mocks.NewTenant(t)
	runDao := dao_mocks.NewProvisioningRun(t)
	builder := builder_mocks.NewInterface(t)
	s := NewTenantLifecycleSubscriber(tenantDao, runDao, builder)

	tenantDao.On("GetById", mock.Anything, uint(42)).Return(creatingTenant(), nil)
	runDao.On("Create", mock.Anything, uint(42)).Return(sampleRun(), nil)
	builder.On("Construct", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("auth realm failed"))
	runDao.On("Complete", mock.Anything, uint(1), models.ProvisioningRunFailed, "auth realm failed").Return(nil)
	tenantDao.On("UpdateStatus", mock.Anything, uint(42), models.TenantFailed).Return(creatingTenant(), nil)

	err := s.handleTenantProvision(context.Background(), sampleProvision())
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Happy path / skip path
// ---------------------------------------------------------------------------

func TestHandleTenantProvision_AlreadyReady_Skipped(t *testing.T) {
	tenantDao := dao_mocks.NewTenant(t)
	runDao := dao_mocks.NewProvisioningRun(t)
	builder := builder_mocks.NewInterface(t)
	s := NewTenantLifecycleSubscriber(tenantDao, runDao, builder)

	readyTenant := &models.Tenant{Model: gorm.Model{ID: 42}, Name: "test-tenant", Status: models.TenantReady}
	tenantDao.On("GetById", mock.Anything, uint(42)).Return(readyTenant, nil)

	err := s.handleTenantProvision(context.Background(), sampleProvision())
	assert.NoError(t, err)
	builder.AssertNotCalled(t, "Construct")
}

func TestRegister_SubscribesBothTopics(t *testing.T) {
	s := newSubscriber(t)

	pubsuberMock := runaipubsub_mocks.NewPubSuber(t)
	pubsuberMock.On("Subscribe", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	runaipubsub.SetMock(pubsuberMock)
	defer runaipubsub.SetMock(nil)

	require.NoError(t, s.Register())
	assert.Equal(t, 2, len(pubsuberMock.Calls), "expected exactly TenantProvision + TenantDestruct subscriptions")
}

// ---------------------------------------------------------------------------
// Provisioning run lifecycle
// ---------------------------------------------------------------------------

func TestHandleTenantProvision_HappyPath_RecordsSuccessfulRun(t *testing.T) {
	tenantDao := dao_mocks.NewTenant(t)
	runDao := dao_mocks.NewProvisioningRun(t)
	builder := builder_mocks.NewInterface(t)
	s := NewTenantLifecycleSubscriber(tenantDao, runDao, builder)

	tenant := creatingTenant()
	tenantDao.On("GetById", mock.Anything, uint(42)).Return(tenant, nil)
	runDao.On("Create", mock.Anything, uint(42)).Return(sampleRun(), nil)
	builder.On("Construct", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	runDao.On("Complete", mock.Anything, uint(1), models.ProvisioningRunSuccess, "").Return(nil)
	tenantDao.On("UpdateStatus", mock.Anything, uint(42), models.TenantReady).Return(tenant, nil)

	err := s.handleTenantProvision(context.Background(), sampleProvision())
	assert.NoError(t, err)
	runDao.AssertCalled(t, "Complete", mock.Anything, uint(1), models.ProvisioningRunSuccess, "")
}

func TestHandleTenantProvision_BuilderError_RunCompletedWithErrorMessage(t *testing.T) {
	tenantDao := dao_mocks.NewTenant(t)
	runDao := dao_mocks.NewProvisioningRun(t)
	builder := builder_mocks.NewInterface(t)
	s := NewTenantLifecycleSubscriber(tenantDao, runDao, builder)

	tenant := creatingTenant()
	tenantDao.On("GetById", mock.Anything, uint(42)).Return(tenant, nil)
	runDao.On("Create", mock.Anything, uint(42)).Return(sampleRun(), nil)
	builder.On("Construct", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("metric_store: connection refused"))
	runDao.On("Complete", mock.Anything, uint(1), models.ProvisioningRunFailed, "metric_store: connection refused").Return(nil)
	tenantDao.On("UpdateStatus", mock.Anything, uint(42), models.TenantFailed).Return(tenant, nil)

	err := s.handleTenantProvision(context.Background(), sampleProvision())
	assert.Error(t, err)
	runDao.AssertCalled(t, "Complete", mock.Anything, uint(1), models.ProvisioningRunFailed, "metric_store: connection refused")
}

func TestHandleTenantProvision_RunCreateFails_StillProceeds(t *testing.T) {
	tenantDao := dao_mocks.NewTenant(t)
	runDao := dao_mocks.NewProvisioningRun(t)
	builder := builder_mocks.NewInterface(t)
	s := NewTenantLifecycleSubscriber(tenantDao, runDao, builder)

	tenant := creatingTenant()
	tenantDao.On("GetById", mock.Anything, uint(42)).Return(tenant, nil)
	runDao.On("Create", mock.Anything, uint(42)).Return(nil, &dao.DaoError{Reason: dao.General, Err: errors.New("db error")})
	builder.On("Construct", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	tenantDao.On("UpdateStatus", mock.Anything, uint(42), models.TenantReady).Return(tenant, nil)

	err := s.handleTenantProvision(context.Background(), sampleProvision())
	assert.NoError(t, err)
}
