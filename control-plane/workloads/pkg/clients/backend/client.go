/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"context"

	temp_backend_settings_client "github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/client"
)

const (
	// TenantSettingGpuResourceOptimization for over-provisioning may look like a typo, but it's not.
	// The RND term is gpu_over_provisioning, which explains the value of the key, but in the UI it is
	// referred to as "GPU Resource Optimization" because Product team requires Oxford dictionary approved names,
	// and turns out that the Oxford definition of "provisioning" is "the act of supplying somebody with
	// something that they need or want". Since "need" and "want" are contradicting terms in the world of
	// Product team, they decided to call it "GPU Resource Optimization", hence the name of the constant.
	TenantSettingGpuResourceOptimization = "tenant.gpu_over_provisioning"

	// TenantSettingHybridSubmission indicates tenant is working with hybrid submission
	TenantSettingHybridSubmission = "tenant.enable_hybrid_submission"

	// TenantSettingAdvancedGpuMetrics indicates tenant has advanced GPU metrics enabled
	TenantSettingAdvancedGpuMetrics = "tenant.enable_advanced_gpu_metrics"
)

type TenantSettingsClient interface {
	GetAdvancedGpuMetricsSetting(ctx context.Context) (bool, error)

	GetGpuResourceOptimization(ctx context.Context) (bool, error)

	GetHybridSubmissionEnabled(ctx context.Context) (bool, error)
}

type DefaultTenantSettingsClient struct {
	settingsClient temp_backend_settings_client.ITempBackendSettingsClient
}

func NewTenantSettingsClient(baseURL string) TenantSettingsClient {
	return &DefaultTenantSettingsClient{
		settingsClient: temp_backend_settings_client.NewTempBackendSettingsClient(baseURL),
	}
}

func (client *DefaultTenantSettingsClient) GetAdvancedGpuMetricsSetting(ctx context.Context) (bool, error) {
	return client.settingsClient.GetTenantSettingsByKey(ctx, TenantSettingAdvancedGpuMetrics)
}

func (client *DefaultTenantSettingsClient) GetGpuResourceOptimization(ctx context.Context) (bool, error) {
	return client.settingsClient.GetTenantSettingsByKey(ctx, TenantSettingGpuResourceOptimization)
}

func (client *DefaultTenantSettingsClient) GetHybridSubmissionEnabled(ctx context.Context) (bool, error) {
	return client.settingsClient.GetTenantSettingsByKey(ctx, TenantSettingHybridSubmission)
}
