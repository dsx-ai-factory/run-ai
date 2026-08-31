/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting"
	"github.com/run-ai/runai/runai-common-packages/setting/manager"
)

const (
	ErrUnknownSettingKey = "unknown setting key"
	ErrValidationFailed  = "validation error"
	WorkloadsCategory    = "workloads"
)

type SettingsManagerInterface interface {
	GetAllSettings(tenantID string) ([]manager.Settinger, error)
	GetSettingByKey(key api.WorkloadsSettingsKeyEnum, tenantID string) (manager.Settinger, error)
	UpdateSettingByKey(ctx context.Context, key api.WorkloadsSettingsKeyEnum, tenantID string, updateRequest api.UpdateWorkloadsSettingByKeyRequest) error
	TransformSettingsToResponse(settingsList []manager.Settinger) *api.WorkloadsSettingsGetResponse
	TransformSettingByKeyToResponse(setting manager.Settinger) (interface{}, error)
}

// SettingsManager handles retrieval, update, and transformation of workloads settings.
type SettingsManager struct {
}

// NewSettingsManager creates a new SettingsManager.
func NewSettingsManager() *SettingsManager {
	return &SettingsManager{}
}

func (s *SettingsManager) GetAllSettings(tenantID string) ([]manager.Settinger, error) {
	return setting_pkg.GetAll(tenantID, WorkloadsCategory)
}

func (s *SettingsManager) GetSettingByKey(key api.WorkloadsSettingsKeyEnum, tenantID string) (manager.Settinger, error) {
	switch key {
	case api.WORKLOADSSETTINGSKEYENUM_ENABLE_POLICIES:
		return setting_pkg.Get[*EnablePolicies](tenantID)
	case api.WORKLOADSSETTINGSKEYENUM_WORKLOAD_SUBMISSION_V2_ENABLED:
		return setting_pkg.Get[*WorkloadSubmissionV2Enabled](tenantID)
	case api.WORKLOADSSETTINGSKEYENUM_ENABLE_TEMPLATES_V2:
		return setting_pkg.Get[*EnableTemplatesV2](tenantID)
	case api.WORKLOADSSETTINGSKEYENUM_ENABLE_HYBRID_SUBMISSION:
		return setting_pkg.Get[*EnableHybridSubmission](tenantID)
	case api.WORKLOADSSETTINGSKEYENUM_GPU_OVER_PROVISIONING:
		return setting_pkg.Get[*GpuOverProvisioning](tenantID)
	case api.WORKLOADSSETTINGSKEYENUM_APPLICATION_SUBMISSION:
		return setting_pkg.Get[*ApplicationSubmission](tenantID)
	default:
		return nil, fmt.Errorf("%s: %s", ErrUnknownSettingKey, key)
	}
}

func (s *SettingsManager) UpdateSettingByKey(ctx context.Context, key api.WorkloadsSettingsKeyEnum, tenantID string, updateRequest api.UpdateWorkloadsSettingByKeyRequest) error {
	if updateRequest.Enabled == nil {
		return fmt.Errorf("%s: enabled field is required", ErrValidationFailed)
	}
	enabled := *updateRequest.Enabled

	switch key {
	case api.WORKLOADSSETTINGSKEYENUM_ENABLE_POLICIES:
		return setting_pkg.Update[*EnablePolicies](tenantID, EnablePolicies{Enabled: enabled}).UpdatedByUser().Execute()
	case api.WORKLOADSSETTINGSKEYENUM_WORKLOAD_SUBMISSION_V2_ENABLED:
		return setting_pkg.Update[*WorkloadSubmissionV2Enabled](tenantID, WorkloadSubmissionV2Enabled{Enabled: enabled}).UpdatedByUser().Execute()
	case api.WORKLOADSSETTINGSKEYENUM_ENABLE_TEMPLATES_V2:
		if enabled {
			hybridSubmission, err := setting_pkg.Get[*EnableHybridSubmission](tenantID)
			if err != nil {
				return fmt.Errorf("failed to check enableHybridSubmission state: %w", err)
			}
			if !hybridSubmission.Enabled {
				return fmt.Errorf("%s: cannot enable enableTemplatesV2 while enableHybridSubmission is disabled", ErrValidationFailed)
			}
		}
		return setting_pkg.Update[*EnableTemplatesV2](tenantID, EnableTemplatesV2{Enabled: enabled}).UpdatedByUser().Execute()
	case api.WORKLOADSSETTINGSKEYENUM_ENABLE_HYBRID_SUBMISSION:
		if err := setting_pkg.Update[*EnableHybridSubmission](tenantID, EnableHybridSubmission{Enabled: enabled}).UpdatedByUser().Execute(); err != nil {
			return err
		}
		if !enabled {
			if err := setting_pkg.Update[*EnableTemplatesV2](tenantID, EnableTemplatesV2{Enabled: false}).UpdatedByUser().Execute(); err != nil {
				return fmt.Errorf("failed to disable enableTemplatesV2 after disabling enableHybridSubmission: %w", err)
			}
		}
		return nil
	case api.WORKLOADSSETTINGSKEYENUM_GPU_OVER_PROVISIONING:
		return setting_pkg.Update[*GpuOverProvisioning](tenantID, GpuOverProvisioning{Enabled: enabled}).UpdatedByUser().Execute()
	case api.WORKLOADSSETTINGSKEYENUM_APPLICATION_SUBMISSION:
		return setting_pkg.Update[*ApplicationSubmission](tenantID, ApplicationSubmission{Enabled: enabled}).UpdatedByUser().Execute()
	default:
		return fmt.Errorf("%s: %s", ErrUnknownSettingKey, key)
	}
}

func (s *SettingsManager) TransformSettingsToResponse(settingsList []manager.Settinger) *api.WorkloadsSettingsGetResponse {
	response := &api.WorkloadsSettingsGetResponse{}

	for _, setting := range settingsList {
		switch v := setting.(type) {
		case *EnablePolicies:
			response.EnablePolicies = api.GetEnablePolicies{
				Value:       api.GetEnableSettingsValue{Enabled: &v.Enabled},
				DisplayName: v.DisplayName(),
				Description: v.Description(),
				Stage:       string(v.Stage()),
			}
		case *WorkloadSubmissionV2Enabled:
			response.WorkloadSubmissionV2Enabled = api.GetWorkloadSubmissionV2Enabled{
				Value:       api.GetEnableSettingsValue{Enabled: &v.Enabled},
				DisplayName: v.DisplayName(),
				Description: v.Description(),
				Stage:       string(v.Stage()),
			}
		case *EnableTemplatesV2:
			response.EnableTemplatesV2 = api.GetEnableTemplatesV2{
				Value:       api.GetEnableSettingsValue{Enabled: &v.Enabled},
				DisplayName: v.DisplayName(),
				Description: v.Description(),
				Stage:       string(v.Stage()),
			}
		case *EnableHybridSubmission:
			response.EnableHybridSubmission = api.GetEnableHybridSubmission{
				Value:       api.GetEnableSettingsValue{Enabled: &v.Enabled},
				DisplayName: v.DisplayName(),
				Description: v.Description(),
				Stage:       string(v.Stage()),
			}
		case *GpuOverProvisioning:
			response.GpuOverProvisioning = api.GetGpuOverProvisioning{
				Value:       api.GetEnableSettingsValue{Enabled: &v.Enabled},
				DisplayName: v.DisplayName(),
				Description: v.Description(),
				Stage:       string(v.Stage()),
			}
		case *ApplicationSubmission:
			response.ApplicationSubmission = api.GetApplicationSubmission{
				Value:       api.GetEnableSettingsValue{Enabled: &v.Enabled},
				DisplayName: v.DisplayName(),
				Description: v.Description(),
				Stage:       string(v.Stage()),
			}
		}
	}

	return response
}

func (s *SettingsManager) TransformSettingByKeyToResponse(setting manager.Settinger) (interface{}, error) {
	switch v := setting.(type) {
	case *EnablePolicies:
		return api.GetEnablePolicies{
			Value:       api.GetEnableSettingsValue{Enabled: &v.Enabled},
			DisplayName: v.DisplayName(),
			Description: v.Description(),
			Stage:       string(v.Stage()),
		}, nil
	case *WorkloadSubmissionV2Enabled:
		return api.GetWorkloadSubmissionV2Enabled{
			Value:       api.GetEnableSettingsValue{Enabled: &v.Enabled},
			DisplayName: v.DisplayName(),
			Description: v.Description(),
			Stage:       string(v.Stage()),
		}, nil
	case *EnableTemplatesV2:
		return api.GetEnableTemplatesV2{
			Value:       api.GetEnableSettingsValue{Enabled: &v.Enabled},
			DisplayName: v.DisplayName(),
			Description: v.Description(),
			Stage:       string(v.Stage()),
		}, nil
	case *EnableHybridSubmission:
		return api.GetEnableHybridSubmission{
			Value:       api.GetEnableSettingsValue{Enabled: &v.Enabled},
			DisplayName: v.DisplayName(),
			Description: v.Description(),
			Stage:       string(v.Stage()),
		}, nil
	case *GpuOverProvisioning:
		return api.GetGpuOverProvisioning{
			Value:       api.GetEnableSettingsValue{Enabled: &v.Enabled},
			DisplayName: v.DisplayName(),
			Description: v.Description(),
			Stage:       string(v.Stage()),
		}, nil
	case *ApplicationSubmission:
		return api.GetApplicationSubmission{
			Value:       api.GetEnableSettingsValue{Enabled: &v.Enabled},
			DisplayName: v.DisplayName(),
			Description: v.Description(),
			Stage:       string(v.Stage()),
		}, nil
	default:
		return nil, fmt.Errorf("%s", ErrUnknownSettingKey)
	}
}
