/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

const (
	APIError = "workload type API call error: %w"
)

type Service struct {
	client Lister
}

type Lister interface {
	List(ctx context.Context, filterBy []string, externalTypesOnly *bool) (*openapi.ListWorkloadTypes200Response, error)
}

func NewService(client Lister) *Service {
	return &Service{client: client}
}

func (s *Service) List(ctx context.Context, cpUrl, token string, fb *filter.Builder, externalTypesOnly *bool) (*Output, error) {
	list, err := s.client.List(ctx, fb.StringList(), externalTypesOnly)
	if err != nil {
		return nil, fmt.Errorf(APIError, err)
	}

	return ParseListResponse(list.Types)
}
