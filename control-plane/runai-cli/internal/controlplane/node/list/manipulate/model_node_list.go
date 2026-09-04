/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	"fmt"
	"strconv"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
)

type BriefNode struct {
	Name          string          `json:"name"`
	Status        string          `json:"status"`
	GpuInfo       openapi.GpuInfo `json:"gpuInfo,omitempty"`
	GpuAllocation string          `json:"gpuAllocation,omitempty"`
	FreeGpu       string          `json:"freeGpu,omitempty"`
	Nodepool      string          `json:"nodePool"`
}

type BriefData struct {
	Nodes     []BriefNode `json:"nodes,omitempty" yaml:"nodes"`
	NextToken *string     `json:"nextToken,omitempty" yaml:"nextToken,omitempty"`
}

type NodesAlignedRow struct {
	Name          string `json:"name,omitempty"`
	Status        string `json:"status,omitempty"`
	GpuType       string `json:"gpuType,omitempty"`
	GpuAllocation string `json:"gpuAllocation,omitempty"`
	FreeGpu       string `json:"freeGpu,omitempty"`
	Nodepool      string `json:"nodepool,omitempty"`
}

type NodesAlignedOutput struct {
	Nodes     []NodesAlignedRow `json:"nodes" yaml:"nodes"`
	NextToken *string           `json:"nextToken,omitempty" yaml:"nextToken,omitempty"`
}

type Data struct {
	Nodes     []openapi.Node `json:"nodes,omitempty" yaml:"nodes"`
	NextToken *string        `json:"nextToken,omitempty" yaml:"nextToken,omitempty"`
}

func newBriefFromNode(node openapi.Node) BriefNode {
	brief := BriefNode{
		Name:   node.Name,
		Status: node.Status,
	}

	if node.GpuInfo.IsSet() {
		gpuInfo := node.GpuInfo.Get()
		brief.GpuInfo = openapi.GpuInfo{
			GpuType:  gpuInfo.GetGpuType(),
			GpuCount: gpuInfo.GetGpuCount(),
		}
	}

	brief.Nodepool = node.NodePool

	return brief
}

// BriefListFromResponseWithTelemetry creates brief nodes with telemetry enrichment
func BriefListFromResponseWithTelemetry(nodeList []openapi.Node, nodeTelemetries map[openapi.NodeTelemetryType]*openapi.TelemetryResponse) []BriefNode {
	freeGpuByNode := parseFreeGpuTelemetry(nodeTelemetries)

	briefList := make([]BriefNode, 0, len(nodeList))
	for _, node := range nodeList {
		brief := newBriefFromNode(node)
		enrichBriefWithTelemetry(&brief, node, freeGpuByNode)
		briefList = append(briefList, brief)
	}

	return briefList
}

// parseFreeGpuTelemetry extracts free GPU data from telemetry response
func parseFreeGpuTelemetry(nodeTelemetries map[openapi.NodeTelemetryType]*openapi.TelemetryResponse) map[string]string {
	freeGpuByNode := make(map[string]string)

	freeGpuTelemetry, exists := nodeTelemetries[openapi.NODETELEMETRYTYPE_FREE_GPUS]
	if !exists || freeGpuTelemetry == nil {
		return freeGpuByNode
	}

	for _, innerValue := range freeGpuTelemetry.Values {
		if len(innerValue.Groups) > 0 {
			freeGpuByNode[innerValue.Groups[0].Value] = innerValue.Value
		}
	}

	return freeGpuByNode
}

// enrichBriefWithTelemetry adds telemetry data to a brief node
func enrichBriefWithTelemetry(brief *BriefNode, node openapi.Node, freeGpuByNode map[string]string) {
	// Set free GPU count
	freeGpus := freeGpuByNode[node.Name]
	if freeGpus == "" {
		freeGpus = NotAvailableString
	}
	brief.FreeGpu = freeGpus

	// Calculate GPU allocation
	brief.GpuAllocation = calculateGpuAllocation(node, freeGpus)
}

// calculateGpuAllocation computes the allocated/total GPU string
func calculateGpuAllocation(node openapi.Node, freeGpus string) string {
	if !node.GpuInfo.IsSet() || freeGpus == NotAvailableString {
		return NotAvailableString
	}

	gpuInfo := node.GpuInfo.Get()
	totalGpus := int(gpuInfo.GetGpuCount())

	freeGpusInt, err := strconv.Atoi(freeGpus)
	if err != nil {
		return NotAvailableString
	}

	allocatedGpus := totalGpus - freeGpusInt
	return fmt.Sprintf(GpuAllocationFormat, allocatedGpus, totalGpus)
}
