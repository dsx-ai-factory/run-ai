/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/manager"
)

// LimitQuotaOverSubscription controls whether projects are prevented from exceeding their department's quota.
type LimitQuotaOverSubscription struct {
	Enabled bool `json:"enabled"`
}

func (e *LimitQuotaOverSubscription) Category() string {
	return OrgUnitCategory
}

func (e *LimitQuotaOverSubscription) Description() string {
	return "Enable to prevent projects' total quota from exceeding the department's quota or being less than the total usage by non-preemptible workloads." +
		" For more information, see the <a href='https://run-ai-docs.nvidia.com/saas/platform-management/runai-scheduler/scheduling/" +
		"concepts-and-principles#over-subscription' target='_blank'>Run:ai scheduler guide</a>"
}

func (e *LimitQuotaOverSubscription) DisplayName() string {
	return "Limit projects from exceeding department quota"
}

func (e *LimitQuotaOverSubscription) Key() string {
	return string(api.ORGUNITSETTINGSKEYENUM_LIMIT_QUOTA_OVER_SUBSCRIPTION)
}

func (e *LimitQuotaOverSubscription) SetDefault() {
	e.Enabled = false
}

func (e *LimitQuotaOverSubscription) Stage() setting_pkg.Stage {
	return setting_pkg.GA
}

func (e *LimitQuotaOverSubscription) Validate(value interface{}) bool {
	return true
}

func (e *LimitQuotaOverSubscription) TransformForMigration(value interface{}) interface{} {
	settingValue, ok := value.(bool)
	if !ok {
		return value
	}
	return LimitQuotaOverSubscription{Enabled: settingValue}
}
