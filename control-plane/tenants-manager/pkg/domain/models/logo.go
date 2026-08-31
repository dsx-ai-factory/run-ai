/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import "gorm.io/gorm"

type Logo struct {
	gorm.Model
	TenantId uint    `gorm:"index;unique"`
	LogoPng  *[]byte `gorm:"not null"`
}
