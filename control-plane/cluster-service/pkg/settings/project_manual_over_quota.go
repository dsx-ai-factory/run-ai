/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/manager"
)

// ProjectManualOverQuota controls whether over-quota weight is enabled for the tenant's node pools.
type ProjectManualOverQuota struct {
	Enabled bool `json:"enabled"`
}

func (e *ProjectManualOverQuota) Category() string {
	return NodePoolCategory
}

func (e *ProjectManualOverQuota) Description() string {
	return "Enable over quota weight to set what proportion of unused resources a project can get on top of its quota." +
		" For more information, see the <a href='https://run-ai-docs.nvidia.com/saas/platform-management/" +
		"runai-scheduler/scheduling/concepts-and-principles#over-quota-weight' target='_blank'>Over quota weight guide</a>"
}

func (e *ProjectManualOverQuota) DisplayName() string {
	return "Over quota weight"
}

func (e *ProjectManualOverQuota) Key() string {
	return string(api.NODEPOOLSETTINGSKEYENUM_PROJECT_MANUAL_OVER_QUOTA)
}

func (e *ProjectManualOverQuota) SetDefault() {
	e.Enabled = false
}

func (e *ProjectManualOverQuota) Stage() setting_pkg.Stage {
	return setting_pkg.GA
}

func (e *ProjectManualOverQuota) Validate(value interface{}) bool {
	return true
}

func (e *ProjectManualOverQuota) TransformForMigration(value interface{}) interface{} {
	settingValue, ok := value.(bool)
	if !ok {
		return value
	}
	return ProjectManualOverQuota{Enabled: settingValue}
}
