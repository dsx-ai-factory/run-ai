/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	backend "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/tree"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

const nonPrivate = false

var SupportTenantMode = false

// apply policy takes care both workspace and training.
func applyPolicy(ctx context.Context, request *tree.ApplyPolicyRequest, options tree.ApplyPolicyOptions) (*dbmodel.PolicyWithOptionalEffective, *api.Error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof(">>> apply %s policy for %s started", request.WorkloadType, request.PolicyMeta.Scope)

	if request.PolicyMeta.Scope == api.SCOPETYPE_SYSTEM {
		return nil, api.NewValidationError("system scope is not supported for policies")
	}

	if !options.Overwrite {
		err := validateResetFields(request.FieldsToReset, request.Defaults, request.Rules)
		if err != nil {
			return nil, err
		}
	}

	request.Requester = auth.SubjectIdOf(ctx)
	request.TenantId = auth.TenantIdOf(ctx)

	scopeParams, scopeErr := NewScopeInfoFromMeta(request.PolicyMeta)
	if scopeErr != nil {
		return nil, scopeErr
	}

	rbacErr := auth.ValidateAccessByResourceTypeAndAction(ctx, scopeParams, backend.RESOURCETYPE_POLICIES, backend.ACTION_UPDATE, nonPrivate)
	if rbacErr != nil {
		return nil, api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))
	}

	// Get hold of the relevant hierarchy tree
	hierarchyTree, treeInitialized := tree.HierarchyTrees.Get(ctx, auth.TenantIdOf(ctx), request.WorkloadType)
	if !treeInitialized {
		err := hierarchyTree.InitializeTenant(ctx, request.TenantId, request.WorkloadType)
		if err != nil {
			return nil, err
		}
	}

	fallbackTree, err := getFallbackTree(ctx, auth.TenantIdOf(ctx), request.WorkloadType)
	if err != nil {
		return nil, err
	}

	if fallbackTree != nil {
		// make sure that all the policies of the fallback exist
		_, err = fallbackTree.HandleGetRequest(ctx, request.ToFallback())
		if err != nil {
			return nil, err
		}
	}

	return hierarchyTree.HandleApplyRequest(ctx, request, options)
}
