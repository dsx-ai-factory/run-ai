/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads_service

import (
	"context"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"

	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/google/uuid"
)

// SimpleMockClient used for testing
type SimpleMockClient struct {
	Inferences []workloadsServiceClient.Inference

	// Excluded from serialization: tests snapshot this struct to compare the created workloads.
	ResourceClaimTemplateWorkloads         *workloadsServiceClient.ResourceClaimTemplateWorkloadsResponse         `json:"-"`
	ResourceClaimTemplateWorkloadsErr      *workloadsServiceClient.Error                                          `json:"-"`
	ResourceClaimTemplateWorkloadsRequests []workloadsServiceClient.InternalResourceClaimTemplateWorkloadsRequest `json:"-"`
}

func (s *SimpleMockClient) CreateInference(ctx context.Context, request *workloadsServiceClient.InferenceCreationRequest) (*workloadsServiceClient.Inference, *workloadsServiceClient.Error) {
	inference := workloadsServiceClient.Inference{
		Name:          request.Name,
		RequestedName: request.Name,
		WorkloadId:    uuid.NewString(),
		ProjectId:     request.ProjectId,
		ClusterId:     request.ClusterId,
		DesiredPhase:  workloadsServiceClient.WORKLOADDESIREDPHASE_RUNNING,
		ActualPhase:   convutils.Ref(workloadsServiceClient.PHASE_CREATING),
		Spec:          request.Spec,
	}

	s.Inferences = append(s.Inferences, inference)

	return &inference, nil
}

// NewSimpleMockClient returns
func NewSimpleMockClient() SimpleMockClient {
	return SimpleMockClient{}
}

// InternalCreateWorkload is a mock implementation used for testing
func (s SimpleMockClient) InternalCreateWorkload(ctx context.Context, notification *workloadsServiceClient.WorkloadCreationFields) (*workloadsServiceClient.Workload, *workloadsServiceClient.Error) {
	// TODO #14092 implement some testing logic here
	return nil, nil
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

// SanitizeWorkloads clears generated values before comparing captured submissions.
func (s *SimpleMockClient) SanitizeWorkloads() {
	for idx := range s.Inferences {
		inference := &s.Inferences[idx]
		inference.WorkloadId = ""
		inference.ClusterId = ""
		inference.ProjectId = ""
		if inference.Spec != nil {
			s.sanitizeAnnotations(inference.Spec.Annotations)
		}
	}
}

func (s *SimpleMockClient) sanitizeAnnotations(annotations []workloadsServiceClient.Annotation) {
	for idx := range annotations {
		if convutils.Deref(annotations[idx].Name.Get(), "") == composer.ComplimentaryAssetInfoAnn {
			annotations[idx].Value = *workloadsServiceClient.NewNullableString(convutils.Ref(""))
			return
		}
	}
}

func (s *SimpleMockClient) Clear() {
	s.Inferences = nil
	s.ResourceClaimTemplateWorkloads = nil
	s.ResourceClaimTemplateWorkloadsErr = nil
	s.ResourceClaimTemplateWorkloadsRequests = nil
}
