/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"gorm.io/gorm"
)

type Cluster struct {
	Id        string `gorm:"primaryKey"`
	TenantId  int32
	DeletedAt gorm.DeletedAt
	Name      string
}

func (Cluster) TableName() string {
	return "clusters"
}
