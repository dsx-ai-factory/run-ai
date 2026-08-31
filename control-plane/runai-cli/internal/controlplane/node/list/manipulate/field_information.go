/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	"fmt"
	"strings"

	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
)

var ColumnMap = map[NodeColumn]FieldInformation{
	ColumnKeyName:          NameField,
	ColumnKeyStatus:        StatusField,
	ColumnKeyGpuType:       GpuTypeField,
	ColumnKeyGpuAllocation: GpuAllocationField,
	ColumnKeyFreeGpu:       FreeGpuField,
	ColumnKeyNodepool:      NodepoolField,
}

var defaultColumns = []NodeColumn{
	ColumnKeyName,
	ColumnKeyStatus,
	ColumnKeyGpuType,
	ColumnKeyGpuAllocation,
	ColumnKeyFreeGpu,
	ColumnKeyNodepool,
}

func GetDefaultColumns() []NodeColumn {
	return defaultColumns
}

type FieldInformation struct {
	columnTitle string
	extractFunc func(node *openapi.Node, nodeRow *NodesAlignedRow)
	formatFunc  func(nodeRow *NodesAlignedRow) string
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
		columnTitle: "Name",
		columnWidth: DefaultWidthName,
		extractFunc: extractNameData,
		formatFunc:  formatNameValue,
	}
	StatusField = FieldInformation{
		columnTitle: "Status",
		columnWidth: DefaultWidthStatus,
		extractFunc: extractStatusData,
		formatFunc:  formatStatusValue,
	}
	GpuTypeField = FieldInformation{
		columnTitle: "GPU Type",
		columnWidth: DefaultWidthGpuType,
		extractFunc: extractGpuTypeData,
		formatFunc:  formatGpuTypeValue,
	}
	GpuAllocationField = FieldInformation{
		columnTitle: "Allocated/Total GPUs",
		columnWidth: DefaultWidthGpuAllocation,
		extractFunc: extractGpuAllocationData,
		formatFunc:  formatGpuAllocationValue,
	}
	FreeGpuField = FieldInformation{
		columnTitle: "Free GPUs",
		columnWidth: DefaultWidthFreeGpu,
		extractFunc: extractFreeGpuData,
		formatFunc:  formatFreeGpuValue,
	}
	NodepoolField = FieldInformation{
		columnTitle: "Nodepool",
		columnWidth: DefaultWidthNodepool,
		extractFunc: extractNodepoolData,
		formatFunc:  formatNodepoolValue,
	}
)

func GetColumns(columns []NodeColumn) []NodeColumn {
	if len(columns) != 0 {
		return columns
	}
	return defaultColumns
}

func ParseNodeColumns(columns []string) ([]NodeColumn, error) {
	if len(columns) == 0 {
		return nil, nil
	}

	result := make([]NodeColumn, 0, len(columns))
	var invalidColumns []string

	for _, columnName := range columns {
		trimmedKey := NodeColumn(strings.ToLower(strings.TrimSpace(columnName)))

		if _, exists := ColumnMap[trimmedKey]; exists {
			result = append(result, trimmedKey)
		} else {
			invalidColumns = append(invalidColumns, fmt.Sprintf("\"%s\"", columnName))
		}
	}

	if len(invalidColumns) > 0 {
		return nil, fmt.Errorf(consts.ErrorInvalidColumns,
			strings.Join(invalidColumns, ", "))
	}

	return result, nil
}

func extractNameData(node *openapi.Node, nodeRow *NodesAlignedRow) {
	nodeRow.Name = node.Name
}

func extractStatusData(node *openapi.Node, nodeRow *NodesAlignedRow) {
	nodeRow.Status = node.Status
}

func extractGpuTypeData(node *openapi.Node, nodeRow *NodesAlignedRow) {
	if node.GpuInfo.IsSet() {
		nodeRow.GpuType = node.GpuInfo.Get().GetGpuType()
	}
}

func extractGpuAllocationData(node *openapi.Node, nodeRow *NodesAlignedRow) {
	// This will be set during enrichment with telemetry data
	nodeRow.GpuAllocation = ""
}

func extractFreeGpuData(node *openapi.Node, nodeRow *NodesAlignedRow) {
	// This will be set during enrichment with telemetry data
	nodeRow.FreeGpu = ""
}

func extractNodepoolData(node *openapi.Node, nodeRow *NodesAlignedRow) {
	nodeRow.Nodepool = node.NodePool
}

func formatNameValue(nodeRow *NodesAlignedRow) string {
	return nodeRow.Name
}

func formatStatusValue(nodeRow *NodesAlignedRow) string {
	return nodeRow.Status
}

func formatGpuTypeValue(nodeRow *NodesAlignedRow) string {
	if nodeRow.GpuType != "" {
		return nodeRow.GpuType
	}
	return noDataFormat
}

func formatGpuAllocationValue(nodeRow *NodesAlignedRow) string {
	if nodeRow.GpuAllocation != "" {
		return nodeRow.GpuAllocation
	}
	return noDataFormat
}

func formatFreeGpuValue(nodeRow *NodesAlignedRow) string {
	if nodeRow.FreeGpu != "" {
		return nodeRow.FreeGpu
	}
	return noDataFormat
}

func formatNodepoolValue(nodeRow *NodesAlignedRow) string {
	if nodeRow.Nodepool != "" {
		return nodeRow.Nodepool
	}
	return noDataFormat
}
