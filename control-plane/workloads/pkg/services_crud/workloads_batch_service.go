package services_crud

import (
	"context"
	"errors"
	"net/http"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
)

type WorkloadsBatchAPIService struct{}

// NewWorkloadsBatchAPIService creates a default api service
func NewWorkloadsBatchAPIService() *WorkloadsBatchAPIService {
	return &WorkloadsBatchAPIService{}
}

// BatchWorkloads - Workload batch operations.
func (s *WorkloadsBatchAPIService) BatchWorkloads(ctx context.Context, request api.WorkloadBatch) (api.ImplResponse, error) {
	err := validateBatchRequest(request)
	if err != nil {
		return api.Response(http.StatusBadRequest, nil), err
	}

	switch request.Action {
	case api.ACTION_DELETE:
		return ChangeDesiredStates(ctx, request.Ids, api.WORKLOADDESIREDPHASE_DELETED)
	case api.ACTION_SUSPEND:
		return ChangeDesiredStates(ctx, request.Ids, api.WORKLOADDESIREDPHASE_STOPPED)
	case api.ACTION_RESUME:
		return ChangeDesiredStates(ctx, request.Ids, api.WORKLOADDESIREDPHASE_RUNNING)
	default:
		return api.Response(http.StatusNotImplemented, nil), errors.New("BatchWorkloads method not implemented")

	}
}

func validateBatchRequest(request api.WorkloadBatch) error {
	if len(request.Ids) == 0 {
		return errors.New("ids cannot be empty")
	}

	if len(request.Ids) > 100 {
		return errors.New("ids cannot be more than 100")
	}
	return nil
}
