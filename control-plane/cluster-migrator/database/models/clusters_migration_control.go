/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

type ClustersMigrationControl struct {
	Version string `gorm:"primaryKey"`
}

func (ClustersMigrationControl) TableName() string {
	return "migration_control"
}
