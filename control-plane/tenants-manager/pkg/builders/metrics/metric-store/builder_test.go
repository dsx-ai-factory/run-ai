/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metric_store

import (
	"context"
	"fmt"
	"testing"

	metric_store "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/business/metric-store"

	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/builders"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/business/errors"
	berrors "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/business/errors"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/business/metric-store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func initTest(t *testing.T) (testSubject *Builder, metricStoreMock *mocks.MetricStoreService) {
	metricStoreMock = mocks.NewMetricStoreService(t)
	testSubject = &Builder{
		metricsInfraService: metricStoreMock,
	}
	return
}

func TestDestruct_HappyFlow(t *testing.T) {
	testSubject, metricStoreMock := initTest(t)
	metricStoreMock.On("DeleteMetricsStoreInfra", mock.Anything, mock.Anything).Return(nil)

	err := testSubject.Destruct(context.Background(), getTenantInfo())
	assert.Nil(t, err)
}

func TestCreate_HappyFlow(t *testing.T) {
	testSubject, metricStoreMock := initTest(t)
	metricStoreMock.On("CreateMetricsStoreInfra", mock.Anything, mock.Anything).Return(nil, nil)
	ctx := context.Background()
	tenantInfo := getTenantInfo()
	err := testSubject.Construct(ctx, tenantInfo, nil)
	request := &metric_store.MetricStoreInfraCreateRequest{
		TenantId:            tenantInfo.Id,
		TenantName:          tenantInfo.Name,
		MimirId:             tenantInfo.MimirId,
		ExistingMetricStore: tenantInfo.MetricStore,
	}
	metricStoreMock.AssertCalled(t, "CreateMetricsStoreInfra", ctx, request)
	assert.Nil(t, err)
}

func TestDestruct_NotFound_Error(t *testing.T) {
	testSubject, metricStoreMock := initTest(t)
	metricStoreMock.On("DeleteMetricsStoreInfra", mock.Anything, mock.Anything).Return(getBuisnessError(errors.NotFound))

	err := testSubject.Destruct(context.Background(), getTenantInfo())
	assert.Nil(t, err)
}

func TestDestruct_TenantNotFound_Error(t *testing.T) {
	testSubject, metricStoreMock := initTest(t)
	metricStoreMock.On("DeleteMetricsStoreInfra", mock.Anything, mock.Anything).Return(getBuisnessError(errors.TenantNotFound))

	err := testSubject.Destruct(context.Background(), getTenantInfo())
	assert.Nil(t, err)
}

func TestDestruct_Error(t *testing.T) {
	testSubject, metricStoreMock := initTest(t)
	metricStoreMock.On("DeleteMetricsStoreInfra", mock.Anything, mock.Anything).Return(getBuisnessError(errors.GeneralError))

	err := testSubject.Destruct(context.Background(), getTenantInfo())
	assert.NotNil(t, err)
}

func getTenantInfo() *builders.TenantInfo {
	return &builders.TenantInfo{
		Name:    "name",
		Id:      1,
		Email:   "email",
		MimirId: "323-222-111",
	}
}

func getBuisnessError(reason berrors.ClientErrorReason) *berrors.BusinessError {
	return &berrors.BusinessError{Reason: reason, Err: fmt.Errorf("error")}
}
