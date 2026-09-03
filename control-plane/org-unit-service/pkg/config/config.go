/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

var Schema = "org_unit_service"

func Init() error {
	logging.Debugf("initializing config")
	if err := registerConfigs(); err != nil {
		return err
	}

	return nil
}

func registerConfigs() error {
	if err := config.Register[*General](); err != nil {
		return fmt.Errorf("failed to register general configuration: %v", err)
	}

	if err := config.Register[*Server](); err != nil {
		return fmt.Errorf("failed to register server configuration: %v", err)
	}

	if err := config.Register[*OvertimeMetrics](); err != nil {
		return fmt.Errorf("failed to register overtime metrics configuration: %v", err)
	}

	if err := config.Register[*Reports](); err != nil {
		return fmt.Errorf("failed to register reports configuration: %v", err)
	}

	if err := config.Register[*SyncJob](); err != nil {
		return fmt.Errorf("failed to register sync job configuration: %v", err)
	}

	if err := config.Register[*DBCleanerConfig](); err != nil {
		return fmt.Errorf("failed to register db cleaner configuration: %v", err)
	}

	return nil
}

type General struct {
	LogLevel string `mapstructure:"log_level"`
}

func (c *General) SetDefaults() {
	c.LogLevel = "info"
}

type Reports struct {
	BucketName string `mapstructure:"s3_bucket"`
}

func (c *Reports) SetDefaults() {
}

type Server struct {
	ServicePort            int    `mapstructure:"service_port"`
	TempBackendSettingsUrl string `mapstructure:"temp_backend_settings_url"`
	TenantsManagerUrl      string `mapstructure:"tenants_manager_url"`
	ClustersServiceUrl     string `mapstructure:"cluster_service_url"`
	WorkloadsServiceUrl    string `mapstructure:"workloads_service_url"`
	DatavolumesServiceUrl  string `mapstructure:"datavolumes_service_url"`
}

func (c *Server) SetDefaults() {
	c.ServicePort = 8080
	c.TempBackendSettingsUrl = "http://runai-backend-temp-backend-settings:8080"
	c.TenantsManagerUrl = "http://runai-backend-tenants-manager:8080"
	c.ClustersServiceUrl = "http://runai-backend-cluster-service:8080"
	c.WorkloadsServiceUrl = "http://runai-backend-workloads:8080"
	c.DatavolumesServiceUrl = "http://runai-backend-datavolumes:8080"
}

type OvertimeMetrics struct {
	Query24hRangeIntervalSeconds     int `mapstructure:"query_24h_range_interval_seconds"`
	Query7dRangeIntervalSeconds      int `mapstructure:"query_7d_range_interval_seconds"`
	Query30dRangeIntervalSeconds     int `mapstructure:"query_30d_range_interval_seconds"`
	Query24hRangeInitialSleepSeconds int `mapstructure:"query_24h_range_initial_sleep_seconds"`
	Query7dRangeInitialSleepSeconds  int `mapstructure:"query_7d_range_initial_sleep_seconds"`
	Query30dRangeInitialSleepSeconds int `mapstructure:"query_30d_range_initial_sleep_seconds"`

	RateLimiterLimit int `mapstructure:"rate_limiter_limit"`
	RateLimiterBurst int `mapstructure:"rate_limiter_burst"`
}

func (c *OvertimeMetrics) SetDefaults() {
	c.Query24hRangeIntervalSeconds = 60 * 60      // 1 hour
	c.Query7dRangeIntervalSeconds = 60 * 60 * 6   // 6 hours
	c.Query30dRangeIntervalSeconds = 60 * 60 * 24 // 24 hours
	c.Query24hRangeInitialSleepSeconds = 10
	c.Query7dRangeInitialSleepSeconds = 21
	c.Query30dRangeInitialSleepSeconds = 33
	c.RateLimiterLimit = 1
	c.RateLimiterBurst = 10
}

type SyncJob struct {
	SyncJobIntervalHours         int `mapstructure:"sync_job_interval_hours"`
	SyncJobTenantIntervalSeconds int `mapstructure:"sync_job_interval_tenant_seconds"`
}

func (c *SyncJob) SetDefaults() {
	c.SyncJobIntervalHours = 24
	c.SyncJobTenantIntervalSeconds = 5
}

type DBCleanerConfig struct {
	CleanerEnabled        bool `mapstructure:"cleaner_enabled"`
	CleanerIntervalHours  int  `mapstructure:"cleaner_interval_hours"`
	SoftDeletedDaysToLive int  `mapstructure:"soft_deleted_days_to_live"`
}

func (c *DBCleanerConfig) SetDefaults() {
	c.CleanerEnabled = true
	c.CleanerIntervalHours = 24
	c.SoftDeletedDaysToLive = 30
}
