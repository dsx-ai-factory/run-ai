/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models_crud

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	TemplatesPresetsTableName = "crud_templates_presets"

	TemplatesPresetsColumnId                = "id"
	TemplatesPresetsColumnPresetsTemplateId = "preset_template_id"
	TemplatesPresetsColumnTenantId          = "tenant_id"
	TemplatesPresetsColumnIsDeleted         = "is_deleted"
	TemplatesPresetsColumnTenantTemplateId  = "tenant_template_id"
)

// TemplatesPresets holds information about system templates that tenants decided to diverge from
// if the tenant deleted the template, it will be marked with is_deleted
// if the tenant decided to fork the template, the forked template id is stored in this table.
type TemplatesPresets struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	PresetTemplateId string
	TenantId         uint
	IsDeleted        bool
	TenantTemplateId *string
}

func (TemplatesPresets) TableName() string {
	return TemplatesPresetsTableName
}

func (t TemplatesPresets) DebugInfo() string {
	if t.TenantTemplateId == nil {
		return fmt.Sprintf("%s -> %d [deleted]", t.PresetTemplateId, t.TenantId)
	} else {
		return fmt.Sprintf("%s -> %d [%s]", t.PresetTemplateId, t.TenantId, *t.TenantTemplateId)
	}
}
