/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
)

// TenantSettingsType* constants define the possible types for a tenant setting value.
const (
	TenantSettingsTypeBoolean = "Boolean"
	TenantSettingsTypeNumber  = "Number"
	TenantSettingsTypeString  = "String"
)

// KeyPresetsInstalled is the tenant_settings key used to track whether preset
// assets have been cloned to a tenant and at which revision.
const KeyPresetsInstalled = "workspaces.preset_installed"

// TenantSettingsRec models a row in the tenant_settings table (assets schema).
// This is a key-value store per tenant, used by the assets service to persist
// configuration flags such as the preset cloning revision.
type TenantSettingsRec struct {
	TenantId int32
	Key      string
	Value    string
	Type     string
}

// TableName returns the unqualified table name within the assets schema.
func (TenantSettingsRec) TableName() string {
	return dbtables.TableNameTenantSettings
}
