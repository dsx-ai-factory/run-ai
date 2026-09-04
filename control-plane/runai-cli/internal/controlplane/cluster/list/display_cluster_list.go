/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli/structs"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/cluster/list/manipulate"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/settings"
	uitable "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/table"
	"gopkg.in/yaml.v3"
)

// SetRaw sets the raw output preference
func (o *Output) SetRaw(raw bool) {
	o.raw = raw
}

// SetColumns sets the columns to display
func (o *Output) SetColumns(columns []string) error {
	if len(columns) == 0 {
		columns = settings.GetInstance().GetClusterColumns()
	}
	clusterColumns, err := manipulate.ParseClusterColumns(columns)
	if err != nil {
		return err
	}

	o.columns = clusterColumns
	return nil
}

type Output struct {
	data    manipulate.Data
	raw     bool
	columns []manipulate.ClusterColumn
}

func NewOutput(clusterList []openapi.ClustersMinimalData, next *structs.NextToken) (*Output, error) {
	var nextToken *string

	if next != nil {
		encoded, errEnc := next.Encode()
		if errEnc != nil {
			return nil, errEnc
		}
		nextToken = &encoded
	}

	return &Output{
		data: manipulate.Data{
			Clusters:  clusterList, // Store originals directly
			NextToken: nextToken,
		},
	}, nil
}

func (o *Output) Subject() string {
	return consts.ClusterStr
}

func (o *Output) Columns() []table.Column {
	columns := o.getActiveColumns()
	var tableColumns []table.Column
	for _, column := range columns {
		fieldInfo := manipulate.ColumnMap[column]
		tableColumns = append(tableColumns, table.Column{
			Title: fieldInfo.GetColumnTitle(),
			Width: fieldInfo.GetColumnWidth(),
		})
	}

	return tableColumns
}

func (o *Output) getActiveColumns() []manipulate.ClusterColumn {
	if len(o.columns) > 0 {
		return o.columns
	}
	return manipulate.GetDefaultColumns()
}

// getBriefClusters converts original clusters to brief format on-demand
func (o *Output) getBriefClusters() []manipulate.BriefCluster {
	if len(o.data.Clusters) == 0 {
		return []manipulate.BriefCluster{}
	}

	briefList, _ := manipulate.BriefListFromResponse(o.data.Clusters)
	return briefList
}

func (o *Output) Data() [][]string {
	columns := o.getActiveColumns()
	clustersAligned := manipulate.CreateAlignedClusters(manipulate.Data{Clusters: o.data.Clusters, NextToken: o.data.NextToken}, columns).Clusters
	return manipulate.CreateTableData(clustersAligned, columns)
}

func (o *Output) GenerateTable() (string, error) {
	display, err := uitable.NewTableDisplay(o).Generate()
	if o.data.NextToken != nil {
		display += fmt.Sprintf("\r\n\n next token: %s", *o.data.NextToken)
	}
	return display, err
}

// getDataStructure returns the appropriate data structure based on raw and format alignment settings
func (o *Output) getDataStructure() interface{} {
	if o.raw {
		return o.data
	}

	formatAlign := settings.GetInstance().GetListFormatAlign()
	if !formatAlign {
		// Use original brief format when format alignment is disabled
		return manipulate.BriefData{
			Clusters:  o.getBriefClusters(),
			NextToken: o.data.NextToken,
		}
	}

	// Use aligned format with column titles when format alignment is enabled
	columns := o.getActiveColumns()
	return manipulate.CreateAlignedClusters(manipulate.Data{Clusters: o.data.Clusters, NextToken: o.data.NextToken}, columns)
}

func (o *Output) GenerateJson() (string, error) {
	data := o.getDataStructure()
	output, err := json.MarshalIndent(data, "", "  ")
	return string(output), err
}

func (o *Output) GenerateYaml() (string, error) {
	data := o.getDataStructure()
	output, err := yaml.Marshal(data)
	return string(output), err
}

func (o *Output) GetList() []manipulate.BriefCluster {
	return o.getBriefClusters()
}
