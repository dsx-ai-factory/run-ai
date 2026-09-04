/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"encoding/json"
	"fmt"

	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"

	cliutils "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/utils"

	"github.com/charmbracelet/bubbles/table"
	listCommon "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/compute/list/manipulate"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/settings"
	uitable "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/table"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type Output struct {
	rawData []openapi.ComputeAsset
	raw     bool
	columns []manipulate.ComputeColumn
}

func ParseListResponse(response *openapi.ComputeListResponse) (*Output, error) {
	if response == nil {
		return nil, fmt.Errorf("response is nil")
	}

	return &Output{
		rawData: response.Entries,
	}, nil
}

func (o *Output) SetRaw(raw bool) {
	o.raw = raw
}

func (o *Output) SetColumns(columns []string) error {
	if len(columns) == 0 {
		columns = settings.GetInstance().GetComputeColumns()
	}
	computeColumns, err := manipulate.ParseComputeColumns(columns)
	if err != nil {
		return err
	}
	o.columns = computeColumns
	return nil
}

func (o *Output) Subject() string {
	return consts.ComputeStr
}

func (o *Output) getActiveColumns() []manipulate.ComputeColumn {
	return manipulate.GetColumns(o.columns)
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

func (o *Output) Data() [][]string {
	columns := o.getActiveColumns()
	aligned := manipulate.CreateAlignedComputes(o.rawData, columns)
	data := manipulate.CreateTableData(aligned.Computes, columns)

	cliutils.MultiColumnSort(data, []cliutils.SortColumn{
		{Index: 0, Order: cliutils.Ascending},
	})

	return data
}

func (o *Output) GenerateTable() (string, error) {
	return uitable.NewTableDisplay(o).Generate()
}

// getDataStructure returns the appropriate data structure based on raw and format alignment settings
func (o *Output) getDataStructure() interface{} {
	if o.raw {
		return manipulate.Data{Computes: o.rawData}
	}

	formatAlign := settings.GetInstance().GetListFormatAlign()
	if !formatAlign {
		briefComputes := make([]manipulate.BriefCompute, 0, len(o.rawData))
		for _, compute := range o.rawData {
			briefComputes = append(briefComputes, manipulate.BriefComputeFromAsset(compute))
		}
		return manipulate.BriefData{Computes: briefComputes}
	}

	columns := o.getActiveColumns()
	return manipulate.CreateAlignedComputes(o.rawData, columns)
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

func (o *Output) GetList() []manipulate.BriefCompute {
	briefComputes := make([]manipulate.BriefCompute, 0, len(o.rawData))
	for _, compute := range o.rawData {
		briefComputes = append(briefComputes, manipulate.BriefComputeFromAsset(compute))
	}
	return briefComputes
}

func HandleColumnsFlags(output listCommon.Output, cmd *cobra.Command, defaultColumnsMsg string) error {
	var err error
	userColumns := output.GetColumns()

	if len(userColumns) > 0 {
		_, err = manipulate.ParseComputeColumns(userColumns)
		if err != nil {
			return err
		}
	}
	if output.IsSetDefaultColumns() {
		err = settings.GetInstance().SetComputeColumns(userColumns)
		if err != nil {
			return err
		}
		if len(userColumns) == 0 {
			cmd.Printf(defaultColumnsMsg, manipulate.GetDefaultColumns())
		} else {
			cmd.Printf(defaultColumnsMsg, userColumns)
		}
	}
	return nil
}
