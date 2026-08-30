package status

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/cluster/meta"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/generated/pkg/api"
)

type ApiService struct{}

// NewApiService creates a default api service
func NewApiService() api.StatusAPIServicer {
	return &ApiService{}
}

func (s *ApiService) Status(_ context.Context) (api.ImplResponse, error) {
	return api.Response(http.StatusOK, meta.Get()), nil
}
