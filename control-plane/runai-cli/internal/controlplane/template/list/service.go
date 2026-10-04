/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli/structs"
	wcgenerated "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

// TemplateLister defines the interface for fetching templates from the API
type TemplateLister interface {
	List(ctx context.Context, filterBy []string, includeAncestors bool, limit, offset *int32) (*wcgenerated.TemplatesListResponse, error)
}

// Service handles template list operations
type Service struct {
	lister    TemplateLister
	clusterId string
}

// NewService creates a new template list service
func NewService(lister TemplateLister, clusterId string) *Service {
	return &Service{
		lister:    lister,
		clusterId: clusterId,
	}
}

// GetListParams holds parameters for fetching template list
type GetListParams struct {
	ProjectId string
	All       bool
	Limit     *int32
	Offset    *int32
}

// GetTemplateList fetches templates and returns displayable output
func (s *Service) GetTemplateList(ctx context.Context, params GetListParams) (*Output, error) {
	var filterBy []string

	// Add project filter if not using -A flag
	var includeAncestors bool
	if !params.All && params.ProjectId != "" {
		filterBy = append(filterBy, "projectId=="+params.ProjectId)
		includeAncestors = true
	}

	// Call the thin client with explicit parameters
	response, err := s.lister.List(ctx, filterBy, includeAncestors, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}

	// Extract templates from raw API response
	templates := response.GetEntries()

	// Convert NextOffset to NextToken if available (Service layer transformation)
	var nextToken *structs.NextToken
	if pagination := response.Pagination; pagination != nil {
		if nextOffset := pagination.NextOffset; nextOffset != nil && *nextOffset > 0 {
			nextOffsetInt32 := int32(*nextOffset)
			nextToken, err = structs.OffsetToNextToken(nextOffsetInt32)
			if err != nil {
				return nil, fmt.Errorf("failed to create next token: %w", err)
			}
		}
	}

	return NewOutput(templates, nextToken)
}
