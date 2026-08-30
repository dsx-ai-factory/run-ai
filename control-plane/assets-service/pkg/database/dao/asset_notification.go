package dao

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	policies "github.com/run-ai/runai/control-plane/workloads-and-policies/clients/policies"
	async_events "github.com/run-ai/runai/runai-application-packages/async-events"
	scope "github.com/run-ai/runai/runai-application-packages/scopeinfo"
	policyService "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
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
