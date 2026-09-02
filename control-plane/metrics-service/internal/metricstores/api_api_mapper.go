/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metricstores

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/generated/pkg/api"
	tenants_api "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client/generated"
)

func APITenantsToAPIMetricStore(extendedMetricStore *tenants_api.ExtendedMetricStore) *api.MetricStore {
	if extendedMetricStore == nil {
		return nil
	}

	return &api.MetricStore{
		TenantId:     extendedMetricStore.TenantId,
		DatasourceId: extendedMetricStore.DatasourceId,
		MetricStoreDetails: &api.MetricStoreDetails{
			Type:  api.MetricStoreType(extendedMetricStore.Type),
			Read:  APITenantsToAPIMetricStoreEndpoint(extendedMetricStore.Read),
			Write: APITenantsToAPIMetricStoreEndpoint(extendedMetricStore.Write),
		},
	}
}

func APITenantsToAPIMetricStoreEndpoint(tenantsMSEndpoint tenants_api.MetricStoreEndpoint) api.MetricStoreEndpoint {
	metricStoreEndpoint := api.MetricStoreEndpoint{
		Url: tenantsMSEndpoint.Url,
	}

	if tenantsMSEndpoint.Auth.Get().HasBasic() {
		if tenantsMSEndpoint.Auth.Get().GetBasic().Username == nil || tenantsMSEndpoint.Auth.Get().GetBasic().Password == nil {
			return metricStoreEndpoint
		}
		metricStoreEndpoint.Auth = &api.MetricStoreEndpointAuth{
			Type: api.METRICSTOREENDPOINTAUTHTYPE_BASIC,
			Basic: &api.MetricStoreEndpointAuthBasic{
				Username: *tenantsMSEndpoint.Auth.Get().GetBasic().Username,
				Password: *tenantsMSEndpoint.Auth.Get().GetBasic().Password,
			},
		}
	}

	return metricStoreEndpoint
}
