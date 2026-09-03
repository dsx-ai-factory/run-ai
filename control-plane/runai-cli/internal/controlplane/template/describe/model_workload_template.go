/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe

import (
	"time"
)

type Template struct {
	Meta       TemplateMeta  `json:"meta" yaml:"meta"`
	Spec       *TemplateSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
	MasterSpec *MasterSpec   `json:"masterSpec,omitempty" yaml:"masterspec,omitempty"`
	Raw        interface{}   `json:"-" yaml:"-"` // Raw API response for JSON/YAML output
}

type TemplateMeta struct {
	ID           string     `json:"id" yaml:"id"`
	Name         string     `json:"name" yaml:"name"`
	Description  *string    `json:"description,omitempty" yaml:"description,omitempty"`
	Type         string     `json:"type" yaml:"type"`
	Scope        string     `json:"scope" yaml:"scope"`
	ScopeID      string     `json:"scopeId" yaml:"scopeid"`
	ClusterID    *string    `json:"clusterId,omitempty" yaml:"clusterid,omitempty"`
	ProjectID    *int32     `json:"projectId,omitempty" yaml:"projectid,omitempty"`
	DepartmentID *string    `json:"departmentId,omitempty" yaml:"departmentid,omitempty"`
	CreatedAt    time.Time  `json:"createdAt" yaml:"createdat"`
	CreatedBy    string     `json:"createdBy" yaml:"createdby"`
	UpdatedAt    *time.Time `json:"updatedAt,omitempty" yaml:"updatedat,omitempty"`
	UpdatedBy    *string    `json:"updatedBy,omitempty" yaml:"updatedby,omitempty"`
}

type TemplateSpec struct {
	Image                *string           `json:"image,omitempty" yaml:"image,omitempty"`
	Command              *string           `json:"command,omitempty" yaml:"command,omitempty"`
	Args                 *string           `json:"args,omitempty" yaml:"args,omitempty"`
	WorkingDir           *string           `json:"workingDir,omitempty" yaml:"workingdir,omitempty"`
	Preemptible          *bool             `json:"preemptible,omitempty" yaml:"preemptible,omitempty"`
	PriorityClass        *string           `json:"priorityClass,omitempty" yaml:"priorityclass,omitempty"`
	NodePools            []string          `json:"nodePools,omitempty" yaml:"nodepools,omitempty"`
	EnvironmentVariables map[string]string `json:"environmentVariables,omitempty" yaml:"environmentvariables,omitempty"`
	Compute              *Compute          `json:"compute,omitempty" yaml:"compute,omitempty"`
	Ports                []Port            `json:"ports,omitempty" yaml:"ports,omitempty"`
	ExposedUrls          []ExposedUrl      `json:"exposedUrls,omitempty" yaml:"exposedurls,omitempty"`
}

type Compute struct {
	GpuDevicesRequest *int32   `json:"gpuDevicesRequest,omitempty" yaml:"gpudevicesrequest,omitempty"`
	GpuMemoryRequest  *string  `json:"gpuMemoryRequest,omitempty" yaml:"gpumemoryrequest,omitempty"`
	GpuMemoryLimit    *string  `json:"gpuMemoryLimit,omitempty" yaml:"gpumemorylimit,omitempty"`
	CpuCoreRequest    *float64 `json:"cpuCoreRequest,omitempty" yaml:"cpucorerequest,omitempty"`
	CpuCoreLimit      *float64 `json:"cpuCoreLimit,omitempty" yaml:"cpucorelimit,omitempty"`
	CpuMemoryRequest  *string  `json:"cpuMemoryRequest,omitempty" yaml:"cpumemoryrequest,omitempty"`
	CpuMemoryLimit    *string  `json:"cpuMemoryLimit,omitempty" yaml:"cpumemorylimit,omitempty"`
}

// Port configuration
type Port struct {
	Container int32  `json:"container" yaml:"container"`
	External  *int32 `json:"external,omitempty" yaml:"external,omitempty"`
}

// ExposedUrl configuration
type ExposedUrl struct {
	Url  string `json:"url" yaml:"url"`
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

// MasterSpec contains master-specific configuration for distributed workloads
type MasterSpec struct {
	Image                *string           `json:"image,omitempty" yaml:"image,omitempty"`
	Command              *string           `json:"command,omitempty" yaml:"command,omitempty"`
	Args                 *string           `json:"args,omitempty" yaml:"args,omitempty"`
	WorkingDir           *string           `json:"workingDir,omitempty" yaml:"workingdir,omitempty"`
	NodePools            []string          `json:"nodePools,omitempty" yaml:"nodepools,omitempty"`
	EnvironmentVariables map[string]string `json:"environmentVariables,omitempty" yaml:"environmentvariables,omitempty"`
	Compute              *Compute          `json:"compute,omitempty" yaml:"compute,omitempty"`
	Ports                []Port            `json:"ports,omitempty" yaml:"ports,omitempty"`
	ExposedUrls          []ExposedUrl      `json:"exposedUrls,omitempty" yaml:"exposedurls,omitempty"`
}
