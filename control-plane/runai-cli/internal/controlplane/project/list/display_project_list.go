/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli/structs"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project/list/manipulate"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/settings"
	uitable "github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/table"
	openapi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	"gopkg.in/yaml.v3"
)

const (
	Percent = 100
)

type BriefData struct {
	Projects  []manipulate.BriefProject `json:"projects,omitempty" yaml:"projects"`
	NextToken *string                   `json:"nextToken,omitempty" yaml:"nextToken,omitempty"`
}

type Output struct {
	data    manipulate.Data
	raw     bool
	columns []manipulate.ProjectColumn
}

func NewOutput(projectList []openapi.Project, next *structs.NextToken) (*Output, error) {
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
			Projects:  projectList,
			NextToken: nextToken,
		},
	}, nil
}

func (o *Output) Subject() string {
	return consts.ProjectStr
}

func (o *Output) Columns() []table.Column {
	columns := manipulate.GetColumns(o.columns)
	var columnsTable []table.Column
	for _, column := range columns {
		columnsTable = append(columnsTable, table.Column{Title: manipulate.ColumnMap[column].GetColumnTitle(), Width: manipulate.ColumnMap[column].GetColumnWidth()})
	}
	return columnsTable
}

func (o *Output) Data() [][]string {
	columns := manipulate.GetColumns(o.columns)
	projectsAligned := manipulate.CreateAlignedProjects(o.data, columns).Projects
	return manipulate.CreateTableData(projectsAligned, columns)
}

func (o *Output) GenerateTable() (string, error) {
	display, err := uitable.NewTableDisplay(o).Generate()
	if o.data.NextToken != nil {
		display += fmt.Sprintf("\r\n\n next token: %s", *o.data.NextToken)
	}
	return display, err
}

func (o *Output) SetRaw(raw bool) {
	o.raw = raw
}

func (o *Output) SetColumns(columns []string) error {
	if len(columns) == 0 {
		columns = settings.GetInstance().GetProjectColumns()
	}
	projectColumns, err := manipulate.ParseProjectColumns(columns)
	if err != nil {
		return err
	}

	o.columns = projectColumns
	return nil
}

func (o *Output) getBriefProjects() []manipulate.BriefProject {
	if len(o.data.Projects) == 0 {
		return []manipulate.BriefProject{}
	}
	briefList, _ := manipulate.BriefListFromResponse(o.data.Projects)
	return briefList
}

func (o *Output) getDataStructure() interface{} {
	if o.raw {
		return o.data
	}

	formatAlign := settings.GetInstance().GetListFormatAlign()
	if !formatAlign {
		return BriefData{
			Projects:  o.getBriefProjects(),
			NextToken: o.data.NextToken,
		}
	}

	columns := manipulate.GetColumns(o.columns)
	return manipulate.CreateAlignedProjects(o.data, columns)
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

func (o *Output) GetList() []manipulate.BriefProject {
	return o.getBriefProjects()
}

func (o *Output) IsEmpty() bool {
	return len(o.data.Projects) == 0
}
