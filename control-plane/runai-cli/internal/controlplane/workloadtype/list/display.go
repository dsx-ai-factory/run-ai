/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"encoding/json"

	"github.com/charmbracelet/bubbles/table"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workloadtype/list/manipulate"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/settings"
	uitable "github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/table"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"gopkg.in/yaml.v3"
)

type BriefData struct {
	WorkloadTypes []manipulate.BriefWorkloadType `json:"workloadTypes,omitempty" yaml:"workloadTypes"`
}

type Output struct {
	data    manipulate.Data
	raw     bool
	columns []manipulate.WorkloadTypeColumn
}

func ParseListResponse(workloadTypes []openapi.WorkloadTypeConfig) (*Output, error) {
	return &Output{
		data: manipulate.Data{
			WorkloadTypes: workloadTypes,
		},
	}, nil
}

func (o *Output) Subject() string {
	return "workload-type"
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
	workloadTypesAligned := manipulate.CreateAlignedWorkloadTypes(o.data, columns).WorkloadTypes
	return manipulate.CreateTableData(workloadTypesAligned, columns)
}

func (o *Output) GenerateTable() (string, error) {
	return uitable.NewTableDisplay(o).Generate()
}

func (o *Output) SetRaw(raw bool) {
	o.raw = raw
}

func (o *Output) SetColumns(columns []string) error {
	if len(columns) == 0 {
		columns = settings.GetInstance().GetWorkloadTypeColumns()
	}
	workloadTypeColumns, err := manipulate.ParseWorkloadTypeColumns(columns)
	if err != nil {
		return err
	}

	o.columns = workloadTypeColumns
	return nil
}

func (o *Output) getBriefWorkloadTypes() []manipulate.BriefWorkloadType {
	if len(o.data.WorkloadTypes) == 0 {
		return []manipulate.BriefWorkloadType{}
	}
	briefList, _ := manipulate.BriefListFromResponse(o.data.WorkloadTypes)
	return briefList
}

func (o *Output) getDataStructure() interface{} {
	if o.raw {
		return o.data
	}

	formatAlign := settings.GetInstance().GetListFormatAlign()
	if !formatAlign {
		return BriefData{
			WorkloadTypes: o.getBriefWorkloadTypes(),
		}
	}

	columns := manipulate.GetColumns(o.columns)
	return manipulate.CreateAlignedWorkloadTypes(o.data, columns)
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

func (o *Output) GetList() []manipulate.BriefWorkloadType {
	return o.getBriefWorkloadTypes()
}

func (o *Output) IsEmpty() bool {
	return len(o.data.WorkloadTypes) == 0
}
