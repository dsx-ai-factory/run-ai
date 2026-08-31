/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	authorizationclient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/run-ai/runai/control-plane/cluster-service/pkg/config"
)

// Helper function to create a secrets directory with files
func createSecretsDir(t *testing.T, secrets map[string]string) string {
	tmpDir, err := os.MkdirTemp("", "cluster-service-secrets-*")
	assert.NoError(t, err)

	for key, value := range secrets {
		secretPath := filepath.Join(tmpDir, key)
		err := os.WriteFile(secretPath, []byte(value), 0o600)
		assert.NoError(t, err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	return tmpDir
}

// Helper function to initialize config for testing
func initConfigForTesting(t *testing.T, secretsDir string) {
	runai_config.FlushAll()

	if secretsDir == "" {
		secretsDir = t.TempDir()
	}

	// Register config types that were moved to their consuming packages
	if err := runai_config.Register[*config.Service](); err != nil {
		t.Fatalf("Failed to register service configuration: %v", err)
	}
	if err := runai_config.Register[*config.Server](); err != nil {
		t.Fatalf("Failed to register server configuration: %v", err)
	}
	if err := runai_config.Register[*config.Environment](); err != nil {
		t.Fatalf("Failed to register environment configuration: %v", err)
	}
	if err := runai_config.Register[*config.ApplicationMetricStore](); err != nil {
		t.Fatalf("Failed to register application metric store configuration: %v", err)
	}
	if err := runai_config.Register[*config.Sync](); err != nil {
		t.Fatalf("Failed to register sync configuration: %v", err)
	}
	if err := runai_config.Register[*config.AsyncClusterDelete](); err != nil {
		t.Fatalf("Failed to register async cluster delete configuration: %v", err)
	}

	config.Init(secretsDir)
}

func TestRegister_Installation_rejectsInvalidInstallationType(t *testing.T) {
	runai_config.FlushAll()
	require.NoError(t, runai_config.SetKbSecretsFolder(t.TempDir()))
	t.Setenv("INSTALLATION_TYPE", "bogus")
	err := runai_config.Register[*config.Installation]()
	require.Error(t, err)
}

func TestInit_CustomEnvVars(t *testing.T) {
	// Set environment variables before initialization
	t.Setenv("SERVICE_PORT", "9090")
	t.Setenv("CLUSTER_CONNECTIVITY_THRESHOLD", "10m")
	t.Setenv("CLUSTER_HOST_TEMPLATE", "custom-{{.Uuid}}")
	t.Setenv("NODEPOOL_SYNC_INTERVAL_MINUTES", "360")
	t.Setenv("ASYNC_CLUSTER_DELETE_ENABLED", "false")
	t.Setenv("WORKLOADS_SVC_URL", "http://custom-workloads:9000")

	initConfigForTesting(t, "")

	// Verify overrides
	serverConfig := runai_config.Get[*config.Server]()
	assert.Equal(t, 9090, serverConfig.ServicePort)

	appConfig := runai_config.Get[*config.App]()
	assert.Equal(t, 10*time.Minute, appConfig.ClusterConnectivityThreshold)
	assert.Equal(t, "custom-{{.Uuid}}", appConfig.ClusterHostTemplate)
	assert.Equal(t, "http://custom-workloads:9000", appConfig.WorkloadsServiceUrl)

	syncConfig := runai_config.Get[*config.Sync]()
	assert.Equal(t, 360, syncConfig.IntervalMinutes)

	asyncDeleteConfig := runai_config.Get[*config.AsyncClusterDelete]()
	assert.False(t, asyncDeleteConfig.Enabled)
}

func TestInit_SecretsFromFiles(t *testing.T) {
	// Create secrets directory with test secrets
	secrets := map[string]string{
		"application_metric_store_write_username": "test-user",
		"application_metric_store_write_password": "test-password",
	}
	secretsDir := createSecretsDir(t, secrets)

	// Set environment variables for non-secret fields
	t.Setenv("APPLICATION_METRIC_STORE_WRITE_URL", "http://metrics.example.com")
	t.Setenv("APPLICATION_METRIC_STORE_LABEL", "custom_label")

	initConfigForTesting(t, secretsDir)

	// Verify secrets are loaded from files
	metricStoreConfig := runai_config.Get[*config.ApplicationMetricStore]()
	assert.Equal(t, "http://metrics.example.com", metricStoreConfig.Url)
	assert.Equal(t, "test-user", metricStoreConfig.Username)
	assert.Equal(t, "test-password", metricStoreConfig.Password)
	assert.Equal(t, "custom_label", metricStoreConfig.Label)
}

func TestRegisterConfigs_Success(t *testing.T) {
	runai_config.FlushAll()

	// Register config types that were moved to their consuming packages
	if err := runai_config.Register[*config.Service](); err != nil {
		t.Fatalf("Failed to register service configuration: %v", err)
	}
	if err := runai_config.Register[*config.Server](); err != nil {
		t.Fatalf("Failed to register server configuration: %v", err)
	}
	if err := runai_config.Register[*config.Environment](); err != nil {
		t.Fatalf("Failed to register environment configuration: %v", err)
	}
	if err := runai_config.Register[*config.ApplicationMetricStore](); err != nil {
		t.Fatalf("Failed to register application metric store configuration: %v", err)
	}
	if err := runai_config.Register[*config.Sync](); err != nil {
		t.Fatalf("Failed to register sync configuration: %v", err)
	}
	if err := runai_config.Register[*config.AsyncClusterDelete](); err != nil {
		t.Fatalf("Failed to register async cluster delete configuration: %v", err)
	}

	// Test that registration succeeds without errors
	config.Init(t.TempDir())

	// Verify all configs are accessible
	assert.NotNil(t, runai_config.Get[*config.Service]())
	assert.NotNil(t, runai_config.Get[*config.Server]())
	assert.NotNil(t, runai_config.Get[*config.Environment]())
	assert.NotNil(t, runai_config.Get[*config.Auth]())
	assert.NotNil(t, runai_config.Get[*config.Backend]())
	assert.NotNil(t, runai_config.Get[*config.App]())
	assert.NotNil(t, runai_config.Get[*config.Installation]())
	assert.NotNil(t, runai_config.Get[*config.CloudFormation]())
	assert.NotNil(t, runai_config.Get[*config.Sync]())
	assert.NotNil(t, runai_config.Get[*config.AsyncClusterDelete]())
	assert.NotNil(t, runai_config.Get[*config.ApplicationMetricStore]())
	assert.NotNil(t, runai_config.Get[*authorizationclient.Config]())
}

func TestInit_FlushAllBetweenTests(t *testing.T) {
	// First registration with custom value
	runai_config.FlushAll()
	t.Setenv("NODEPOOL_SYNC_INTERVAL_MINUTES", "180")

	// Register config types
	if err := runai_config.Register[*config.Sync](); err != nil {
		t.Fatalf("Failed to register sync configuration: %v", err)
	}
	config.Init(t.TempDir())

	syncConfig1 := runai_config.Get[*config.Sync]()
	assert.Equal(t, 180, syncConfig1.IntervalMinutes)

	// Flush and re-register with different value
	runai_config.FlushAll()
	t.Setenv("NODEPOOL_SYNC_INTERVAL_MINUTES", "240")

	// Re-register config types
	if err := runai_config.Register[*config.Sync](); err != nil {
		t.Fatalf("Failed to register sync configuration: %v", err)
	}
	config.Init(t.TempDir())

	syncConfig2 := runai_config.Get[*config.Sync]()
	assert.Equal(t, 240, syncConfig2.IntervalMinutes)

	// Verify the configs are different instances after flush
	assert.NotEqual(t, syncConfig1.IntervalMinutes, syncConfig2.IntervalMinutes)
}
