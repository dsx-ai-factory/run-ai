/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting/manager"
)

// TenantIsCspManaged indicates whether the tenant is managed by a Cloud Service Provider (e.g., DGXC).
type TenantIsCspManaged struct {
	Enabled bool `json:"enabled"`
}

func (e *TenantIsCspManaged) Category() string {
	return AuthorizationCategory
}

func (e *TenantIsCspManaged) Description() string {
	return "Indicates whether the tenant is managed by a Cloud Service Provider."
}

func (e *TenantIsCspManaged) DisplayName() string {
	return "Tenant is CSP managed"
}

func (e *TenantIsCspManaged) Key() string {
	return string(api.AUTHORIZATIONSETTINGSKEYENUM_TENANT_IS_CSP_MANAGED)
}

func (e *TenantIsCspManaged) SetDefault() {
	e.Enabled = false
}

func (e *TenantIsCspManaged) Stage() setting_pkg.Stage {
	return setting_pkg.Alpha
}

func (e *TenantIsCspManaged) Validate(value interface{}) bool {
	return true
}

func (e *TenantIsCspManaged) TransformForMigration(value interface{}) interface{} {
	settingValue, ok := value.(bool)
	if !ok {
		return value
	}
	return TenantIsCspManaged{
		Enabled: settingValue,
	}
}
