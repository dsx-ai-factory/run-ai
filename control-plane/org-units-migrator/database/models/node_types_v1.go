/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import "github.com/google/uuid"

type NodeType struct {
	ID          int       `gorm:"column:id;primaryKey"`
	TenantID    uint      `gorm:"column:tenant_id"`
	ClusterUUID uuid.UUID `gorm:"column:cluster_uuid"`
	Name        string    `gorm:"column:name"`
}

func (NodeType) TableName() string {
	return "node_types"
}
