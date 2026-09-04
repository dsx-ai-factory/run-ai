/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"os"

	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

// ServerConfig holds server-related configurations
type ServerConfig struct {
	ServicePort           int `mapstructure:"service_port"`
	PoliciesPerBulk       int `mapstructure:"policies_per_bulk"`
	SyncApiWorkers        int `mapstructure:"sync_api_workers"`
	RetryMaxIntervalSec   int `mapstructure:"retry_max_interval_sec"`
	RetryMaxElapsedSec    int `mapstructure:"retry_max_elapsed_sec"`
	DaysLimitForFirstSync int `mapstructure:"days_limit_for_first_sync"`
}

// SetDefaults sets default values for ServerConfig
func (sc *ServerConfig) SetDefaults() {
	sc.ServicePort = 8080
	sc.PoliciesPerBulk = 50
	sc.SyncApiWorkers = 20
	sc.RetryMaxIntervalSec = 10
	sc.RetryMaxElapsedSec = 60
	sc.DaysLimitForFirstSync = 7
}

// ServicesConfig holds configuration for external services
type ServicesConfig struct {
	AssetsServiceUrl  string `mapstructure:"assets_service_url"`
	ClusterServiceUrl string `mapstructure:"cluster_service_url"`
	OrgUnitServiceUrl string `mapstructure:"org_unit_service_url"`
	LogLevel          string `mapstructure:"log_level"`
}

// SetDefaults sets default values for ServicesConfig
func (sc *ServicesConfig) SetDefaults() {
	sc.AssetsServiceUrl = "http://runai-backend-assets-service:8080"
	sc.ClusterServiceUrl = "http://runai-backend-cluster-service:8080"
	sc.OrgUnitServiceUrl = "http://runai-backend-org-unit-service:8080"
	sc.LogLevel = "info"
}

// CacheConfig defines general caching configuration
type CacheConfig struct {
	ExpirationSec int `mapstructure:"expiration_sec"`
	CleanupSec    int `mapstructure:"cleanup_sec"`
}

// registerConfigs registers the individual configuration structs
func registerConfigs() {
	if err := runai_config.Register[*ServerConfig](); err != nil {
		logging.Fatalf("Failed to register server config: %v", err)
	}

	if err := runai_config.Register[*ServicesConfig](); err != nil {
		logging.Fatalf("Failed to register services config: %v", err)
	}

}

// Init initializes the configuration
func Init() {
	registerConfigs()
}

// MockConfig sets mock environment variables for testing
func MockConfig() {
	os.Setenv("SECRETS_MOUNT_PATH", ".")
	os.Setenv("PUBSUB_ENABLED", "false")

	runai_config.Set(&ServerConfig{
		PoliciesPerBulk: 5,
	})
}
