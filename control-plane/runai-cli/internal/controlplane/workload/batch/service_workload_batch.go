package batch

import (
	"context"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

type Batcher interface {
	BatchDelete(ctx context.Context, ids ...string) (*openapi.BatchResponse, error)
}

type Service struct {
	workloadClient Batcher
}

func NewService(workloadClient Batcher) *Service {
	return &Service{workloadClient: workloadClient}
}

func (s *Service) Delete(ctx context.Context, ids ...string) (*openapi.BatchResponse, error) {
	batch, err := s.workloadClient.BatchDelete(ctx, ids...)
	if err != nil {
		return nil, err
	}
	return batch, nil
}
