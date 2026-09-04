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
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/environment/list/manipulate"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/settings"
	uitable "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/table"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type Output struct {
	data    manipulate.Data
	raw     bool
	columns []manipulate.EnvironmentColumn
}

func ParseListResponse(response *openapi.EnvironmentListResponse) (*Output, error) {
	if response == nil {
		return nil, fmt.Errorf("response is nil")
	}

	return &Output{
		data: manipulate.Data{Environments: response.Entries},
	}, nil
}

// SetRaw sets the raw output preference
func (o *Output) SetRaw(raw bool) {
	o.raw = raw
}

// SetColumns sets the columns to display
func (o *Output) SetColumns(columns []string) error {
	if len(columns) == 0 {
		columns = settings.GetInstance().GetEnvironmentColumns()
	}
	environmentColumns, err := manipulate.ParseEnvironmentColumns(columns)
	if err != nil {
		return err
	}
	o.columns = environmentColumns
	return nil
}

func (o *Output) Subject() string {
	return consts.EnvironmentStr
}

func (o *Output) getActiveColumns() []manipulate.EnvironmentColumn {
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
	aligned := manipulate.CreateAlignedEnvironments(o.data.Environments, columns)
	data := manipulate.CreateTableData(aligned.Environments, columns)

	cliutils.MultiColumnSort(data, []cliutils.SortColumn{
		{Index: 0, Order: cliutils.Ascending},
	})

	return data
}

func (o *Output) GenerateTable() (string, error) {
	return uitable.NewTableDisplay(o).Generate()
}

// getBriefEnvironments converts original environments to brief format on-demand
func (o *Output) getBriefEnvironments() []manipulate.BriefEnvironment {
	if len(o.data.Environments) == 0 {
		return []manipulate.BriefEnvironment{}
	}

	var briefEnvs []manipulate.BriefEnvironment
	for _, env := range o.data.Environments {
		briefEnvs = append(briefEnvs, manipulate.BriefEnvironmentFromAsset(env))
	}
	return briefEnvs
}

// getDataStructure returns the appropriate data structure based on raw and format alignment settings
func (o *Output) getDataStructure() interface{} {
	if o.raw {
		return o.data
	}

	formatAlign := settings.GetInstance().GetListFormatAlign()
	if !formatAlign {
		// Use original brief format when format alignment is disabled
		return manipulate.BriefData{Environments: o.getBriefEnvironments()}
	}

	// Use aligned format with column titles when format alignment is enabled
	columns := o.getActiveColumns()
	return manipulate.CreateAlignedEnvironments(o.data.Environments, columns)
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

func (o *Output) GetList() []manipulate.BriefEnvironment {
	return o.getBriefEnvironments()
}

// HandleColumnsFlags validates and handles --columns and --set-default-columns flags
func HandleColumnsFlags(output listCommon.Output, cmd *cobra.Command, defaultColumnsMsg string) error {
	var err error
	userColumns := output.GetColumns()

	if len(userColumns) > 0 {
		_, err = manipulate.ParseEnvironmentColumns(userColumns)
		if err != nil {
			return err
		}
	}
	if output.IsSetDefaultColumns() {
		err = settings.GetInstance().SetEnvironmentColumns(userColumns)
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
