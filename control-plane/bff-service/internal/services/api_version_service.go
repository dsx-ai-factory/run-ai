/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/internal/generated/pkg/api"
)

// VersionAPIService is a service that implements the logic for the VersionAPIServicer
// This service should implement the business logic for every endpoint for the VersionAPI API.
// Include any external packages or services that will be required by this service.
type VersionAPIService struct {
	version string
}

// NewVersionAPIService creates a default api service
func NewVersionAPIService(version string) *VersionAPIService {
	return &VersionAPIService{version: version}
}

// GetVersion - get version
func (s *VersionAPIService) GetVersion(ctx context.Context) (api.ImplResponse, error) {
	return api.Response(http.StatusOK, api.VersionResponse{Version: s.version}), nil
}
