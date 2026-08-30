package huggingface

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/run-ai/runai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

// API parameter constants
const (
	ParamSearch = "search"
	ParamLimit  = "limit"
	ParamSort   = "sort"
	ParamFull   = "full"
)

// API parameter values
const (
	SortDownloads     = "downloads"
	SortTrendingScore = "trendingScore"
	LimitTrending     = "10"  // Limit for trending models
	LimitDownload     = "100" // Limit for download models
	FullTrue          = "true"
	PipelineTagParam  = "pipeline_tag"
	TextGeneration    = "text-generation"
)

// Numeric constants
const (
	MaxModelsLimit = 100 // Maximum number of models to return
)

// Model represents a model from the HuggingFace API
type Model struct {
	ID        string      `json:"id"`
	Gated     interface{} `json:"gated,omitempty"`
	Downloads int32       `json:"downloads"`
	Trending  bool        `json:"-"` // Internal flag, not from HuggingFace API
}

// ModelSearchInterface defines the interface for searching HuggingFace models
type ModelSearchInterface interface {
	SearchModels(ctx context.Context, query string) ([]Model, *api.Error)
	FetchDefaultModels(ctx context.Context) ([]Model, *api.Error)
}

// ModelSearchService implements the ModelSearchInterface
type ModelSearchService struct {
	httpClient *http.Client
	baseURL    string
}

// NewModelSearchService creates a new ModelSearchService
func NewModelSearchService() ModelSearchInterface {
	return &ModelSearchService{
		httpClient: &http.Client{},
		baseURL:    "https://huggingface.co/api/models",
	}
}

// NewModelSearchServiceWithConfig creates a new ModelSearchService with custom configuration
func NewModelSearchServiceWithConfig(httpClient *http.Client, baseURL string) ModelSearchInterface {
	return &ModelSearchService{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// buildSearchParams creates parameters for search queries
func (s *ModelSearchService) buildSearchParams(query string) map[string]string {
	return map[string]string{
		ParamSearch: query,
		ParamLimit:  LimitDownload,
		ParamSort:   SortDownloads,
		ParamFull:   FullTrue,
	}
}

// buildDefaultParams creates base parameters for default model queries (trending/downloaded)
func (s *ModelSearchService) buildDefaultParams(sortBy, limit string) map[string]string {
	return map[string]string{
		ParamSort:        sortBy,
		ParamLimit:       limit,
		ParamFull:        FullTrue,
		PipelineTagParam: TextGeneration,
	}
}

// buildTrendingParams creates parameters for trending models
func (s *ModelSearchService) buildTrendingParams() map[string]string {
	return s.buildDefaultParams(SortTrendingScore, LimitTrending)
}

// buildDownloadedParams creates parameters for downloaded models
func (s *ModelSearchService) buildDownloadedParams() map[string]string {
	return s.buildDefaultParams(SortDownloads, LimitDownload)
}

// SearchModels searches for models based on the query
func (s *ModelSearchService) SearchModels(ctx context.Context, query string) ([]Model, *api.Error) {
	params := s.buildSearchParams(query)
	return s.fetchHuggingFaceModels(ctx, params)
}

// FetchDefaultModels fetches trending and most downloaded models
func (s *ModelSearchService) FetchDefaultModels(ctx context.Context) ([]Model, *api.Error) {
	logger := logging.WithCtx(runaictx.LoggerHook(ctx))

	// Fetch trending models
	trendingModels, trendingErr := s.fetchHuggingFaceModels(ctx, s.buildTrendingParams())
	if trendingErr != nil {
		logger.Warnf("Failed to fetch trending models with %s: %v", SortTrendingScore, trendingErr)
		trendingModels = []Model{} // Set empty trending models
	}

	for i := range trendingModels {
		trendingModels[i].Trending = true
	}

	// Fetch downloaded models
	downloadedModels, downloadedErr := s.fetchHuggingFaceModels(ctx, s.buildDownloadedParams())
	if downloadedErr != nil {
		// If we can't get downloaded models, but we have trending models, return just trending
		if len(trendingModels) > 0 {
			logger.Warnf("Failed to fetch downloaded models but have trending models: %v", downloadedErr)
			return trendingModels, nil
		}
		return nil, downloadedErr
	}

	// Deduplicate and combine: trending first, then remaining downloaded models
	result := DeduplicateAndCombine(trendingModels, downloadedModels)

	logger.Infof("Returning %d models (%d trending, %d downloaded)", len(result), len(trendingModels), len(downloadedModels))
	return result, nil
}

// fetchHuggingFaceModels makes the actual API call to HuggingFace
func (s *ModelSearchService) fetchHuggingFaceModels(ctx context.Context, params map[string]string) ([]Model, *api.Error) {
	logger := logging.WithCtx(runaictx.LoggerHook(ctx))

	// Build query parameters
	urlParams := url.Values{}
	for key, value := range params {
		urlParams.Add(key, value)
	}

	fullURL := fmt.Sprintf("%s?%s", s.baseURL, urlParams.Encode())

	// Debug logging
	logger.Debugf("HuggingFace API request URL: %s", fullURL)

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to create request", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "failed to connect to external service", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read the response body for more detailed error information
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, api.NewError(resp.StatusCode, "external service error", fmt.Errorf("HuggingFace API failed with status %d", resp.StatusCode))
		}
		return nil, api.NewError(resp.StatusCode, "external service error", fmt.Errorf("HuggingFace API failed with status %d: %s", resp.StatusCode, string(bodyBytes)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to read response body", err)
	}

	// Debug logging
	logger.Debugf("HuggingFace API response body: %s", string(body))

	var models []Model
	if err := json.Unmarshal(body, &models); err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to parse response", err)
	}

	logger.Debugf("HuggingFace API returned %d models", len(models))
	return models, nil
}

// DeduplicateAndCombine combines trending and downloaded models, removing duplicates
func DeduplicateAndCombine(trending, downloaded []Model) []Model {
	seen := make(map[string]bool)
	var result []Model

	// Add all trending models first (they're all unique since result is empty)
	for _, model := range trending {
		if len(result) == MaxModelsLimit {
			return result
		}
		result = append(result, model)
		seen[model.ID] = true
	}

	// Add unique downloaded models
	result = addUniqueModels(result, downloaded, seen)

	return result
}

// addUniqueModels adds unique models to the result slice up to MaxModelsLimit
func addUniqueModels(result []Model, models []Model, seen map[string]bool) []Model {
	for _, model := range models {
		if len(result) == MaxModelsLimit {
			break
		}
		if !seen[model.ID] {
			result = append(result, model)
			seen[model.ID] = true
		}
	}
	return result
}

// ConvertToAPIResponse converts Model slice to API response format
func ConvertToAPIResponse(models []Model) api.HuggingFaceModelSearchResults {
	// Initialize as empty slice instead of nil to ensure consistent JSON serialization
	resources := make([]api.HuggingFaceModelItem, 0, len(models))

	for _, model := range models {
		gated := isModelGated(model.Gated)
		resources = append(resources, api.HuggingFaceModelItem{
			Id:        model.ID,
			Downloads: model.Downloads,
			Gated:     gated,
			Trending:  model.Trending,
		})
	}

	return api.HuggingFaceModelSearchResults{
		Resources: resources,
	}
}

// isModelGated determines if a model is gated based on the gated field value
func isModelGated(gated interface{}) bool {
	if gated == nil {
		return false
	}

	// In production, HuggingFace API returns string values ("auto", "manual", or false)
	if v, ok := gated.(string); ok {
		return v != "" && v != "false"
	}

	// Handle boolean values (in case API changes or for tests)
	if v, ok := gated.(bool); ok {
		return v
	}

	return false
}
