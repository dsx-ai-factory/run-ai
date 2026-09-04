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
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/datasource/list/manipulate"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/settings"
	uitable "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/table"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type Output struct {
	rawData []openapi.DatasourceListResponseEntry
	raw     bool
	columns []manipulate.DatasourceColumn
}

func ParseListResponse(response *openapi.DatasourceListResponse) (*Output, error) {
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
		columns = settings.GetInstance().GetDatasourceColumns()
	}
	datasourceColumns, err := manipulate.ParseDatasourceColumns(columns)
	if err != nil {
		return err
	}
	o.columns = datasourceColumns
	return nil
}

func (o *Output) Subject() string {
	return consts.DatasourceStr
}

func (o *Output) getActiveColumns() []manipulate.DatasourceColumn {
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
	aligned := manipulate.CreateAlignedDatasources(o.rawData, columns)
	data := manipulate.CreateTableData(aligned.Datasources, columns)

	cliutils.MultiColumnSort(data, []cliutils.SortColumn{
		{Index: 0, Order: cliutils.Ascending},
	})

	return data
}

func (o *Output) GenerateTable() (string, error) {
	return uitable.NewTableDisplay(o).Generate()
}

// getBriefDatasources converts original datasources to brief format on-demand
func (o *Output) getBriefDatasources() []manipulate.BriefDatasource {
	if len(o.rawData) == 0 {
		return []manipulate.BriefDatasource{}
	}

	var briefDS []manipulate.BriefDatasource
	for _, ds := range o.rawData {
		briefDS = append(briefDS, manipulate.BriefDatasourceFromEntry(ds))
	}
	return briefDS
}

// getDataStructure returns the appropriate data structure based on raw and format alignment settings
func (o *Output) getDataStructure() interface{} {
	if o.raw {
		return manipulate.Data{Datasources: o.rawData}
	}

	formatAlign := settings.GetInstance().GetListFormatAlign()
	if !formatAlign {
		// Use original brief format when format alignment is disabled
		return manipulate.BriefData{Datasources: o.getBriefDatasources()}
	}

	// Use aligned format when format alignment is enabled
	columns := o.getActiveColumns()
	return manipulate.CreateAlignedDatasources(o.rawData, columns)
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

func (o *Output) GetList() []manipulate.BriefDatasource {
	return o.getBriefDatasources()
}

// HandleColumnsFlags validates and handles --columns and --set-default-columns flags
func HandleColumnsFlags(output listCommon.Output, cmd *cobra.Command, defaultColumnsMsg string) error {
	var err error
	userColumns := output.GetColumns()

	if len(userColumns) > 0 {
		_, err = manipulate.ParseDatasourceColumns(userColumns)
		if err != nil {
			return err
		}
	}
	if output.IsSetDefaultColumns() {
		err = settings.GetInstance().SetDatasourceColumns(userColumns)
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
