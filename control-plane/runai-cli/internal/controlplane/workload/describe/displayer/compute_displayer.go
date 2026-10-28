/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/field"
)

// Type definitions
type Compute struct {
	GPUComputeMode     *string            `json:"gpuComputeMode,omitempty" yaml:"gpucomputemode,omitempty"`
	RequestedResources ComputeResources   `json:"requestedResources"`
	AllocatedResources AllocatedResources `json:"allocatedResources"`
	RequestedNodePools []string           `json:"requestedNodePools,omitempty"`
	CurrentNodePools   []string           `json:"currentNodePools,omitempty"`
}

type AllocatedResources struct {
	GPUCompute float64 `json:"gpuCompute"`
	CPUCores   float64 `json:"cpuCores"`
	GPUMemory  string  `json:"gpuMemory"`
	CPUMemory  string  `json:"cpuMemory"`
}

type ComputeResources struct {
	GPUCompute RequestedLimitedResources `json:"gpuCompute"`
	CPUCores   RequestedLimitedResources `json:"cpuCores"`
	GPUMemory  RequestedLimitQuantity    `json:"gpuMemory"`
	CPUMemory  RequestedLimitQuantity    `json:"cpuMemory"`
}

type RequestedLimitedResources struct {
	Requested float64  `json:"requested"`
	Limit     *float64 `json:"limit,omitempty"`
}

type RequestedLimitQuantity struct {
	Requested string `json:"request"`
	Limit     string `json:"limit"`
}

// Displayer
type ComputeDisplayer struct {
	compute Compute
}

func NewComputeSection(compute Compute) *Section {
	return &Section{
		Displayer: &ComputeDisplayer{
			compute: compute,
		},
	}
}

func (c *ComputeDisplayer) Subject() string {
	return TitleCompute
}

func (c *ComputeDisplayer) DisplayBeforeTable(indent string) {
	if mode := field.StringOrEmpty(c.compute.GPUComputeMode); mode != "" {
		fmt.Printf("%s%-*s %s\n", indent, GeneralAndRuntimeColumnPadding, FieldGPUComputeMode+":", mode) //nolint:forbidigo
	}
}

func (c *ComputeDisplayer) Columns() []table.Column {
	return []table.Column{
		{Title: "Resource", Width: ResourceColumnWidth},
		{Title: "Requested", Width: RequestedColumnWidth},
		{Title: "Limit", Width: LimitColumnWidth},
		{Title: "Allocated", Width: AllocatedColumnWidth},
	}
}

func (c *ComputeDisplayer) Data() [][]string {
	return [][]string{
		c.getResourceRow(FieldGPUCompute, c.compute.RequestedResources.GPUCompute, c.compute.AllocatedResources.GPUCompute),
		c.getResourceRow(FieldCPUCores, c.compute.RequestedResources.CPUCores, c.compute.AllocatedResources.CPUCores),
		c.getMemoryRow(FieldGPUMemory, c.compute.RequestedResources.GPUMemory, c.compute.AllocatedResources.GPUMemory),
		c.getMemoryRow(FieldCPUMemory, c.compute.RequestedResources.CPUMemory, c.compute.AllocatedResources.CPUMemory),
	}
}

func (c *ComputeDisplayer) getResourceRow(name string, requested RequestedLimitedResources, allocated float64) []string {
	return []string{
		name,
		field.Float64OrEmpty(&requested.Requested),
		field.Float64OrEmpty(requested.Limit),
		field.Float64OrEmpty(&allocated),
	}
}

func (c *ComputeDisplayer) getMemoryRow(name string, requested RequestedLimitQuantity, allocated string) []string {
	return []string{
		name,
		requested.Requested,
		requested.Limit,
		allocated,
	}
}
