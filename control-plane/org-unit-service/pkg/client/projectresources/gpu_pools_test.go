/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package projectresources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/stretchr/testify/require"
)

func TestProjectGPUPoolIDs(t *testing.T) {
	pages := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pages++
		if r.URL.Path != "/api/v1/org-unit/projects/node-pool-resources" ||
			r.URL.Query().Get("filterBy") != "orgUnitId==42,gpuAccessible==true" {
			http.Error(w, "unexpected project filters", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("offset") == "0" {
			_, _ = w.Write([]byte(`{"nodePoolResources":[{"orgUnitId":"42","nodePoolId":"first"}],"next":1}`))
		} else {
			_, _ = w.Write([]byte(`{"nodePoolResources":[{"orgUnitId":"42","nodePoolId":"second"}]}`))
		}
	}))
	defer server.Close()
	conf := openapi.NewConfiguration()
	conf.Servers = openapi.ServerConfigurations{{URL: server.URL}}
	client := openapi.NewAPIClient(conf)
	pools, err := GPUEnabledPoolIDs(context.Background(), client.ProjectsAPI, "42")
	require.NoError(t, err)
	require.Equal(t, []string{"first", "second"}, pools)
	require.Equal(t, 2, pages)
}
