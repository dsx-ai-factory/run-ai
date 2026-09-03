/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool

import (
	"encoding/json"
	"fmt"

	cliutils "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/utils"

	"github.com/charmbracelet/bubbles/table"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api"
	nodes "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/node/list/manipulate"
	table2 "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/table"
	"gopkg.in/yaml.v3"
)

const (
	detected    = "Detected"
	notDetected = "Not Detected"
)

const (
	nameColumnIndex = 0
)

type Data struct {
	Nodepools []*BriefNodepool `json:"nodepools,omitempty" yaml:"nodepools"`
	NextToken *string          `json:"nextToken,omitempty" yaml:"nextToken,omitempty"`
}

type Output struct {
	data Data
}

func NewOutput(nodepools *openapi.GetNodepools200Response, nodes *openapi.Nodes, telemetry api.NodeTelemetriesByType) (*Output, error) {
	briefList, err := BriefListFromResponse(nodepools, nodes, telemetry)
	if err != nil {
		return nil, err
	}
	return &Output{Data{
		Nodepools: briefList,
		NextToken: nil,
	}}, nil
}

func (o *Output) Subject() string {
	return consts.NodeStr
}

func (o *Output) Columns() []table.Column {
	return []table.Column{
		{Title: "Name", Width: nodes.DefaultWidthName},
		{Title: "Status", Width: nodes.DefaultWidthStatus},
		{Title: "Allocated/Total GPUs", Width: nodes.DefaultWidthGpuAllocation},
		{Title: "Nodes", Width: nodes.DefaultWidthGpuCount},
		{Title: "MNNVL detection", Width: nodes.DefaultWidthMNNVLDetection},
		{Title: "Network Topology", Width: nodes.DefaultWidthNetworkTopology},
	}
}

func (o *Output) Data() [][]string {
	var data [][]string

	for _, brief := range o.data.Nodepools {
		var row []string
		row = append(row, brief.Name)
		row = append(row, brief.Status)
		row = append(row, brief.GpuAllocation)
		row = append(row, fmt.Sprintf("%d", len(brief.Nodes)))
		row = append(row, detectionString(brief.MNNVLDetected))
		row = append(row, brief.NetworkTopologyName)
		data = append(data, row)
	}

	cliutils.MultiColumnSort(data, []cliutils.SortColumn{
		{Index: nameColumnIndex, Order: cliutils.Ascending},
	})

	return data
}

func (o *Output) GenerateTable() (string, error) {
	return table2.NewTableDisplay(o).Generate()
}

func (o *Output) GenerateJson() (string, error) {
	output, err := json.MarshalIndent(o.data, "", "  ")
	return string(output), err
}

func (o *Output) GenerateYaml() (string, error) {
	output, err := yaml.Marshal(o.data)
	return string(output), err
}

func detectionString(detection bool) string {
	if detection {
		return detected
	}
	return notDetected
}
