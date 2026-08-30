package services

import (
	"context"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/query_types"
)

// DataVolumesAccessor is an interface to access data volumes
type DataVolumesAccessor interface {
	ListDataVolumes(ctx context.Context, queryParams *query_types.QueryParams) ([]api.DatasourceListResponseEntry, error)
}
