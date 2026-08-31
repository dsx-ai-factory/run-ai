/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package status

import (
	"context"
	"errors"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/generated/pkg/api"
	workloadDomain "github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/workload"
)

// StatusApiService is a service that implements the logic for the ClusterAPIApiServicer
// This service should implement the business logic for every endpoint for the ClusterAPIApi API.
// Include any external packages or services that will be required by this service.
type StatusApiService struct{}

// NewStatusApiService creates a default api service
func NewStatusApiService() api.ClusterAPIAPIServicer {
	return &StatusApiService{}
}

// GetWorkloadStatus - Get the status of a workload
func (s *StatusApiService) GetWorkloadStatus(
	ctx context.Context, projectName, workloadType, workloadName, workloadFramework string,
) (api.ImplResponse, error) {
	// Retrieve the workload from the context
	workload := workloadDomain.FromContext(ctx)
	if workload == nil {
		// should never happen
		msg := "workload not found in context"
		log.Error().Msg(msg)

		return api.Response(http.StatusInternalServerError, nil), errors.New(msg)
	}

	response := api.WorkloadStatus{Status: workload.Phase}

	log.Info().Msg("GetWorkloadStatus completed successfully")

	return api.Response(http.StatusOK, response), nil
}
