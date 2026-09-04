/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	ClusterConfigTable = "cluster_configs"
)

type ClusterSystemConfigData = api.ClusterSystemConfigData

type ClusterUserConfigData = api.ClusterUserConfigData

type ClusterConfigData = api.ClusterFullConfigData

type ClusterConfig struct {
	Common    `gorm:"embedded"`
	DeletedAt gorm.DeletedAt
	ID        uuid.UUID           `gorm:"primaryKey"`
	ClusterID uuid.UUID           `gorm:"not null"`
	TenantID  int32               `gorm:"not null"`
	Config    ClusterConfigData   `gorm:"serializer:json; type:jsonb;not null;"`
	Status    ClusterConfigStatus `gorm:"serializer:json; type:jsonb;not null;"`
}

func (ClusterConfig) TableName() string {
	return ClusterConfigTable
}

func NewClusterConfig(tenantId int32, clusterId string) *ClusterConfig {
	uuidClusterId := uuid.MustParse(clusterId)
	cc := &ClusterConfig{
		ID:        uuidClusterId,
		TenantID:  tenantId,
		ClusterID: uuidClusterId,
	}
	return cc
}
