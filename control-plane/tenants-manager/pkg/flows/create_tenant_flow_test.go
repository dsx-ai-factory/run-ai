/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package flows

import (
	"context"
	"net/http"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/business/tenant/mocks"
	tm_config "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/domain/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

const (
	tenantName = "tenantName"
)

func initCreateTenantFlowTest(t *testing.T) (tenantProviderMock *mocks.TenantProvider, testSubject *CreateTenantFlow) {
	tenantProviderMock = mocks.NewTenantProvider(t)
	testSubject = &CreateTenantFlow{
		tenantProvider: tenantProviderMock,
		tenantConf:     &tm_config.Tenant{Name: tenantName},
	}
	return
}

func TestCheckTenantProvisionState_TenantCreating_Waits(t *testing.T) {
	tenantProviderMock, _ := initCreateTenantFlowTest(t)
	tenantResp := getTenantResponse(tenantName, models.TenantCreating)
	tenantProviderMock.On("GetTenantByName", context.Background(), tenantName).Return(tenantResp, nil)
	action, existing := checkTenantProvisionState(context.Background(), tenantProviderMock, tenantName)
	assert.Equal(t, actionWait, action)
	assert.Equal(t, tenantResp, existing)
}

func TestCheckTenantProvisionState_TenantReady_Skips(t *testing.T) {
	tenantProviderMock, _ := initCreateTenantFlowTest(t)
	tenantResp := getTenantResponse(tenantName, models.TenantReady)
	tenantProviderMock.On("GetTenantByName", context.Background(), tenantName).Return(tenantResp, nil)
	action, existing := checkTenantProvisionState(context.Background(), tenantProviderMock, tenantName)
	assert.Equal(t, actionSkip, action)
	assert.Equal(t, tenantResp, existing)
}

func TestCheckTenantProvisionState_TenantNotFound_Creates(t *testing.T) {
	tenantProviderMock, _ := initCreateTenantFlowTest(t)
	tenantProviderMock.On("GetTenantByName", context.Background(), tenantName).Return(nil, &api.Error{
		Code:    http.StatusNotFound,
		Message: "Tenant not found",
	})
	action, existing := checkTenantProvisionState(context.Background(), tenantProviderMock, tenantName)
	assert.Equal(t, actionCreate, action)
	assert.Nil(t, existing)
}

func TestCheckTenantProvisionState_TenantError_Skips(t *testing.T) {
	tenantProviderMock, _ := initCreateTenantFlowTest(t)
	tenantProviderMock.On("GetTenantByName", context.Background(), tenantName).Return(nil, &api.Error{
		Code:    http.StatusInternalServerError,
		Message: "Internal server error",
	})
	action, existing := checkTenantProvisionState(context.Background(), tenantProviderMock, tenantName)
	assert.Equal(t, actionSkip, action)
	assert.Nil(t, existing)
}

func TestCheckTenantProvisionState_TenantFailed_RequeuesAndWaits(t *testing.T) {
	tenantProviderMock, _ := initCreateTenantFlowTest(t)
	tenantResp := getTenantResponse(tenantName, models.TenantFailed)
	tenantProviderMock.On("GetTenantByName", context.Background(), tenantName).Return(tenantResp, nil)
	action, existing := checkTenantProvisionState(context.Background(), tenantProviderMock, tenantName)
	assert.Equal(t, actionRequeueAndWait, action)
	assert.Equal(t, tenantResp, existing)
}

func getTenantResponse(tenantName string, status models.TenantStatus) *models.Tenant {
	return &models.Tenant{
		Model: gorm.Model{
			ID: uint(1),
		},
		Name:   tenantName,
		Status: status,
	}
}
