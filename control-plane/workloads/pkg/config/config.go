/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"time"

	notifyclient "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

type ServerConfig struct {
	ViewRefreshRateMilliseconds int  `mapstructure:"view_refresh_rate_milli_seconds"`
	UsePartialMaterializedView  bool `mapstructure:"use_partial_materialized_view"`
}

func (c *ServerConfig) SetDefaults() {
	c.ViewRefreshRateMilliseconds = 0
	c.UsePartialMaterializedView = true
}

type DBCleanerConfig struct {
	CleanerIntervalHours           int `mapstructure:"cleaner_interval_hours"`
	DeletedWorkloadsDaysToLive     int `mapstructure:"deleted_workloads_days_to_live"`
	FinishedWorkloadsDaysToLive    int `mapstructure:"finished_workloads_days_to_live"`
	CreatingWorkloadsMinutesToLive int `mapstructure:"creating_workloads_minutes_to_live"`
}

func (c *DBCleanerConfig) SetDefaults() {
	c.CleanerIntervalHours = 24
	c.DeletedWorkloadsDaysToLive = 14
	c.FinishedWorkloadsDaysToLive = 28
	c.CreatingWorkloadsMinutesToLive = 60
}

type Notify struct {
	BatchSize                 int    `mapstructure:"batch_size"`
	BatchIntervalSec          int    `mapstructure:"batch_interval_sec"`
	NotifyUrl                 string `mapstructure:"notify_url"`
	ConfigUrl                 string `mapstructure:"config_url"`
	PhaseChangeIntervalMillis int    `mapstructure:"phase_change_interval_millis"`
	TimeoutIntervalMillis     int    `mapstructure:"timeout_interval_millis"`
}

func (c *Notify) SetDefaults() {
	c.BatchSize = notifyclient.DEFAULT_NOTIFY_BATCH_SIZE
	c.BatchIntervalSec = notifyclient.DEFAULT_NOTIFY_BATCH_INTERVAL_SEC
	c.NotifyUrl = "http://runai-backend-notifications-service:5000"
	c.ConfigUrl = "http://runai-backend-notifications-service:8080"
	c.PhaseChangeIntervalMillis = 300000
	c.TimeoutIntervalMillis = 300000
}

type WorkloadsSyncConfig struct {
	BulkSize int `mapstructure:"workloads_sync_bulk_size"`
}

func (c *WorkloadsSyncConfig) SetDefaults() {
	c.BulkSize = 50
}

type ApplicationsSyncConfig struct {
	BulkSize int `mapstructure:"applications_sync_bulk_size"`
}

func (c *ApplicationsSyncConfig) SetDefaults() {
	c.BulkSize = 50
}

type WorkloadsBatchConfig struct {
	BatchSize int `mapstructure:"workloads_batch_size"`
}

func (c *WorkloadsBatchConfig) SetDefaults() {
	c.BatchSize = 50
}

type WatchDog struct {
	WakeupIntervalMin   int    `mapstructure:"wakeup_interval_min"`
	MaxCreationDuration string `mapstructure:"max_creation_duration"`
}

func (c *WatchDog) SetDefaults() {
	c.WakeupIntervalMin = 1
	c.MaxCreationDuration = "3 minutes"
}

type MetricsExporter struct {
	IntervalSeconds     int   `mapstructure:"interval_seconds"`
	MultiTenancyEnabled *bool `mapstructure:"multi_tenancy_enabled"`
}

func (c *MetricsExporter) SetDefaults() {
	c.IntervalSeconds = 30
}

type MetricsWriter struct {
	IntervalMinutes              int `mapstructure:"interval_minutes"`
	MinimumTenantIntervalSeconds int `mapstructure:"min_tenant_interval_seconds"`
	MaximumTenantsInChunk        int `mapstructure:"max_tenants_in_chunk"`
}

func (c *MetricsWriter) SetDefaults() {
	c.IntervalMinutes = 20
	c.MinimumTenantIntervalSeconds = 5
	c.MaximumTenantsInChunk = 2
}

type ServicesConfig struct {
	LogLevel                     string                      `mapstructure:"log_level"`
	AuthorizationUrl             string                      `mapstructure:"authorization_url"`
	AssetsServiceUrl             string                      `mapstructure:"assets_service_url"`
	PolicyServiceUrl             string                      `mapstructure:"policy_service_url"`
	ObjectsTrackerUrl            string                      `mapstructure:"objects_tracker_url"`
	ClusterServiceUrl            string                      `mapstructure:"cluster_service_url"`
	ClustersCache                checker.ClustersCacheConfig `mapstructure:"clusters_cache"`
	TempBackendSettingsUrl       string                      `mapstructure:"temp_backend_settings_url"`
	OrgUnitUrl                   string                      `mapstructure:"org_unit_url"`
	DataVolumeUrl                string                      `mapstructure:"data_volume_url"`
	TenantsManagerUrl            string                      `mapstructure:"tenants_manager_url"`
	WorkloadsManagerServiceUrl   string                      `mapstructure:"workloads_manager_service_url"`
	AnyWorkloadSubmissionEnabled bool                        `mapstructure:"any_workload_submission"`

	AnyWorkloadWorkspaceSubmission   bool `mapstructure:"any_workload_workspace_submission"`
	AnyWorkloadTrainingSubmission    bool `mapstructure:"any_workload_training_submission"`
	AnyWorkloadDistributedSubmission bool `mapstructure:"any_workload_distributed_submission"`
}

func (c *ServicesConfig) SetDefaults() {
	c.LogLevel = "info"
	c.AuthorizationUrl = "http://runai-backend-authorization:8080"
	c.AssetsServiceUrl = "http://runai-backend-assets-service:8080"
	c.PolicyServiceUrl = "http://runai-backend-policy-service:8080"
	c.ObjectsTrackerUrl = "http://runai-backend-k8s-objects-tracker:8080"
	c.TempBackendSettingsUrl = "http://runai-backend-temp-backend-settings:8080"
	c.OrgUnitUrl = "http://runai-backend-org-unit-service:8080"
	c.DataVolumeUrl = "http://runai-backend-datavolumes:8080"
	c.TenantsManagerUrl = "http://runai-backend-tenants-manager:8080"
	c.ClusterServiceUrl = "http://runai-backend-cluster-service:8080"
	c.ClustersCache.ExpirationSec = 600
	c.ClustersCache.CleanupSec = 1200
	c.WorkloadsManagerServiceUrl = "http://runai-backend-workloads-manager:8080"
	c.AnyWorkloadSubmissionEnabled = false
	c.AnyWorkloadWorkspaceSubmission = false
	c.AnyWorkloadTrainingSubmission = false
	c.AnyWorkloadDistributedSubmission = false
}

// ChartUploadConfig governs the AI application local chart upload feature: how uploaded charts are
// served to the cluster (publicly reachable base URL), the upload size limit, and how long an
// unclaimed chart is retained before cleanup.
type ChartUploadConfig struct {
	PublicBaseURL             string `mapstructure:"chart_public_base_url"`
	MaxUploadBytes            int64  `mapstructure:"chart_max_upload_bytes"`
	RetentionMinutes          int    `mapstructure:"chart_retention_minutes"`
	InspectionCacheTTLSeconds int    `mapstructure:"chart_inspection_cache_ttl_seconds"`
}

func (c *ChartUploadConfig) SetDefaults() {
	c.MaxUploadBytes = 10 * 1024 * 1024
	c.RetentionMinutes = 24 * 60
	c.InspectionCacheTTLSeconds = 600
}

type CacheConfig struct {
	ExpirationSec int `mapstructure:"expiration_sec"`
	CleanupSec    int `mapstructure:"cleanup_sec"`
}

func (cc *CacheConfig) SetDefaults() {
	cc.ExpirationSec = 600
	cc.CleanupSec = 1200
}

// Expiration returns the cache entry TTL.
func (cc *CacheConfig) Expiration() time.Duration {
	return time.Duration(cc.ExpirationSec) * time.Second
}

// OrgUnitCache configures the TTL cache used to resolve org-unit (department) names.
type OrgUnitCache struct {
	CacheConfig `mapstructure:",squash"`
}

func registerConfigs() {
	if err := runai_config.Register[*ServerConfig](); err != nil {
		logging.Fatalf("Failed to register server config: %v", err)
	}
	if err := runai_config.Register[*DBCleanerConfig](); err != nil {
		logging.Fatalf("Failed to register db cleaner config: %v", err)
	}
	if err := runai_config.RegisterRoot[*Notify](); err != nil {
		logging.Fatalf("Failed to register notify client config: %v", err)
	}
	if err := runai_config.Register[*WorkloadsSyncConfig](); err != nil {
		logging.Fatalf("Failed to register workloads sync config: %v", err)
	}
	if err := runai_config.Register[*ApplicationsSyncConfig](); err != nil {
		logging.Fatalf("Failed to register applications sync config: %v", err)
	}
	if err := runai_config.Register[*WorkloadsBatchConfig](); err != nil {
		logging.Fatalf("Failed to register workloads batch config: %v", err)
	}
	if err := runai_config.RegisterRoot[*WatchDog](); err != nil {
		logging.Fatalf("Failed to register watchdog config: %v", err)
	}
	if err := runai_config.RegisterRoot[*MetricsWriter](); err != nil {
		logging.Fatalf("Failed to register metrics writer config: %v", err)
	}
	if err := runai_config.RegisterRoot[*MetricsExporter](); err != nil {
		logging.Fatalf("Failed to register metrics exporter config: %v", err)
	}
	if err := runai_config.Register[*ServicesConfig](); err != nil {
		logging.Fatalf("Failed to register services config: %v", err)
	}
	if err := runai_config.Register[*ChartUploadConfig](); err != nil {
		logging.Fatalf("Failed to register chart upload config: %v", err)
	}
	if err := runai_config.RegisterRoot[*OrgUnitCache](); err != nil {
		logging.Fatalf("Failed to register org unit cache config: %v", err)
	}
}

// Init initializes configuration
func Init() {
	registerConfigs()
}
