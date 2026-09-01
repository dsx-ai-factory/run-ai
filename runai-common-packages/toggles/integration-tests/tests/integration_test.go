//go:build integration
// +build integration

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/toggles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	adminURL   string
	adminToken string
)

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// unleashAdminRequest sends an HTTP request to the Unleash admin API.
func unleashAdminRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody *bytes.Buffer
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(data)
	} else {
		reqBody = &bytes.Buffer{}
	}

	req, err := http.NewRequest(method, adminURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", adminToken)
	return http.DefaultClient.Do(req)
}

// seedFeatureFlags creates test feature flags in the default project via admin API.
func seedFeatureFlags(t *testing.T) {
	t.Helper()

	// Create an enabled feature.
	createFeature(t, "test-enabled-feature", true)

	// Create a disabled feature (created but not toggled on).
	createFeature(t, "test-disabled-feature", false)
}

func createFeature(t *testing.T, name string, enabled bool) {
	t.Helper()

	resp, err := unleashAdminRequest("POST", "/api/admin/projects/default/features", map[string]interface{}{
		"name": name,
		"type": "release",
	})
	require.NoError(t, err)
	resp.Body.Close()
	// 201 Created or 409 Conflict (already exists) are both fine.
	require.True(t, resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict,
		"unexpected status creating feature %s: %d", name, resp.StatusCode)

	if enabled {
		// Add default strategy.
		resp, err = unleashAdminRequest("POST",
			fmt.Sprintf("/api/admin/projects/default/features/%s/environments/development/strategies", name),
			map[string]string{"name": "default"})
		require.NoError(t, err)
		resp.Body.Close()

		// Toggle on.
		resp, err = unleashAdminRequest("POST",
			fmt.Sprintf("/api/admin/projects/default/features/%s/environments/development/on", name), nil)
		require.NoError(t, err)
		resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode,
			"unexpected status enabling feature %s: %d", name, resp.StatusCode)
	}
}

func setUpIntegrationTest(t *testing.T) {
	t.Helper()

	adminURL = getEnvOrDefault("UNLEASH_ADMIN_URL", "http://localhost:4242")
	adminToken = getEnvOrDefault("UNLEASH_ADMIN_TOKEN", "*:*.toggles-admin-token")
	unleashURL := getEnvOrDefault("UNLEASH_URL", "http://localhost:4242/api")
	apiKey := getEnvOrDefault("UNLEASH_API_KEY", "*:development.toggles-client-token")

	// Seed feature flags before initializing the client, so the first poll picks them up.
	seedFeatureFlags(t)

	enabled := true
	cfg := &toggles.Config{
		Enabled:                &enabled,
		Url:                    unleashURL,
		Environment:            toggles.FFSEDev,
		ApiKey:                 apiKey,
		RefreshIntervalSeconds: 2,
		ServiceName:            "integration-test",
		ProjectName:            "default",
	}

	if err := toggles.Init(cfg); err != nil {
		t.Fatalf("Error initializing toggles: %v", err)
	}

	// Wait for the client to complete its initial feature fetch.
	time.Sleep(4 * time.Second)
}

func tearDownIntegrationTest(t *testing.T) {
	t.Helper()

	if err := toggles.Close(); err != nil {
		t.Fatalf("Error closing toggles: %v", err)
	}
}

// --- Tests ---

func TestIsEnabled_EnabledFeature(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest(t)

	ctx := context.Background()
	assert.True(t, toggles.IsEnabled(ctx, "test-enabled-feature"))
}

func TestIsEnabled_DisabledFeature(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest(t)

	ctx := context.Background()
	assert.False(t, toggles.IsEnabled(ctx, "test-disabled-feature"))
}

func TestIsEnabled_NonexistentFeature(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest(t)

	ctx := context.Background()
	assert.False(t, toggles.IsEnabled(ctx, "does-not-exist"))
}

func TestIsEnabled_WithTenantId(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest(t)

	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "tenant-123")
	assert.True(t, toggles.IsEnabled(ctx, "test-enabled-feature"))
	assert.False(t, toggles.IsEnabled(ctx, "test-disabled-feature"))
}

func TestInitAndClose_Lifecycle(t *testing.T) {
	setUpIntegrationTest(t)

	ctx := context.Background()
	assert.True(t, toggles.IsEnabled(ctx, "test-enabled-feature"))

	tearDownIntegrationTest(t)

	// After close, the dummy toggler is used — always returns true.
	assert.True(t, toggles.IsEnabled(ctx, "test-disabled-feature"))
	assert.True(t, toggles.IsEnabled(ctx, "does-not-exist"))
}
