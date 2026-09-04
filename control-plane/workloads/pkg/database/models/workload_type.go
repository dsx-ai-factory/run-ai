/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkloadTypeSource string

const (
	WorkloadTypeSourceNative   WorkloadTypeSource = "native"
	WorkloadTypeSourceCustom   WorkloadTypeSource = "custom"
	WorkloadTypeSourceExternal WorkloadTypeSource = "external"
)

type WorkloadType struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string
	CreatedAt time.Time
	CreatedBy *string
	Group     string
	Source    WorkloadTypeSource
}

func (WorkloadType) TableName() string {
	return "workload_types"
}

func (wt *WorkloadType) BeforeCreate(tx *gorm.DB) (err error) {
	wt.ID = uuid.New()
	return
}

func (WorkloadType) FieldsConfiguration() queryapi.APIFieldsMapping {
	return queryapi.APIFieldsMapping{
		"name":  queryapi.NewFieldMapping("name"),
		"group": queryapi.NewFieldMapping("\"group\""),
	}
}

type NativeWorkloadTypeConfig struct {
	ID           uuid.UUID
	Name         string
	TenantId     string
	CategoryName string
	PriorityName string
}

type WorkloadTypeStatus struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	WorkloadTypeId uuid.UUID `gorm:"type:uuid;index"`
	ClusterId      uuid.UUID `gorm:"type:uuid;index"`
	TenantId       string    `gorm:"index"`
	UpdatedAt      time.Time

	Phase      string
	Conditions ConditionArray `gorm:"type:jsonb"`
	Version    string
}

func (WorkloadTypeStatus) TableName() string {
	return "workload_type_statuses"
}
