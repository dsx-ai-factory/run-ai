package clusters

import (
	"context"
	"fmt"
	"net/http"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
)

type InternalApiServiceMock struct {
	called   int
	response []openapi.ClusterHistory
	err      error
}

func NewInternalApiServiceMock(response []openapi.ClusterHistory, err error) *InternalApiServiceMock {
	return &InternalApiServiceMock{called: 0, response: response, err: err}
}

func (a *InternalApiServiceMock) GetClustersHistories(ctx context.Context) openapi.ApiGetClustersHistoriesRequest {
	return openapi.ApiGetClustersHistoriesRequest{}
}

func (a *InternalApiServiceMock) GetClustersHistoriesExecute(r openapi.ApiGetClustersHistoriesRequest) ([]openapi.ClusterHistory, *http.Response, error) {
	a.called++
	return a.response, nil, a.err
}

func (a *InternalApiServiceMock) SetClusterAccessControl(ctx context.Context, clusterUuid string) openapi.ApiSetClusterAccessControlRequest {
	return openapi.ApiSetClusterAccessControlRequest{}
}

func (a *InternalApiServiceMock) SetClusterAccessControlExecute(r openapi.ApiSetClusterAccessControlRequest) (*http.Response, error) {
	return nil, fmt.Errorf("not implemented")
}
