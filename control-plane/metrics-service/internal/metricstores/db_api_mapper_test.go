/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metricstores

import (
	"testing"

	"github.com/run-ai/runai/control-plane/metrics-service/internal/database/models"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/generated/pkg/api"
	"github.com/stretchr/testify/require"
)

func TestAPIToDBMetricStoreModel(t *testing.T) {
	ms := &api.MetricStore{
		TenantId:     1001,
		DatasourceId: "some_datasource_id",
		MetricStoreDetails: &api.MetricStoreDetails{
			Type: api.METRICSTORETYPE_GRAFANALABS,
			Read: api.MetricStoreEndpoint{
				Url: "url",
				Auth: &api.MetricStoreEndpointAuth{
					Type:  api.METRICSTOREENDPOINTAUTHTYPE_BASIC,
					Basic: &api.MetricStoreEndpointAuthBasic{Username: "username", Password: "password"},
				},
			},
			Write: api.MetricStoreEndpoint{
				Url: "url",
				Auth: &api.MetricStoreEndpointAuth{
					Type:  api.METRICSTOREENDPOINTAUTHTYPE_BASIC,
					Basic: &api.MetricStoreEndpointAuthBasic{Username: "username", Password: "password"},
				},
			},
		},
	}

	expected := &models.MetricStore{
		TenantId:     1001,
		DatasourceId: "some_datasource_id",
		MetricStoreDetails: models.MetricStoreDetails{
			Type: models.MetricStoreType(api.METRICSTORETYPE_GRAFANALABS),
			Read: models.MetricStoreEndpoint{
				Url: "url",
				Auth: &models.MetricStoreEndpointAuth{
					Type:  models.AuthType(api.METRICSTOREENDPOINTAUTHTYPE_BASIC),
					Basic: &models.AuthBasic{Username: "username", Password: "password"},
				},
				Headers: map[string]string{},
			},
			Write: models.MetricStoreEndpoint{
				Url: "url",
				Auth: &models.MetricStoreEndpointAuth{
					Type:  models.AuthType(api.METRICSTOREENDPOINTAUTHTYPE_BASIC),
					Basic: &models.AuthBasic{Username: "username", Password: "password"},
				},
				Headers: map[string]string{},
			},
		},
	}

	result := APIToDBMetricStoreModel(ms)

	require.Equal(t, expected, result)
}

func TestAPIToDBMetricStoreModel_AuthNil(t *testing.T) {
	ms := &api.MetricStore{
		TenantId:     1001,
		DatasourceId: "some_datasource_id",
		MetricStoreDetails: &api.MetricStoreDetails{
			Type: api.METRICSTORETYPE_GRAFANALABS,
			Read: api.MetricStoreEndpoint{
				Url:  "url",
				Auth: nil,
			},
			Write: api.MetricStoreEndpoint{
				Url:  "url",
				Auth: nil,
			},
		},
	}

	expected := &models.MetricStore{
		TenantId:     1001,
		DatasourceId: "some_datasource_id",
		MetricStoreDetails: models.MetricStoreDetails{
			Type: models.MetricStoreType(api.METRICSTORETYPE_GRAFANALABS),
			Read: models.MetricStoreEndpoint{
				Url:     "url",
				Auth:    nil,
				Headers: map[string]string{},
			},
			Write: models.MetricStoreEndpoint{
				Url:     "url",
				Auth:    nil,
				Headers: map[string]string{},
			},
		},
	}

	result := APIToDBMetricStoreModel(ms)

	require.Equal(t, expected, result)
}

func TestAPIToDBMetricStoreModel_EndpointNil(t *testing.T) {
	ms := &api.MetricStore{
		TenantId:           1001,
		DatasourceId:       "some_datasource_id",
		MetricStoreDetails: &api.MetricStoreDetails{Type: api.METRICSTORETYPE_GRAFANALABS},
	}

	expected := &models.MetricStore{
		TenantId:     1001,
		DatasourceId: "some_datasource_id",
		MetricStoreDetails: models.MetricStoreDetails{
			Type: models.MetricStoreType(api.METRICSTORETYPE_GRAFANALABS),
			Read: models.MetricStoreEndpoint{
				Url:     "",
				Auth:    nil,
				Headers: map[string]string{},
			},
			Write: models.MetricStoreEndpoint{
				Url:     "",
				Auth:    nil,
				Headers: map[string]string{},
			},
		},
	}

	result := APIToDBMetricStoreModel(ms)

	require.Equal(t, expected, result)
}

func TestAPIToDBMetricStoreModel_Nil(t *testing.T) {
	require.Nil(t, APIToDBMetricStoreModel(nil))
}

func TestDBToAPIMetricStore(t *testing.T) {
	ms := &models.MetricStore{
		TenantId:     1001,
		DatasourceId: "some_datasource_id",
		MetricStoreDetails: models.MetricStoreDetails{
			Type: models.MSTMimir,
			Read: models.MetricStoreEndpoint{
				Url: "url",
				Auth: &models.MetricStoreEndpointAuth{
					Type: models.ATBasic,
					Basic: &models.AuthBasic{
						Username: "username",
						Password: "password",
					},
				},
				Headers: map[string]string{
					"something": "something_else",
				},
			},
			Write: models.MetricStoreEndpoint{
				Url: "url",
				Auth: &models.MetricStoreEndpointAuth{
					Type: models.ATBasic,
					Basic: &models.AuthBasic{
						Username: "username",
						Password: "password",
					},
				},
				Headers: map[string]string{
					"something": "something_else",
				},
			},
		},
	}

	expected := &api.MetricStore{
		TenantId:     1001,
		DatasourceId: "some_datasource_id",
		MetricStoreDetails: &api.MetricStoreDetails{
			Type: api.METRICSTORETYPE_MIMIR,
			Read: api.MetricStoreEndpoint{
				Url: "url",
				Auth: &api.MetricStoreEndpointAuth{
					Type:  api.MetricStoreEndpointAuthType(models.ATBasic),
					Basic: &api.MetricStoreEndpointAuthBasic{Username: "username", Password: "password"},
				},
				Headers: map[string]string{
					"something": "something_else",
				},
			},
			Write: api.MetricStoreEndpoint{
				Url: "url",
				Auth: &api.MetricStoreEndpointAuth{
					Type:  api.MetricStoreEndpointAuthType(models.ATBasic),
					Basic: &api.MetricStoreEndpointAuthBasic{Username: "username", Password: "password"},
				},
				Headers: map[string]string{
					"something": "something_else",
				},
			},
		},
	}

	result := DBToAPIMetricStore(ms)

	require.Equal(t, expected, result)
}

func TestDBToAPIMetricStore_Nil(t *testing.T) {
	require.Nil(t, DBToAPIMetricStore(nil))
}

func TestDBToAPIMetricStore_AuthNil(t *testing.T) {
	ms := &models.MetricStore{
		TenantId:     1001,
		DatasourceId: "some_datasource_id",
		MetricStoreDetails: models.MetricStoreDetails{
			Type: models.MSTMimir,
			Read: models.MetricStoreEndpoint{
				Url:  "url",
				Auth: nil,
				Headers: map[string]string{
					"something": "something_else",
				},
			},
			Write: models.MetricStoreEndpoint{
				Url:  "url",
				Auth: nil,
				Headers: map[string]string{
					"something": "something_else",
				},
			},
		},
	}

	expected := &api.MetricStore{
		TenantId:     1001,
		DatasourceId: "some_datasource_id",
		MetricStoreDetails: &api.MetricStoreDetails{
			Type: api.METRICSTORETYPE_MIMIR,
			Read: api.MetricStoreEndpoint{
				Url:  "url",
				Auth: nil,
				Headers: map[string]string{
					"something": "something_else",
				},
			},
			Write: api.MetricStoreEndpoint{
				Url:  "url",
				Auth: nil,
				Headers: map[string]string{
					"something": "something_else",
				},
			},
		},
	}

	result := DBToAPIMetricStore(ms)

	require.Equal(t, expected, result)
}

func TestDBToAPIMetricStore_AuthAuthNil(t *testing.T) {
	ms := &models.MetricStore{
		TenantId:     1001,
		DatasourceId: "some_datasource_id",
		MetricStoreDetails: models.MetricStoreDetails{
			Type: models.MSTMimir,
			Read: models.MetricStoreEndpoint{
				Url: "url",
				Auth: &models.MetricStoreEndpointAuth{
					Type:  models.ATBasic,
					Basic: nil,
				},
				Headers: map[string]string{
					"something": "something_else",
				},
			},
			Write: models.MetricStoreEndpoint{
				Url: "url",
				Auth: &models.MetricStoreEndpointAuth{
					Type:  models.ATBasic,
					Basic: nil,
				},
				Headers: map[string]string{
					"something": "something_else",
				},
			},
		},
	}

	expected := &api.MetricStore{
		TenantId:     1001,
		DatasourceId: "some_datasource_id",
		MetricStoreDetails: &api.MetricStoreDetails{
			Type: api.METRICSTORETYPE_MIMIR,
			Read: api.MetricStoreEndpoint{
				Url: "url",
				Auth: &api.MetricStoreEndpointAuth{
					Type:  api.METRICSTOREENDPOINTAUTHTYPE_BASIC,
					Basic: nil,
				},
				Headers: map[string]string{
					"something": "something_else",
				},
			},
			Write: api.MetricStoreEndpoint{
				Url: "url",
				Auth: &api.MetricStoreEndpointAuth{
					Type:  api.METRICSTOREENDPOINTAUTHTYPE_BASIC,
					Basic: nil,
				},
				Headers: map[string]string{
					"something": "something_else",
				},
			},
		},
	}

	result := DBToAPIMetricStore(ms)

	require.Equal(t, expected, result)
}

func TestDBToAPIMetricStore_EndpointNil(t *testing.T) {
	ms := &models.MetricStore{
		TenantId:           1001,
		DatasourceId:       "some_datasource_id",
		MetricStoreDetails: models.MetricStoreDetails{Type: models.MSTMimir},
	}

	expected := &api.MetricStore{
		TenantId:     1001,
		DatasourceId: "some_datasource_id",
		MetricStoreDetails: &api.MetricStoreDetails{
			Type: api.METRICSTORETYPE_MIMIR,
			Read: api.MetricStoreEndpoint{
				Url:     "",
				Auth:    nil,
				Headers: map[string]string{},
			},
			Write: api.MetricStoreEndpoint{
				Url:     "",
				Auth:    nil,
				Headers: map[string]string{},
			},
		},
	}

	result := DBToAPIMetricStore(ms)

	require.Equal(t, expected, result)
}
