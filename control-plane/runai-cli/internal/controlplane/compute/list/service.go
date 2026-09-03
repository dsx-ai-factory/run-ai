/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"context"
	"fmt"

	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
)

const (
	APIError = "compute API call error: %w"
)

type Service struct {
	client Lister
}

type Lister interface {
	List(ctx context.Context, projectID *int32, clusterID *string) (*openapi.ComputeListResponse, error)
}

func NewService(client Lister) *Service {
	return &Service{client: client}
}

func (cs *Service) List(ctx context.Context, projectID *int32, clusterID *string) (*Output, error) {
	list, err := cs.client.List(ctx, projectID, clusterID)
	if err != nil {
		return nil, fmt.Errorf(APIError, err)
	}
	return ParseListResponse(list)
}
