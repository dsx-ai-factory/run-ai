/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package providers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	wc_openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"

	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/business/providers"
	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/business/settings"
)

func newWorkloadsBulkTestServer(resp map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func newWorkloadsClient(serverURL string) *wc_openapi.APIClient {
	cfg := wc_openapi.NewConfiguration()
	cfg.Servers = wc_openapi.ServerConfigurations{{URL: serverURL}}
	return wc_openapi.NewAPIClient(cfg)
}

func boolPtr(v bool) *bool { return &v }

func TestWorkloadsProvider_FetchSettings_AllSettings(t *testing.T) {
	server := newWorkloadsBulkTestServer(map[string]interface{}{
		"enablePolicies":              map[string]interface{}{"value": map[string]interface{}{"enabled": true}},
		"workloadSubmissionV2Enabled": map[string]interface{}{"value": map[string]interface{}{"enabled": false}},
		"enableTemplatesV2":           map[string]interface{}{"value": map[string]interface{}{"enabled": true}},
		"enableHybridSubmission":      map[string]interface{}{"value": map[string]interface{}{"enabled": true}},
		"gpuOverProvisioning":         map[string]interface{}{"value": map[string]interface{}{"enabled": true}},
	})
	defer server.Close()

	provider := providers.NewWorkloadsProvider(newWorkloadsClient(server.URL))
	result, err := provider.FetchSettings(context.Background(), 1)

	assert.NoError(t, err)
	assert.Contains(t, result, settings.SettingKeyEnablePolicies)
	assert.Equal(t, true, result[settings.SettingKeyEnablePolicies].Value)
	assert.Contains(t, result, settings.SettingKeyWorkloadSubmissionV2Enabled)
	assert.Equal(t, false, result[settings.SettingKeyWorkloadSubmissionV2Enabled].Value)
	assert.Contains(t, result, settings.SettingKeyEnableTemplatesV2)
	assert.Equal(t, true, result[settings.SettingKeyEnableTemplatesV2].Value)
	assert.Contains(t, result, settings.SettingKeyEnableHybridSubmission)
	assert.Equal(t, true, result[settings.SettingKeyEnableHybridSubmission].Value)
	assert.Contains(t, result, settings.SettingKeyGpuOverProvisioning)
	assert.Equal(t, true, result[settings.SettingKeyGpuOverProvisioning].Value)
}

func TestWorkloadsProvider_FetchSettings_WithMetadata(t *testing.T) {
	server := newWorkloadsBulkTestServer(map[string]interface{}{
		"enablePolicies": map[string]interface{}{
			"value":       map[string]interface{}{"enabled": true},
			"displayName": "Enable Policies",
			"description": "Toggle policy enforcement",
			"stage":       "GA",
		},
		"workloadSubmissionV2Enabled": map[string]interface{}{
			"value":       map[string]interface{}{"enabled": false},
			"displayName": "Workload Submission V2",
		},
		"enableTemplatesV2":      map[string]interface{}{"value": map[string]interface{}{"enabled": true}},
		"enableHybridSubmission": map[string]interface{}{"value": map[string]interface{}{"enabled": true}},
		"gpuOverProvisioning":    map[string]interface{}{"value": map[string]interface{}{"enabled": true}},
	})
	defer server.Close()

	provider := providers.NewWorkloadsProvider(newWorkloadsClient(server.URL))
	result, err := provider.FetchSettings(context.Background(), 1)

	assert.NoError(t, err)

	policies := result[settings.SettingKeyEnablePolicies]
	assert.Equal(t, true, policies.Value)
	assert.Equal(t, "Enable Policies", policies.DisplayName)
	assert.Equal(t, "Toggle policy enforcement", policies.Description)
	assert.Equal(t, "GA", policies.Stage)

	submission := result[settings.SettingKeyWorkloadSubmissionV2Enabled]
	assert.Equal(t, false, submission.Value)
	assert.Equal(t, "Workload Submission V2", submission.DisplayName)
	assert.Empty(t, submission.Description)
	assert.Empty(t, submission.Stage)

	templates := result[settings.SettingKeyEnableTemplatesV2]
	assert.Equal(t, true, templates.Value)
	assert.Empty(t, templates.DisplayName)
}

func TestWorkloadsProvider_FetchSettings_MixedValues(t *testing.T) {
	server := newWorkloadsBulkTestServer(map[string]interface{}{
		"enablePolicies":              map[string]interface{}{"value": map[string]interface{}{"enabled": false}},
		"workloadSubmissionV2Enabled": map[string]interface{}{"value": map[string]interface{}{"enabled": true}},
		"enableTemplatesV2":           map[string]interface{}{"value": map[string]interface{}{"enabled": false}},
		"enableHybridSubmission":      map[string]interface{}{"value": map[string]interface{}{"enabled": false}},
		"gpuOverProvisioning":         map[string]interface{}{"value": map[string]interface{}{"enabled": false}},
	})
	defer server.Close()

	provider := providers.NewWorkloadsProvider(newWorkloadsClient(server.URL))
	result, err := provider.FetchSettings(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, false, result[settings.SettingKeyEnablePolicies].Value)
	assert.Equal(t, true, result[settings.SettingKeyWorkloadSubmissionV2Enabled].Value)
	assert.Equal(t, false, result[settings.SettingKeyEnableTemplatesV2].Value)
	assert.Equal(t, false, result[settings.SettingKeyEnableHybridSubmission].Value)
	assert.Equal(t, false, result[settings.SettingKeyGpuOverProvisioning].Value)
}

func TestWorkloadsProvider_FetchSettings_Unreachable(t *testing.T) {
	provider := providers.NewWorkloadsProvider(newWorkloadsClient("http://localhost:1"))
	result, err := provider.FetchSettings(context.Background(), 1)

	assert.NoError(t, err)
	assert.Empty(t, result)
}
