/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"context"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type SyncStatusRec struct {
	Id string `gorm:"not null;"`

	TenantId  int32  `gorm:"not null;"`
	ClusterId string `gorm:"not null;"`

	AssetId  string `gorm:"not null;"`
	PrevName string
	Spec     JsonBMapFromString

	SentToAgentApply  int64
	SentToAgentDelete int64
	ProcessedByAgent  int64
	Message           string
}

var CredentialAssetsToSync = []api.AssetKind{
	api.ASSETKIND_PASSWORD,
	api.ASSETKIND_ACCESS_KEY,
	api.ASSETKIND_GENERIC_SECRET,
	api.ASSETKIND_DOCKER_REGISTRY,
	api.ASSETKIND_NGC_API_KEY,
}
var credentialsAssetsToSyncSet = datastructs.NewSet(CredentialAssetsToSync...)

var CredentialAssetsToSyncWithoutNgcApiKey = []api.AssetKind{
	api.ASSETKIND_PASSWORD,
	api.ASSETKIND_ACCESS_KEY,
	api.ASSETKIND_GENERIC_SECRET,
	api.ASSETKIND_DOCKER_REGISTRY,
}

var DatasourceAssetsToSync = []api.AssetKind{
	api.ASSETKIND_PVC,
}
var dataSourceAssetsToSync = datastructs.NewSet(DatasourceAssetsToSync...)

func (SyncStatusRec) TableName() string {
	return dbtables.TableNameSyncStatus
}

// DetermineStatus Note -> This function assumes 1:1 between asset and synced cluster. it is valid to assume
// so, because we do not sync tenant/system scoped assets. once we synchronize tenant scoped assets we will
// have to introduce a new status info object in the API, allowing to specify status per cluster
//
// There are a few scenarios
// (1) asset update, agent did not "get" this update yet. status should be creating/deleting
// (2) agent received the update, but did not process it yet: status should be creating/deleting
// (3) agent received the update and processed it
//
//	  (3a) there is an error message -> status should be: create/delete failure, with the message
//	  (3b) there is no error
//	 		Create: the status should be determined based on k8sObjectsTracker
//		    Delete: status should be "deleting", until the asset will be permanently removed
func (ssr *SyncStatusRec) DetermineStatus(ctx context.Context, meta *api.AssetMeta, status **api.AssetClusterStatusInfo) {
	recentCountSent := max(ssr.SentToAgentApply, ssr.SentToAgentDelete)

	if ssr.ProcessedByAgent > recentCountSent {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("asset %s cluster %s: processed by agent %d is greater than recent update we sent to it %d, this cannot be",
			ssr.AssetId, ssr.ClusterId, ssr.ProcessedByAgent, recentCountSent)
		return
	}

	//
	//   there was an update that the agent did not receive (1), or an update sent to the agent, but it is still processing (2)
	//
	if meta.UpdateCount > recentCountSent || ssr.ProcessedByAgent < recentCountSent {
		*status = &api.AssetClusterStatusInfo{}

		if meta.DeletedAt.IsZero() {
			if ssr.ProcessedByAgent == 0 {
				(*status).Status = api.ASSETSYNCSTATUSENUM_CREATING
			} else {
				(*status).Status = api.ASSETSYNCSTATUSENUM_UPDATING
			}
		} else {
			(*status).Status = api.ASSETSYNCSTATUSENUM_DELETING
		}
		return
	}

	// if the agent processed the recent sync event, then we are no longer doing sync
	// if the sync failed with an error message (3a), update it in the status. otherwise we do not update
	// any status and in this case the status will be based on any issue obtained by k8sObjectsTracker
	if ssr.Message != "" {
		*status = &api.AssetClusterStatusInfo{
			Message: ssr.Message,
		}

		if meta.DeletedAt.IsZero() {
			(*status).Status = api.ASSETSYNCSTATUSENUM_FAILED
			if credentialsAssetsToSyncSet.Has(meta.Kind) {
				(*status).Message = "The credential couldn't be synchronized to the cluster"
			}
			if dataSourceAssetsToSync.Has(meta.Kind) {
				(*status).Message = "The data source couldn't be synchronized to the cluster"
			}
		} else {
			(*status).Status = api.ASSETSYNCSTATUSENUM_DELETION_FAILED
		}
	}

	// there was no error of agent processing (3b)
	if !meta.DeletedAt.IsZero() {
		*status = &api.AssetClusterStatusInfo{
			Status: api.ASSETSYNCSTATUSENUM_DELETING,
		}
	}
}
