package services

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/catalog-service/pkg/services/huggingface"
)

// inferencesAPIService is a service that implements the logic for the InferencesAPIServicer
// This service should implement the business logic for every endpoint for the InferencesAPI API.
// Include any external packages or services that will be required by this service.

type InferencesAPIService struct {
	modelSearchService huggingface.ModelSearchInterface
}

// NewInferencesAPIService creates a default api service
func NewInferencesAPIService() *InferencesAPIService {
	return &InferencesAPIService{
		modelSearchService: huggingface.NewModelSearchService(),
	}
}

// SearchHuggingFaceCatalog - Search Hugging Face models.
func (s *InferencesAPIService) SearchHuggingFaceCatalog(ctx context.Context, query string) (api.ImplResponse, error) {
	var allModels []huggingface.Model
	var apiErr *api.Error

	if query == "" {
		// Fetch trending and most downloaded models
		allModels, apiErr = s.modelSearchService.FetchDefaultModels(ctx)
	} else {
		// Search for specific query
		allModels, apiErr = s.modelSearchService.SearchModels(ctx, query)
	}

	if apiErr != nil {
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	// Convert to API response format
	response := huggingface.ConvertToAPIResponse(allModels)

	return api.Response(http.StatusOK, response), nil
}
