/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	"github.com/spf13/viper"
)

var responses map[string]openapi.GetMetrics200Response

func SetMockMetricsServer(responsesMap map[string]openapi.GetMetrics200Response) {
	responses = responsesMap
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body openapi.MetricInformation
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Failed to decode JSON body", http.StatusBadRequest)
			return
		}

		var jsonResponse []byte
		if response, ok := responses[body.Name]; ok {
			jsonResponse, _ = json.Marshal(response)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
	}))
	viper.Set("metrics_server_host", strings.TrimPrefix(mockServer.URL, "http://"))
}

func AddResponsesForMockMetricsServer(responsesMap map[string]openapi.GetMetrics200Response) {
	for key, value := range responsesMap {
		responses[key] = value
	}
}
