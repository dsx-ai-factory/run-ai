/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package internal

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/run-ai/runai/runai-common-packages/logging"

	metrics_client "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
	metrics_api "github.com/run-ai/runai/control-plane/metrics-service/pkg/client/generated"
	tenants_manager_api "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client/generated"
)

type Status struct {
	Failed          int
	AlreadyMigrated int
	Succeed         int
}

type Migrator struct {
	tenantsClient      TenantsClienter
	metricStoresClient metrics_client.MetricStoresClienter
	status             Status
}

func NewMigrator(tc TenantsClienter, mc metrics_client.MetricStoresClienter) *Migrator {
	return &Migrator{
		tenantsClient:      tc,
		metricStoresClient: mc,
		status:             Status{},
	}
}

func (m *Migrator) Migrate(ctx context.Context) {
	defer func() {
		logging.Infof("migration complitted, succeed: %d, already migrated: %d failed %d", m.status.Succeed, m.status.AlreadyMigrated, m.status.Failed)
	}()

	tenantsData, err := m.tenantsClient.GetTenants(ctx)
	if err != nil {
		logging.Fatalf("migration failed to get tenants. err: %+v", err)
	}

	for _, tenantData := range tenantsData {
		select {
		case <-ctx.Done():
			logging.Infof("got cancelation signal, terminating")
		default:
			logging.Infof("migrating tenant [id:%d,name:%s]", tenantData.GetId(), tenantData.Name)
			err := m.migrateTenant(ctx, &tenantData)
			if err == nil {
				m.status.Succeed += 1
				logging.Infof("tenant [id:%d,name:%s] migration completed", tenantData.GetId(), tenantData.Name)
				continue
			}

			var apiErr *metrics_api.Error
			if errors.As(err, &apiErr) && apiErr.Code == http.StatusConflict {
				m.status.AlreadyMigrated += 1
				logging.Infof("tenant [id:%d,name:%s] already migrated", tenantData.GetId(), tenantData.Name)

				continue
			}

			m.status.Failed += 1
			logging.Errorf("tenant [id:%d,name:%s] migration failed, err: %+v", tenantData.GetId(), tenantData.Name, err)
		}
	}
}

func (m *Migrator) migrateTenant(ctx context.Context, tenantData *tenants_manager_api.Tenant) error {

	metricStore, err := m.tenantsClient.GetMetricStore(ctx, tenantData.GetId())
	if err != nil {
		return errors.Wrap(err, "failed to get metric store from tenants manager")
	}
	var newMetricStore *metrics_api.MetricStore
	if metricStore == nil {
		logging.Infof("no metric store for tenant, will try creating provisioned metric store: %d", tenantData.GetId())
		newMetricStore = metrics_api.NewMetricStore(tenantData.GetId())
	} else {
		newMetricStore, err = mapTenantMetricStoreToApi(tenantData.GetId(), metricStore)
	}
	if err != nil {
		return err
	}

	_, err = m.metricStoresClient.CreateMetricStore(ctx, newMetricStore)

	if err != nil {
		return errors.Wrap(err, "failed to create metric store")
	}

	return err
}

func mapTenantMetricStoreToApi(tenantId int32, metricStore *tenants_manager_api.ExtendedMetricStore) (*metrics_api.MetricStore, error) {
	newMetricStore := metrics_api.NewMetricStore(tenantId)
	typ, err := metrics_api.NewMetricStoreTypeFromValue(metricStore.Type)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to parse metric store type: %s", metricStore.Type))
	}

	newMetricStore.SetDatasourceId(metricStore.DatasourceId)
	newMetricStore.SetMetricStoreDetails(metrics_api.MetricStoreDetails{
		Type:  *typ,
		Read:  *getMetricStoreEndpoint(metricStore.Read),
		Write: *getMetricStoreEndpoint(metricStore.Write),
	})
	return newMetricStore, nil
}

func getMetricStoreEndpoint(endpoint tenants_manager_api.MetricStoreEndpoint) *metrics_api.MetricStoreEndpoint {
	ep := &metrics_api.MetricStoreEndpoint{
		Url:     endpoint.Url,
		Auth:    metrics_api.NullableMetricStoreEndpointAuth{},
		Headers: &endpoint.Headers,
	}
	if !endpoint.Auth.IsSet() {
		return ep
	}

	auth := metrics_api.NewMetricStoreEndpointAuth()
	ep.Auth.Set(auth)
	if !endpoint.Auth.Get().HasBasic() {
		return ep
	}

	auth.SetType(metrics_api.METRICSTOREENDPOINTAUTHTYPE_BASIC)
	if basic := endpoint.Auth.Get().GetBasic(); basic.Username != nil && basic.Password != nil {
		auth.SetBasic(*metrics_api.NewMetricStoreEndpointAuthBasic(*basic.Username, *basic.Password))
	}
	return ep
}
