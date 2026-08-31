/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/queryapi"
	"gorm.io/gorm"
)

// AllAuthenticatedUsers is the reserved value stored in both the subject_type and the subject_id
// columns for all-authenticated-users rules (which carry no real subject). It is intentionally
// absent from the base SubjectType create enum — reachable only through the dedicated
// authenticated-users endpoint, never the regular create body. As the subject_id it satisfies the
// not-null column and, via the unique index, keeps one such rule per role and scope.
const AllAuthenticatedUsers = "all-authenticated-users"

type AccessRule struct {
	Common         `gorm:"embedded"`
	ID             uint                    `gorm:"primarykey"`
	SubjectId      string                  `gorm:"uniqueIndex:idx_subject_role_resource,where:deleted_at is null;not null"`
	SubjectType    api.SubjectTypeExtended `gorm:"not null"`
	RoleID         uint                    `gorm:"uniqueIndex:idx_subject_role_resource,where:deleted_at is null;not null"`
	Role           Role                    `gorm:"constraint:OnDelete:CASCADE;foreignKey:RoleID"`
	ScopeId        string                  `gorm:"uniqueIndex:idx_subject_role_resource,where:deleted_at is null;not null"`
	ScopeType      api.ScopeType           `gorm:"uniqueIndex:idx_subject_role_resource,where:deleted_at is null;not null"`
	ScopeName      string                  `gorm:"not null"`
	TenantId       uint                    `gorm:"index"`
	ClusterUuid    string                  `gorm:"index;not null"`
	DeletedAt      gorm.DeletedAt          `gorm:"index"`
	Phase          string                  `gorm:""`
	PhaseMessage   string                  `gorm:""`
	PhaseByCluster PhaseByCluster          `gorm:"type:jsonb"`
}

func (AccessRule) SortAndFilterFieldsToColumn() queryapi.APIFieldsMapping {
	return queryapi.APIFieldsMapping{
		"createdAt":       queryapi.NewFieldMapping("access_rules.created_at"),
		"createdBy":       queryapi.NewFieldMapping("access_rules.created_by"),
		"subjectIdFilter": queryapi.NewFieldMapping("subject_id"),
		"subjectId":       queryapi.NewFieldMapping("subject_id"),
		"subjectType":     queryapi.NewFieldMapping("subject_type"),
		"roleId":          queryapi.NewFieldMapping("role_id"),
		"roleName":        queryapi.NewFieldMapping("roles.name"),
		"scopeId":         queryapi.NewFieldMapping("access_rules.scope_id"),
		"scopeType":       queryapi.NewFieldMapping("access_rules.scope_type"),
		"scopeName":       queryapi.NewFieldMapping("access_rules.scope_name"),
		"deletedAt":       queryapi.NewFieldMapping("deleted_at"),
		"phase":           queryapi.NewFieldMapping("phase"),
	}
}

// PhaseByCluster - mapping clusterId -> to the phase in that cluster
type PhaseByCluster map[string]string

func (c PhaseByCluster) Value() (driver.Value, error) {
	valueString, err := json.Marshal(c)
	return string(valueString), err
}

func (c *PhaseByCluster) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &c); err != nil {
		return err
	}
	return nil
}
