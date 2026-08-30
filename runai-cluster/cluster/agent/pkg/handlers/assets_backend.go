package handlers

import (
	"github.com/run-ai/runai/runai-application-packages/convutils"
	assetsApi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/config"
)

type AssetsUpdateFromBackend struct {
	RequestInfo  *assetsApi.RequestInfo
	ResponseInfo *assetsApi.ResponseInfo
	SystemInfo   *assetsApi.SystemInfo
	Assets       []AssetsFromBackend
}

type AssetsFromBackend struct {
	Meta        assetsApi.AssetMetaForSync `json:"meta"`
	UpdateCount int64                      `json:"updateCount"`
	IsDeleted   bool
	refIndex    int
	Spec        interface{}
}

func (ah *AssetsHandler) GetUpdatesFromBackend() (*AssetsUpdateFromBackend, *assetsApi.Error) {
	syncReport, _ := syncReportPerAssetType.Load(ah.assetGroup)

	var updates AssetsUpdateFromBackend

	switch ah.assetGroup {
	case AssetGroupDatasources:
		resp, err := ah.backendClient.AssetsServiceClient.DatasourceSyncUpdates(ah.backendClient.Context(), config.Config().ClusterUuid, ah.syncCounter, syncReport.theReport)
		if err != nil {
			return nil, err
		}

		updates = AssetsUpdateFromBackend{
			RequestInfo:  resp.RequestInfo,
			ResponseInfo: resp.ResponseInfo,
			SystemInfo:   resp.SystemInfo,
			Assets:       make([]AssetsFromBackend, 0, len(resp.Assets)),
		}

		for idx, asset := range resp.Assets {
			updates.Assets = append(updates.Assets, AssetsFromBackend{
				Meta:        asset.Meta,
				UpdateCount: asset.UpdateCount,
				IsDeleted:   convutils.BoolIsTrue(asset.IsDeleted.Get()),
				Spec:        asset.Spec,
				refIndex:    idx,
			})
		}

	case AssetGroupCredentials:
		resp, err := ah.backendClient.AssetsServiceClient.CredentialsSyncUpdates(ah.backendClient.Context(), config.Config().ClusterUuid, ah.syncCounter, syncReport.theReport)
		if err != nil {
			return nil, err
		}

		updates = AssetsUpdateFromBackend{
			RequestInfo:  resp.RequestInfo,
			ResponseInfo: resp.ResponseInfo,
			SystemInfo:   resp.SystemInfo,
			Assets:       make([]AssetsFromBackend, 0, len(resp.Assets)),
		}

		for _, asset := range resp.Assets {
			updates.Assets = append(updates.Assets, AssetsFromBackend{
				Meta:        asset.Meta,
				UpdateCount: asset.UpdateCount,
				IsDeleted:   convutils.BoolIsTrue(asset.IsDeleted.Get()),
				Spec:        asset.Spec,
			})
		}
	}

	return &updates, nil
}
