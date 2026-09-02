/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	authorizationClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"os"
	"sync"
)

var once sync.Once

type ServerConfig struct {
	ServicePort           int `mapstructure:"service_port"`
	RetryMaxIntervalSec   int `mapstructure:"retry_max_interval_sec"`
	RetryMaxElapsedSec    int `mapstructure:"retry_max_elapsed_sec"`
	DaysLimitForFirstSync int `mapstructure:"days_limit_for_first_sync"`

	CleanupRefreshRateHours int `mapstructure:"cleanup_refresh_rate_hours"`
	CleanupKeepAssetsDays   int `mapstructure:"cleanup_keep_assets_days"`
}

func (sc *ServerConfig) SetDefaults() {
	sc.ServicePort = 8080
	sc.RetryMaxIntervalSec = 10
	sc.RetryMaxElapsedSec = 60
	sc.DaysLimitForFirstSync = 7
	sc.CleanupKeepAssetsDays = 30
	sc.CleanupRefreshRateHours = 24
}

type ServicesConfig struct {
	ClusterServiceUrl string `mapstructure:"cluster_service_url"`
	OrgUnitUrl        string `mapstructure:"org_unit_service_url"`
	AuthorizationUrl  string `mapstructure:"auth_service_url"`
}

func (sc *ServicesConfig) SetDefaults() {
	sc.ClusterServiceUrl = "http://runai-backend-cluster-service:8080"
	sc.OrgUnitUrl = "http://runai-backend-org-unit-service:8080"
	sc.AuthorizationUrl = "http://runai-backend-authorization:8080"
}

type ClusterCache struct {
	CacheConfig `mapstructure:",squash"`
}

func (cc *ClusterCache) SetDefaults() {
	cc.ExpirationSec = 600
	cc.CleanupSec = 1200
}

type CacheConfig struct {
	ExpirationSec int `mapstructure:"expiration_sec"`
	CleanupSec    int `mapstructure:"cleanup_sec"`
}

type ProjectsCache struct {
	CacheConfig     `mapstructure:",squash"`
	MaxRecordSizeKB int `mapstructure:"cache_projects_max_record_size_kb"`
}

func (pc *ProjectsCache) SetDefaults() {
	pc.ExpirationSec = 600
	pc.CleanupSec = 1200
	pc.MaxRecordSizeKB = 300
}

type K8sObjectsTrackerConfig struct {
	LogLevel             string `mapstructure:"log_level"`
	KeyDelimiter         string `mapstructure:"key_delimiter"`
	K8sClientConfigQPS   int    `mapstructure:"k8s_client_config_qps"`
	K8sClientConfigBurst int    `mapstructure:"k8s_client_config_burst"`
}

func (asc *K8sObjectsTrackerConfig) SetDefaults() {
	asc.LogLevel = "info"
	asc.KeyDelimiter = "_"
	asc.K8sClientConfigQPS = 600
	asc.K8sClientConfigBurst = 3000
}

func (sc *ServicesConfig) AuthorizationClientConfig() *authorizationClient.Config {
	return &authorizationClient.Config{Url: sc.AuthorizationUrl}
}

func registerConfigs() {
	if err := runai_config.Register[*ServerConfig](); err != nil {
		logging.Fatalf("Failed to register server config: %v", err)
	}

	if err := runai_config.Register[*ServicesConfig](); err != nil {
		logging.Fatalf("Failed to register services config: %v", err)
	}

	if err := runai_config.RegisterRoot[*ClusterCache](); err != nil {
		logging.Fatalf("Failed to initialize clusters cache config: %v", err)
	}

	if err := runai_config.RegisterRoot[*ProjectsCache](); err != nil {
		logging.Fatalf("Failed to initialize projects cache config: %v", err)
	}

	if err := runai_config.Register[*K8sObjectsTrackerConfig](); err != nil {
		logging.Fatalf("Failed to register k8s-objects-tracker config: %v", err)
	}
}

func Init() {
	once.Do(func() {
		logging.Debugf("Initialising environment variables")
		registerConfigs()
	})
}

func MockConfig() {
	// Point the secrets folder at an existing dir (no secret files) so config.Init's
	// SetKbSecretsFolder doesn't fail on the missing default /etc/secrets in tests.
	os.Setenv("SECRETS_MOUNT_PATH", ".")
	// PubSub toggle for tests (runaipubsub reads PUBSUB_ENABLED)
	os.Setenv("PUBSUB_ENABLED", "false")

	// Log level
	os.Setenv("LOG_LEVEL", "info")
}
