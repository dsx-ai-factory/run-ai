/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/queryapi"

	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	"gorm.io/gorm"
)

const (
	ClusterTable = "clusters"
)

var ClusterMinimalQueryParser = queryapi.NewParser(clusterMinimalColumnMapping())

type ClusterStatus = api.ClusterReportedStatus

type Cluster struct {
	Common            `gorm:"embedded"`
	ID                uuid.UUID `gorm:"primaryKey"`
	Name              string    `gorm:"not null"`
	TenantId          int32     `gorm:"not null"`
	Domain            string    `gorm:"not null"`
	DeletedAt         gorm.DeletedAt
	LastLiveness      *time.Time
	Status            ClusterStatus `gorm:"serializer:json;type:jsonb"`
	Version           string        `gorm:"not null"`
	DeleteRequestedAt *time.Time
}

var BasicVerbosityClusterFields = []string{"id", "name", "tenant_id", "created_at", "version", "delete_requested_at"}

func (Cluster) TableName() string {
	return ClusterTable
}

func (Cluster) PrimaryKey() string {
	return "id"
}

func (c *Cluster) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return
}

func (c *Cluster) ToMinimalData(clusterState api.ClusterState) api.ClustersMinimalData {
	var version *string
	if c.Version != "" {
		version = convutils.Ref(c.Version)
	}
	minimalData := api.ClustersMinimalData{
		Uuid:               c.ID.String(),
		Name:               c.Name,
		TenantId:           c.TenantId,
		Domain:             convutils.Ref(c.Domain),
		Platform:           c.Status.Platform,
		DependenciesStatus: c.Status.Dependencies,
		State:              clusterState,
		Version:            version,
	}
	if c.Status.Config != nil {
		minimalData.IngressClass = c.Status.Config.IngressClass
	}
	return minimalData
}

// clusterMinimalColumnMapping restricts filtering and sorting to only the fields
// available in the metadata response.
func clusterMinimalColumnMapping() queryapi.APIFieldsMapping {
	return queryapi.APIFieldsMapping{
		"name": queryapi.NewFieldMapping("name"),
	}
}
