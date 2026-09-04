/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"

	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/clients"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

const (
	Name01          = "workload-01"
	Name02          = "workload-02"
	Name03          = "workload-03"
	Name04          = "workload-04"
	ProjectNameP300 = "p300"
	ProjectNameP301 = "p301"
	ProjectNameP303 = "p303"
	ProjectIdP301   = "301"
	ProjectIdP303   = "303"
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
		Type:        api.Inference.String(),
		RequestedPods: &openapi.RequestedPods{
			Max: *openapi.NewNullableInt32(ptr.To(int32(1))),
			Min: *openapi.NewNullableInt32(ptr.To(int32(0))),
		},
	}
	workload02 = openapi.Workload{
		TenantId:    1,
		RunningPods: 1,
		Phase:       openapi.PHASE_CREATING,
		Name:        Name02,
		Id:          Name02,
		ClusterId:   clusterId,
		ProjectName: ProjectNameP301,
		ProjectId:   ProjectIdP301,
		Namespace:   fmt.Sprintf("ns-%s", ProjectIdP301),
		Type:        api.Inference.String(),
		RequestedPods: &openapi.RequestedPods{
			Max: *openapi.NewNullableInt32(ptr.To(int32(3))),
			Min: *openapi.NewNullableInt32(ptr.To(int32(1))),
		},
	}
	workload03 = openapi.Workload{
		TenantId:    1,
		RunningPods: 0,
		Phase:       openapi.PHASE_CREATING,
		Name:        Name03,
		Id:          Name03,
		ClusterId:   clusterId,
		ProjectName: ProjectNameP303,
		ProjectId:   ProjectIdP303,
		Namespace:   fmt.Sprintf("ns-%s", ProjectIdP303),
		Type:        api.Inference.String(),
	}
	workload04Nim = openapi.Workload{
		TenantId:    1,
		RunningPods: 2,
		Phase:       openapi.PHASE_RUNNING,
		Name:        Name04,
		Id:          Name04,
		ClusterId:   clusterId,
		ProjectName: ProjectNameP301,
		ProjectId:   ProjectIdP301,
		Namespace:   fmt.Sprintf("ns-%s", ProjectIdP301),
		Type:        api.BackendNimServiceWorkloadType,
		RequestedPods: &openapi.RequestedPods{
			Number: *openapi.NewNullableInt32(ptr.To(int32(2))),
		},
	}

	ListWl01N02N03N04 = []openapi.Workload{
		workload01,
		workload02,
		workload03,
		workload04Nim,
	}
	ListWl01N02N04 = []openapi.Workload{
		workload01,
		workload02,
		workload04Nim,
	}
	ListWl01N02N03 = []openapi.Workload{
		workload01,
		workload02,
		workload03,
	}
	ListWl01N02 = []openapi.Workload{
		workload01,
		workload02,
	}

	inferenceListFilter = api.Inference.QueryParam() + "||" + api.DistributedInference.QueryParam() + "||" + api.BackendNimServiceWorkloadType

	FilterList = []clients.MockWorkloadFilter{
		{
			// Filter: inference list (standard + distributed + NIM)
			Filter: filter.NewBuilder().WithClusterID(clusterId).WithWorkloadType(inferenceListFilter).StringList(),
			Response: openapi.GetWorkloads200Response{
				Next:      nil,
				Workloads: ListWl01N02N03N04,
			},
		},
		{
			// Updated to match the new inference list command that queries both Inference and DistributedInference types
			Filter: filter.NewBuilder().WithClusterID(clusterId).WithWorkloadType(inferenceListFilter).WithProject(ProjectNameP301).StringList(),
			Response: openapi.GetWorkloads200Response{
				Next:      nil,
				Workloads: ListWl01N02N04,
			},
		},
		{
			// Project p300 does not have workloads; returns empty list
			Filter: filter.NewBuilder().WithClusterID(clusterId).WithWorkloadType(inferenceListFilter).WithProject(ProjectNameP300).StringList(),
			Response: openapi.GetWorkloads200Response{
				Next:      nil,
				Workloads: []openapi.Workload{},
			},
		},
		{
			// Project p302 with empty workloads list (for testing empty list display)
			Filter: filter.NewBuilder().WithClusterID(clusterId).WithWorkloadType(inferenceListFilter).WithProject(clients.ProjectNameP302).StringList(),
			Response: openapi.GetWorkloads200Response{
				Next:      nil,
				Workloads: []openapi.Workload{},
			},
		},
	}
)
