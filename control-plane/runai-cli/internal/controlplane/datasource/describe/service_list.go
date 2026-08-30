package describe

import (
	"context"
	"fmt"

	assetsgenerated "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

type Lister interface {
	List(ctx context.Context, name *string, projectId *int32, clusterId *string, dsType *string) (*assetsgenerated.DatasourceListResponse, error)
}

func FindByName(ctx context.Context, lister Lister, name string, projectID *int32, clusterID *string, dsType *string) (*assetsgenerated.DatasourceListResponseEntry, error) {
	response, err := lister.List(ctx, &name, projectID, clusterID, dsType)
	if err != nil {
		return nil, fmt.Errorf("failed to search for datasource: %w", err)
	}

	for _, ds := range response.Entries {
		if ds.Meta.Name == name {
			return &ds, nil
		}
	}

	if dsType != nil && *dsType != "" {
		return nil, fmt.Errorf("datasource '%s' of type '%s' not found", name, *dsType)
	}
	return nil, fmt.Errorf("datasource '%s' not found", name)
}

func ListByName(ctx context.Context, lister Lister, clusterId, datasourceName string, projectId *int32, dsType *string) (*Datasource, error) {
	response, err := lister.List(ctx, &datasourceName, projectId, &clusterId, dsType)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch datasource: %w", err)
	}

	entries := response.GetEntries()
	if len(entries) == 0 {
		return nil, fmt.Errorf("datasource '%s' does not exist", datasourceName)
	}

	rawDatasource := entries[0]

	displayDatasource, err := buildDatasourceFromApiResponse(&rawDatasource)
	if err != nil {
		return nil, fmt.Errorf("failed to build datasource: %w", err)
	}

	return displayDatasource, nil
}
