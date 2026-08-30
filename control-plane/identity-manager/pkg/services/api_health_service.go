package services

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/generated/pkg/api"
)

type HealthApiService struct{}

func NewHealthApiService() api.HealthAPIServicer {
	return &HealthApiService{}
}

func (s *HealthApiService) Health(ctx context.Context) (api.ImplResponse, error) {
	return api.Response(http.StatusOK, "I'm alive!"), nil
}
