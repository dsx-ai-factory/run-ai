/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"github.com/google/uuid"
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
	"gorm.io/gorm"
)

const (
	PrioritiesTableName       = "priorities"
	PriorityStatusesTableName = "priority_statuses"
	PrioritiesClusterViewName = "priorities_cluster_view"
	PriorityStatusPreload     = "Status"
)

func (Priority) TableName() string {
	return PrioritiesTableName
}

type Priority struct {
	dao_utils.BaseModel
	Name      string
	Value     int
	Source    string
	ScopeType string
	ScopeId   string
	TenantId  string
	Status    []PriorityStatus `gorm:"foreignKey:PriorityID"`
}

func (PriorityClusterView) TableName() string {
	return PrioritiesClusterViewName
}

type PriorityClusterView struct {
	dao_utils.BaseModel
	Name      string
	Value     int
	Source    string
	ScopeType string
	ScopeId   string
	TenantId  string
	StatusId  uuid.UUID
	Phase     string
	ClusterId string
}

func (PriorityStatus) TableName() string {
	return PriorityStatusesTableName
}

type PriorityStatus struct {
	ID         uuid.UUID
	PriorityID uuid.UUID
	ClusterId  uuid.UUID
	Phase      string
}

func (ps *PriorityStatus) BeforeCreate(tx *gorm.DB) (err error) {
	emptyId := uuid.UUID{}
	if ps.ID == emptyId {
		ps.ID = uuid.New()
	}
	return
}

func (Priority) FilterFieldsToColumn() map[string]string {
	return map[string]string{
		"name":   "name",
		"source": "source",
		"value":  "value",
	}
}

func (PriorityClusterView) FilterFieldsToColumn() map[string]string {
	return map[string]string{
		"name":   "name",
		"source": "source",
		"value":  "value",
		"phase":  "phase",
	}
}
