/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
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
