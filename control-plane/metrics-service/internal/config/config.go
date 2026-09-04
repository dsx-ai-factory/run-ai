/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/clusters"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/datasources"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/metrics"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/metricstores"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/providers"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/reports"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/tenants"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
)

func Init(configPath string, secretsFolderPath string) error {
	if err := config.SetConfigFiles(configPath); err != nil {
		return err
	}

	if err := config.SetKbSecretsFolder(secretsFolderPath); err != nil {
		return err
	}

	if err := registerConfigs(); err != nil {
		return err
	}

	return nil
}

type ServiceConfiguration struct {
	LogLevel string `mapstructure:"log_level"`
	UseMocks bool   `mapstructure:"use_mocks"`
}

func (c *ServiceConfiguration) SetDefaults() {
	c.LogLevel = "info"
}

func registerConfigs() error {
	if err := config.Register[*ServiceConfiguration](); err != nil {
		return fmt.Errorf("failed to register service configuration: %v", err)
	}

	if err := config.Register[*metrics.MetricsConfiguration](); err != nil {
		return fmt.Errorf("failed to register metrics configuration: %v", err)
	}

	if err := config.RegisterRoot[*reports.ReportsConfiguration](); err != nil {
		return fmt.Errorf("failed to register reports configuration: %v", err)
	}

	if err := config.RegisterRoot[*clusters.ClustersConfiguration](); err != nil {
		return fmt.Errorf("failed to register clusters configuration: %v", err)
	}

	if err := config.RegisterRoot[*tenants.TenantsConfiguration](); err != nil {
		return fmt.Errorf("failed to register tenants configuration: %v", err)
	}

	if err := config.RegisterRoot[*metricstores.MetricStoresConfiguration](); err != nil {
		return fmt.Errorf("failed to register metricstores configuration: %v", err)
	}

	if err := config.Register[*metricstores.DatasourceIdConfig](); err != nil {
		return fmt.Errorf("failed to register datasource id configuration: %v", err)
	}

	if err := config.RegisterRoot[*datasources.DatasourcesConfiguration](); err != nil {
		return fmt.Errorf("failed to register datasources configuration: %v", err)
	}

	if err := config.RegisterRoot[*database.CleanerConfiguration](); err != nil {
		return fmt.Errorf("failed to register cleaner configuration: %v", err)
	}

	if err := config.RegisterRoot[*database.CronJobsConfiguration](); err != nil {
		return fmt.Errorf("failed to register db cron jobs configuration: %v", err)
	}

	if err := config.RegisterRoot[*providers.DatasourceProviderConfig](); err != nil {
		return fmt.Errorf("failed to register datasource provider configuration: %v", err)
	}

	return nil
}
