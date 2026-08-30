package services

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/control-plane/bff-service/internal/generated/pkg/api"
)

// HealthAPIService is a service that implements the logic for the HealthAPIServicer
// This service should implement the business logic for every endpoint for the HealthAPI API.
// Include any external packages or services that will be required by this service.
type HealthAPIService struct{}

// NewHealthAPIService creates a default api service
func NewHealthAPIService() *HealthAPIService {
	return &HealthAPIService{}
}

// Health - health check
func (s *HealthAPIService) Health(ctx context.Context) (api.ImplResponse, error) {
	return api.Response(http.StatusOK, "I'm alive!"), nil
}
