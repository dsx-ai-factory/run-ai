/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

type MigrationControl struct {
	Version  string `gorm:"uniqueIndex:v_tenant"`
	TenantId *int32 `gorm:"uniqueIndex:v_tenant"`
	Status   MigrationStatus
}

func (MigrationControl) TableName() string {
	return "org_units_migration_control"
}
