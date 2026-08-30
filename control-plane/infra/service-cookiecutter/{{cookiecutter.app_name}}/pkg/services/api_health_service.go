package services

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/control-plane/{{cookiecutter.app_name}}/internal/generated/pkg/api"
)

// HealthApiService is a service that implements the logic for the HealthApiServicer
// This service should implement the business logic for every endpoint for the HealthApi API.
// Include any external packages or services that will be required by this service.
type HealthApiService struct {
}

// NewHealthApiService creates a default api service
func NewHealthApiService() api.HealthApiServicer {
	return &HealthApiService{}
}

// Health - health check
func (s *HealthApiService) Health(ctx context.Context) (api.ImplResponse, error) {
	return api.Response(http.StatusOK, "I'm alive!"), nil
}
