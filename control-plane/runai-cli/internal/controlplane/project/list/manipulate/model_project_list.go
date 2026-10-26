/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"k8s.io/utils/ptr"
)

type BriefNodepoolAssignedGPU struct {
	Deserved float64 `json:"deserved,omitempty"`
}

type BriefNodePoolAssignedResources struct {
	Gpu BriefNodepoolAssignedGPU `json:"gpu,omitempty"`
}

type BriefQuotaStatusAllocation struct {
	Gpu float64 `json:"gpu,omitempty"`
}

type BriefQuotaStatuses struct {
	Allocated BriefQuotaStatusAllocation `json:"allocated,omitempty"`
}

type BriefProjectStatus struct {
	Phase         string               `json:"phase,omitempty"`
	QuotaStatuses []BriefQuotaStatuses `json:"nodePoolQuotaStatuses,omitempty"`
}

type BriefDepartment struct {
	DepartmentName string `json:"name"`
}

type BriefProject struct {
	Id                 string `json:"id"`
	Name               string `json:"name"`
	BriefDepartment    `json:"parent,omitempty"`
	Resources          []BriefNodePoolAssignedResources `json:"resources,omitempty"`
	BriefProjectStatus `json:"status"`
}

type ProjectsAlignedRow struct {
	Name            string  `json:"name,omitempty"`
	Status          string  `json:"status,omitempty"`
	Department      string  `json:"department,omitempty"`
	GPUQuota        float64 `json:"gpuQuota,omitempty"`
	AllocatedGPUs   float64 `json:"allocatedGPUs,omitempty"`
	AllocationRatio float64 `json:"allocationRatio,omitempty"`
}

type ProjectsAlignedOutput struct {
	Projects  []ProjectsAlignedRow `json:"projects" yaml:"projects"`
	NextToken *string              `json:"nextToken,omitempty" yaml:"nextToken,omitempty"`
}

type Data struct {
	Projects  []openapi.Project `json:"projects,omitempty" yaml:"projects"`
	NextToken *string           `json:"nextToken,omitempty" yaml:"nextToken,omitempty"`
}

func newBriefFromProject(project openapi.Project) BriefProject {
	brief := BriefProject{
		Id:   project.Id,
		Name: project.Name,
	}

	if project.Parent.IsSet() {
		brief.DepartmentName = project.Parent.Get().Name
	}

	for _, resource := range project.Resources {
		assignedResources := BriefNodePoolAssignedResources{Gpu: BriefNodepoolAssignedGPU{Deserved: resource.Gpu.Deserved}}
		brief.Resources = append(brief.Resources, assignedResources)
	}

	if project.Status.Phase != nil || len(project.Status.NodePoolQuotaStatuses) != 0 {
		brief.Phase = string(project.Status.GetPhase())
		for _, quataStatus := range project.Status.NodePoolQuotaStatuses {
			if quataStatus.GetAllocated().Gpu != nil {
				brief.QuotaStatuses = append(brief.QuotaStatuses,
					BriefQuotaStatuses{Allocated: BriefQuotaStatusAllocation{Gpu: ptr.Deref[float64](quataStatus.GetAllocated().Gpu, 0)}})
			}
		}
	}

	return brief
}

func BriefListFromResponse(projectList []openapi.Project) ([]BriefProject, error) {
	var briefList = make([]BriefProject, 0)

	for _, project := range projectList {
		briefList = append(briefList, newBriefFromProject(project))
	}

	return briefList, nil
}
