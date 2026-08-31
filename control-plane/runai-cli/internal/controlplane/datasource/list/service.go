/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"context"
	"fmt"

	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

const (
	APIError = "datasource API call error: %w"
)

type Service struct {
	client Lister
}

type Lister interface {
	List(ctx context.Context, name *string, projectID *int32, clusterID *string, dsType *string) (*openapi.DatasourceListResponse, error)
}

func NewService(client Lister) *Service {
	return &Service{client: client}
}

func (s *Service) List(ctx context.Context, projectID *int32, clusterID *string, dsType *string) (*Output, error) {
	list, err := s.client.List(ctx, nil, projectID, clusterID, dsType)
	if err != nil {
		return nil, fmt.Errorf(APIError, err)
	}
	return ParseListResponse(list)
}
