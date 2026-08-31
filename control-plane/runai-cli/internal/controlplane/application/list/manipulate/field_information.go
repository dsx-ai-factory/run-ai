/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	"fmt"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

var ColumnMap = map[AIApplicationColumn]FieldInformation{
	ColumnKeyName:            NameField,
	ColumnKeyType:            TypeField,
	ColumnKeyStatus:          StatusField,
	ColumnKeyProject:         ProjectField,
	ColumnKeyGPUComputeReq:   GPUComputeReqField,
	ColumnKeyGPUComputeAlloc: GPUComputeAllocField,
	ColumnKeyGPUMemoryReq:    GPUMemoryReqField,
	ColumnKeyGPUMemoryAlloc:  GPUMemoryAllocField,
	ColumnKeyCPUComputeReq:   CPUComputeReqField,
	ColumnKeyCPUComputeAlloc: CPUComputeAllocField,
	ColumnKeyCPUMemoryReq:    CPUMemoryReqField,
	ColumnKeyCPUMemoryAlloc:  CPUMemoryAllocField,
}

var defaultColumns = []AIApplicationColumn{
	ColumnKeyName,
	ColumnKeyType,
	ColumnKeyStatus,
	ColumnKeyProject,
	ColumnKeyGPUComputeReq,
	ColumnKeyGPUComputeAlloc,
	ColumnKeyGPUMemoryReq,
	ColumnKeyGPUMemoryAlloc,
}

func GetDefaultColumns() []AIApplicationColumn {
	return defaultColumns
}

type FieldInformation struct {
	columnTitle string
	extractFunc func(app *openapi.AIApplication, row *AIApplicationAlignedRow)
	formatFunc  func(row *AIApplicationAlignedRow) string
	columnWidth int
}

func (f FieldInformation) GetColumnWidth() int {
	return f.columnWidth
}

func (f FieldInformation) GetColumnTitle() string {
	return f.columnTitle
}

var (
	NameField = FieldInformation{
		columnTitle: "AI Application",
		columnWidth: DefaultWidthName,
		extractFunc: func(app *openapi.AIApplication, row *AIApplicationAlignedRow) {
			row.Name = app.GetName()
		},
		formatFunc: func(row *AIApplicationAlignedRow) string { return row.Name },
	}
	TypeField = FieldInformation{
		columnTitle: "Type",
		columnWidth: DefaultWidthType,
		extractFunc: func(app *openapi.AIApplication, row *AIApplicationAlignedRow) {
			row.Type = app.GetType()
		},
		formatFunc: func(row *AIApplicationAlignedRow) string {
			if row.Type == "" {
				return noDataFormat
			}
			return row.Type
		},
	}
	StatusField = FieldInformation{
		columnTitle: "Status",
		columnWidth: DefaultWidthStatus,
		extractFunc: func(app *openapi.AIApplication, row *AIApplicationAlignedRow) {
			row.Status = app.GetActionStatus()
		},
		formatFunc: func(row *AIApplicationAlignedRow) string {
			if row.Status == "" {
				return noDataFormat
			}
			return row.Status
		},
	}
	ProjectField = FieldInformation{
		columnTitle: "Project",
		columnWidth: DefaultWidthProject,
		extractFunc: func(app *openapi.AIApplication, row *AIApplicationAlignedRow) {
			row.Project = app.GetProjectName()
		},
		formatFunc: func(row *AIApplicationAlignedRow) string {
			if row.Project == "" {
				return noDataFormat
			}
			return row.Project
		},
	}
	GPUComputeReqField = FieldInformation{
		columnTitle: "GPU Compute Req",
		columnWidth: DefaultWidthGPUCompute,
		extractFunc: func(app *openapi.AIApplication, row *AIApplicationAlignedRow) {
			if app.RequestedResources.IsSet() && app.RequestedResources.Get() != nil {
				req := app.RequestedResources.Get()
				if req.Gpu.IsSet() && req.Gpu.Get() != nil && req.Gpu.Get().HasRequest() {
					row.GPUComputeReq = fmt.Sprintf("%.2f", req.Gpu.Get().GetRequest())
				}
			}
		},
		formatFunc: func(row *AIApplicationAlignedRow) string {
			if row.GPUComputeReq == "" {
				return noDataFormat
			}
			return row.GPUComputeReq
		},
	}
	GPUComputeAllocField = FieldInformation{
		columnTitle: "GPU Compute Alloc",
		columnWidth: DefaultWidthGPUCompute,
		extractFunc: func(app *openapi.AIApplication, row *AIApplicationAlignedRow) {
			if app.AllocatedResources.IsSet() && app.AllocatedResources.Get() != nil {
				alloc := app.AllocatedResources.Get()
				if alloc.Gpu.IsSet() && alloc.Gpu.Get() != nil {
					row.GPUComputeAlloc = fmt.Sprintf("%.2f", *alloc.Gpu.Get())
				}
			}
		},
		formatFunc: func(row *AIApplicationAlignedRow) string {
			if row.GPUComputeAlloc == "" {
				return noDataFormat
			}
			return row.GPUComputeAlloc
		},
	}
	GPUMemoryReqField = FieldInformation{
		columnTitle: "GPU Memory Req",
		columnWidth: DefaultWidthGPUMemory,
		extractFunc: func(app *openapi.AIApplication, row *AIApplicationAlignedRow) {
			if app.RequestedResources.IsSet() && app.RequestedResources.Get() != nil {
				req := app.RequestedResources.Get()
				if req.GpuMemory.IsSet() && req.GpuMemory.Get() != nil {
					row.GPUMemoryReq = req.GpuMemory.Get().GetRequest()
				}
			}
		},
		formatFunc: func(row *AIApplicationAlignedRow) string {
			if row.GPUMemoryReq == "" {
				return noDataFormat
			}
			return row.GPUMemoryReq
		},
	}
	GPUMemoryAllocField = FieldInformation{
		columnTitle: "GPU Memory Alloc",
		columnWidth: DefaultWidthGPUMemory,
		extractFunc: func(app *openapi.AIApplication, row *AIApplicationAlignedRow) {
			if app.AllocatedResources.IsSet() && app.AllocatedResources.Get() != nil {
				alloc := app.AllocatedResources.Get()
				if alloc.GpuMemory.IsSet() && alloc.GpuMemory.Get() != nil {
					row.GPUMemoryAlloc = *alloc.GpuMemory.Get()
				}
			}
		},
		formatFunc: func(row *AIApplicationAlignedRow) string {
			if row.GPUMemoryAlloc == "" {
				return noDataFormat
			}
			return row.GPUMemoryAlloc
		},
	}
	CPUComputeReqField = FieldInformation{
		columnTitle: "CPU Compute Req",
		columnWidth: DefaultWidthCPUCompute,
		extractFunc: func(app *openapi.AIApplication, row *AIApplicationAlignedRow) {
			if app.RequestedResources.IsSet() && app.RequestedResources.Get() != nil {
				req := app.RequestedResources.Get()
				if req.Cpu.IsSet() && req.Cpu.Get() != nil && req.Cpu.Get().HasRequest() {
					row.CPUComputeReq = fmt.Sprintf("%.2f", req.Cpu.Get().GetRequest())
				}
			}
		},
		formatFunc: func(row *AIApplicationAlignedRow) string {
			if row.CPUComputeReq == "" {
				return noDataFormat
			}
			return row.CPUComputeReq
		},
	}
	CPUComputeAllocField = FieldInformation{
		columnTitle: "CPU Compute Alloc",
		columnWidth: DefaultWidthCPUCompute,
		extractFunc: func(app *openapi.AIApplication, row *AIApplicationAlignedRow) {
			if app.AllocatedResources.IsSet() && app.AllocatedResources.Get() != nil {
				alloc := app.AllocatedResources.Get()
				if alloc.Cpu.IsSet() && alloc.Cpu.Get() != nil {
					row.CPUComputeAlloc = fmt.Sprintf("%.2f", *alloc.Cpu.Get())
				}
			}
		},
		formatFunc: func(row *AIApplicationAlignedRow) string {
			if row.CPUComputeAlloc == "" {
				return noDataFormat
			}
			return row.CPUComputeAlloc
		},
	}
	CPUMemoryReqField = FieldInformation{
		columnTitle: "CPU Memory Req",
		columnWidth: DefaultWidthCPUMemory,
		extractFunc: func(app *openapi.AIApplication, row *AIApplicationAlignedRow) {
			if app.RequestedResources.IsSet() && app.RequestedResources.Get() != nil {
				req := app.RequestedResources.Get()
				if req.CpuMemory.IsSet() && req.CpuMemory.Get() != nil {
					row.CPUMemoryReq = req.CpuMemory.Get().GetRequest()
				}
			}
		},
		formatFunc: func(row *AIApplicationAlignedRow) string {
			if row.CPUMemoryReq == "" {
				return noDataFormat
			}
			return row.CPUMemoryReq
		},
	}
	CPUMemoryAllocField = FieldInformation{
		columnTitle: "CPU Memory Alloc",
		columnWidth: DefaultWidthCPUMemory,
		extractFunc: func(app *openapi.AIApplication, row *AIApplicationAlignedRow) {
			if app.AllocatedResources.IsSet() && app.AllocatedResources.Get() != nil {
				alloc := app.AllocatedResources.Get()
				if alloc.CpuMemory.IsSet() && alloc.CpuMemory.Get() != nil {
					row.CPUMemoryAlloc = *alloc.CpuMemory.Get()
				}
			}
		},
		formatFunc: func(row *AIApplicationAlignedRow) string {
			if row.CPUMemoryAlloc == "" {
				return noDataFormat
			}
			return row.CPUMemoryAlloc
		},
	}
)
