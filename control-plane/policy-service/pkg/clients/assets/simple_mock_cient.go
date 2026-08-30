package assets

import (
	"context"

	"github.com/run-ai/runai/runai-application-packages/datastructs"
	assetsService "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

// SimpleMockClient only supports fixed assets.
type SimpleMockClient struct {
	Assets assetsService.DatasourceListResponse
}

func NewSimpleMockClient() *SimpleMockClient {
	return &SimpleMockClient{
		Assets: assetsService.DatasourceListResponse{
			Entries: make([]assetsService.DatasourceListResponseEntry, 0),
		},
	}
}

func (smc *SimpleMockClient) WithAsset(asset assetsService.DatasourceListResponseEntry) *SimpleMockClient {
	smc.Assets.Entries = append(smc.Assets.Entries, asset)
	return smc
}

func (smc *SimpleMockClient) ListDatasourcesByIds(ctx context.Context, ids []string) (*assetsService.DatasourceListResponse, *assetsService.Error) {
	if len(ids) == 0 {
		return &smc.Assets, nil
	}

	idsSet := datastructs.NewSet(ids...)

	result := make([]assetsService.DatasourceListResponseEntry, 0, len(ids))

	for _, asset := range smc.Assets.Entries {
		if idsSet.Has(asset.Meta.Id) {
			result = append(result, asset)
		}
	}

	return assetsService.NewDatasourceListResponse(result), nil
}

type SimpleMockClientFactory struct {
	*SimpleMockClient
}

func (smc SimpleMockClientFactory) Client() ClientInterface {
	return smc.SimpleMockClient
}
