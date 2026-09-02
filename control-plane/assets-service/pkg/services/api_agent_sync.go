/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/encryption"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type UnifiedAssetSyncResponse struct {
	RequestInfo  api.RequestInfo
	ResponseInfo api.ResponseInfo
	SystemInfo   api.SystemInfo

	CredentialAssets []api.CredentialsUpdate
	DatasourceAssets []api.DatasourceUpdate
}

func (uas *UnifiedAssetSyncResponse) ToCredentials() api.CredentialsUpdatesResp {
	return api.CredentialsUpdatesResp{
		RequestInfo:  uas.RequestInfo,
		ResponseInfo: uas.ResponseInfo,
		SystemInfo:   uas.SystemInfo,
		Assets:       uas.CredentialAssets,
	}
}

func (uas *UnifiedAssetSyncResponse) ToDatasource() api.DatasourceUpdatesResp {
	return api.DatasourceUpdatesResp{
		RequestInfo:  uas.RequestInfo,
		ResponseInfo: uas.ResponseInfo,
		SystemInfo:   uas.SystemInfo,
		Assets:       uas.DatasourceAssets,
	}
}

func (s *SyncApiService) handleAgentSyncUpdates(ctx context.Context,
	clusterId string,
	after int64,
	clusterSyncReport *api.ClusterSyncReport,
	assetKinds []api.AssetKind,
	needToDecrypt bool) (*UnifiedAssetSyncResponse, *api.Error,
) {
	tenantId := auth.TenantIdOf(ctx)

	//
	//  if cluster sent updates of previous synced assets, update them first
	//
	if clusterSyncReport != nil {
		dao.HandleAgentSyncUpdates(ctx, tenantId, clusterId, clusterSyncReport.Updates)
	}

	// Check if the cluster supports private credentials
	capabilitiesChecker := checker.NewClusterCapabilitiesChecker(clusterId)
	supportsPrivateCredentials, checkerErr := capabilitiesChecker.ValidateClusterCapability(ctx, checker.MinVersionOfUserCredentials)
	if checkerErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to check private credentials capability: %v", checkerErr)
		return nil, api.NewError(http.StatusInternalServerError, "Failed to check cluster capabilities", checkerErr)
	}

	updates, dbErr := dao.ListAssetsForSync(ctx, tenantId, clusterId, assetKinds, after, runai_config.Get[*config.ServicesConfig]().SyncBulkSize, supportsPrivateCredentials)
	if dbErr != nil {
		return nil, dbErr
	}

	syncResponse := UnifiedAssetSyncResponse{
		RequestInfo: api.RequestInfo{
			AfterUpdate: after,
		},
	}

	syncedUpdates, err := s.prepareAssetSyncUpdatesResp(ctx, needToDecrypt, &syncResponse, updates)
	if err != nil {
		return nil, err
	}

	err = dao.UpdateSyncStatusSendToAgent(ctx, tenantId, clusterId, syncedUpdates)
	if err != nil {
		return nil, err
	}

	macCount, countErr := dao.GetHighestUpdateCount(ctx)
	if countErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to get highest update count: %s", countErr.Message)
	} else {
		syncResponse.SystemInfo.HighestUpdate = macCount
	}

	return &syncResponse, nil
}

// prepareAssetSyncUpdatesResp gets the list of updates that the DB query finds.
func (s *SyncApiService) prepareAssetSyncUpdatesResp(ctx context.Context, needToDecrypt bool, syncResponse *UnifiedAssetSyncResponse,
	updates []dbmodel.AssetsToSync) ([]dbmodel.AssetsToSync, *api.Error,
) {
	if len(updates) != 0 {
		syncResponse.ResponseInfo.LowestUpdate = updates[0].UpdateCount
		syncResponse.ResponseInfo.HighestUpdate = updates[len(updates)-1].UpdateCount
	}

	var tenantKeys *encryption.TenantEncryptionKeys

	if needToDecrypt {
		var keysErr *api.Error

		tenantKeys, keysErr = encryption.GetTenantEncryptionKeys(ctx)
		if keysErr != nil {
			return nil, keysErr
		}
	}

	//
	// copy the updates to the result and also keep aside all the "real" updates, those that do not belong to "existing"
	// assets. we do not filter that in the query (see comment in the dao function for further info) so we have to do it here
	//
	toSync := make([]dbmodel.AssetsToSync, 0, len(updates))

	for _, update := range updates {
		var credentialsUpdate *dbmodel.AssetsToSync

		switch update.Kind {
		case api.ASSETKIND_PVC:
			if update.AssetsPvc == nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("asset %s is missing pvc spec", update.Id)
				continue
			} else if convutils.BoolIsTrue(update.AssetsPvc.ExistingPvc) {
				continue
			} else if convutils.BoolIsTrue(update.AutoDelete) {
				// Workload volumes (autoDelete=true) are created and owned by the workload runner,
				// not by the asset sync agent.
				continue
			}

			toSync = append(toSync, update)
			syncResponse.DatasourceAssets = append(syncResponse.DatasourceAssets, update.ToDatasourceUpdate())
			continue

		case api.ASSETKIND_ACCESS_KEY:
			if update.AssetsAccessKey == nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("asset %s is missing accessKey spec", update.Id)
				continue
			} else if convutils.IsStringHasValue(update.AssetsAccessKey.ExistingSecretName) {
				continue
			}

			credentialsUpdate = &update

		case api.ASSETKIND_PASSWORD:
			if update.AssetsPassword == nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("asset %s is missing password spec", update.Id)
				continue
			} else if convutils.IsStringHasValue(update.AssetsPassword.ExistingSecretName) {
				continue
			}

			credentialsUpdate = &update

		case api.ASSETKIND_DOCKER_REGISTRY:
			if update.AssetsDockerRegistry == nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("asset %s is missing docker registry spec", update.Id)
				continue
			} else if convutils.IsStringHasValue(update.AssetsDockerRegistry.ExistingSecretName) {
				continue
			}

			credentialsUpdate = &update

		case api.ASSETKIND_GENERIC_SECRET:
			if update.AssetsGenericSecret == nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("asset %s is missing generic secret spec", update.Id)
				continue
			} else if convutils.IsStringHasValue(update.AssetsGenericSecret.ExistingSecretName) {
				continue
			}

			credentialsUpdate = &update
		case api.ASSETKIND_NGC_API_KEY:
			if update.AssetsNgcApiKey == nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("asset %s is missing ngc api key spec", update.Id)
				continue
			}
			credentialsUpdate = &update
		default:
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("unsupported kind %s for asset %s", update.Kind, update.Id)
		}

		if credentialsUpdate != nil {
			toSync = append(toSync, update)

			converted, err := update.ToCredentialsUpdate(ctx, tenantKeys)
			if err != nil {
				return nil, err
			}

			syncResponse.CredentialAssets = append(syncResponse.CredentialAssets, *converted)
		}
	}

	return toSync, nil
}
