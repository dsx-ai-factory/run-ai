/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"

	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/config"
	hf "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/huggingface"
)

// InferencesApiService is a service that implements the logic for the InferencesAPIServicer
// This service should implement the business logic for every endpoint for the InferencseApi API.
// Include any external packages or services that will be required by this service.
type InferencesApiService struct {
	hfModelInfoApi                         hf.ModelInfoApiInterface
	semaphoreConcurrentHuggingFaceRequests chan struct{} // a semaphore that limits the number of concurrent requests to the model hugging face info api
}

// NewInferencesApiService creates a default api service
func NewInferencesApiService() api.InferencesAPIServicer {
	hfc := runai_config.Get[*config.HuggingFaceConfig]()
	hfModelApiInterface := hf.NewModelInfoApi(
		hfc.ModelInfoUrl,
		hfc.ModelFileUrl,
	)

	// use mock object if requested (for testing)
	if hfc.ModelInfoUrl == "/mock" {
		hfModelApiInterface = hf.NewModelInfoApiMock()
	}

	return &InferencesApiService{
		hfModelInfoApi:                         hfModelApiInterface,
		semaphoreConcurrentHuggingFaceRequests: make(chan struct{}, hfc.MaxConcurrentRequests),
	}
}

func (s *InferencesApiService) GetHuggingFaceModelInfo(ctx context.Context, huggingFaceModelInfoRequest api.HuggingFaceModelInfoRequest) (api.ImplResponse, error) {
	// we limit the max concurrent requests of this type to avoid overloading the hugging face api.
	// the limit is enforced by a semaphore and is configured in the config

	// try to acquire semaphore without blocking. instead, we return an error if the max concurrency is reached.
	select {
	case s.semaphoreConcurrentHuggingFaceRequests <- struct{}{}:
		defer func() { <-s.semaphoreConcurrentHuggingFaceRequests }() // release semaphore when done
		modelInfo, err := s.hfModelInfoApi.GetModelInfo(ctx, huggingFaceModelInfoRequest)
		if err != nil {
			return NewErrorResponse(err), nil
		}

		return api.Response(http.StatusOK, modelInfo), nil
	default:
		return NewErrorResponse(api.NewError(http.StatusServiceUnavailable, "too many requests", nil)), nil
	}
}
