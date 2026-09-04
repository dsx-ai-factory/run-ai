/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"fmt"

	"github.com/google/uuid"

	authorizationClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
)

const (
	ObjectsTable = "objects"

	ObjectsIdColumn           = "id"
	ObjectsTypeColumn         = "type"
	ObjectsNameColumn         = "name"
	ObjectsProjectIdColumn    = "project_id"
	ObjectsDepartmentIdColumn = "department_id"
	ObjectsClusterIdColumn    = "cluster_id"
	ObjectsTenantIdColumn     = "tenant_id"
	ObjectsSpecColumn         = "spec"
	ObjectsUniqueConstraint   = "unique_tenant_type_cluster_name_project"
	ObjectScopeTypeColumn     = "scope_type"
)

// ObjectType is defined as an "Enum" to enable storing in a single table all types of objects.
// Until now, we had to write a new table per object and a dedicated DAO
type ObjectType string

const (
	ObjectTypeStorageClass = "StorageClass"
)

// ObjectFilters is a set of filters we can filter by.
type ObjectFilters struct {
	TenantId     int32
	Type         string
	ProjectId    string
	ClusterId    uuid.UUID
	DepartmentId uuid.UUID
	ScopeType    authorizationClient.ScopeType
	Name         *string
}

func (s *ObjectFilters) TableName() string {
	return ObjectsTable
}

func NewObject(filters *ObjectFilters) *Object {
	var result Object
	if filters != nil {
		result.ObjectFilters = *filters
	}

	return &result
}

type Object struct {
	ObjectFilters `gorm:"embedded"`
	Spec          string `gorm:"type:JsonBmap"`
}

func (s *Object) TableName() string {
	return ObjectsTable
}

func (s *Object) TenantIdColumn() string {
	return ObjectsTenantIdColumn
}

func (s *Object) ClusterIdColumn() string {
	return ObjectsClusterIdColumn
}

func (s *Object) ScopeTypeColumn() string {
	return ObjectScopeTypeColumn
}

func (s *Object) ScopeIdColumn() string {
	switch s.ScopeType {
	case authorizationClient.SCOPETYPE_PROJECT:
		return ObjectsProjectIdColumn
	case authorizationClient.SCOPETYPE_DEPARTMENT:
		return ObjectsDepartmentIdColumn
	case authorizationClient.SCOPETYPE_CLUSTER:
		return ObjectsClusterIdColumn
	case authorizationClient.SCOPETYPE_TENANT:
		return ObjectsTenantIdColumn
	}
	return ""
}

func (s *Object) String() string {
	return fmt.Sprintf("Object type [%s] tenant[%d] cluster[%s] Department [%v] Spec [%v]",
		s.Type,
		s.TenantId,
		s.ClusterId,
		s.DepartmentId,
		s.Spec)
}

func (s *Object) ErrorMessage(action string) string {
	return fmt.Sprintf("failed to %s object: %s", action, s)
}
