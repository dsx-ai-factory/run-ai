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
	Inferences  []workloadsServiceClient.Inference
	Workspaces  []workloadsServiceClient.Workspace
	Trainings   []workloadsServiceClient.Training
	Distributed []workloadsServiceClient.DistributedWorkload

	// Excluded from serialization: tests snapshot this struct to compare the created workloads.
	ResourceClaimTemplateWorkloads         *workloadsServiceClient.ResourceClaimTemplateWorkloadsResponse         `json:"-"`
	ResourceClaimTemplateWorkloadsErr      *workloadsServiceClient.Error                                          `json:"-"`
	ResourceClaimTemplateWorkloadsRequests []workloadsServiceClient.InternalResourceClaimTemplateWorkloadsRequest `json:"-"`
}

func (s *SimpleMockClient) CreateWorkspace(ctx context.Context, request *workloadsServiceClient.WorkspaceCreationRequest) (*workloadsServiceClient.Workspace, *workloadsServiceClient.SubmissionErrorV2) {
	workspace := workloadsServiceClient.Workspace{
		Name:          request.Name,
		RequestedName: request.Name,
		WorkloadId:    uuid.NewString(),
		ProjectId:     request.ProjectId,
		ClusterId:     request.ClusterId,
		DesiredPhase:  workloadsServiceClient.WORKLOADDESIREDPHASE_RUNNING,
		ActualPhase:   convutils.Ref(workloadsServiceClient.PHASE_CREATING),
		Spec:          request.Spec,
	}

	s.Workspaces = append(s.Workspaces, workspace)

	return &workspace, nil
}

func (s *SimpleMockClient) CreateTraining(ctx context.Context, request *workloadsServiceClient.TrainingCreationRequest) (*workloadsServiceClient.Training, *workloadsServiceClient.SubmissionErrorV2) {
	training := workloadsServiceClient.Training{
		Name:          request.Name,
		RequestedName: request.Name,
		WorkloadId:    uuid.NewString(),
		ProjectId:     request.ProjectId,
		ClusterId:     request.ClusterId,
		DesiredPhase:  workloadsServiceClient.WORKLOADDESIREDPHASE_RUNNING,
		ActualPhase:   convutils.Ref(workloadsServiceClient.PHASE_CREATING),
		Spec:          request.Spec,
	}

	s.Trainings = append(s.Trainings, training)

	return &training, nil
}

func (s *SimpleMockClient) CreateDistributed(ctx context.Context, request *workloadsServiceClient.DistributedCreationRequest) (*workloadsServiceClient.DistributedWorkload, *workloadsServiceClient.SubmissionErrorV2) {
	distributed := workloadsServiceClient.DistributedWorkload{
		Name:          request.Name,
		RequestedName: request.Name,
		WorkloadId:    uuid.NewString(),
		ProjectId:     request.ProjectId,
		ClusterId:     request.ClusterId,
		DesiredPhase:  workloadsServiceClient.WORKLOADDESIREDPHASE_RUNNING,
		ActualPhase:   convutils.Ref(workloadsServiceClient.PHASE_CREATING),
		Spec:          request.Spec,
		MasterSpec:    request.MasterSpec,

		MasterSpecSameAsWorker: request.MasterSpecSameAsWorker,
	}

	s.Distributed = append(s.Distributed, distributed)

	return &distributed, nil
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

// SanitizeWorkloads remove from the workloads that were created all fields that have values which
// changes from test to test otherwise test will be flaky
func (s *SimpleMockClient) SanitizeWorkloads() {
	if s.Inferences != nil {
		for idx := range s.Inferences {
			s.Inferences[idx].WorkloadId = ""
			s.Inferences[idx].ClusterId = ""
			s.Inferences[idx].ProjectId = ""
			if s.Inferences[idx].Spec != nil {
				s.sanitizeAnnotations(s.Inferences[idx].Spec.Annotations)
			}
		}
	}

	if s.Workspaces != nil {
		for idx := range s.Workspaces {
			s.Workspaces[idx].WorkloadId = ""
			s.Workspaces[idx].ClusterId = ""
			s.Workspaces[idx].ProjectId = ""
			if s.Workspaces[idx].Spec != nil {
				s.sanitizeAnnotations(s.Workspaces[idx].Spec.Annotations)
			}
		}
	}

	if s.Trainings != nil {
		for idx := range s.Trainings {
			s.Trainings[idx].WorkloadId = ""
			s.Trainings[idx].ClusterId = ""
			s.Trainings[idx].ProjectId = ""
			if s.Trainings[idx].Spec != nil {
				s.sanitizeAnnotations(s.Trainings[idx].Spec.Annotations)
			}
		}
	}

	if s.Distributed != nil {
		for idx := range s.Distributed {
			s.Distributed[idx].WorkloadId = ""
			s.Distributed[idx].ClusterId = ""
			s.Distributed[idx].ProjectId = ""
			if s.Distributed[idx].Spec != nil {
				s.sanitizeAnnotations(s.Distributed[idx].Spec.Annotations)
			}
			if convutils.HasValue(s.Distributed[idx].MasterSpec) {
				s.sanitizeAnnotations(s.Distributed[idx].MasterSpec.Get().Annotations)
			}
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
	s.Workspaces = nil
	s.Trainings = nil
	s.Distributed = nil
	s.ResourceClaimTemplateWorkloads = nil
	s.ResourceClaimTemplateWorkloadsErr = nil
	s.ResourceClaimTemplateWorkloadsRequests = nil
}
