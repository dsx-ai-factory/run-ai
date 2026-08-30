package assets

import (
	"context"

	assetsService "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

type ClientInterface interface {
	ListDatasourcesByIds(ctx context.Context, ids []string) (*assetsService.DatasourceListResponse, *assetsService.Error)
}
