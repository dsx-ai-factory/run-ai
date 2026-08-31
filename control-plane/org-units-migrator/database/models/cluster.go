/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"fmt"

	"gorm.io/gorm"
)

const ClustersSchema = "cluster_service"

type Cluster struct {
	Id        string `gorm:"primaryKey"`
	TenantId  int32
	DeletedAt gorm.DeletedAt
	Version   string
}

func (Cluster) TableName() string {
	return fmt.Sprintf("%s.clusters", ClustersSchema)
}
