/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"time"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm/clause"
)

type BackendClusterRow struct {
	ClusterUUID  uuid.UUID
	ClusterName  string
	TenantId     int32
	Domain       string
	Version      string
	LastLiveness *int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Cluster struct {
	ID           uuid.UUID `gorm:"primaryKey"`
	Name         string
	TenantId     int32
	Domain       string
	Version      string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
	LastLiveness *time.Time
}

func CreateV2ClusterModel(db orm.DB, row BackendClusterRow) {
	clusterSpec := Cluster{
		ID:           row.ClusterUUID,
		Name:         row.ClusterName,
		TenantId:     row.TenantId,
		Domain:       row.Domain,
		Version:      row.Version,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		LastLiveness: convertBigintToTs(row.LastLiveness),
	}

	_, err := db.
		Clauses(clause.OnConflict{
			DoNothing: true,
		}).
		Create(&clusterSpec)
	if err != nil {
		logging.Errorf("Failed to create cluster model. err: %+v", err)
		panic(err)
	}
}

func convertBigintToTs(liveness *int64) *time.Time {
	if liveness == nil {
		return nil
	}

	newLiveness := time.UnixMilli(*liveness)
	return &newLiveness
}
