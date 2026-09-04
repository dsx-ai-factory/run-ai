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

const ReportsTable = "reports"

type Report struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt        time.Time
	CreatedBy        string
	Name             string
	Description      string
	TenantId         int32
	Status           string
	StatusUpdatedAt  time.Time
	StatusMessage    string
	Start            time.Time
	End              time.Time
	Labels           LabelArray       `gorm:"type:jsonb"`
	AdditionalFields AdditionalFields `gorm:"type:jsonb"`
	Metrics          []*Metric        `gorm:"->;foreignKey:ReportID;references:ID"`
}

func (m *Report) TableName() string {
	return ReportsTable
}

func (m *Report) BeforeCreate(tx *gorm.DB) (err error) {
	emptyId := uuid.UUID{}
	if m.ID == emptyId {
		m.ID = uuid.New()
	}
	return
}

func (m *Report) SortAndFilterFieldsToColumn() queryapi.APIFieldsMapping {
	return queryapi.APIFieldsMapping{
		"name":      queryapi.NewFieldMapping("name"),
		"createdAt": queryapi.NewFieldMapping("created_at"),
		"createdBy": queryapi.NewFieldMapping("created_by"),
		"status":    queryapi.NewFieldMapping("status"),
	}
}
