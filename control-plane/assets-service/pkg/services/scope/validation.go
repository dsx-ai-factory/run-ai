/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package scope

import (
	"context"
	"errors"
	"fmt"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/clients/org_unit"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-cluster/common/utils"
	"net/http"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
)

// FieldValidation type states if a field is forbidden, mandatory or optional
type FieldValidation string

const (
	FieldForbidden FieldValidation = "should not be provided"
	FieldMandatory FieldValidation = "is mandatory"
	FieldOptional  FieldValidation = "is optional"
)

type ScopeFields string

const (
	ScopeFieldProjectId    ScopeFields = "projectId"
	ScopeFieldDepartmentId ScopeFields = "departmentId"
	ScopeFieldClusterId    ScopeFields = "clusterId"
)

var scopeValidationsPerScope = map[api.Scope]map[ScopeFields]FieldValidation{
	api.SCOPE_PROJECT: {
		ScopeFieldProjectId:    FieldMandatory,
		ScopeFieldDepartmentId: FieldForbidden,
		ScopeFieldClusterId:    FieldOptional, // we ignore its value, but allow the field for backward compatibility
	},
	api.SCOPE_DEPARTMENT: {
		ScopeFieldProjectId:    FieldForbidden,
		ScopeFieldDepartmentId: FieldMandatory,
		ScopeFieldClusterId:    FieldForbidden,
	},
	api.SCOPE_CLUSTER: {
		ScopeFieldProjectId:    FieldForbidden,
		ScopeFieldDepartmentId: FieldForbidden,
		ScopeFieldClusterId:    FieldMandatory,
	},
	api.SCOPE_TENANT: {
		ScopeFieldProjectId:    FieldForbidden,
		ScopeFieldDepartmentId: FieldForbidden,
		ScopeFieldClusterId:    FieldForbidden,
	},
	api.SCOPE_SYSTEM: {
		ScopeFieldProjectId:    FieldForbidden,
		ScopeFieldDepartmentId: FieldForbidden,
		ScopeFieldClusterId:    FieldForbidden,
	},
}

func ValidateScopeParameters(request *api.AssetCreationRequest) *api.Error {
	valueGiven := map[ScopeFields]bool{
		ScopeFieldProjectId:    request.ProjectId != nil,
		ScopeFieldDepartmentId: request.DepartmentId != nil,
		ScopeFieldClusterId:    request.ClusterId != nil,
	}

	fields := scopeValidationsPerScope[request.Scope]

	for field, fieldValidation := range fields {
		switch fieldValidation {
		case FieldMandatory:
			if !valueGiven[field] {
				return api.NewValidationError(fmt.Sprintf("%s %s for %s scoped assets", field, fieldValidation, request.Scope))
			}
		case FieldForbidden:
			if valueGiven[field] {
				return api.NewValidationError(fmt.Sprintf("%s %s for %s scoped assets", field, fieldValidation, request.Scope))
			}
		}
	}

	return nil
}

// GetNameValidationQuery returns a query to validate name uniqueness, when updating an asset - provide the asset id.
func GetNameValidationQuery(
	ctx context.Context,
	assetName string,
	assetScope api.Scope,
	projectId *int32,
	departmentId *string,
	clusterId *string,
	private *bool,
	assetKind api.AssetKind,
	id *string,
) (*database.ParametrizedQuery, *api.Error) {
	nameValidationQuery := database.NewParametrizedQuery()

	nameValidationQuery.Add(fmt.Sprintf("%v.kind = ? AND %v.name = ?", dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), assetKind, assetName)

	// this case is for update: if we are updating an asset we want to exclude
	// this asset from the validation
	if id != nil {
		nameValidationQuery.Add(fmt.Sprintf(" AND %v.id != ?", dbtables.TableNameAssetsMeta), *id)
	}

	// for system scope asset we are only checking that no other system scoped asset
	// has the same name. we allow non-system assets to have the same name as
	// a system scoped asset
	if assetScope == api.SCOPE_SYSTEM {
		nameValidationQuery.Add(fmt.Sprintf(` AND %v.scope = ?`, dbtables.TableNameAssetsMeta), api.SCOPE_SYSTEM)
		return nameValidationQuery, nil
	}

	tenantId := auth.TenantIdOf(ctx)

	// for tenant asset we don't allow the same name to be given to any other asset in any scope under this tenant.
	if assetScope == api.SCOPE_TENANT {
		var privateAsset = convutils.Deref(private, false)
		nameValidationQuery.Add(fmt.Sprintf(` AND %v.tenant_id = ? AND COALESCE(%v.%v, FALSE) = ?`,
			dbtables.TableNameAssetsMeta,
			dbtables.TableNameAssetsMeta, dbtables.AssetMetaColumnPrivate), tenantId, privateAsset)
		if privateAsset {
			subjectId := auth.SubjectIdOf(ctx)
			nameValidationQuery.Add(fmt.Sprintf(` AND %v.created_by = ?`, dbtables.TableNameAssetsMeta), subjectId)
		}
		return nameValidationQuery, nil
	}

	// for cluster scope, we do not allow a tenant-scoped-assets or any other scope under the same cluster to have the same name.
	nameValidationQuery.Add(fmt.Sprintf(` AND ((%v.scope = ? AND %v.tenant_id = ? and COALESCE(%v.%v, FALSE) = FALSE)`,
		dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta, dbtables.AssetMetaColumnPrivate), api.SCOPE_TENANT, tenantId)
	if assetScope == api.SCOPE_CLUSTER {
		nameValidationQuery.Add(fmt.Sprintf(` OR %v.cluster_id = ?)`, dbtables.TableNameAssetsMeta), *clusterId)
		return nameValidationQuery, nil
	}

	if assetScope == api.SCOPE_DEPARTMENT {
		if clusterId == nil {
			ancestors, err := org_unit.Get().GetDepartmentAncestors(ctx, *departmentId)
			if err != nil {
				return nil, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to get ancestors for department: %v", *departmentId), errors.New(err.Message))
			}
			clusterId = utils.Ref(ancestors.GetClusterUuid())
		}
		nameValidationQuery.Add(fmt.Sprintf(` OR (%v.scope = ? AND %v.cluster_id = ?)`, dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), api.SCOPE_CLUSTER, *clusterId)
		nameValidationQuery.Add(fmt.Sprintf(` OR (%v.scope = ? AND %v.department_id = ?)`, dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), api.SCOPE_DEPARTMENT, *departmentId)

		// in addition to the ancestors of the department, the name cannot be given to any assets in projects of that department.
		departmentProjects, orgErr := org_unit.Get().GetProjectsMetadataByDepartment(ctx, *departmentId)
		if orgErr != nil {
			return nil, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to get projects of department: %s", *departmentId), errors.New(orgErr.Message))
		}

		var departmentProjectIds []string
		for _, project := range departmentProjects {
			departmentProjectIds = append(departmentProjectIds, project.Id)
		}

		if len(departmentProjectIds) > 0 {
			nameValidationQuery.Add(fmt.Sprintf(` OR (%v.scope = ? AND %v.project_id IN ?)`, dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), api.SCOPE_PROJECT, departmentProjectIds)
		}
	}

	if assetScope == api.SCOPE_PROJECT {
		if departmentId == nil || clusterId == nil {
			ancestors, err := org_unit.Get().GetProjectAncestors(ctx, fmt.Sprintf("%v", *projectId))
			if err != nil {
				return nil, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to get ancestors for project: %v", *projectId), errors.New(err.Message))
			}
			departmentId = utils.Ref(ancestors.GetDepartmentId())
			clusterId = utils.Ref(ancestors.GetClusterUuid())
		}
		nameValidationQuery.Add(fmt.Sprintf(` OR (%v.scope = ? AND %v.cluster_id = ?)`, dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), api.SCOPE_CLUSTER, *clusterId)
		nameValidationQuery.Add(fmt.Sprintf(` OR (%v.scope = ? AND %v.department_id = ?)`, dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), api.SCOPE_DEPARTMENT, *departmentId)
		nameValidationQuery.Add(fmt.Sprintf(` OR (%v.scope = ? AND %v.project_id = ?)`, dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), api.SCOPE_PROJECT, *projectId)
	}

	nameValidationQuery.Add(")")
	return nameValidationQuery, nil
}

// GetClaimNameValidationQuery returns a query to validate claim name uniqueness for PVC assets.
// This checks that no other PVC asset in the same scope branch has the same claim name.
// PVC assets are only supported at cluster, department, and project scope.
// Note: projectId, departmentId, and clusterId are validated by the caller before calling this function.
func GetClaimNameValidationQuery(ctx context.Context, claimName string, assetScope api.Scope, projectId int32, departmentId string, clusterId string) (*database.ParametrizedQuery, *api.Error) {
	tenantId := auth.TenantIdOf(ctx)

	claimNameValidationQuery := database.NewParametrizedQuery()

	// Build base query: filter by tenant first, then kind, and claim name
	claimNameValidationQuery.Add(fmt.Sprintf("%v.tenant_id = ? AND %v.kind = ? AND %v.claim_name = ? AND (",
		dbtables.TableNameAssetsMeta,
		dbtables.TableNameAssetsMeta,
		dbtables.TableNameAssetsPVC), tenantId, api.ASSETKIND_PVC, claimName)

	if assetScope == api.SCOPE_CLUSTER {
		// For cluster scope: check all assets in this cluster (cluster + departments + projects)
		claimNameValidationQuery.Add(fmt.Sprintf(`%v.cluster_id = ?)`, dbtables.TableNameAssetsMeta), clusterId)
		return claimNameValidationQuery, nil
	}

	if assetScope == api.SCOPE_DEPARTMENT {
		if clusterId == "" {
			ancestors, err := org_unit.Get().GetDepartmentAncestors(ctx, departmentId)
			if err != nil {
				return nil, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to get ancestors for department: %v", departmentId), errors.New(err.Message))
			}
			clusterId = ancestors.GetClusterUuid()
		}
		// For department scope: check cluster-level assets, this department, and all its projects
		claimNameValidationQuery.Add(fmt.Sprintf(`(%v.scope = ? AND %v.cluster_id = ?)`, dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), api.SCOPE_CLUSTER, clusterId)
		claimNameValidationQuery.Add(fmt.Sprintf(` OR (%v.scope = ? AND %v.department_id = ?)`, dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), api.SCOPE_DEPARTMENT, departmentId)

		// in addition to the ancestors of the department, the claim name cannot be given to any assets in projects of that department.
		departmentProjects, orgErr := org_unit.Get().GetProjectsMetadataByDepartment(ctx, departmentId)
		if orgErr != nil {
			return nil, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to get projects of department: %s", departmentId), errors.New(orgErr.Message))
		}

		var departmentProjectIds []string
		for _, project := range departmentProjects {
			departmentProjectIds = append(departmentProjectIds, project.Id)
		}

		if len(departmentProjectIds) > 0 {
			claimNameValidationQuery.Add(fmt.Sprintf(` OR (%v.scope = ? AND %v.project_id IN ?)`, dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), api.SCOPE_PROJECT, departmentProjectIds)
		}
	}

	if assetScope == api.SCOPE_PROJECT {
		if departmentId == "" || clusterId == "" {
			ancestors, err := org_unit.Get().GetProjectAncestors(ctx, fmt.Sprintf("%v", projectId))
			if err != nil {
				return nil, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to get ancestors for project: %v", projectId), errors.New(err.Message))
			}
			departmentId = ancestors.GetDepartmentId()
			clusterId = ancestors.GetClusterUuid()
		}
		// For project scope: check cluster-level assets, department-level assets, and this project
		claimNameValidationQuery.Add(fmt.Sprintf(`(%v.scope = ? AND %v.cluster_id = ?)`, dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), api.SCOPE_CLUSTER, clusterId)
		claimNameValidationQuery.Add(fmt.Sprintf(` OR (%v.scope = ? AND %v.department_id = ?)`, dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), api.SCOPE_DEPARTMENT, departmentId)
		claimNameValidationQuery.Add(fmt.Sprintf(` OR (%v.scope = ? AND %v.project_id = ?)`, dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), api.SCOPE_PROJECT, projectId)
	}

	claimNameValidationQuery.Add(")")
	return claimNameValidationQuery, nil
}
