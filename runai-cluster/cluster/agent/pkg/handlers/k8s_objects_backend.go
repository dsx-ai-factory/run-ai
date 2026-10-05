/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config"
)

// K8sObjectsUpdateFromBackend mirrors AssetsUpdateFromBackend for the k8s-objects family of assets.
// We implement it as a separate type because k8s-objects assets carry a typed K8sObjectSpec whereas the
// legacy assets family uses an untyped Spec interface{} field with per-kind type assertions in the builder.
type K8sObjectsUpdateFromBackend struct {
	RequestInfo  *assetsApi.RequestInfo
	ResponseInfo *assetsApi.ResponseInfo
	SystemInfo   *assetsApi.SystemInfo
	Assets       []K8sObjectFromBackend
}

// K8sObjectFromBackend is the per-asset info.
// The Spec field is nil-safe: with IsDeleted=true it can be nil, since
// the agent only needs name/namespace to delete the cluster object.
type K8sObjectFromBackend struct {
	Meta        assetsApi.AssetMetaForSync
	UpdateCount int64
	IsDeleted   bool
	Spec        *assetsApi.K8sObjectSpec
}

// GetK8sObjectsUpdatesFromBackend fetches the next batch of k8s-objects
// deltas from the control plane and attaches the previous sync report in
// the POST body — identical cursor/report pattern to the credentials and
// datasources handlers (see assets_backend.go).
func (h *K8sObjectsHandler) GetK8sObjectsUpdatesFromBackend() (*K8sObjectsUpdateFromBackend, *assetsApi.Error) {
	syncReport, _ := k8sObjectsSyncReport.Load(k8sObjectsSyncReportKey)

	resp, err := h.backendClient.AssetsServiceClient.K8sObjectsSyncUpdates(
		h.backendClient.Context(), config.Config().ClusterUuid, h.syncCounter, syncReport.theReport)
	if err != nil {
		return nil, err
	}

	updates := &K8sObjectsUpdateFromBackend{
		RequestInfo:  resp.RequestInfo,
		ResponseInfo: resp.ResponseInfo,
		SystemInfo:   resp.SystemInfo,
		Assets:       make([]K8sObjectFromBackend, 0, len(resp.Assets)),
	}

	for _, asset := range resp.Assets {
		updates.Assets = append(updates.Assets, K8sObjectFromBackend{
			Meta:        asset.Meta,
			UpdateCount: asset.UpdateCount,
			IsDeleted:   convutils.BoolIsTrue(asset.IsDeleted.Get()),
			Spec:        asset.Spec,
		})
	}

	return updates, nil
}
