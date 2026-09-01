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

	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/scopes"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/tree"
	async_events "github.com/run-ai/runai/runai-application-packages/async-events"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
	assetsApi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"golang.org/x/exp/maps"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

func (pa *PolicyApiService) AssetChangedHook(ctx context.Context, assetId string, notification api.AssetChangeNotification) (api.ImplResponse, error) {
	err := assetHook(ctx, assetId, notification)
	if err != nil {
		if err.Details != "" {
			return api.Response(http.StatusBadRequest, api.NewError(http.StatusBadRequest, err.Message, errors.New(err.Details))), nil
		} else {
			return api.Response(http.StatusBadRequest, api.NewError(http.StatusBadRequest, err.Message, nil)), nil
		}
	}

	return api.Response(http.StatusNoContent, nil), nil
}

func AssetModified(ctx context.Context, event *async_events.AssetModified) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("ignore AssetModified event for asset %s: assuming it is handled by the internal API", event.AssetId)

	return nil
}

func AssetDeleted(ctx context.Context, event *async_events.AssetDeleted) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("handling AssetDeleted event for asset %s", event.AssetId)

	// filter out assets that cannot be used as imposed assets
	switch assetsApi.AssetKind(event.Kind) {
	case assetsApi.ASSETKIND_COMPUTE,
		assetsApi.ASSETKIND_ENVIRONMENT,
		assetsApi.ASSETKIND_ACCESS_KEY,
		assetsApi.ASSETKIND_PASSWORD,
		assetsApi.ASSETKIND_DOCKER_REGISTRY,
		assetsApi.ASSETKIND_GENERIC_SECRET,
		assetsApi.ASSETKIND_REGISTRY:
		return nil
	}

	notification := api.AssetChangeNotification{
		ChangeType: api.ASSETCHANGETYPE_DELETED,
		ScopeType:  api.ScopeType(event.ScopeType),
		ScopeId:    convutils.Ref(event.ScopeId),
	}

	err := assetHook(ctx, event.AssetId, notification)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed handling AssetDeleted event for asset %s: %v", event.AssetId, err.Message)
	} else {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("AssetDeleted event for asset %s handled successfuly", event.AssetId)
	}

	return nil
}

func assetHook(ctx context.Context, assetId string, notification api.AssetChangeNotification) *api.Error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("[%d] received an update that asset %s has been changed", auth.TenantIdOf(ctx), assetId)

	tenantId := auth.TenantIdOf(ctx)

	orgUnit := scopes.OrgUnit{
		Scope: notification.ScopeType,
		Id:    convutils.Deref(notification.ScopeId, ""),
	}

	//
	//  the bulletproof order would be to first lock the relevant nodes in the tree and then to search for policies
	//  containing the asset. however, the apply-policy reads the imposed asset updated value so the chance that
	//  someone is updating an asset and at the same time (milliseconds difference) adds it as imposed to a policy is
	//  so rare, that it would be a real waste to lock the tree before we even know if the asset is relevant. if
	//  such rear case ever happens an additional update to the policy will fix it.
	//
	policies, err := dao.FindGivenPoliciesOfOrgUnitAllTypes(ctx, runaidb.DB().WithContext(ctx), orgUnit.Scope, orgUnit.ClusterId(),
		orgUnit.DepartmentId(), orgUnit.ProjectId(ctx), dbmodel.NotToIncludeDeletedPolicies)
	if err != nil {
		return err
	}

	if len(policies) == 0 {
		return nil
	}

	updates := make([]*tree.ApplyPolicyRequest, 0, len(policies))

	// typically policy works on single workload type, but here we have to try all supported
	// policy types because the same imposed assets can take part in both of them
	for _, policy := range policies {
		if policy.ImposedAssets == nil {
			continue
		}

		assetFoundInPolicy := false
		numberOfImposedAssets := 0

		for _, replica := range []workloadsApi.ReplicaType{workloadsApi.REPLICATYPE_WORKER, workloadsApi.REPLICATYPE_MASTER} {
			var imposedAssetIds *[]string

			if replica == workloadsApi.REPLICATYPE_WORKER {
				imposedAssetIds = policy.ImposedAssets.Worker
			} else {
				imposedAssetIds = policy.ImposedAssets.Master
			}

			if imposedAssetIds == nil || len(*imposedAssetIds) == 0 {
				continue
			}

			//
			//   if the asset is not included in the imposed assets of this policy nothing to do
			//
			imposedAssetsSet := datastructs.NewSet(*imposedAssetIds...)
			if !imposedAssetsSet.Has(assetId) {
				continue
			}

			//
			//   for deleted asset - remove the asset from the list, for modified asset we just update the list
			//   to force re-calculation of the effective
			//
			if notification.ChangeType == api.ASSETCHANGETYPE_DELETED {
				delete(imposedAssetsSet, assetId)
				*imposedAssetIds = maps.Keys(imposedAssetsSet)
			}

			numberOfImposedAssets = len(*imposedAssetIds)
			assetFoundInPolicy = true
		}

		if assetFoundInPolicy {
			workloadType := tree.WorkloadType(policy.WorkloadType)

			hierarchyTree, treeInitialized := tree.HierarchyTrees.Get(ctx, tenantId, workloadType)
			if !treeInitialized {
				err := hierarchyTree.InitializeTenant(ctx, tenantId, workloadType)
				if err != nil {
					logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("[%d:%s] failed to load policies", tenantId, workloadType)
					continue
				}
			}

			request := tree.GenericPolicyRequest{
				Requester:    auth.SubjectIdOf(ctx),
				TenantId:     auth.TenantIdOf(ctx),
				WorkloadType: workloadType,
				PolicyMeta:   *policy.ToWorkspaceCreationFields(),
			}

			if len(policy.RulesV2) == 0 && len(policy.DefaultsV2) == 0 && numberOfImposedAssets == 0 {
				//
				//   if the asset was the only element in the policy, the policy should be deleted. we do not
				//   keep empty policies
				//
				err := hierarchyTree.HandleDeleteRequest(ctx, &request, tree.DeleteRequestOpt{DeleteSubTree: false})
				if err != nil {
					logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("[%d:%s] failed to delete policy %s following removal of imposed assets: %+v", tenantId, policy.WorkloadType, policy.ID.String(), *err)
				}
			} else {
				updateRequest := tree.ApplyPolicyRequest{
					GenericPolicyRequest: request,
					ImposedAssets:        policy.ImposedAssets,
				}

				//
				//   for updates, we first check that the update is valid. it is possible, for example, that
				//   this update conflicts a policy where this asset is set as an imposed asset
				//
				_, applyError := hierarchyTree.HandleApplyRequest(ctx, &updateRequest, tree.ApplyPolicyOptions{
					Overwrite:    false,
					ValidateOnly: true,
				})

				if applyError != nil {
					message := fmt.Sprintf("failed to update asset %s: update conflicts %s policy [%s], in which it is used to set up an imposedAsset",
						assetId, policy.WorkloadType, policy.OrgUnit().String())

					return api.NewError(http.StatusBadRequest, message, errors.New(applyError.Message))
				}

				//   keep aa list of the updates that passed validation, we'll run them once we finish validating
				//   all the policies where this asset is set as an imposedAsset
				updates = append(updates, &updateRequest)
			}
		}
	}

	//
	//  finally, run the updates again this time with real update not only validation
	//
	for _, update := range updates {
		hierarchyTree, treeInitialized := tree.HierarchyTrees.Get(ctx, tenantId, update.WorkloadType)
		if !treeInitialized {
			err := hierarchyTree.InitializeTenant(ctx, tenantId, update.WorkloadType)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("[%d:%s] failed to load policies", tenantId, update.WorkloadType)
				continue
			}
		}

		_, applyError := hierarchyTree.HandleApplyRequest(ctx, update, tree.ApplyPolicyOptions{Overwrite: false})

		if applyError != nil {
			message := fmt.Sprintf("[%d] failed to update asset %s due to conflict with %s policy, in which it is used to set up an imposedAsset",
				tenantId, assetId, update.WorkloadType)
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("%s", message)
		}
	}

	return nil
}
