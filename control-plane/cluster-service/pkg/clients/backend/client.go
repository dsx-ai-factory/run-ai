/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"context"
	"errors"
	"fmt"

	metrics_client "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
	metrics_api "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	temp_backend_settings "github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/client"
)

type TenantSettingsClient interface {
	GetOPRSetting(ctx context.Context) (bool, error)
	GetTenantMetricStore(ctx context.Context, tenantId int32) (*metrics_api.MetricStore, error)
}

type DefaultTenantSettingsClient struct {
	settingsClient    temp_backend_settings.ITempBackendSettingsClient
	metricStoreClient metrics_client.MetricStoresClienter
}

func NewTenantSettingsClient(tempBackendSettingsUrl string) TenantSettingsClient {
	return &DefaultTenantSettingsClient{
		settingsClient:    temp_backend_settings.NewTempBackendSettingsClient(tempBackendSettingsUrl),
		metricStoreClient: metrics_client.GetMetricStoresClient(),
	}
}

func (client *DefaultTenantSettingsClient) GetOPRSetting(ctx context.Context) (bool, error) {
	return client.getSetting(ctx, "tenant.gpu_over_provisioning")
}

func (client *DefaultTenantSettingsClient) getSetting(ctx context.Context, settingKey string) (bool, error) {
	enabled, err := client.settingsClient.GetTenantSettingsByKey(ctx, settingKey)
	if err != nil {
		return false, fmt.Errorf("failed to get setting %s: %w", settingKey, err)
	}
	return enabled, nil
}

func (client *DefaultTenantSettingsClient) GetTenantMetricStore(ctx context.Context, tenantId int32) (*metrics_api.MetricStore, error) {
	metricStore, err := client.metricStoreClient.GetMetricStore(ctx, tenantId)
	if err != nil {
		var metricsApiErr *metrics_api.Error
		if errors.As(err, &metricsApiErr) {
			return nil, fmt.Errorf("metrics API error (code %d): %s", metricsApiErr.Code, metricsApiErr.Message)
		}
		return nil, fmt.Errorf("failed to get tenant metric store: %w", err)
	}

	return metricStore, nil
}
