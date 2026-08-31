/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	"errors"
	"fmt"
	"strconv"

	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api"
	cliutils "github.com/run-ai/runai/control-plane/runai-cli/internal/utils"
	"k8s.io/utils/ptr"
)

const (
	NotAvailableString  = "N/A"
	GpuAllocationFormat = "%d/%d"
)

func CreateAlignedNodes(nodesRaw Data, columns []NodeColumn) NodesAlignedOutput {
	nodesAligned := make([]NodesAlignedRow, 0)
	for _, node := range nodesRaw.Nodes {
		nodeRow := &NodesAlignedRow{}
		for _, column := range columns {
			ColumnMap[column].extractFunc(ptr.To(node), nodeRow)
		}
		nodesAligned = append(nodesAligned, *nodeRow)
	}
	return NodesAlignedOutput{
		Nodes:     nodesAligned,
		NextToken: nodesRaw.NextToken,
	}
}

func CreateTableData(nodesAligned []NodesAlignedRow, columns []NodeColumn) [][]string {
	var data [][]string

	for _, node := range nodesAligned {
		var nodeRow []string
		for _, column := range columns {
			nodeRow = append(nodeRow, ColumnMap[column].formatFunc(ptr.To(node)))
		}
		data = append(data, nodeRow)
	}

	// Find the index of the name column in the current column order
	nameIndex := 0
	for i, column := range columns {
		if column == ColumnKeyName {
			nameIndex = i
			break
		}
	}

	cliutils.MultiColumnSort(data, []cliutils.SortColumn{
		{Index: nameIndex, Order: cliutils.Ascending},
	})

	return data
}

// CreateAlignedNodesWithTelemetry creates aligned nodes from raw data enriched with telemetry
func CreateAlignedNodesWithTelemetry(nodesRaw Data, nodeTelemetries api.NodeTelemetriesByType, columns []NodeColumn) NodesAlignedOutput {
	// Parse telemetry data
	freeGpuByNode, err := ParseTelemetryResponse(nodeTelemetries[openapi.NODETELEMETRYTYPE_FREE_GPUS])
	if err != nil {
		// If telemetry fails, fall back to basic processing
		freeGpuByNode = make(map[string]string)
	}

	nodesAligned := make([]NodesAlignedRow, 0)
	for _, node := range nodesRaw.Nodes {
		nodeRow := &NodesAlignedRow{}

		// Extract basic node data
		for _, column := range columns {
			ColumnMap[column].extractFunc(ptr.To(node), nodeRow)
		}

		// Enrich with telemetry data
		freeGpus, ok := freeGpuByNode[node.Name]
		if !ok {
			freeGpus = NotAvailableString
		}
		nodeRow.FreeGpu = freeGpus

		// Calculate GPU allocation
		if node.GpuInfo.IsSet() {
			gpuInfo := node.GpuInfo.Get()
			allocation, _ := ConstructGpuAllocation(gpuInfo.GetGpuCount(), freeGpus)
			nodeRow.GpuAllocation = allocation
			nodeRow.GpuType = gpuInfo.GetGpuType()
		} else {
			nodeRow.GpuAllocation = NotAvailableString
			nodeRow.GpuType = NotAvailableString
		}

		nodesAligned = append(nodesAligned, *nodeRow)
	}
	return NodesAlignedOutput{
		Nodes:     nodesAligned,
		NextToken: nodesRaw.NextToken,
	}
}

// ParseTelemetryResponse parses telemetry response into a map
func ParseTelemetryResponse(response *openapi.TelemetryResponse) (map[string]string, error) {
	if response == nil {
		return make(map[string]string), nil
	}

	parsedResponse := make(map[string]string)
	for _, innerValue := range response.Values {
		if len(innerValue.Groups) == 0 {
			return nil, errors.New("failed to parse telemetry response")
		}
		parsedResponse[innerValue.Groups[0].Value] = innerValue.Value
	}
	return parsedResponse, nil
}

// ConstructGpuAllocation calculates GPU allocation string
func ConstructGpuAllocation(totalGpuCount int32, freeGpus string) (string, error) {
	if len(freeGpus) == 0 || freeGpus == NotAvailableString || totalGpuCount == 0 {
		return NotAvailableString, nil
	}

	numericFreeGpus, err := strconv.Atoi(freeGpus)
	if err != nil {
		return "", err
	}
	numericAllocatedGpus := int(totalGpuCount) - numericFreeGpus

	return fmt.Sprintf(GpuAllocationFormat, numericAllocatedGpus, totalGpuCount), nil
}
