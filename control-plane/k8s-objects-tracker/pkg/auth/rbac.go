/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"context"
	"net/http"

	"k8s.io/utils/ptr"

	authorizationClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/clients/authorization"
	orgUnit "github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/clients/org-unit"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

var readAction = authorizationClient.ACTION_READ

func CheckReadPermission(ctx context.Context, resourceType authorizationClient.ResourceType,
	ancestorsResolver *orgUnit.AncestorsResolver,
) *api.Error {
	scopeType := authorizationClient.ScopeType(ancestorsResolver.ScopeType())
	scopeId := ancestorsResolver.ScopeId()

	rawPermittedScopes, err := authorization.Get().CheckPermittedScopes(ctx, resourceType, &readAction)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("check permitted scopes failed: %s: %v", err.Message, err.Details)
		return err
	}

	if len(rawPermittedScopes.Read.Tenants) > 0 {
		return nil
	}

	permittedScopes := NewPermittedScopes(&rawPermittedScopes.Read)

	//
	//   round 1: "naive" validation. department scope VS allowed departments, project scope VS allowed projects
	//
	switch scopeType {
	case authorizationClient.SCOPETYPE_CLUSTER:
		if permittedScopes.isAllowedCluster(scopeId) {
			return nil
		}

	case authorizationClient.SCOPETYPE_DEPARTMENT:
		if permittedScopes.isAllowedDepartment(scopeId) {
			return nil
		}

	case authorizationClient.SCOPETYPE_PROJECT:
		if permittedScopes.isAllowedProject(scopeId) {
			return nil
		}
	}

	//
	//   round 2: we did not have a direct match, so we need to check the ancestors of the asset to see
	//   maybe the permission can be granted on the basis of the ancestors.
	//
	ancestors, getErr := ancestorsResolver.Get(ctx)
	if getErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to obtain ancestors of %s/%s: %v", scopeType, scopeId, getErr.Message)
		return err
	}

	switch scopeType {
	case authorizationClient.SCOPETYPE_DEPARTMENT:
		if permittedScopes.isAllowedCluster(ptr.Deref(ancestors.ClusterUuid.Get(), "")) {
			return nil
		}
	case authorizationClient.SCOPETYPE_PROJECT:
		if permittedScopes.isAllowedDepartment(ptr.Deref(ancestors.DepartmentId.Get(), "")) ||
			permittedScopes.isAllowedCluster(ptr.Deref(ancestors.ClusterUuid.Get(), "")) {
			return nil
		}
	}

	return api.NewError(http.StatusForbidden, "no permissions", nil)
}
