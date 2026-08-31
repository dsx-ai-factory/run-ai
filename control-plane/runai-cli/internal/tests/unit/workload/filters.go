/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"

	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/clients"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

const (
	Name01          = "workload-01"
	Name02          = "workload-02"
	Name03          = "workload-03"
	Name03tag       = "workload-03tag"
	Name04          = "workload-04"
	ProjectNameP301 = "p301"
	ProjectIdP301   = "301"
	clusterId       = "uuid-cluster-1"
)

var (
	workload01 = openapi.Workload{
		TenantId:    1,
		RunningPods: 0,
		Phase:       openapi.PHASE_CREATING,
		Name:        Name01,
		Id:          Name01,
		ClusterId:   clusterId,
		ProjectName: ProjectNameP301,
		ProjectId:   ProjectIdP301,
		Namespace:   fmt.Sprintf("ns-%s", ProjectIdP301),
		Type:        api.Workspace.String(),
		RequestedPods: &openapi.RequestedPods{
			Number: *openapi.NewNullableInt32(ptr.To(int32(1))),
		},
	}
	workload02 = openapi.Workload{
		TenantId:    1,
		RunningPods: 0,
		Phase:       openapi.PHASE_CREATING,
		Name:        Name02,
		Id:          Name02,
		ClusterId:   clusterId,
		ProjectName: ProjectNameP301,
		ProjectId:   ProjectIdP301,
		Namespace:   fmt.Sprintf("ns-%s", ProjectIdP301),
		Type:        api.Training.String(),
		RequestedPods: &openapi.RequestedPods{
			Number: *openapi.NewNullableInt32(ptr.To(int32(1))),
		},
	}
	workload03 = openapi.Workload{
		TenantId:    1,
		RunningPods: 0,
		Phase:       openapi.PHASE_CREATING,
		Name:        Name03,
		Id:          Name03,
		ClusterId:   clusterId,
		ProjectName: ProjectNameP301,
		ProjectId:   ProjectIdP301,
		Namespace:   fmt.Sprintf("ns-%s", ProjectIdP301),
		Type:        api.Training.String(),
		RequestedPods: &openapi.RequestedPods{
			Number: *openapi.NewNullableInt32(ptr.To(int32(1))),
		},
	}
	workload03tag = openapi.Workload{
		TenantId:    1,
		RunningPods: 0,
		Phase:       openapi.PHASE_CREATING,
		Name:        Name03,
		Id:          Name03tag,
		ClusterId:   clusterId,
		ProjectName: ProjectNameP301,
		ProjectId:   ProjectIdP301,
		Namespace:   fmt.Sprintf("ns-%s", ProjectIdP301),
		Type:        api.Workspace.String(),
		RequestedPods: &openapi.RequestedPods{
			Number: *openapi.NewNullableInt32(ptr.To(int32(1))),
		},
	}
	workload04 = openapi.Workload{
		TenantId:    1,
		RunningPods: 0,
		Phase:       openapi.PHASE_CREATING,
		Name:        Name04,
		Id:          Name04,
		ClusterId:   clusterId,
		ProjectName: ProjectNameP301,
		ProjectId:   ProjectIdP301,
		Namespace:   fmt.Sprintf("ns-%s", ProjectIdP301),
		Type:        api.Training.String(),
	}

	ListWl01N02 = []openapi.Workload{
		workload01,
		workload02,
	}

	FilterList = []clients.MockWorkloadFilter{
		{
			Filter: filter.NewBuilder().WithClusterID(clusterId).StringList(),
			Response: openapi.GetWorkloads200Response{
				Next:      nil,
				Workloads: ListWl01N02,
			},
		},
		{
			Filter: filter.NewBuilder().WithClusterID(clusterId).WithProject(ProjectNameP301).StringList(),
			Response: openapi.GetWorkloads200Response{
				Next:      nil,
				Workloads: []openapi.Workload{workload01},
			},
		},
		{
			Filter: filter.NewBuilder().
				WithClusterID(clusterId).
				WithProject(ProjectNameP301).
				AppendName("workload-01").
				AppendName("workload-02").
				StringList(),
			Response: openapi.GetWorkloads200Response{
				Next:      nil,
				Workloads: ListWl01N02,
			},
		},
		{
			Filter: filter.NewBuilder().
				WithClusterID(clusterId).
				WithProject(ProjectNameP301).
				AppendName("workload-01").
				AppendName("workload-03").
				StringList(),
			Response: openapi.GetWorkloads200Response{
				Workloads: []openapi.Workload{workload01, workload03},
			},
		},
		{
			Filter: filter.NewBuilder().
				WithClusterID(clusterId).
				WithProject(ProjectNameP301).
				AppendName("workload-04").
				StringList(),
			Response: openapi.GetWorkloads200Response{
				Workloads: []openapi.Workload{workload04},
			},
		},
		{
			Filter: filter.NewBuilder().
				WithClusterID(clusterId).
				WithProject(ProjectNameP301).
				AppendName("workload-03").
				StringList(),
			Response: openapi.GetWorkloads200Response{
				Workloads: []openapi.Workload{workload03, workload03tag},
			},
		},
	}

	BatchList = []clients.MockWorkloadBatch{
		{
			Ids:      []string{Name01, Name02},
			Response: &openapi.BatchResponse{},
			Err:      nil,
		},
		{
			Ids: []string{Name01, Name03},
			Response: &openapi.BatchResponse{
				Id: "bla",
				Failed: []openapi.BatchResponseFailureItem{
					{
						Id:      Name03,
						Code:    2343,
						Message: "You are not permitted",
					},
				},
			},
			Err: nil,
		},
		{
			Ids: []string{Name04},
			Err: fmt.Errorf("big problem"),
		},
	}
)
