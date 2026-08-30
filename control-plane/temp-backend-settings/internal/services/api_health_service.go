package services

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/control-plane/temp-backend-settings/internal/generated/pkg/api"
)

type HealthAPIService struct{}

func NewHealthAPIService() *HealthAPIService {
	return &HealthAPIService{}
}

func (s *HealthAPIService) Health(ctx context.Context) (api.ImplResponse, error) {
	return api.Response(http.StatusOK, "I'm alive!"), nil
}
