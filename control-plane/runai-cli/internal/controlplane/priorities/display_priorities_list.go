/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package priorities

import (
	"encoding/json"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"

	cliutils "github.com/run-ai/runai/control-plane/runai-cli/internal/utils"

	"github.com/charmbracelet/bubbles/table"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	uitable "github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/table"
	"gopkg.in/yaml.v3"
)

const (
	DefaultWidth = 10

	TitlePriorityName = "Priority Name"
)

const (
	nameColumnIndex = 0
)

type Data struct {
	PrioritiesNames []Brief `json:"prioritiesNames,omitempty" yaml:"prioritiesNames"`
}

type Output struct {
	data Data
}

func NewOutput(response *openapi.GetWorkloadClusterPriorities200Response) (*Output, error) {
	briefList, err := BriefListFromResponse(response.Priorities)
	if err != nil {
		return nil, err
	}
	return &Output{
		data: Data{PrioritiesNames: briefList},
	}, nil
}

func (o *Output) Subject() string {
	return consts.ClusterStr
}

func (o *Output) Columns() []table.Column {
	return []table.Column{
		{Title: TitlePriorityName, Width: DefaultWidth},
	}
}

func (o *Output) Data() [][]string {
	var (
		data    [][]string
		managed string
	)

	for _, brief := range o.data.PrioritiesNames {
		var row []string
		row = append(row, brief.Name)
		row = append(row, managed)
		data = append(data, row)
	}

	cliutils.MultiColumnSort(data, []cliutils.SortColumn{
		{Index: nameColumnIndex, Order: cliutils.Ascending},
	})

	return data
}

func (o *Output) GenerateTable() (string, error) {
	return uitable.NewTableDisplay(o).Generate()
}

func (o *Output) GenerateJson() (string, error) {
	output, err := json.MarshalIndent(o.data, "", "  ")
	return string(output), err
}

func (o *Output) GenerateYaml() (string, error) {
	output, err := yaml.Marshal(o.data)
	return string(output), err
}

func (o *Output) GetList() []Brief {
	return o.data.PrioritiesNames
}
