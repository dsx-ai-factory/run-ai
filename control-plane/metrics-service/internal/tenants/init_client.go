/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tenants

import (
	"fmt"
	"net/http"

	"github.com/jarcoal/httpmock"
	api "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

func InitClient(useMocks bool) *TenantsClient {
	client := NewTenantsClient()
	if useMocks {
		mockHttpClients(client)
		mockHttpEndpoints()
	}
	return client
}

func mockHttpClients(client *TenantsClient) {
	logging.Infof("Mocking tenants client http clients")
	httpmock.ActivateNonDefault(client.client.GetConfig().HTTPClient)
}

func mockHttpEndpoints() {
	metricStore := &api.ExtendedMetricStore{
		Type:         "thanos",
		DatasourceId: "test",
		Read: api.MetricStoreEndpoint{
			Url: "http://prometheus:9090",
			Headers: map[string]string{
				"X-Scope-OrgID": "123",
			},
		},
	}
	mockMetricsStore(1, http.StatusOK, metricStore)
}

func mockMetricsStore(tenantId int32, statusCode int, body any) {
	conf := config.Get[*TenantsConfiguration]()
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/internal/api/v1/tenants/%v/extended-metric-store", conf.ServiceUrl, tenantId), httpmock.NewJsonResponderOrPanic(statusCode, body))
}
