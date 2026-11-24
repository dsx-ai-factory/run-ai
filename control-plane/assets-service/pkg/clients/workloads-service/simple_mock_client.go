/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads_service

import (
	"context"

	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

// SimpleMockClient used for testing
type SimpleMockClient struct {
	ResourceClaimTemplateWorkloads         *workloadsServiceClient.ResourceClaimTemplateWorkloadsResponse
	ResourceClaimTemplateWorkloadsErr      *workloadsServiceClient.Error
	ResourceClaimTemplateWorkloadsRequests []workloadsServiceClient.InternalResourceClaimTemplateWorkloadsRequest
}

// InternalGetResourceClaimTemplateWorkloads records the request and returns the configured response or error.
// With neither configured it returns an empty list.
func (s *SimpleMockClient) InternalGetResourceClaimTemplateWorkloads(ctx context.Context,
	request *workloadsServiceClient.InternalResourceClaimTemplateWorkloadsRequest, offset int32, limit int32,
) (*workloadsServiceClient.ResourceClaimTemplateWorkloadsResponse, *workloadsServiceClient.Error) {
	s.ResourceClaimTemplateWorkloadsRequests = append(s.ResourceClaimTemplateWorkloadsRequests, *request)
	if s.ResourceClaimTemplateWorkloadsErr != nil {
		return nil, s.ResourceClaimTemplateWorkloadsErr
	}
	if s.ResourceClaimTemplateWorkloads != nil {
		return s.ResourceClaimTemplateWorkloads, nil
	}
	return &workloadsServiceClient.ResourceClaimTemplateWorkloadsResponse{
		Workloads: []workloadsServiceClient.ResourceClaimTemplateWorkload{},
	}, nil
}

// SimpleMockClientFactory is the mock factory for testing
type SimpleMockClientFactory struct {
	*SimpleMockClient
}

// NewSimpleMockClientFactory returns a mock factory
func NewSimpleMockClientFactory() SimpleMockClientFactory {
	return SimpleMockClientFactory{
		SimpleMockClient: &SimpleMockClient{},
	}
}

// Client  returns a mock workload service client
func (f SimpleMockClientFactory) Client() ClientInterface {
	return f.SimpleMockClient
}
