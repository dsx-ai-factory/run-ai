package health

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/generated/pkg/api"
)

// ApiService is a service that implements the logic for the HealthApiServicer
// This service should implement the business logic for every endpoint for the HealthApi API.
// Include any external packages or services that will be required by this service.
type ApiService struct{}

// NewApiService creates a default api service
func NewApiService() api.HealthAPIServicer {
	return &ApiService{}
}

// Health - health check
func (s *ApiService) Health(ctx context.Context) (api.ImplResponse, error) {
	return api.Response(http.StatusOK, http.StatusText(http.StatusOK)), nil
}
