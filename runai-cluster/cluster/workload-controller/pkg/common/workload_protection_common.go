/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	PrivilegedWorkloadCrudPermissionsField = "spec.privilegedWorkloadCrudPermissions"
)

var RolesWithPermissionToCreateProject = map[string]bool{
	"system-administrator":     true,
	"department-administrator": true,
	"editor":                   true,
}

func PrivilegedWorkloadCrudPermissionsIndexer(object client.Object) []string {
	accessRule, ok := object.(*v1alpha1.AccessRule)
	if !ok {
		return []string{}
	}

	if accessRule.Spec.ScopeType == v1alpha1.ProjectScope {
		return []string{}
	}

	// currently it's hard-coded; in the future we will have to dynamically understand
	// if it has permissions to CRUD workloads or not
	_, found := RolesWithPermissionToCreateProject[accessRule.Spec.RoleName]
	if !found {
		return []string{}
	}

	return []string{accessRule.Spec.SubjectId}
}
