/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"gorm.io/gorm"
)

const (
	resourceIssuesTable           = "resource_issues"
	TenantIdColumn                = "tenant_id"
	ClusterIdColumn               = "cluster_id"
	ScopeTypeColumn               = "scope_type"
	ScopeIdColumn                 = "scope_id"
	ResourceIdColumn              = "resource_id"
	resourceIssuesTenantIdColumn  = "tenant_id"
	resourceIssuesClusterIdColumn = "cluster_id"
	resourceIssuesScopeTypeColumn = "scope_type"
	resourceIssuesScopeIdColumn   = "scope_id"
)

type ResourceIssueFilters struct {
	TenantId   int32
	ClusterId  *string
	ScopeType  *string
	ScopeId    *string
	ResourceId *string
	Type       *api.ClusterResourceType
}

func (filters *ResourceIssueFilters) TableName() string {
	return resourceIssuesTable
}

type ResourceIssue struct {
	ID                   uuid.UUID `gorm:"primaryKey"`
	ResourceIssueFilters `gorm:"embedded"`
	Issue                string
}

func NewResourceIssue(filters *ResourceIssueFilters) *ResourceIssue {
	var result ResourceIssue
	if filters != nil {
		result.ResourceIssueFilters = *filters
	}

	return &result
}

func (ri *ResourceIssue) BeforeCreate(_ *gorm.DB) error {
	ri.ID = uuid.New()
	return nil
}

func (ri *ResourceIssue) TableName() string {
	return resourceIssuesTable
}

func (ri *ResourceIssue) TenantIdColumn() string {
	return resourceIssuesTenantIdColumn
}

func (ri *ResourceIssue) ClusterIdColumn() string {
	return resourceIssuesClusterIdColumn
}

func (ri *ResourceIssue) ScopeTypeColumn() string {
	return resourceIssuesScopeTypeColumn
}

func (ri *ResourceIssue) ScopeIdColumn() string {
	return resourceIssuesScopeIdColumn
}

func (ri *ResourceIssue) ErrorMessage(title string) string {
	return fmt.Sprintf("failed to %s resource issue %s", title, ri.details())
}

func (ri *ResourceIssue) details() string {
	return fmt.Sprintf("tenant[%d] cluster[%s] %s[%s] resource issue [%s]",
		ri.TenantId,
		convutils.Deref(ri.ClusterId, ""),
		convutils.Deref(ri.ScopeType, ""),
		convutils.Deref(ri.ScopeId, ""),
		convutils.Deref(ri.ResourceId, ""))
}
