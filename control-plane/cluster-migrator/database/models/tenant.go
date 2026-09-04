/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"fmt"

	"gorm.io/gorm"
)

const TenantsSchema = "tenants_manager"

type Tenant struct {
	Id        int32 `gorm:"primaryKey"`
	DeletedAt gorm.DeletedAt
	Name      string
}

func (Tenant) TableName() string {
	return fmt.Sprintf("%s.tenants", TenantsSchema)
}
