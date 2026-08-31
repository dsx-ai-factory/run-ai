/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package query_types

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

// ApplyRbacRules adds rbac filtering to queryParams. this is used by database list functions, which adds
// "where" clauses based on this filtering
func ApplyRbacRules(ctx context.Context, queryParams *QueryParams) (*auth.PermittedScopes, *api.Error) {
	queryParams.AllowedOrgUnits = NewAllowedOrgUnits()

	permittedScopes, err := auth.LoadAllowedOrgUnits(ctx)
	if err != nil {
		return nil, api.NewError(int(err.Code), err.Message, errors.New(err.Details))
	}

	if queryParams.ClusterId == nil {
		// if requested to see all the clusters, restrict the results to all the clusters the user is allowed to see
		queryParams.AllowedOrgUnits.AllowedClusters = permittedScopes.AllowedClustersArr()
	} else {
		// if requested single cluster, either add this cluster or nothing if not allowed for this cluster
		for _, allowedCluster := range permittedScopes.AllowedClustersArr() {
			if allowedCluster == *queryParams.ClusterId {
				queryParams.AllowedOrgUnits.AllowedClusters = append(queryParams.AllowedOrgUnits.AllowedClusters, allowedCluster)
				break
			}
		}
	}

	if queryParams.DepartmentId == nil {
		// if requested to see all the departments, restrict the results to all the departments the user is allowed to see
		queryParams.AllowedOrgUnits.AllowedDepartments = permittedScopes.AllowedDepartmentsArr()
	} else {
		// if requested single department, either add this department or nothing if not allowed for this department
		for _, allowedDepartment := range permittedScopes.AllowedDepartmentsArr() {
			if allowedDepartment == *queryParams.DepartmentId {
				queryParams.AllowedOrgUnits.AllowedDepartments = append(queryParams.AllowedOrgUnits.AllowedDepartments, allowedDepartment)
				break
			}
		}
	}

	if queryParams.ProjectId == nil {
		// if requested to see all the project, restrict the results to all the departments the user is allowed to see
		queryParams.AllowedOrgUnits.AllowedProjects = permittedScopes.AllowedProjectsArr()
	} else {
		projectIdAsString := strconv.Itoa(int(*queryParams.ProjectId))

		// if requested single project, either add this project or nothing if not allowed for this project
		// if this explanation sounds familiar it's because it's the same for department, ten lines above... :-)
		for _, allowedProject := range permittedScopes.AllowedProjectsArr() {
			if allowedProject == projectIdAsString {
				queryParams.AllowedOrgUnits.AllowedProjects = append(queryParams.AllowedOrgUnits.AllowedProjects, allowedProject)
				break
			}
		}
	}

	return permittedScopes, nil
}

func ApplyRbacRulesGetById(ctx context.Context, assetMeta api.AssetMeta) (*auth.PermittedScopes, []string, *api.Error) {
	permittedScopes, err := auth.LoadAllowedOrgUnits(ctx)
	if err != nil {
		return nil, nil, api.NewError(int(err.Code), err.Message, errors.New(err.Details))
	}

	switch assetMeta.Scope {
	case api.SCOPE_PROJECT:
		projectIdStr := fmt.Sprintf("%d", convutils.Deref(assetMeta.ProjectId, 0))
		if permittedScopes.AllowedProjects().Has(projectIdStr) {
			return permittedScopes, permittedScopes.AllowedProjectsArr(), nil
		}

	case api.SCOPE_DEPARTMENT:
		if permittedScopes.AllowedDepartments().Has(convutils.Deref(assetMeta.DepartmentId, "")) {
			return permittedScopes, permittedScopes.AllowedProjectsArr(), nil
		}

	case api.SCOPE_CLUSTER:
		if permittedScopes.AllowedCluster().Has(convutils.Deref(assetMeta.ClusterId, "")) {
			return permittedScopes, permittedScopes.AllowedProjectsArr(), nil
		}

	case api.SCOPE_TENANT:
		if auth.TenantIdOf(ctx) == convutils.Deref(assetMeta.TenantId, 0) {
			return permittedScopes, permittedScopes.AllowedProjectsArr(), nil
		}

	case api.SCOPE_SYSTEM:
		return permittedScopes, permittedScopes.AllowedProjectsArr(), nil
	}

	return permittedScopes, nil, api.NewError(int(auth.ForbiddenError.Code), auth.ForbiddenError.Message, errors.New(auth.ForbiddenError.Details))
}
