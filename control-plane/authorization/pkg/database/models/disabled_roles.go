/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

type DisabledRole struct {
	RoleId    int64     `gorm:""`
	ScopeType RoleScope `gorm:""`
	ScopeId   string    `gorm:""`
}
