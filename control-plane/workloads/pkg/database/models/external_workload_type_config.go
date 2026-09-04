/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"

	"github.com/google/uuid"
)

type ExternalWorkloadTypeConfig struct {
	ID             uuid.UUID
	TenantId       *string
	ScopeType      string
	WorkloadTypeId uuid.UUID
	CategoryID     *uuid.UUID
	PriorityID     *uuid.UUID
	Preemptibility *string
	CreatedAt      time.Time
	CreatedBy      *string
	UpdatedAt      time.Time
	UpdatedBy      *string
}

func (ExternalWorkloadTypeConfig) TableName() string {
	return "external_workloads_type_config"
}

type ExternalWorkloadTypeConfigView struct {
	ID             uuid.UUID
	Name           string
	Source         string
	TenantId       *string
	ScopeType      string
	CategoryId     uuid.UUID
	CategoryName   string
	PriorityId     uuid.UUID
	PriorityName   string
	Preemptibility string
	CreatedAt      time.Time
	CreatedBy      *string
	UpdatedAt      time.Time
	UpdatedBy      *string

	Group string
	// Kartas - an array of WorkloadTypeKarta per-version (renamed from ResourceInterfaces in v2.26)
	Kartas WorkloadTypeKartaArray `gorm:"column:kartas;type:json"`

	// not part of the actual view - can be fetched separately
	Statuses []WorkloadTypeStatus `gorm:"-"`
}

func (ExternalWorkloadTypeConfigView) TableName() string {
	return "external_workloads_type_config_view"
}

func (ExternalWorkloadTypeConfigView) FieldsConfiguration() queryapi.APIFieldsMapping {
	return WorkloadType{}.FieldsConfiguration()
}
