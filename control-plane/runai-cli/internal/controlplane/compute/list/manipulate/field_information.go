/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

const (
	DefaultWidthName       = 20
	DefaultWidthScope      = 10
	DefaultWidthGPUDevices = 15
	DefaultWidthGPUMemory  = 15
	DefaultWidthCPUMemory  = 15

	columnTitleName       = "Name"
	columnTitleScope      = "Scope"
	columnTitleGPUDevices = "GPU Devices"
	columnTitleGPUMemory  = "GPU Memory"
	columnTitleCPUMemory  = "CPU Memory"
)

type FieldInformation struct {
	columnTitle string
	columnWidth int
	extractFunc func(*openapi.ComputeAsset, *ComputesAlignedRow)
	formatFunc  func(*ComputesAlignedRow) string
}

func (f FieldInformation) GetColumnTitle() string {
	return f.columnTitle
}

func (f FieldInformation) GetColumnWidth() int {
	return f.columnWidth
}

func (f FieldInformation) HasExtractFunc() bool {
	return f.extractFunc != nil
}

func extractNameData(asset *openapi.ComputeAsset, row *ComputesAlignedRow) {
	row.Name = asset.Meta.Name
}

func extractScopeData(asset *openapi.ComputeAsset, row *ComputesAlignedRow) {
	row.Scope = string(asset.Meta.Scope)
}

func formatNameValue(row *ComputesAlignedRow) string {
	return row.Name
}

func formatScopeValue(row *ComputesAlignedRow) string {
	return row.Scope
}

func formatGPUDevicesValue(row *ComputesAlignedRow) string {
	return formatGPUDevices(row.Spec)
}

func formatGPUMemoryValue(row *ComputesAlignedRow) string {
	return formatGPUMemory(row.Spec)
}

func formatCPUMemoryValue(row *ComputesAlignedRow) string {
	return formatCPUMemory(row.Spec)
}

var (
	NameField = FieldInformation{
		columnTitle: columnTitleName,
		columnWidth: DefaultWidthName,
		extractFunc: extractNameData,
		formatFunc:  formatNameValue,
	}

	ScopeField = FieldInformation{
		columnTitle: columnTitleScope,
		columnWidth: DefaultWidthScope,
		extractFunc: extractScopeData,
		formatFunc:  formatScopeValue,
	}

	GPUDevicesField = FieldInformation{
		columnTitle: columnTitleGPUDevices,
		columnWidth: DefaultWidthGPUDevices,
		formatFunc:  formatGPUDevicesValue,
	}

	GPUMemoryField = FieldInformation{
		columnTitle: columnTitleGPUMemory,
		columnWidth: DefaultWidthGPUMemory,
		formatFunc:  formatGPUMemoryValue,
	}

	CPUMemoryField = FieldInformation{
		columnTitle: columnTitleCPUMemory,
		columnWidth: DefaultWidthCPUMemory,
		formatFunc:  formatCPUMemoryValue,
	}
)

var ColumnMap = map[ComputeColumn]FieldInformation{
	ColumnKeyName:       NameField,
	ColumnKeyScope:      ScopeField,
	ColumnKeyGPUDevices: GPUDevicesField,
	ColumnKeyGPUMemory:  GPUMemoryField,
	ColumnKeyCPUMemory:  CPUMemoryField,
}
