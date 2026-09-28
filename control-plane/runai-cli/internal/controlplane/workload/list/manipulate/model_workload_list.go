/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	"time"

	workloadservice "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"k8s.io/utils/ptr"
)

type AllocatedResources struct {
	Gpu float64 `json:"gpu,omitempty"`
}

type Pods struct {
	RunningPods      *int32 `json:"running,omitempty" yaml:"running,omitempty"`
	RequestedPods    *int32 `json:"requested,omitempty" yaml:"requested,omitempty"`
	RequestedPodsMin *int32 `json:"requestedMin,omitempty" yaml:"requestedMin,omitempty"`
	RequestedPodsMax *int32 `json:"requestedMax,omitempty" yaml:"requestedMax,omitempty"`
	Parallelism      *int32 `json:"parallelism,omitempty" yaml:"parallelism,omitempty"`
	Completions      *int32 `json:"completions,omitempty" yaml:"completions,omitempty"`
}

type Brief struct {
	Id                   string                         `json:"id"`
	Name                 string                         `json:"name"`
	Type                 string                         `json:"type"`
	Phase                *workloadservice.Phase         `json:"phase,omitempty"`
	Project              string                         `json:"projectName"`
	Preemptible          *bool                          `json:"preemptible,omitempty"`
	RunningPods          int32                          `json:"runningPods"`
	RequestedPods        *workloadservice.RequestedPods `json:"requestedPods,omitempty"`
	AllocatedResources   *AllocatedResources            `json:"allocatedResources,omitempty"`
	DistributedFramework *string                        `json:"distributedFramework,omitempty"`
}

type WorkloadsAlignedRow struct {
	Name                    string                 `json:"name,omitempty"`
	Type                    string                 `json:"type,omitempty"`
	DistributedFramework    *string                `json:"framework,omitempty" yaml:"framework,omitempty"`
	Status                  *workloadservice.Phase `json:"status,omitempty" yaml:"status,omitempty"`
	Project                 string                 `json:"projectName,omitempty"`
	Preemptible             *bool                  `json:"preemptible,omitempty" yaml:"preemptible,omitempty"`
	Pods                    *Pods                  `json:"pods,omitempty" yaml:"pods,omitempty"`
	GpuAllocation           float64                `json:"gpuAllocation,omitempty" yaml:"gpuAllocation,omitempty"`
	Category                string                 `json:"category,omitempty" yaml:"category,omitempty"`
	Priority                int32                  `json:"priority,omitempty" yaml:"priority,omitempty"`
	Department              string                 `json:"department,omitempty" yaml:"department,omitempty"`
	SubmittedBy             string                 `json:"submittedBy,omitempty" yaml:"submittedBy,omitempty"`
	CreatedAt               *time.Time             `json:"createdAt,omitempty" yaml:"createdAt,omitempty"`
	TotalPendingTimeSeconds int32                  `json:"totalPendingTimeSeconds,omitempty" yaml:"totalPendingTimeSeconds,omitempty"`
	AllocatedGPUMemory      string                 `json:"allocatedGPUMemory,omitempty" yaml:"allocatedGPUMemory,omitempty"`
	AllocatedCPUCompute     float64                `json:"allocatedCPUCompute,omitempty" yaml:"allocatedCPUCompute,omitempty"`
	AllocatedCPUMemory      string                 `json:"allocatedCPUMemory,omitempty" yaml:"allocatedCPUMemory,omitempty"`
	CompletedAt             *time.Time             `json:"completedAt,omitempty" yaml:"completedAt,omitempty"`
	Urls                    string                 `json:"urls,omitempty" yaml:"urls,omitempty"`
	Image                   string                 `json:"image,omitempty" yaml:"image,omitempty"`
	CurrentNodepool         string                 `json:"currentNodepool,omitempty" yaml:"currentNodepool,omitempty"`
	RequestedNodepool       string                 `json:"requestedNodepool,omitempty" yaml:"requestedNodepool,omitempty"`
}

func newBriefFromWorkload(workload workloadservice.Workload, deleted bool) Brief {
	brief := Brief{
		Id:                   workload.Id,
		Name:                 workload.Name,
		Type:                 workload.Type,
		Phase:                nil, // Will be set below if not deleted
		Project:              workload.ProjectName,
		Preemptible:          workload.Preemptible.Get(),
		RunningPods:          workload.RunningPods,
		RequestedPods:        workload.RequestedPods,
		DistributedFramework: workload.DistributedFramework,
	}

	// Only set phase if not deleted
	if !deleted {
		brief.Phase = ptr.To(workload.Phase)
	}

	if workload.AllocatedResources.IsSet() && workload.AllocatedResources.Get().HasGpu() {
		brief.AllocatedResources = &AllocatedResources{Gpu: workload.AllocatedResources.Get().GetGpu()}
	}
	return brief
}

func BriefListFromResponse(workloadList []workloadservice.Workload, deleted bool) ([]Brief, error) {
	var briefList = make([]Brief, 0)

	for _, wl := range workloadList {
		briefList = append(briefList, newBriefFromWorkload(wl, deleted))
	}

	return briefList, nil
}

// WorkloadsAlignedOutput represents workloads output with column titles as keys
type WorkloadsAlignedOutput struct {
	Workloads []WorkloadsAlignedRow `json:"workloads" yaml:"workloads"`
	NextToken *string               `json:"nextToken,omitempty" yaml:"nextToken,omitempty"`
}

type Data struct {
	Workloads []workloadservice.Workload `json:"workloads," yaml:"workloads"` // Store originals, convert when needed
	NextToken *string                    `json:"nextToken,omitempty" yaml:"nextToken,omitempty"`
}
