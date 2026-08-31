/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"github.com/run-ai/runai/control-plane/datavolumes-service/internal/generated/pkg/api"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting/manager"
)

// EnableDataVolumes represents the setting that controls whether data volumes creation is allowed.
type EnableDataVolumes struct {
	Enabled bool `json:"enabled"`
}

func (e *EnableDataVolumes) Category() string {
	return DatavolumesCategory
}

func (e *EnableDataVolumes) Description() string {
	return "Enable to allow creation of data volumes to share underlying data across scopes. For more information, see the Data volumes guide"
}

func (e *EnableDataVolumes) DisplayName() string {
	return "Data volumes"
}

func (e *EnableDataVolumes) Key() string {
	return string(api.DATAVOLUMESSETTINGSKEYENUM_ENABLE_DATA_VOLUMES)
}

func (e *EnableDataVolumes) SetDefault() {
	e.Enabled = true
}

func (e *EnableDataVolumes) Stage() setting_pkg.Stage {
	return setting_pkg.GA
}

// Validate is required by the Settinger interface; no custom validation needed for this setting.
func (e *EnableDataVolumes) Validate(value interface{}) bool {
	return true
}

func (e *EnableDataVolumes) TransformForMigration(value interface{}) interface{} {
	settingValue, ok := value.(bool)
	if !ok {
		return value
	}
	return EnableDataVolumes{
		Enabled: settingValue,
	}
}
