/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
)

const fairShareSimulationProjectType = "project"

func (o OrgUnitApiService) SimulateNodePoolFairShare(ctx context.Context, _ string, request api.FairShareSimulationRequest) (api.ImplResponse, error) {
	// The spec can't require department or project read depending on orgUnitType, so the handler checks it.
	resourceType := authorization_api.RESOURCETYPE_DEPARTMENT
	if request.OrgUnitType == fairShareSimulationProjectType {
		resourceType = authorization_api.RESOURCETYPE_PROJECT
	}
	scopes, _, err := o.AuthorizationClient.GetPermittedScopes(ctx, resourceType, ptr.To(authorization_api.ACTION_READ))
	if err != nil {
		return HandleError(fmt.Errorf("get permitted %s scopes: %w", resourceType, err), http.StatusServiceUnavailable,
			"Failed to get permitted scopes"), nil
	}
	if !hasAnyScope(scopes.Read) {
		return api.Response(http.StatusForbidden, api.NewError(http.StatusForbidden,
			fmt.Sprintf("simulating %ss requires read on %ss", request.OrgUnitType, resourceType), nil)), nil
	}
	return api.Response(http.StatusNotImplemented, api.NewError(http.StatusNotImplemented, "fair share simulation is not implemented yet", nil)), nil
}

func hasAnyScope(scopes authorization_api.PermittedScopes) bool {
	return ptr.Deref(scopes.System, false) || len(scopes.Tenants) > 0 || len(scopes.Clusters) > 0 ||
		len(scopes.Departments) > 0 || len(scopes.Projects) > 0
}
