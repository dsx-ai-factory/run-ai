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

	metricerrors "github.com/run-ai/runai/control-plane/metrics-service/internal/errors"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/services"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/services/mocks"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const tenantId, overrideTenantId = int32(100123), int32(123)

func TestGetMetricStores_MissingTenantId(t *testing.T) {
	mss := services.NewMetricStoresAPIService(nil)
	res, err := mss.GetMetricStores(context.Background(), -1)
	require.Error(t, err)
	require.EqualError(t, err, "missing tenant id")
	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestGetMetricStores_BadTenantId(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "user@run.ai")
	mss := services.NewMetricStoresAPIService(nil)
	res, err := mss.GetMetricStores(ctx, -1)
	require.Error(t, err)
	require.EqualError(t, err, "missing tenant id")
	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestGetMetricStores_FailedToGetMetricStore(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
	msd := mocks.NewMetricStoreManager(t)
	mss := services.NewMetricStoresAPIService(msd)

	msd.On("GetByTenantId", ctx, tenantId).Return(nil, fmt.Errorf("failed"))
	res, err := mss.GetMetricStores(ctx, -1)
	require.Error(t, err)
	require.EqualError(t, err, "{Code:500 Message:Failed to fetch metric store Details:}")
	require.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestGetMetricStores_StoreDoesntExist(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
	msd := mocks.NewMetricStoreManager(t)
	mss := services.NewMetricStoresAPIService(msd)

	msd.On("GetByTenantId", ctx, overrideTenantId).Return(nil, nil)
	res, err := mss.GetMetricStores(ctx, overrideTenantId)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.Len(t, res.Body.([]api.MetricStore), 0)
}

func TestGetMetricStores(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
	msd := mocks.NewMetricStoreManager(t)
	mss := services.NewMetricStoresAPIService(msd)

	msd.On("GetByTenantId", ctx, overrideTenantId).Return(&api.MetricStore{TenantId: overrideTenantId, DatasourceId: "some_id"}, nil)
	res, err := mss.GetMetricStores(ctx, overrideTenantId)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, overrideTenantId, res.Body.([]api.MetricStore)[0].TenantId)
}

func TestCreateMetricStore(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
	ms := api.MetricStore{TenantId: overrideTenantId, DatasourceId: "some_id"}
	msd := mocks.NewMetricStoreManager(t)
	mss := services.NewMetricStoresAPIService(msd)

	msd.On("Create", ctx, mock.Anything).Return(&ms, nil)
	res, err := mss.CreateMetricStore(ctx, api.MetricStore{TenantId: overrideTenantId, DatasourceId: "some_id"})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, overrideTenantId, res.Body.(*api.MetricStore).TenantId)
}

func TestCreateMetricStore_Failed(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
	msd := mocks.NewMetricStoreManager(t)
	mss := services.NewMetricStoresAPIService(msd)

	msd.On("Create", ctx, mock.Anything).Return(nil, fmt.Errorf("failed to create or something"))
	res, err := mss.CreateMetricStore(ctx, api.MetricStore{DatasourceId: "some_id"})
	require.Error(t, err)
	require.EqualError(t, err, "{Code:500 Message:Failed to create metric store Details:}")
	require.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestCreateMetricStore_AlreadyExists(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
	managerMock := mocks.NewMetricStoreManager(t)
	apiService := services.NewMetricStoresAPIService(managerMock)

	managerMock.On("Create", ctx, mock.Anything).Return(nil, metricerrors.ErrMetricStoreAlreadyExists)
	res, err := apiService.CreateMetricStore(ctx, api.MetricStore{TenantId: overrideTenantId, DatasourceId: "some_id"})
	require.EqualError(t, err, "metric store already exists")
	require.Equal(t, http.StatusConflict, res.Code)
}

func TestDeleteMetricStores(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
	msd := mocks.NewMetricStoreManager(t)
	mss := services.NewMetricStoresAPIService(msd)

	msd.On("DeleteByTenantId", ctx, tenantId).Return(nil)
	res, err := mss.DeleteMetricStores(ctx, tenantId)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, res.Code)
}

func TestDeleteMetricStores_Failed(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
	msd := mocks.NewMetricStoreManager(t)
	mss := services.NewMetricStoresAPIService(msd)

	msd.On("DeleteByTenantId", ctx, tenantId).Return(fmt.Errorf("failed"))
	res, err := mss.DeleteMetricStores(ctx, tenantId)
	require.Error(t, err)
	require.EqualError(t, err, "failed to delete metric stores of tenant 100123")
	require.Equal(t, http.StatusInternalServerError, res.Code)
}
