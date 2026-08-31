/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"fmt"
	"strconv"
	"time"

	nodes "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/node"

	"github.com/golang/mock/gomock"
	clusterOpenapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/nodepool"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"
	cpmocks "github.com/run-ai/runai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/department"
	nodepool2 "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/nodepool"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli"
	openapi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	"k8s.io/utils/ptr"
)

const (
	ProjectOneName    = "p301"
	ProjectOneNS      = "ns-p301"
	ProjectOneID      = "301"
	ProjectTwoName    = "p302"
	ProjectTwoNS      = "ns-p302"
	ProjectTwoID      = "302"
	ClusterUuid       = "uuid-cluster-1"
	DefaultNodepool   = "default"
	AnotherPoolName   = "another-nodepool"
	DefaultDepartment = "default"
	AnotherDepartment = "another-department"
	FreeGpu           = "100"

	NilPoolName = "nil"

	GpuCount      = 321
	NodeName      = "my name is dudu"
	OtherNodeName = "my name is yaakov"
	Status        = "Ready"
	GpuType       = "jensen"
)

var ProjectNames = []string{ProjectOneName, ProjectTwoName}

var ProjectOne = openapi.Project{
	Name:               ProjectOneName,
	RequestedNamespace: *openapi.NewNullableString(ptr.To(ProjectOneNS)),
	Id:                 ProjectOneID,
	CreatedAt:          time.Unix(10, 0),
}

var ProjectTwo = openapi.Project{
	Name:               ProjectTwoName,
	RequestedNamespace: *openapi.NewNullableString(ptr.To(ProjectTwoNS)),
	Id:                 ProjectTwoID,
	CreatedAt:          time.Unix(10, 0),
}

var projList = &openapi.GetProjects200Response{
	Projects: []openapi.Project{
		ProjectOne,
		ProjectTwo,
	},
}

var createdProj = ProjectOne

func getMockNodepoolList() *clusterOpenapi.GetNodepools200Response {
	poolStatus := clusterOpenapi.NODEPOOLPHASE_READY
	emptyPoolStatus := clusterOpenapi.NODEPOOLPHASE_EMPTY
	nilPoolStatus := clusterOpenapi.NODEPOOLPHASE_UNSCHEDULABLE
	return &clusterOpenapi.GetNodepools200Response{Nodepools: []clusterOpenapi.Nodepool{
		{
			Name:  DefaultNodepool,
			Phase: &poolStatus,
			Status: &clusterOpenapi.NodepoolSyncFieldsStatus{
				GpuNetworkAccelerationDetected: ptr.To(true),
			},
			Id: "4400001",
		},
		{
			Name:  AnotherPoolName,
			Phase: &emptyPoolStatus,
			Status: &clusterOpenapi.NodepoolSyncFieldsStatus{
				GpuNetworkAccelerationDetected: ptr.To(false),
			},
			Id: "4400002",
		},
		{
			Name:  NilPoolName,
			Phase: &nilPoolStatus,
			Id:    "4400003",
		},
	}}
}

func getMockNodepoolMinimalList() *clusterOpenapi.GetNodepoolsMinimal200Response {
	return &clusterOpenapi.GetNodepoolsMinimal200Response{
		Nodepools: []clusterOpenapi.NodepoolMinimal{
			{
				Name: DefaultNodepool,
				Id:   "4400001",
			},
			{
				Name: AnotherPoolName,
				Id:   "4400002",
			},
			{
				Name: NilPoolName,
				Id:   "4400003",
			},
		},
	}
}

func getMockDepartmentList() *openapi.GetDepartments200Response {
	return &openapi.GetDepartments200Response{
		Next: nil,
		Departments: []openapi.DepartmentForList{
			{Name: DefaultDepartment},
			{Name: AnotherDepartment},
		},
	}
}

func getMockNodeList() *clusterOpenapi.Nodes {
	ns := []clusterOpenapi.Node{
		{
			Status:      Status,
			GpuInfo:     *clusterOpenapi.NewNullableGpuInfo(clusterOpenapi.NewGpuInfo(GpuType, GpuCount)),
			Name:        NodeName,
			ClusterUuid: ClusterUuid,
			NodePool:    DefaultNodepool,
		},
		{
			Status:      Status,
			GpuInfo:     *clusterOpenapi.NewNullableGpuInfo(clusterOpenapi.NewGpuInfo(GpuType, GpuCount)),
			Name:        OtherNodeName,
			ClusterUuid: ClusterUuid,
			NodePool:    DefaultNodepool,
		},
	}
	return &clusterOpenapi.Nodes{Nodes: ns}
}

func getMockTelemetryResponse(nodeNameToFreeGpu map[string]string) api.NodeTelemetriesByType {
	values := make([]clusterOpenapi.TelemetryResponseValuesInner, 0)
	for nodeName, freeGpu := range nodeNameToFreeGpu {
		values = append(values, clusterOpenapi.TelemetryResponseValuesInner{
			Value: freeGpu,
			Groups: []clusterOpenapi.TelemetryResponseValuesInnerGroupsInner{
				{
					Key:   "Node",
					Value: nodeName,
				},
			},
		})
	}

	telemetryResponse := &clusterOpenapi.TelemetryResponse{
		Type:   string(clusterOpenapi.NODETELEMETRYTYPE_FREE_GPUS),
		Values: values,
	}

	return api.NodeTelemetriesByType{clusterOpenapi.NODETELEMETRYTYPE_FREE_GPUS: telemetryResponse}
}

func MockProject(mockCtrl *gomock.Controller) {
	mockProjectLister := cpmocks.NewMockProjectLister(mockCtrl)
	mockProjectCreator := cpmocks.NewMockProjectCreator(mockCtrl)
	mockNodepoolClient := cpmocks.NewMockNodepoolClient(mockCtrl)
	mockNodeClient := cpmocks.NewMockNodeClient(mockCtrl)
	mockDepartmentClient := cpmocks.NewMockDepartmentClient(mockCtrl)

	for _, proj := range projList.Projects {
		uniqueRes := &openapi.GetProjects200Response{Projects: []openapi.Project{proj}}
		mockProjectLister.EXPECT().List(gomock.Any(), int32(2), int32(0), gomock.InAnyOrder(filter.NewBuilder().WithClusterID(ClusterUuid).AppendName(proj.Name).StringList())).
			Return(uniqueRes, nil).AnyTimes()
	}
	// unknown project name will catch empty list
	mockProjectLister.EXPECT().List(gomock.Any(), int32(2), int32(0), gomock.Any()).
		Return(&openapi.GetProjects200Response{}, nil).AnyTimes()

	mockProjectLister.EXPECT().List(gomock.Any(), int32(11), int32(0), gomock.Any()).Return(projList, nil).AnyTimes()
	mockProjectLister.EXPECT().List(gomock.Any(), int32(101), int32(0), gomock.Any()).Return(projList, nil).AnyTimes()
	mockProjectCreator.EXPECT().CreateProject(gomock.Any(), gomock.Any()).Return(&createdProj, nil).AnyTimes()

	nodepoolList := getMockNodepoolList()
	mockNodepoolClient.EXPECT().GetNodepoolList(gomock.Any(), filter.NewBuilder().WithClusterID(ClusterUuid).StringList()).Return(nodepoolList, nil).AnyTimes()

	nodepoolMinimalList := getMockNodepoolMinimalList()
	mockNodepoolClient.EXPECT().GetNodepoolMinimalList(gomock.Any(), filter.NewBuilder().WithClusterID(ClusterUuid).StringList()).Return(nodepoolMinimalList, nil).AnyTimes()

	mockNodes := getMockNodeList()

	nodeTelemetryResponse := getMockTelemetryResponse(map[string]string{NodeName: FreeGpu, OtherNodeName: strconv.Itoa(GpuCount)})
	mockNodeClient.EXPECT().GetNodesTelemetry(gomock.Any(), ClusterUuid, nodes.RequiredNodeTelemetryTypes).Return(nodeTelemetryResponse, nil).AnyTimes()

	mockNodeClient.EXPECT().GetNodeList(gomock.Any(), gomock.Any()).Return(mockNodes, nil).AnyTimes()

	departmentList := getMockDepartmentList()
	mockDepartmentClient.EXPECT().GetDepartmentList(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(departmentList, nil).AnyTimes()

	bns := nodepool2.NewBackendNodepoolService(mockNodepoolClient, mockNodeClient)
	nodepool.SetBackendNodepoolService(bns)
	project.SetLister(mockProjectLister)
	project.SetCreator(mockProjectCreator)
	department.SetClient(mockDepartmentClient)
}

func MockProjectWithList(mockCtrl *gomock.Controller, projectList []string) {
	var projArray []openapi.Project
	mockProject := cpmocks.NewMockProjectLister(mockCtrl)
	fb := filter.NewBuilder()

	projArray = append(projArray, ProjectOne, ProjectTwo)
	for _, proj := range projectList {
		projStruct := openapi.Project{Name: proj,
			RequestedNamespace: *openapi.NewNullableString(ptr.To(proj)),
			Id:                 proj,
			CreatedAt:          time.Now()}
		projArray = append(projArray, projStruct)
		uniqueRes := &openapi.GetProjects200Response{Projects: []openapi.Project{projStruct}}
		mockProject.EXPECT().List(gomock.Any(), int32(2), int32(0), gomock.InAnyOrder(fb.WithClusterID(ClusterUuid).ClearNames().AppendName(proj).StringList())).Return(uniqueRes, nil).AnyTimes()
	}
	resAll := &openapi.GetProjects200Response{Projects: projArray}

	mockProject.EXPECT().List(gomock.Any(), int32(2), int32(0), gomock.InAnyOrder(fb.WithClusterID(ClusterUuid).ClearNames().AppendName(ProjectOne.Name).StringList())).
		Return(&openapi.GetProjects200Response{Projects: []openapi.Project{ProjectOne}}, nil).AnyTimes()
	mockProject.EXPECT().List(gomock.Any(), int32(2), int32(0), gomock.InAnyOrder(fb.WithClusterID(ClusterUuid).ClearNames().AppendName(ProjectTwo.Name).StringList())).
		Return(&openapi.GetProjects200Response{Projects: []openapi.Project{ProjectTwo}}, nil).AnyTimes()
	mockProject.EXPECT().List(gomock.Any(), int32(101), int32(0), gomock.Any()).Return(resAll, nil).AnyTimes()
	project.SetLister(mockProject)
}

func ProjectSet(projectName string, envMap map[string]string) error {
	_, err := cli.RunCommand(fmt.Sprintf("project set %s", projectName), envMap)
	return err
}
