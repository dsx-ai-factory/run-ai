/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/errors"
	"github.com/stretchr/testify/mock"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/services"
	mocks "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/services/mocks"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/assert"
)

func TestGetTenantIdFromCtx_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "123")
	tenantId, err := services.GetTenantIdFromCtx(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int32(123), tenantId)
}

func TestGetTenantIdFromCtx_NotInCtx(t *testing.T) {
	ctx := context.Background()
	_, err := services.GetTenantIdFromCtx(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get tenant id from context")
}

func TestGetTenantIdFromCtx_FailedToParse(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "abc")
	_, err := services.GetTenantIdFromCtx(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse tenant id abc")
}

func TestGetTenantId_Override(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "234")
	tenantId, err := services.GetTenantId(ctx, 123)
	assert.NoError(t, err)
	assert.Equal(t, int32(123), tenantId)
}

func TestGetTenantId_NoOverride(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "123")
	tenantId, err := services.GetTenantId(ctx, 0)
	assert.NoError(t, err)
	assert.Equal(t, int32(123), tenantId)
}

func TestGetMetrics_NoTenantId(t *testing.T) {
	res, err := services.NewMetricsApiService(nil, nil, nil).GetMetrics(context.Background(), api.MetricInformation{}, 0)
	assert.Error(t, err)
	assert.Equal(t, res.Code, http.StatusBadRequest)
}

func TestGetMetrics_TenantNotExist(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "123")
	cg := mocks.NewTenantsClienter(t)
	cg.On("TenantExist", ctx, int32(123)).Return(false, nil)
	res, err := services.NewMetricsApiService(nil, nil, cg).GetMetrics(ctx, api.MetricInformation{}, 0)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.Code)
}

func TestGetMetrics_ClusterVersionNotSupported(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "123")
	cg := mocks.NewTenantsClienter(t)
	mqb := mocks.NewMetricQueryBuilder(t)
	cg.On("TenantExist", ctx, int32(123)).Return(true, nil)
	mqb.On("Build", ctx, mock.Anything).Return("", fmt.Errorf("%w %s", errors.ErrClusterVersionNotSupported, "TestGetMetrics_ClusterVersionNotSupported"))
	res, err := services.NewMetricsApiService(mqb, nil, cg).GetMetrics(ctx, api.MetricInformation{}, 0)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.True(t, errors.IsClusterVersionNotSupportedErr(err))
}

func TestGetMetrics_DataSourceHostNotFound(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "123")
	cg := mocks.NewTenantsClienter(t)
	mc := mocks.NewMetricQuerier(t)
	mqb := mocks.NewMetricQueryBuilder(t)

	cg.On("TenantExist", ctx, int32(123)).Return(true, nil)
	mqb.On("Build", ctx, mock.Anything).Return("TestGetMetrics_DataSourceHostNotFound_query", nil)
	mc.On("Query", ctx, int32(123), "TestGetMetrics_DataSourceHostNotFound_query", mock.Anything).Return(nil, fmt.Errorf(errors.NoSuchHost))

	res, err := services.NewMetricsApiService(mqb, mc, cg).GetMetrics(ctx, api.MetricInformation{}, 0)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), `failed to get metric data for tenant 123 query TestGetMetrics_DataSourceHostNotFound_query err: `+errors.NoSuchHost)
	assert.Equal(t, http.StatusMisdirectedRequest, res.Code)
}

func TestGetMetrics_FailedToGetDataSource(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "123")
	cg := mocks.NewTenantsClienter(t)
	mc := mocks.NewMetricQuerier(t)
	mqb := mocks.NewMetricQueryBuilder(t)

	cg.On("TenantExist", ctx, int32(123)).Return(true, nil)
	mqb.On("Build", ctx, mock.Anything).Return("TestGetMetrics_DataSourceHostNotFound_query", nil)
	mc.On("Query", ctx, int32(123), "TestGetMetrics_DataSourceHostNotFound_query", mock.Anything).Return(nil, fmt.Errorf(errors.MetricStoreNotFound))

	res, err := services.NewMetricsApiService(mqb, mc, cg).GetMetrics(ctx, api.MetricInformation{}, 0)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), `failed to get metric data for tenant 123 query TestGetMetrics_DataSourceHostNotFound_query err: `+errors.MetricStoreNotFound)
	assert.Equal(t, http.StatusMisdirectedRequest, res.Code)
}
