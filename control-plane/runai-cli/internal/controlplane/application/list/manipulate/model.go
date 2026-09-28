/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

type BriefAIApplication struct {
	Name         string `json:"name" yaml:"name"`
	Type         string `json:"type,omitempty" yaml:"type,omitempty"`
	ActionStatus string `json:"actionStatus,omitempty" yaml:"actionStatus,omitempty"`
	ProjectName  string `json:"projectName,omitempty" yaml:"projectName,omitempty"`
}

type AIApplicationAlignedRow struct {
	Name            string `json:"name,omitempty" yaml:"name,omitempty"`
	Type            string `json:"type,omitempty" yaml:"type,omitempty"`
	Status          string `json:"status,omitempty" yaml:"status,omitempty"`
	Project         string `json:"project,omitempty" yaml:"project,omitempty"`
	GPUComputeReq   string `json:"gpuComputeReq,omitempty" yaml:"gpuComputeReq,omitempty"`
	GPUComputeAlloc string `json:"gpuComputeAlloc,omitempty" yaml:"gpuComputeAlloc,omitempty"`
	GPUMemoryReq    string `json:"gpuMemoryReq,omitempty" yaml:"gpuMemoryReq,omitempty"`
	GPUMemoryAlloc  string `json:"gpuMemoryAlloc,omitempty" yaml:"gpuMemoryAlloc,omitempty"`
	CPUComputeReq   string `json:"cpuComputeReq,omitempty" yaml:"cpuComputeReq,omitempty"`
	CPUComputeAlloc string `json:"cpuComputeAlloc,omitempty" yaml:"cpuComputeAlloc,omitempty"`
	CPUMemoryReq    string `json:"cpuMemoryReq,omitempty" yaml:"cpuMemoryReq,omitempty"`
	CPUMemoryAlloc  string `json:"cpuMemoryAlloc,omitempty" yaml:"cpuMemoryAlloc,omitempty"`
}

type AIApplicationAlignedOutput struct {
	AIApplications []AIApplicationAlignedRow `json:"aiApplications" yaml:"aiApplications"`
	NextToken      *string                   `json:"nextToken,omitempty" yaml:"nextToken,omitempty"`
}

type Data struct {
	AIApplications []openapi.AIApplication `json:"aiApplications,omitempty" yaml:"aiApplications"`
	NextToken      *string                 `json:"nextToken,omitempty" yaml:"nextToken,omitempty"`
}

func newBriefFromAIApplication(app openapi.AIApplication) BriefAIApplication {
	brief := BriefAIApplication{
		Name:         app.GetName(),
		Type:         app.GetType(),
		ActionStatus: app.GetActionStatus(),
		ProjectName:  app.GetProjectName(),
	}
	return brief
}

func BriefListFromResponse(apps []openapi.AIApplication) []BriefAIApplication {
	result := make([]BriefAIApplication, 0, len(apps))
	for _, app := range apps {
		result = append(result, newBriefFromAIApplication(app))
	}
	return result
}
