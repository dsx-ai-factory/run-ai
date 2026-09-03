/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metricstores

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	metricerrors "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/errors"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/generated/pkg/api"
	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
)

const externalDatasourceId = "external"

var releaseSanitizer = regexp.MustCompile(`[^a-zA-Z0-9]+`)

type Manager struct {
	metricStoresDao         MetricStoresDao
	tenantMetricStoreGetter TenantMetricStoreGetter
	provider                DatasourceProvider
}

func NewManager(metricStoreDeleter MetricStoresDao, tenantMetricStoreGetter TenantMetricStoreGetter, provider DatasourceProvider) *Manager {
	registerCacheStores()
	return &Manager{
		metricStoresDao:         metricStoreDeleter,
		tenantMetricStoreGetter: tenantMetricStoreGetter,
		provider:                provider,
	}
}

func (m *Manager) SubscribeToTenantDeleteEvent() {
	runaipubsub.SubscribeAsync[events.TenantDelete](func(ctx context.Context, event *events.TenantDelete) error {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("received tenant delete event for tenant %s", event.TenantId)
		tenantId, err := strconv.Atoi(event.TenantId)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to convert tenant id to int: %v", err)
			return err
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("deleting metric stores for tenant %d", tenantId)
		return m.DeleteByTenantId(ctx, int32(tenantId))
	})
}

func (m *Manager) Create(ctx context.Context, metricStore api.MetricStore) (*api.MetricStore, error) {
	existing, err := m.metricStoresDao.GetByTenantId(ctx, metricStore.TenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to fetch existing metric store from DB for tenant %d: %v. Proceeding with create; provider API may return 409 if it already exists.", metricStore.TenantId, err)
	}
	if existing != nil {
		return nil, fmt.Errorf("%w: metric store already exists", metricerrors.ErrMetricStoreAlreadyExists)
	}

	if metricStore.MetricStoreDetails != nil {
		return m.createExternal(ctx, metricStore)
	}
	return m.createProvisioned(ctx, metricStore)
}

func (m *Manager) GetByTenantId(ctx context.Context, tenantId int32) (*api.MetricStore, error) {
	store, err := cache.Get[api.MetricStore](ctx, strconv.Itoa(int(tenantId)))
	if store != nil {
		return store, nil
	}
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get metric store for tenant %d from cache, err: %v", tenantId, err)
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Cache miss for metric store for tenant %d", tenantId)

	store, dbErr := m.getMetricStoreFromDB(ctx, tenantId)
	if dbErr != nil {
		logging.Errorf("Failed to get metric store for tenant %d from DB, err: %v", tenantId, dbErr)
	}
	if store != nil {
		if err := cache.Set[api.MetricStore](ctx, strconv.Itoa(int(tenantId)), *store); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to set metric store for tenant %d in cache, err: %v", tenantId, err)
		}
		return store, nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Metric store for tenant %d not found in DB, getting from tenants manager", tenantId)
	store = m.getMetricStoreFromTenantsManager(ctx, tenantId)
	if store == nil {
		return nil, dbErr
	}

	return store, nil
}

func (m *Manager) getMetricStoreFromDB(ctx context.Context, tenantId int32) (*api.MetricStore, error) {
	metricStore, err := m.metricStoresDao.GetByTenantId(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	return DBToAPIMetricStore(metricStore), nil
}

func (m *Manager) getMetricStoreFromTenantsManager(ctx context.Context, tenantId int32) *api.MetricStore {
	metricStore, err := m.tenantMetricStoreGetter.GetExtendedMetricStore(ctx, tenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get metric store for tenant %d from tenants manager, err: %v", tenantId, err)
		return nil
	}
	return APITenantsToAPIMetricStore(metricStore)
}

func (m *Manager) DeleteByTenantId(ctx context.Context, tenantId int32) error {
	existing, err := m.metricStoresDao.GetByTenantId(ctx, tenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get metric store for tenant %d before deletion: %v", tenantId, err)
		return err
	}
	if existing != nil && existing.DatasourceId != externalDatasourceId {
		if err := m.provider.Delete(ctx, existing.DatasourceId); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to delete datasource %s for tenant %d, proceeding with DB delete: %v", existing.DatasourceId, tenantId, err)
		}
	}

	if err := m.metricStoresDao.Delete(ctx, tenantId); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to delete metric stores for tenant %d: %v", tenantId, err)
		return err
	}
	if err := cache.Del[api.MetricStore](ctx, strconv.Itoa(int(tenantId))); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to delete metric store for tenant %d from cache, err: %v", tenantId, err)
	}
	return nil
}

func (m *Manager) createExternal(ctx context.Context, metricStore api.MetricStore) (*api.MetricStore, error) {
	dbModel := APIToDBMetricStoreModel(&metricStore)
	if dbModel.DatasourceId == "" {
		dbModel.DatasourceId = externalDatasourceId
	}
	saved, err := m.metricStoresDao.Create(ctx, *dbModel)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to save external metric store for tenant %d: %v", metricStore.TenantId, err)
		return nil, err
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("External metric store created for tenant %d", metricStore.TenantId)
	return DBToAPIMetricStore(saved), nil
}

func (m *Manager) createProvisioned(ctx context.Context, metricStore api.MetricStore) (*api.MetricStore, error) {
	datasourceId := generateDatasourceId(metricStore.TenantId)
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Provisioning datasource %s for tenant %d", datasourceId, metricStore.TenantId)
	ds, err := m.provider.Create(ctx, datasourceId)
	if err != nil {
		return nil, fmt.Errorf("failed to create datasource: %w", err)
	}

	dbModel := DatasourceToDBMetricStore(metricStore.TenantId, datasourceId, ds)
	saved, err := m.metricStoresDao.Create(ctx, *dbModel)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to save provisioned metric store for tenant %d, rolling back datasource %s: %v", metricStore.TenantId, datasourceId, err)
		if delErr := m.provider.Delete(ctx, datasourceId); delErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to clean up orphan datasource %s: %v. Manual cleanup required", datasourceId, delErr)
		}
		return nil, err
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Provisioned metric store created for tenant %d (datasource %s)", metricStore.TenantId, datasourceId)
	return DBToAPIMetricStore(saved), nil
}

func generateDatasourceId(tenantId int32) string {
	cfg := config.Get[*DatasourceIdConfig]()
	release := strings.ToLower(releaseSanitizer.ReplaceAllString(cfg.Release, ""))
	return fmt.Sprintf("%s%d%s", cfg.Env, tenantId, release)
}

func registerCacheStores() {
	cfg := config.Get[*MetricStoresConfiguration]()
	if err := cache.RegisterStore[api.MetricStore](options.Ttl(time.Duration(cfg.Cache.TTLMinutes)*time.Minute), options.MaxStoreSizeMB(10), options.MaxRecordSizeKB(1)); err != nil {
		logging.Errorf("failed to create metric store cache: %v", err)
	}
}
