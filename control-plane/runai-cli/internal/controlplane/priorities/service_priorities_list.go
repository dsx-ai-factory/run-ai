/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package priorities

import (
	"context"
	"fmt"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

const (
	ListAPIError = "priorities API call error: %w"
)

type Lister interface {
	ListPriorities(ctx context.Context, clusterId string) (*openapi.GetWorkloadClusterPriorities200Response, error)
}

func (pls *Service) GetWorkloadClusterPriorities(ctx context.Context, clusterId string) (*openapi.GetWorkloadClusterPriorities200Response, error) {
	return pls.client.ListPriorities(ctx, clusterId)
}

func (pls *Service) GetWorkloadClusterPrioritiesBrief(ctx context.Context, clusterId string) (*Output, error) {
	list, err := pls.GetWorkloadClusterPriorities(ctx, clusterId)
	if err != nil {
		return nil, fmt.Errorf(ListAPIError, err)
	}
	return NewOutput(list)
}

type Service struct {
	client Lister
}

func NewPrioritiesService(client Lister) *Service {
	return &Service{client: client}
}
