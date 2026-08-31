/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/field"
	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

func ParseComputeColumns(columnNames []string) ([]ComputeColumn, error) {
	var columns []ComputeColumn
	for _, name := range columnNames {
		normalizedName := strings.ToLower(strings.TrimSpace(name))
		column := ComputeColumn(normalizedName)

		if _, exists := ColumnMap[column]; !exists {
			return nil, fmt.Errorf("invalid column name: %s", name)
		}
		columns = append(columns, column)
	}
	return columns, nil
}

func BriefComputeFromAsset(asset openapi.ComputeAsset) BriefCompute {
	scope := string(asset.Meta.Scope)
	if scope == "" {
		scope = "Unknown"
	}

	return BriefCompute{
		Name:       asset.Meta.Name,
		Scope:      scope,
		GPUDevices: formatGPUDevices(asset.Spec),
		GPUMemory:  formatGPUMemory(asset.Spec),
		CPUMemory:  formatCPUMemory(asset.Spec),
	}
}

func CreateAlignedComputes(computes []openapi.ComputeAsset, columns []ComputeColumn) ComputesAlignedData {
	var aligned []ComputesAlignedRow

	for _, compute := range computes {
		row := ComputesAlignedRow{
			Spec: compute.Spec,
		}
		for _, col := range columns {
			fieldInfo := ColumnMap[col]
			if fieldInfo.HasExtractFunc() {
				fieldInfo.extractFunc(&compute, &row)
			}
		}
		populateAlignedRowFields(&row, columns)
		aligned = append(aligned, row)
	}

	return ComputesAlignedData{Computes: aligned}
}

func populateAlignedRowFields(row *ComputesAlignedRow, columns []ComputeColumn) {
	for _, col := range columns {
		fieldInfo := ColumnMap[col]
		value := fieldInfo.formatFunc(row)

		switch col {
		case ColumnKeyGPUDevices:
			row.GPUDevices = value
		case ColumnKeyGPUMemory:
			row.GPUMemory = value
		case ColumnKeyCPUMemory:
			row.CPUMemory = value
		}
	}
}

func CreateTableData(computes []ComputesAlignedRow, columns []ComputeColumn) [][]string {
	var data [][]string

	for _, computeRow := range computes {
		var row []string
		for _, col := range columns {
			fieldInfo := ColumnMap[col]
			value := fieldInfo.formatFunc(&computeRow)
			if value == "" {
				value = "-"
			}
			row = append(row, value)
		}
		data = append(data, row)
	}

	return data
}

func GetColumns(requested []ComputeColumn) []ComputeColumn {
	if len(requested) > 0 {
		return requested
	}
	return GetDefaultColumns()
}

func GetDefaultColumns() []ComputeColumn {
	return DefaultColumns
}

func formatGPUDevices(spec openapi.ComputeAssetSpec) string {
	if spec.GpuDevicesRequest.IsSet() && spec.GpuDevicesRequest.Get() != nil {
		gpuDevices := *spec.GpuDevicesRequest.Get()
		if gpuDevices > 0 {
			return field.Int32OrEmpty(&gpuDevices)
		}
	}
	if spec.GpuPortionRequest.IsSet() && spec.GpuPortionRequest.Get() != nil {
		gpuPortion := *spec.GpuPortionRequest.Get()
		if gpuPortion > 0 {
			return fmt.Sprintf("%g", gpuPortion)
		}
	}
	return ""
}

func formatGPUMemory(spec openapi.ComputeAssetSpec) string {
	if spec.GpuMemoryRequest.IsSet() && spec.GpuMemoryRequest.Get() != nil {
		return *spec.GpuMemoryRequest.Get()
	}
	if spec.GpuPortionRequest.IsSet() && spec.GpuPortionRequest.Get() != nil {
		gpuPortion := *spec.GpuPortionRequest.Get()
		if gpuPortion > 0 && gpuPortion <= 1 {
			return fmt.Sprintf("%.0f%%", gpuPortion*100)
		}
	}
	return ""
}

func formatCPUMemory(spec openapi.ComputeAssetSpec) string {
	if spec.CpuMemoryRequest.IsSet() && spec.CpuMemoryRequest.Get() != nil {
		return *spec.CpuMemoryRequest.Get()
	}
	return ""
}
