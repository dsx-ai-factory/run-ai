/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package adapters_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	wc_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"

	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/business/adapters"
	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/business/settings"
)

const testWorkloadsTenantID = 42

type workloadsTestServer struct {
	server     *httptest.Server
	lastMethod string
	lastPath   string
	lastBody   map[string]interface{}
	statusCode int
}

func newWorkloadsAdapterTestServer(statusCode int) *workloadsTestServer {
	ts := &workloadsTestServer{statusCode: statusCode}
	ts.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ts.lastMethod = r.Method
		ts.lastPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		if len(body) > 0 {
			_ = json.Unmarshal(body, &ts.lastBody)
		}
		w.WriteHeader(ts.statusCode)
	}))
	return ts
}

func newWorkloadsAdapterWithServer(ts *workloadsTestServer) *adapters.WorkloadsAdapter {
	cfg := wc_openapi.NewConfiguration()
	cfg.Servers = wc_openapi.ServerConfigurations{{URL: ts.server.URL}}
	client := wc_openapi.NewAPIClient(cfg)
	return &adapters.WorkloadsAdapter{SettingsAPI: client.SettingsAPI}
}

func TestWorkloadsAdapter_UpdateSetting_EnablePolicies(t *testing.T) {
	ts := newWorkloadsAdapterTestServer(http.StatusOK)
	defer ts.server.Close()
	adapter := newWorkloadsAdapterWithServer(ts)

	err := adapter.UpdateSetting(context.Background(), testWorkloadsTenantID, settings.SettingKeyEnablePolicies, true)

	assert.NoError(t, err)
	assert.Equal(t, http.MethodPut, ts.lastMethod)
	assert.Equal(t, "/api/v1/workloads/settings/enablePolicies", ts.lastPath)
	assert.Equal(t, true, ts.lastBody["enabled"])
}

func TestWorkloadsAdapter_UpdateSetting_EnablePolicies_False(t *testing.T) {
	ts := newWorkloadsAdapterTestServer(http.StatusOK)
	defer ts.server.Close()
	adapter := newWorkloadsAdapterWithServer(ts)

	err := adapter.UpdateSetting(context.Background(), testWorkloadsTenantID, settings.SettingKeyEnablePolicies, false)

	assert.NoError(t, err)
	assert.Equal(t, false, ts.lastBody["enabled"])
}

func TestWorkloadsAdapter_UpdateSetting_WorkloadSubmissionV2Enabled(t *testing.T) {
	ts := newWorkloadsAdapterTestServer(http.StatusOK)
	defer ts.server.Close()
	adapter := newWorkloadsAdapterWithServer(ts)

	err := adapter.UpdateSetting(context.Background(), testWorkloadsTenantID, settings.SettingKeyWorkloadSubmissionV2Enabled, true)

	assert.NoError(t, err)
	assert.Equal(t, "/api/v1/workloads/settings/workloadSubmissionV2Enabled", ts.lastPath)
	assert.Equal(t, true, ts.lastBody["enabled"])
}

func TestWorkloadsAdapter_UpdateSetting_EnableTemplatesV2(t *testing.T) {
	ts := newWorkloadsAdapterTestServer(http.StatusOK)
	defer ts.server.Close()
	adapter := newWorkloadsAdapterWithServer(ts)

	err := adapter.UpdateSetting(context.Background(), testWorkloadsTenantID, settings.SettingKeyEnableTemplatesV2, true)

	assert.NoError(t, err)
	assert.Equal(t, "/api/v1/workloads/settings/enableTemplatesV2", ts.lastPath)
	assert.Equal(t, true, ts.lastBody["enabled"])
}

func TestWorkloadsAdapter_UpdateSetting_EnableHybridSubmission(t *testing.T) {
	ts := newWorkloadsAdapterTestServer(http.StatusOK)
	defer ts.server.Close()
	adapter := newWorkloadsAdapterWithServer(ts)

	err := adapter.UpdateSetting(context.Background(), testWorkloadsTenantID, settings.SettingKeyEnableHybridSubmission, true)

	assert.NoError(t, err)
	assert.Equal(t, "/api/v1/workloads/settings/enableHybridSubmission", ts.lastPath)
	assert.Equal(t, true, ts.lastBody["enabled"])
}

func TestWorkloadsAdapter_UpdateSetting_GpuOverProvisioning(t *testing.T) {
	ts := newWorkloadsAdapterTestServer(http.StatusOK)
	defer ts.server.Close()
	adapter := newWorkloadsAdapterWithServer(ts)

	err := adapter.UpdateSetting(context.Background(), testWorkloadsTenantID, settings.SettingKeyGpuOverProvisioning, true)

	assert.NoError(t, err)
	assert.Equal(t, "/api/v1/workloads/settings/gpuOverProvisioning", ts.lastPath)
	assert.Equal(t, true, ts.lastBody["enabled"])
}

func TestWorkloadsAdapter_UpdateSetting_ServerError(t *testing.T) {
	ts := newWorkloadsAdapterTestServer(http.StatusInternalServerError)
	defer ts.server.Close()
	adapter := newWorkloadsAdapterWithServer(ts)

	err := adapter.UpdateSetting(context.Background(), testWorkloadsTenantID, settings.SettingKeyEnablePolicies, true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "enablePolicies")
}

func TestWorkloadsAdapter_UpdateSetting_UnknownKey(t *testing.T) {
	ts := newWorkloadsAdapterTestServer(http.StatusOK)
	defer ts.server.Close()
	adapter := newWorkloadsAdapterWithServer(ts)

	err := adapter.UpdateSetting(context.Background(), testWorkloadsTenantID, "unknown.key", true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown workloads setting key")
}
