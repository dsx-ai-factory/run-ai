/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"fmt"
	rcpconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config"
)

func TestMetricsExporter(t *testing.T) {
	mockEnv(t)

	exporter := GetMetricsExporter()
	assert.NotNil(t, exporter)
	exporter.Init()

	server := httptest.NewServer(promhttp.Handler())
	defer server.Close()

	testBackendSyncCounter(t, exporter, server.URL, 3)
}

func testBackendSyncCounter(t *testing.T, exporter *MetricsExporter, serverURL string,
	expectedPullCounter int) {
	validateMetricValues(t, 0)

	for i := 0; i < expectedPullCounter; i++ {
		exporter.IncBackendSyncCounter(Pull)
	}
	validateMetricValues(t, expectedPullCounter)

	response, err := http.Get(fmt.Sprintf("%s%s", serverURL, config.Config().MetricsPath))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	validateMetricValues(t, expectedPullCounter)
}

func validateMetricValues(t *testing.T, expectedPullCounter int) {
	metricFamily, err := prometheus.DefaultGatherer.Gather()
	assert.NoError(t, err)

	for _, mf := range metricFamily {
		if *mf.Name != syncDurationMetricName {
			continue
		}
		for _, m := range mf.Metric {
			for _, label := range m.Label {
				if *label.Name != "type" {
					continue
				}
				switch *label.Value {
				case string(Pull):
					assert.Equal(t, float64(expectedPullCounter), *m.Counter.Value)
				}
			}
		}
	}
}

func mockEnv(t *testing.T) {
	t.Setenv("CLUSTER_DOMAIN", "1")
	t.Setenv("CLUSTER_VERSION", "1")
	t.Setenv("ENDPOINT", "1")
	t.Setenv("METRICS_PORT", "8080")
	t.Setenv("METRICS_PATH", "/metrics")

	ag, _ := rcpconfig.GetR[*config.AgentConfig]()
	ag.ClientName = "1"
	ag.ClusterUuid = "1"
	ag.ClientSecret = "1"
	_ = rcpconfig.Set(ag)
}
