/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	cpworkloadmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane/workload"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	nimclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/golang/mock/gomock"
)

// DescribeMocker sets up mocks for describe command testing
type DescribeMocker struct {
	mockCtrl             *gomock.Controller
	MockWorkloadClient   *cpworkloadmocks.MockWorkloadClient
	MockWorkloadV2Getter *cpworkloadmocks.MockWorkloadV2Getter
	MockPodsClient       *cpworkloadmocks.MockPodsClient
	MockEventClient      *cpworkloadmocks.MockEventClient
	fixtures             *DescribeFixtures
	workloadType         WorkloadType
}

// NewDescribeMocker creates a new describe mocker with the given fixtures
func NewDescribeMocker(mockCtrl *gomock.Controller, workloadType WorkloadType) (*DescribeMocker, error) {
	fixtures, err := LoadDescribeFixtures(workloadType)
	if err != nil {
		return nil, err
	}

	dm := &DescribeMocker{
		mockCtrl:     mockCtrl,
		fixtures:     fixtures,
		workloadType: workloadType,
	}
	dm.setupMocks()
	return dm, nil
}

// NewDescribeMockerWithFixtures creates a describe mocker with pre-loaded fixtures
func NewDescribeMockerWithFixtures(mockCtrl *gomock.Controller, fixtures *DescribeFixtures, workloadType WorkloadType) *DescribeMocker {
	dm := &DescribeMocker{
		mockCtrl:     mockCtrl,
		fixtures:     fixtures,
		workloadType: workloadType,
	}
	dm.setupMocks()
	return dm
}

// setupMocks sets up all the mocks needed for describe tests
func (dm *DescribeMocker) setupMocks() {
	dm.setupWorkloadClientMock()
	dm.setupPodsClientMock()
	dm.setupEventsClientMock()
}

// GetFixtures returns the loaded fixtures
func (dm *DescribeMocker) GetFixtures() *DescribeFixtures {
	return dm.fixtures
}

// GetWorkloadName returns the workload name from fixtures
func (dm *DescribeMocker) GetWorkloadName() string {
	if dm.fixtures.WorkloadDetailed != nil {
		return dm.fixtures.WorkloadDetailed.Name
	}
	return ""
}

// GetWorkloadID returns the workload ID from fixtures
func (dm *DescribeMocker) GetWorkloadID() string {
	if dm.fixtures.WorkloadDetailed != nil {
		return dm.fixtures.WorkloadDetailed.Id
	}
	return ""
}

// GetProjectName returns the project name from fixtures
func (dm *DescribeMocker) GetProjectName() string {
	if dm.fixtures.WorkloadDetailed != nil {
		return dm.fixtures.WorkloadDetailed.ProjectName
	}
	return ""
}

// GetClusterID returns the cluster ID from fixtures
func (dm *DescribeMocker) GetClusterID() string {
	if dm.fixtures.WorkloadDetailed != nil {
		return dm.fixtures.WorkloadDetailed.ClusterId
	}
	return ""
}

// CreateDescribeService creates a describe.Service with the mocked clients
func (dm *DescribeMocker) CreateDescribeService() *describe.Service {
	return describe.NewService(dm.MockWorkloadClient, dm.MockWorkloadV2Getter,
		dm.MockPodsClient, dm.MockEventClient)
}

func (dm *DescribeMocker) setupWorkloadClientMock() {
	dm.MockWorkloadClient = cpworkloadmocks.NewMockWorkloadClient(dm.mockCtrl)

	// Mock ListSlim - return the workload list from fixtures
	dm.MockWorkloadClient.EXPECT().
		ListSlim(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(client.ToSlimResponse(dm.fixtures.WorkloadList), nil).
		AnyTimes()

	// Mock GetWorkload - return the detailed workload from fixtures
	if dm.fixtures.WorkloadDetailed != nil {
		dm.MockWorkloadClient.EXPECT().
			GetWorkload(gomock.Any(), dm.fixtures.WorkloadDetailed.Id).
			Return(dm.fixtures.WorkloadDetailed, nil).
			AnyTimes()

		// Also mock with gomock.Any() for cases where ID comes from list
		dm.MockWorkloadClient.EXPECT().
			GetWorkload(gomock.Any(), gomock.Any()).
			Return(dm.fixtures.WorkloadDetailed, nil).
			AnyTimes()
	}

	// Mock type-specific getters based on workload type
	dm.setupTypeSpecificMocks()
}

func (dm *DescribeMocker) setupTypeSpecificMocks() {
	switch dm.workloadType {
	case TrainingFixture:
		if dm.fixtures.TrainingSpec != nil {
			dm.MockWorkloadClient.EXPECT().
				GetTraining(gomock.Any(), gomock.Any()).
				Return(dm.fixtures.TrainingSpec, nil).
				AnyTimes()
		} else {
			// Return empty training spec if not in fixtures
			dm.MockWorkloadClient.EXPECT().
				GetTraining(gomock.Any(), gomock.Any()).
				Return(&openapi.Training{Spec: openapi.NewTrainingSpecSpec()}, nil).
				AnyTimes()
		}
	case WorkspaceFixture:
		if dm.fixtures.WorkspaceSpec != nil {
			dm.MockWorkloadClient.EXPECT().
				GetWorkspace(gomock.Any(), gomock.Any()).
				Return(dm.fixtures.WorkspaceSpec, nil).
				AnyTimes()
		} else {
			dm.MockWorkloadClient.EXPECT().
				GetWorkspace(gomock.Any(), gomock.Any()).
				Return(&openapi.Workspace{Spec: openapi.NewWorkspaceSpecSpec()}, nil).
				AnyTimes()
		}
	case InferenceFixture:
		if dm.fixtures.InferenceSpec != nil {
			dm.MockWorkloadClient.EXPECT().
				GetInference(gomock.Any(), gomock.Any()).
				Return(dm.fixtures.InferenceSpec, nil).
				AnyTimes()
		} else {
			dm.MockWorkloadClient.EXPECT().
				GetInference(gomock.Any(), gomock.Any()).
				Return(&openapi.Inference{Spec: openapi.NewInferenceSpecSpec()}, nil).
				AnyTimes()
		}
	case DistributedFixture:
		if dm.fixtures.DistributedSpec != nil {
			dm.MockWorkloadClient.EXPECT().
				GetDistributed(gomock.Any(), gomock.Any()).
				Return(dm.fixtures.DistributedSpec, nil).
				AnyTimes()
		} else {
			dm.MockWorkloadClient.EXPECT().
				GetDistributed(gomock.Any(), gomock.Any()).
				Return(&openapi.DistributedWorkload{Spec: openapi.NewDistributedSpecSpec()}, nil).
				AnyTimes()
		}
	case DistributedInferenceFixture:
		if dm.fixtures.DistributedInfSpec != nil {
			dm.MockWorkloadClient.EXPECT().
				GetDistributedInference(gomock.Any(), gomock.Any()).
				Return(dm.fixtures.DistributedInfSpec, nil).
				AnyTimes()
		} else {
			dm.MockWorkloadClient.EXPECT().
				GetDistributedInference(gomock.Any(), gomock.Any()).
				Return(&openapi.DistributedInference{Spec: openapi.NewDistributedInferenceSpecSpec()}, nil).
				AnyTimes()
		}
	case NimFixture, NimAutoscaleFixture, NimMultiNodeFixture:
		if dm.fixtures.NimSpec != nil {
			dm.MockWorkloadClient.EXPECT().
				GetNimService(gomock.Any(), gomock.Any()).
				Return(dm.fixtures.NimSpec, nil).
				AnyTimes()
		} else {
			dm.MockWorkloadClient.EXPECT().
				GetNimService(gomock.Any(), gomock.Any()).
				Return(&nimclient.NimServiceResponse{}, nil).
				AnyTimes()
		}
	}
}

func (dm *DescribeMocker) setupPodsClientMock() {
	dm.MockPodsClient = cpworkloadmocks.NewMockPodsClient(dm.mockCtrl)

	dm.MockPodsClient.EXPECT().
		ListByLimit(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(dm.fixtures.Pods, nil).
		AnyTimes()
}

func (dm *DescribeMocker) setupEventsClientMock() {
	dm.MockEventClient = cpworkloadmocks.NewMockEventClient(dm.mockCtrl)

	dm.MockEventClient.EXPECT().
		ListByLimit(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(dm.fixtures.Events, nil).
		AnyTimes()
}
