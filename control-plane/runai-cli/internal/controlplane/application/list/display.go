/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli/structs"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/application/list/manipulate"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/settings"
	uitable "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/table"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"gopkg.in/yaml.v3"
)

type BriefData struct {
	AIApplications []manipulate.BriefAIApplication `json:"aiApplications,omitempty" yaml:"aiApplications"`
	NextToken      *string                         `json:"nextToken,omitempty" yaml:"nextToken,omitempty"`
}

type Output struct {
	data    manipulate.Data
	raw     bool
	columns []manipulate.AIApplicationColumn
}

func NewOutput(apps []openapi.AIApplication, next *structs.NextToken) (*Output, error) {
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
			AIApplications: apps,
			NextToken:      nextToken,
		},
	}, nil
}

func (o *Output) Subject() string {
	return consts.AIApplicationStr
}

func (o *Output) Columns() []table.Column {
	columns := manipulate.GetColumns(o.columns)
	var columnsTable []table.Column
	for _, column := range columns {
		columnsTable = append(columnsTable, table.Column{
			Title: manipulate.ColumnMap[column].GetColumnTitle(),
			Width: manipulate.ColumnMap[column].GetColumnWidth(),
		})
	}
	return columnsTable
}

func (o *Output) Data() [][]string {
	columns := manipulate.GetColumns(o.columns)
	aligned := manipulate.CreateAlignedApplications(o.data, columns)
	return manipulate.CreateTableData(aligned.AIApplications, columns)
}

func (o *Output) GenerateTable() (string, error) {
	display, err := uitable.NewTableDisplay(o).Generate()
	if err != nil {
		return "", err
	}
	if o.data.NextToken != nil {
		display += fmt.Sprintf("\r\n\n next token: %s", *o.data.NextToken)
	}
	return display, nil
}

func (o *Output) SetRaw(raw bool) {
	o.raw = raw
}

func (o *Output) SetColumns(columns []string) error {
	if len(columns) == 0 {
		columns = settings.GetInstance().GetAIApplicationColumns()
	}
	appColumns, err := manipulate.ParseAIApplicationColumns(columns)
	if err != nil {
		return err
	}
	o.columns = appColumns
	return nil
}

func (o *Output) getBriefApplications() []manipulate.BriefAIApplication {
	if len(o.data.AIApplications) == 0 {
		return []manipulate.BriefAIApplication{}
	}
	return manipulate.BriefListFromResponse(o.data.AIApplications)
}

func (o *Output) GetList() []manipulate.BriefAIApplication {
	return o.getBriefApplications()
}

func (o *Output) getDataStructure() interface{} {
	if o.raw {
		return o.data
	}

	if !settings.GetInstance().GetListFormatAlign() {
		return BriefData{
			AIApplications: o.getBriefApplications(),
			NextToken:      o.data.NextToken,
		}
	}

	columns := manipulate.GetColumns(o.columns)
	return manipulate.CreateAlignedApplications(o.data, columns)
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

func (o *Output) IsEmpty() bool {
	return len(o.data.AIApplications) == 0
}
