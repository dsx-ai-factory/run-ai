/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/manager"
)

// CPUResourcesQuota controls whether resource quota based on CPU resources is enabled for the tenant's node pools.
type CPUResourcesQuota struct {
	Enabled bool `json:"enabled"`
}

func (e *CPUResourcesQuota) Category() string {
	return NodePoolCategory
}

func (e *CPUResourcesQuota) Description() string {
	return "Enable resource quota based on CPU resources.\nFor more information, see the " +
		"<a href='https://run-ai-docs.nvidia.com/saas/platform-management/aiinitiatives/organization/" +
		"departments' target='_blank'>Departments guide</a> and the <a href='https://run-ai-docs.nvidia.com/" +
		"saas/platform-management/aiinitiatives/organization/projects' target='_blank'>Projects guide</a>"
}

func (e *CPUResourcesQuota) DisplayName() string {
	return "CPU quota"
}

func (e *CPUResourcesQuota) Key() string {
	return string(api.NODEPOOLSETTINGSKEYENUM_CPU_RESOURCES_QUOTA)
}

func (e *CPUResourcesQuota) SetDefault() {
	e.Enabled = false
}

func (e *CPUResourcesQuota) Stage() setting_pkg.Stage {
	return setting_pkg.GA
}

func (e *CPUResourcesQuota) Validate(value interface{}) bool {
	return true
}

func (e *CPUResourcesQuota) TransformForMigration(value interface{}) interface{} {
	settingValue, ok := value.(bool)
	if !ok {
		return value
	}
	return CPUResourcesQuota{Enabled: settingValue}
}
