/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

type NodepoolsMigrationControl struct {
	Version  string `gorm:"uniqueIndex:np_v_tenant"`
	TenantId *int32 `gorm:"uniqueIndex:np_v_tenant"`
	Status   MigrationStatus
}

func (NodepoolsMigrationControl) TableName() string {
	return "nodepools_migration_control"
}
