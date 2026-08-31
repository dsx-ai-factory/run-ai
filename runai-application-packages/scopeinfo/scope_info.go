/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package scopeinfo

import (
	"fmt"
)

type Scope string

const (
	System     Scope = "system"
	Tenant     Scope = "tenant"
	Cluster    Scope = "cluster"
	Department Scope = "department"
	Project    Scope = "project"
)

// ScopeInfo is a utility class which allowed us to org unit id in unified manner for all the scopes
// unlike AssetMeta, for example, in which department and project id are stored separately.
type ScopeInfo struct {
	Scope   Scope
	ScopeId string
}

func ScopeInfoOf(scope Scope, projectId *int32, departmentId *string, clusterId *string, tenantId *int32) *ScopeInfo {
	result := ScopeInfo{
		Scope: scope,
	}
	if scope == Project {
		result.ScopeId = fmt.Sprintf("%d", *projectId)
	}
	if scope == Department {
		result.ScopeId = *departmentId
	}
	if scope == Cluster {
		result.ScopeId = *clusterId
	}
	if scope == Tenant && tenantId != nil {
		result.ScopeId = fmt.Sprintf("%d", *tenantId)
	}
	return &result
}

func NewSystemScope() *ScopeInfo {
	return &ScopeInfo{Scope: System}
}

func NewTenantScope() *ScopeInfo {
	return &ScopeInfo{Scope: Tenant}
}

func NewClusterScope(clusterId string) *ScopeInfo {
	return &ScopeInfo{Scope: Cluster, ScopeId: clusterId}
}

func NewDepartmentScope(departmentId string) *ScopeInfo {
	return &ScopeInfo{Scope: Department, ScopeId: departmentId}
}

func NewProjectScope(projectId string) *ScopeInfo {
	return &ScopeInfo{Scope: Project, ScopeId: projectId}
}
