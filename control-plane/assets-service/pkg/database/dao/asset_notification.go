/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	policies "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	async_events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	scope "github.com/dsx-ai-factory/run-ai/runai-application-packages/scopeinfo"
	policyService "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
)

func ValidateAssetVsPolicy(ctx context.Context, assetId string, scopeInfo *scope.ScopeInfo) *policyService.Error {
	// we need to consult policy service if the asset change does not validate the policy
	// this is only done for data sources
	return policies.Get().Client().NotifyAssetModified(ctx, assetId, policyService.ScopeType(scopeInfo.Scope), scopeInfo.ScopeId)
}

func NotifyAssetModified(ctx context.Context, assetId string, kind api.AssetKind, scopeInfo *scope.ScopeInfo) {
	tenantId := auth.TenantIdOf(ctx)

	err := runaipubsub.Publish[async_events.AssetModified](
		ctx,
		async_events.NewAssetModified(
			fmt.Sprintf("%d", tenantId),
			assetId,
			string(kind),
			string(scopeInfo.Scope),
			scopeInfo.ScopeId,
		))
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to publish asset modified for %s: %v", assetId, err)
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("update of asset %s populated", assetId)
}

func NotifyAssetDeleted(ctx context.Context, assetId string, kind api.AssetKind, scopeInfo *scope.ScopeInfo) {
	tenantId := auth.TenantIdOf(ctx)

	err := runaipubsub.Publish[async_events.AssetDeleted](
		ctx,
		async_events.NewAssetDeleted(
			fmt.Sprintf("%d", tenantId),
			assetId,
			string(kind),
			string(scopeInfo.Scope),
			scopeInfo.ScopeId,
		))
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to publish asset deleted for %s: %v", assetId, err)
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("deletion of asset %s populated", assetId)
}

func ScopeInfoOfAssetMeta(meta *api.AssetMeta) *scope.ScopeInfo {
	return scope.ScopeInfoOf(scope.Scope(meta.Scope), meta.ProjectId, meta.DepartmentId, meta.ClusterId, meta.TenantId)
}
