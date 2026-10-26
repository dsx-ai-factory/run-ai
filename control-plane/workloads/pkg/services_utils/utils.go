/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	authClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	authClientApi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	"golang.org/x/exp/slices"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type OrgUnitKey struct {
	Scope authClientApi.ScopeType
	Id    string
}

// CheckWorkloadOwnershipProtection checks if the workload is protected by ownership protection and if the user has the
// required permissions to perform the action, which means either:
// 1. the user is the owner of the workload
// 2. the user has permissions for the ancestor scope (e.g. the department that contains the given project)
func CheckWorkloadOwnershipProtection(ctx context.Context, clusterId string, departmentId string, submittedBy string) *api.Error {
	clusterConfig, err := checker.NewClusterCapabilitiesChecker(clusterId).ClusterConfiguration(ctx)
	if err != nil {
		return api.NewError(http.StatusServiceUnavailable, err.Error(), nil)
	}

	return CheckWorkloadOwnershipProtectionWithClusterConfig(ctx, clusterConfig, clusterId, departmentId, submittedBy)
}

// CheckWorkloadOwnershipProtectionWithClusterConfig is CheckWorkloadOwnershipProtection with the cluster configuration
// already fetched. Reading it is a round trip to the shared cache, so callers that check many workloads fetch it once
// per cluster instead of once per workload.
func CheckWorkloadOwnershipProtectionWithClusterConfig(ctx context.Context, clusterConfig *checker.ClusterConfiguration,
	clusterId string, departmentId string, submittedBy string) *api.Error {
	// if ownership protection is disabled or the user is the owner of the workload then leave
	if (!clusterConfig.WorkloadOwnershipProtection) || (submittedBy == *runaictx.GetString(ctx, runaictx.CfSubjectID)) {
		return nil
	}

	clusterChecker := checker.NewClusterCapabilitiesChecker(clusterId)
	clusterSupportsNewOwnershipProtection, err := clusterChecker.ValidateClusterCapability(ctx, checker.MinVersionOfNewWorkloadOwnershipProtection)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, err.Error(), nil)
	}

	if clusterSupportsNewOwnershipProtection {
		hasPermission, err := hasCreateProjectPermissionInDepartment(ctx, clusterId, departmentId)
		if err != nil {
			return api.NewError(http.StatusInternalServerError, err.Error(), nil)
		}

		if !hasPermission {
			return api.NewError(http.StatusForbidden, "insufficient permissions", nil)
		}
	} else {
		hasPermission, err := hasAncestorsPermission(ctx, clusterId, departmentId)
		if err != nil {
			return api.NewError(http.StatusInternalServerError, err.Error(), nil)
		}

		if !hasPermission {
			return api.NewError(http.StatusForbidden, "insufficient permissions", nil)
		}
	}

	return nil
}

func hasAncestorsPermission(ctx context.Context, clusterId string, departmentId string) (bool, error) {
	permittedScopes, err := authClient.ParsePermittedScopes(ctx)
	if err != nil {
		return false, err
	}

	if permittedScopes.GetSystem() {
		return true, nil
	}

	if permittedScopes.Tenants != nil {
		tenantId, err := ParseTenantId(ctx)
		if err != nil {
			return false, errors.New(err.Message)
		}
		if slices.Contains(permittedScopes.Tenants, strconv.Itoa(int(tenantId))) {
			return true, nil
		}
	}

	if permittedScopes.Clusters != nil {
		if slices.Contains(permittedScopes.Clusters, clusterId) {
			return true, nil
		}
	}

	if permittedScopes.Departments != nil {
		if slices.Contains(permittedScopes.Departments, departmentId) {
			return true, nil
		}
	}

	return false, nil
}

func hasCreateProjectPermissionInDepartment(ctx context.Context, clusterId string, departmentId string) (bool, error) {
	authorizationClient := authClient.NewAuthorizationClient()
	if authorizationClient == nil {
		return false, errors.New("authorization client is nil")
	}

	projectPermittedScopeActions, err := authorizationClient.CalculatePermittedScopes(ctx, authClientApi.GetPermittedScopesRequest{
		ResourceType: authClientApi.RESOURCETYPE_PROJECT,
		Action:       *authClientApi.NewNullableString((*string)(convutils.Ref(authClientApi.ACTION_CREATE))),
	})
	if err != nil {
		return false, err
	}

	projectCreationPermittedScopes := projectPermittedScopeActions.Create
	if slices.Contains(projectCreationPermittedScopes.Departments, departmentId) {
		return true, nil
	}

	if slices.Contains(projectCreationPermittedScopes.Clusters, clusterId) {
		return true, nil
	}

	tenantId, parseErr := ParseTenantId(ctx)
	if parseErr != nil {
		return false, errors.New(parseErr.Message)
	}
	if slices.Contains(projectCreationPermittedScopes.Tenants, strconv.Itoa(int(tenantId))) {
		return true, nil
	}

	return false, nil
}

func ValidateClusterSupportsCapability(ctx context.Context, clusterId string, capabilityToCheck int) *workloadsApi.SubmissionErrorV2 {
	capabilitiesChecker := checker.NewClusterCapabilitiesChecker(clusterId)

	clusterSupportFlatWorkloads, checkerError := capabilitiesChecker.ValidateClusterCapability(ctx, capabilityToCheck)
	if checkerError != nil {
		return workloadsApi.NewSubmissionErrorV2(http.StatusServiceUnavailable, checkerError.Error())
	}

	if !clusterSupportFlatWorkloads {
		return workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, fmt.Sprintf("this API is incompatible with the version of cluster '%s'", clusterId))
	}

	return nil
}

// ValidateClusterSupportsGenericWorkloads validates if agent supports generic workload submission
func ValidateClusterSupportsGenericWorkloads(ctx context.Context, clusterId string) *workloadsApi.SubmissionErrorV2 {
	return ValidateClusterSupportsCapability(ctx, clusterId, checker.MinVersionOfGenericWorkloadsSupport)
}
