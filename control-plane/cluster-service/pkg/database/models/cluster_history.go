/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

const (
	ClusterHistoryTable = "cluster_histories"
)

type ClusterHistory struct {
	gorm.Model
	TenantId        int32     `gorm:"not null"`
	ClusterID       uuid.UUID `gorm:"primaryKey"`
	Version         string    `gorm:"primaryKey"`
	UpdateTimestamp time.Time
	Cluster         Cluster `gorm:"foreignKey:ClusterID"`
}

func (ClusterHistory) TableName() string {
	return ClusterHistoryTable
}

func (c ClusterHistory) MarshalBinary() ([]byte, error) {
	return json.Marshal(ClusterHistory{
		TenantId:        c.TenantId,
		ClusterID:       c.ClusterID,
		Version:         c.Version,
		UpdateTimestamp: c.UpdatedAt,
	})
}

type ClusterHistories []ClusterHistory

func (c ClusterHistories) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}
