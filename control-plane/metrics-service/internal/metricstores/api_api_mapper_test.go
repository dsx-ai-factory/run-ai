/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metricstores

import (
	"testing"

	"github.com/run-ai/runai/control-plane/metrics-service/internal/generated/pkg/api"
	tenants_api "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client/generated"
	"github.com/stretchr/testify/require"
)

func TestAPITenantsToAPIMetricStore(t *testing.T) {
	username := "test_username"
	password := "test_password"
	basicAuth := tenants_api.MetricStoreEndpointAuthBasic{Username: &username, Password: &password}
	auth := tenants_api.MetricStoreEndpointAuth{}
	auth.SetBasic(basicAuth)

	extendedMetricStore := &tenants_api.ExtendedMetricStore{
		DatasourceId: "test_datasource_id",
		Type:         "mimir",
		Read: tenants_api.MetricStoreEndpoint{
			Url:  "https://read.example.com",
			Auth: *tenants_api.NewNullableMetricStoreEndpointAuth(&auth),
		},
		Write: tenants_api.MetricStoreEndpoint{
			Url:  "https://write.example.com",
			Auth: *tenants_api.NewNullableMetricStoreEndpointAuth(&auth),
		},
	}

	expected := &api.MetricStore{
		DatasourceId: "test_datasource_id",
		MetricStoreDetails: &api.MetricStoreDetails{
			Type: api.MetricStoreType("mimir"),
			Read: api.MetricStoreEndpoint{
				Url: "https://read.example.com",
				Auth: &api.MetricStoreEndpointAuth{
					Type: api.METRICSTOREENDPOINTAUTHTYPE_BASIC,
					Basic: &api.MetricStoreEndpointAuthBasic{
						Username: "test_username",
						Password: "test_password",
					},
				},
			},
			Write: api.MetricStoreEndpoint{
				Url: "https://write.example.com",
				Auth: &api.MetricStoreEndpointAuth{
					Type: api.METRICSTOREENDPOINTAUTHTYPE_BASIC,
					Basic: &api.MetricStoreEndpointAuthBasic{
						Username: "test_username",
						Password: "test_password",
					},
				},
			},
		},
	}

	result := APITenantsToAPIMetricStore(extendedMetricStore)

	require.Equal(t, expected, result)
}

func TestAPITenantsToAPIMetricStore_Nil(t *testing.T) {
	result := APITenantsToAPIMetricStore(nil)
	require.Nil(t, result)
}

func TestAPITenantsToAPIMetricStore_NoBasicAuth(t *testing.T) {
	auth := tenants_api.MetricStoreEndpointAuth{}
	extendedMetricStore := &tenants_api.ExtendedMetricStore{
		DatasourceId: "test_datasource_id",
		Type:         "grafanalabs",
		Read: tenants_api.MetricStoreEndpoint{
			Url:  "https://read.example.com",
			Auth: *tenants_api.NewNullableMetricStoreEndpointAuth(&auth),
		},
		Write: tenants_api.MetricStoreEndpoint{
			Url:  "https://write.example.com",
			Auth: *tenants_api.NewNullableMetricStoreEndpointAuth(&auth),
		},
	}

	expected := &api.MetricStore{
		DatasourceId: "test_datasource_id",
		MetricStoreDetails: &api.MetricStoreDetails{
			Type: api.MetricStoreType("grafanalabs"),
			Read: api.MetricStoreEndpoint{
				Url: "https://read.example.com",
			},
			Write: api.MetricStoreEndpoint{
				Url: "https://write.example.com",
			},
		},
	}

	result := APITenantsToAPIMetricStore(extendedMetricStore)

	require.Equal(t, expected, result)
}

func TestAPITenantsToAPIMetricStoreEndpoint(t *testing.T) {
	username := "endpoint_username"
	password := "endpoint_password"
	basicAuth := tenants_api.MetricStoreEndpointAuthBasic{Username: &username, Password: &password}
	auth := tenants_api.MetricStoreEndpointAuth{}
	auth.SetBasic(basicAuth)

	tenantEndpoint := tenants_api.MetricStoreEndpoint{
		Url:  "https://endpoint.example.com",
		Auth: *tenants_api.NewNullableMetricStoreEndpointAuth(&auth),
	}

	expected := api.MetricStoreEndpoint{
		Url: "https://endpoint.example.com",
		Auth: &api.MetricStoreEndpointAuth{
			Type: api.METRICSTOREENDPOINTAUTHTYPE_BASIC,
			Basic: &api.MetricStoreEndpointAuthBasic{
				Username: "endpoint_username",
				Password: "endpoint_password",
			},
		},
	}

	result := APITenantsToAPIMetricStoreEndpoint(tenantEndpoint)

	require.Equal(t, expected, result)
}

func TestAPITenantsToAPIMetricStoreEndpoint_NoBasicAuth(t *testing.T) {
	auth := tenants_api.MetricStoreEndpointAuth{}
	tenantEndpoint := tenants_api.MetricStoreEndpoint{
		Url:  "https://endpoint.example.com",
		Auth: *tenants_api.NewNullableMetricStoreEndpointAuth(&auth),
	}

	expected := api.MetricStoreEndpoint{
		Url: "https://endpoint.example.com",
	}

	result := APITenantsToAPIMetricStoreEndpoint(tenantEndpoint)

	require.Equal(t, expected, result)
}

func TestAPITenantsToAPIMetricStoreEndpoint_EmptyBasicAuth(t *testing.T) {
	auth := tenants_api.MetricStoreEndpointAuth{
		Basic: *tenants_api.NewNullableMetricStoreEndpointAuthBasic(&tenants_api.MetricStoreEndpointAuthBasic{}),
	}
	tenantEndpoint := tenants_api.MetricStoreEndpoint{
		Url:  "https://endpoint.example.com",
		Auth: *tenants_api.NewNullableMetricStoreEndpointAuth(&auth),
	}

	expected := api.MetricStoreEndpoint{
		Url: "https://endpoint.example.com",
	}

	result := APITenantsToAPIMetricStoreEndpoint(tenantEndpoint)

	require.Equal(t, expected, result)
}

func TestAPITenantsToAPIMetricStoreEndpoint_EmptyUrl(t *testing.T) {
	username := "test_user"
	password := "test_pass"
	basicAuth := tenants_api.MetricStoreEndpointAuthBasic{Username: &username, Password: &password}
	auth := tenants_api.MetricStoreEndpointAuth{}
	auth.SetBasic(basicAuth)

	tenantEndpoint := tenants_api.MetricStoreEndpoint{
		Url:  "",
		Auth: *tenants_api.NewNullableMetricStoreEndpointAuth(&auth),
	}

	expected := api.MetricStoreEndpoint{
		Url: "",
		Auth: &api.MetricStoreEndpointAuth{
			Type: api.METRICSTOREENDPOINTAUTHTYPE_BASIC,
			Basic: &api.MetricStoreEndpointAuthBasic{
				Username: "test_user",
				Password: "test_pass",
			},
		},
	}

	result := APITenantsToAPIMetricStoreEndpoint(tenantEndpoint)

	require.Equal(t, expected, result)
}
