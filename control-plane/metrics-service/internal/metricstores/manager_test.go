/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metricstores_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/run-ai/runai/control-plane/metrics-service/internal/database/models"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/metricstores"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/metricstores/mocks"
	tenantsApi "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client/generated"
	cache_testing "github.com/run-ai/runai/runai-common-packages/cache/v2/testing"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
	pubsubMocks "github.com/run-ai/runai/runai-common-packages/runaipubsub/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func initManager(t *testing.T) (pubsubMock *pubsubMocks.PubSuber, cacheMock *cache_testing.MockCacher, daoMock *mocks.MetricStoresDao, tenantMetricStoreGetterMock *mocks.TenantMetricStoreGetter, providerMock *mocks.DatasourceProvider, manager *metricstores.Manager) {
	pubsubMock = pubsubMocks.NewPubSuber(t)
	runaipubsub.SetMock(pubsubMock)
	daoMock = mocks.NewMetricStoresDao(t)
	tenantMetricStoreGetterMock = mocks.NewTenantMetricStoreGetter(t)
	providerMock = mocks.NewDatasourceProvider(t)

	cacheMock = cache_testing.SetMock(t)
	cacheMock.On("RegisterStore", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	config.Set(&metricstores.MetricStoresConfiguration{
		Cache: &metricstores.CacheConfiguration{TTLMinutes: 60},
	})
	config.Set(&metricstores.DatasourceIdConfig{
		Env:     "test",
		Release: "runai-backend",
	})
	t.Cleanup(func() { config.FlushAll() })

	manager = metricstores.NewManager(daoMock, tenantMetricStoreGetterMock, providerMock)
	return
}

func TestDeleter_SubscribeToTenantDeleteEvent(t *testing.T) {
	pubsubMock, _, _, _, _, manager := initManager(t)
	pubsubMock.On("SubscribeAsync", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	manager.SubscribeToTenantDeleteEvent()
	time.Sleep(time.Second * 2) // wait for the subscriber to subscribe
}

func TestManager_Create_ExternalWithDatasourceId(t *testing.T) {
	_, _, daoMock, _, _, manager := initManager(t)
	daoMock.On("GetByTenantId", mock.Anything, int32(1)).Return(nil, nil)
	daoMock.On("Create", mock.Anything, mock.MatchedBy(func(m models.MetricStore) bool {
		return m.TenantId == 1 && m.DatasourceId == "migrator-id"
	})).Return(&models.MetricStore{TenantId: 1, DatasourceId: "migrator-id"}, nil)

	res, err := manager.Create(context.Background(), api.MetricStore{
		TenantId:           1,
		DatasourceId:       "migrator-id",
		MetricStoreDetails: &api.MetricStoreDetails{Type: api.METRICSTORETYPE_MIMIR},
	})
	require.NoError(t, err)
	require.Equal(t, "migrator-id", res.DatasourceId)
}

func TestManager_Create_ExternalNoDatasourceId(t *testing.T) {
	_, _, daoMock, _, _, manager := initManager(t)
	daoMock.On("GetByTenantId", mock.Anything, int32(1)).Return(nil, nil)
	daoMock.On("Create", mock.Anything, mock.MatchedBy(func(m models.MetricStore) bool {
		return m.DatasourceId == "external"
	})).Return(&models.MetricStore{TenantId: 1, DatasourceId: "external"}, nil)

	res, err := manager.Create(context.Background(), api.MetricStore{
		TenantId:           1,
		MetricStoreDetails: &api.MetricStoreDetails{Type: api.METRICSTORETYPE_MIMIR},
	})
	require.NoError(t, err)
	require.Equal(t, "external", res.DatasourceId)
}

func TestManager_Create_GenerateDatasource(t *testing.T) {
	_, _, daoMock, _, providerMock, manager := initManager(t)
	daoMock.On("GetByTenantId", mock.Anything, int32(42)).Return(nil, nil)

	expectedDatasourceId := "test42runaibackend"
	providerMock.On("Create", mock.Anything, expectedDatasourceId).Return(&metricstores.Datasource{
		Type:  metricstores.DatasourceTypeGrafanaLabs,
		Read:  metricstores.DatasourceEndpoint{Url: "http://read"},
		Write: metricstores.DatasourceEndpoint{Url: "http://write"},
	}, nil)
	daoMock.On("Create", mock.Anything, mock.MatchedBy(func(m models.MetricStore) bool {
		return m.TenantId == 42 && m.DatasourceId == expectedDatasourceId && m.MetricStoreDetails.Type == models.MSTGrafanaLabs
	})).Return(&models.MetricStore{TenantId: 42, DatasourceId: expectedDatasourceId, MetricStoreDetails: models.MetricStoreDetails{Type: models.MSTGrafanaLabs}}, nil)

	res, err := manager.Create(context.Background(), api.MetricStore{TenantId: 42})
	require.NoError(t, err)
	require.Equal(t, expectedDatasourceId, res.DatasourceId)
}

func TestManager_Create_GenerateDatasource_ProviderError(t *testing.T) {
	_, _, daoMock, _, providerMock, manager := initManager(t)
	daoMock.On("GetByTenantId", mock.Anything, int32(1)).Return(nil, nil)
	providerMock.On("Create", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("provider boom"))

	_, err := manager.Create(context.Background(), api.MetricStore{TenantId: 1})
	require.ErrorContains(t, err, "failed to create datasource")
}

func TestManager_GetByTenantId_Cached(t *testing.T) {
	_, cacheMock, _, _, _, manager := initManager(t)
	cacheMock.On("Get", mock.Anything, cache_testing.KeyFromType[api.MetricStore](), "1").Return(&api.MetricStore{TenantId: 1, DatasourceId: "test"}, nil)

	store, err := manager.GetByTenantId(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, &api.MetricStore{TenantId: 1, DatasourceId: "test"}, store)
}

func TestManager_GetByTenantId_NotCached_FromDB(t *testing.T) {
	_, cacheMock, daoMock, _, _, manager := initManager(t)
	cacheMock.On("Get", mock.Anything, cache_testing.KeyFromType[api.MetricStore](), "1").Return(nil, nil)
	cacheMock.On("Set", mock.Anything, cache_testing.KeyFromType[api.MetricStore](), "1", mock.Anything).Return(nil)
	daoMock.On("GetByTenantId", mock.Anything, int32(1)).Return(&models.MetricStore{TenantId: 1, DatasourceId: "test"}, nil)
	_, err := manager.GetByTenantId(context.Background(), 1)
	require.NoError(t, err)
}

func TestManager_GetByTenantId_NotCached_FromTenantsManager(t *testing.T) {
	_, cacheMock, daoMock, tenantMetricStoreGetterMock, _, manager := initManager(t)
	cacheMock.On("Get", mock.Anything, cache_testing.KeyFromType[api.MetricStore](), "1").Return(nil, nil)
	daoMock.On("GetByTenantId", mock.Anything, int32(1)).Return(nil, nil)
	tenantMetricStoreGetterMock.On("GetExtendedMetricStore", mock.Anything, int32(1)).Return(&tenantsApi.ExtendedMetricStore{Type: "mimir", DatasourceId: "test"}, nil)
	_, err := manager.GetByTenantId(context.Background(), 1)
	require.NoError(t, err)
}

func TestManager_GetByTenantId_NotFound(t *testing.T) {
	_, cacheMock, daoMock, tenantMetricStoreGetterMock, _, manager := initManager(t)
	cacheMock.On("Get", mock.Anything, cache_testing.KeyFromType[api.MetricStore](), "1").Return(nil, nil)
	daoMock.On("GetByTenantId", mock.Anything, int32(1)).Return(nil, nil)
	tenantMetricStoreGetterMock.On("GetExtendedMetricStore", mock.Anything, int32(1)).Return(nil, nil)

	store, err := manager.GetByTenantId(context.Background(), 1)
	require.NoError(t, err)
	require.Nil(t, store)
}

func TestManager_DeleteByTenantId_External(t *testing.T) {
	_, cacheMock, daoMock, _, _, manager := initManager(t)
	daoMock.On("GetByTenantId", mock.Anything, int32(1)).Return(&models.MetricStore{TenantId: 1, DatasourceId: "external"}, nil)
	daoMock.On("Delete", mock.Anything, int32(1)).Return(nil)
	cacheMock.On("Del", mock.Anything, cache_testing.KeyFromType[api.MetricStore](), "1").Return(nil)

	err := manager.DeleteByTenantId(context.Background(), 1)
	require.NoError(t, err)
}

func TestManager_DeleteByTenantId_Provisioned(t *testing.T) {
	_, cacheMock, daoMock, _, providerMock, manager := initManager(t)
	daoMock.On("GetByTenantId", mock.Anything, int32(1)).Return(&models.MetricStore{TenantId: 1, DatasourceId: "runaitest1runaibackend"}, nil)
	providerMock.On("Delete", mock.Anything, "runaitest1runaibackend").Return(nil)
	daoMock.On("Delete", mock.Anything, int32(1)).Return(nil)
	cacheMock.On("Del", mock.Anything, cache_testing.KeyFromType[api.MetricStore](), "1").Return(nil)

	err := manager.DeleteByTenantId(context.Background(), 1)
	require.NoError(t, err)
}

func TestManager_DeleteByTenantId_ProviderError_StillDeletes(t *testing.T) {
	_, cacheMock, daoMock, _, providerMock, manager := initManager(t)
	daoMock.On("GetByTenantId", mock.Anything, int32(1)).Return(&models.MetricStore{TenantId: 1, DatasourceId: "runaitest1runaibackend"}, nil)
	providerMock.On("Delete", mock.Anything, "runaitest1runaibackend").Return(fmt.Errorf("provider error"))
	daoMock.On("Delete", mock.Anything, int32(1)).Return(nil)
	cacheMock.On("Del", mock.Anything, cache_testing.KeyFromType[api.MetricStore](), "1").Return(nil)

	err := manager.DeleteByTenantId(context.Background(), 1)
	require.NoError(t, err)
}

func TestManager_DeleteByTenantId_DaoDeleteError(t *testing.T) {
	_, _, daoMock, _, _, manager := initManager(t)
	daoMock.On("GetByTenantId", mock.Anything, int32(1)).Return(&models.MetricStore{TenantId: 1, DatasourceId: "external"}, nil)
	daoMock.On("Delete", mock.Anything, int32(1)).Return(fmt.Errorf("dao error"))

	err := manager.DeleteByTenantId(context.Background(), 1)
	require.Error(t, err)
}
