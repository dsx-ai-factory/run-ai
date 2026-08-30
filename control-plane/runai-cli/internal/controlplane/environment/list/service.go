package list

import (
	"context"
	"fmt"

	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

const (
	APIError = "environment API call error: %w"
)

type Service struct {
	client Lister
}

type Lister interface {
	List(ctx context.Context, projectID *int32, clusterID *string, name *string, workloadType *string) (*openapi.EnvironmentListResponse, error)
}

func NewService(client Lister) *Service {
	return &Service{client: client}
}

func (es *Service) List(ctx context.Context, projectID *int32, clusterID *string) (*Output, error) {
	list, err := es.client.List(ctx, projectID, clusterID, nil, nil)
	if err != nil {
		return nil, fmt.Errorf(APIError, err)
	}
	return ParseListResponse(list)
}
