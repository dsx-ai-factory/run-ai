/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	orgUnitApi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"
	assetsScope "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/scope"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
)

var DefaultCheckerClusterConfig = &checker.ClustersCacheConfig{ExpirationSec: 600, CleanupSec: 10}

func ValidateAssetCreateRequestMeta(ctx context.Context, meta *api.AssetCreationRequest, kind api.AssetKind) *api.Error {
	if meta == nil {
		return api.NewValidationError("request meta is not provided")
	}

	scopeErr := assetsScope.ValidateScopeParameters(meta)
	if scopeErr != nil {
		return scopeErr
	}

	if meta.Scope == api.SCOPE_PROJECT || meta.Scope == api.SCOPE_DEPARTMENT {

		var resourceId string

		switch meta.Scope {
		case api.SCOPE_PROJECT:
			resourceId = fmt.Sprintf("%v", *meta.ProjectId)
		case api.SCOPE_DEPARTMENT:
			resourceId = *meta.DepartmentId
		}

		//
		//   complete the request with the cluster ID of the project/department
		//
		orgScopeType, orgErr := orgUnitApi.NewScopeTypeFromValue(string(meta.Scope))
		if orgErr != nil {
			return api.NewError(http.StatusInternalServerError, fmt.Sprintf("invalid scope type: %s", meta.Scope), orgErr)
		}
		resp, err := org_unit.Get().GetAncestorsOfScope(ctx, *orgScopeType, resourceId)
		if err != nil {
			return api.NewError(int(err.GetCode()), "failed to interrogate the project", errors.New(err.Message))
		}

		meta.ClusterId = resp.ClusterUuid.Get()
	}

	if clusterErr := ValidateClusterIdBelongsToTenant(ctx, meta.ClusterId); clusterErr != nil {
		return clusterErr
	}

	if kind == api.ASSETKIND_ENVIRONMENT {
		if meta.WorkloadSupportedTypes == nil {
			return api.NewValidationError("mandatory parameter workloadSupportedTypes is not provided")
		} else {
			validationError := validateSelectedType(meta.WorkloadSupportedTypes)
			if validationError != nil {
				return validationError
			}
		}
	}

	if kind != api.ASSETKIND_ENVIRONMENT && meta.WorkloadSupportedTypes != nil {
		return api.NewValidationError(fmt.Sprintf("workloadSupportedTypes is not supported for kind %s", kind))
	}

	var private = convutils.BoolIsTrue(meta.Private)
	// Private assets is supported only for docker registry and generic secret
	if private && kind != api.ASSETKIND_DOCKER_REGISTRY && kind != api.ASSETKIND_GENERIC_SECRET && kind != api.ASSETKIND_NGC_API_KEY {
		return api.NewValidationError(fmt.Sprintf("private asset isn't supported for kind %s", kind))
	}
	// Verify that private asset is in tenant scope
	if private && meta.Scope != api.SCOPE_TENANT {
		return api.NewValidationError(fmt.Sprintf("private asset is not supported in %v scope", meta.Scope))
	}

	return nil
}

func ContainedAssetsToAssetRef(containedAssets []dbmodel.AssetsContainedInWorkload, dataSourceOverridesById map[string]*api.DataSourceOverrides) api.AssetsRef {
	result := api.AssetsRef{}

	for _, asset := range containedAssets {
		switch asset.Kind {
		case api.ASSETKIND_ENVIRONMENT:
			result.Environment = api.EnvironmentAssetRef{
				Id:   asset.Id,
				Name: asset.Name,
			}
		case api.ASSETKIND_COMPUTE:
			result.Compute = &api.AssetsRefCompute{
				Id:   asset.Id,
				Name: asset.Name,
			}
		default:
			dataSourceRef := api.DatasourceRef{
				Id:        asset.Id,
				Name:      asset.Name,
				Kind:      asset.Kind,
				Overrides: dataSourceOverridesById[asset.Id],
			}
			result.Datasources = append(result.Datasources, dataSourceRef)
		}
	}

	return result
}

func safeStrPtr(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// ValidateClusterIdBelongsToTenant validates that the given clusterId exists and is owned
// by the tenant of the caller (as derived from the JWT/context).
//
// Uses /api/v1/clusters/{clusterUuid}/minimal (clusters-minimal:read) instead of
// /api/v1/clusters/{uuid} (cluster:read) because asset creators are not guaranteed to
// have the latter permission.
//
// Fails closed: returns 403 Forbidden whenever ownership cannot be confirmed — i.e. when
// the cluster does not exist, belongs to another tenant, or the cluster service cannot be
// reached. The same response is used in every failure case to avoid leaking information
// about foreign clusters or upstream service health.
// Returns nil when clusterId is nil or empty (no validation is needed for tenant/system scope).
func ValidateClusterIdBelongsToTenant(ctx context.Context, clusterId *string) *api.Error {
	if convutils.IsEmptyOrNull(clusterId) {
		return nil
	}

	callerTenantId := auth.TenantIdOf(ctx)

	minimalCluster, apiErr := cluster.Get().Client().GetClusterMinimalDataById(ctx, *clusterId)
	if apiErr == nil && minimalCluster.GetTenantId() == callerTenantId {
		return nil
	}

	return api.NewError(http.StatusForbidden, auth.ForbiddenError.Message, nil)
}
