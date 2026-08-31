/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"encoding/json"
	"fmt"

	listCommon "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/list"
	"github.com/spf13/cobra"

	"github.com/charmbracelet/bubbles/table"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli/structs"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/template/list/manipulate"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/settings"
	uitable "github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/table"
	wcgenerated "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"gopkg.in/yaml.v3"
)

const DefaultTemplateColumnsSet = "Default template columns set to: %s\n"

func (o *Output) SetRaw(raw bool) {
	o.raw = raw
}

func (o *Output) SetColumns(columns []string) error {
	if len(columns) == 0 {
		columns = settings.GetInstance().GetTemplateColumns()
	}
	templateColumns, err := manipulate.ParseTemplateColumns(columns)
	if err != nil {
		return err
	}

	o.columns = templateColumns
	return nil
}

type Output struct {
	data    manipulate.Data
	raw     bool
	columns []manipulate.TemplateColumn
}

func NewOutput(templateList []wcgenerated.SupersetTemplate, next *structs.NextToken) (*Output, error) {
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
			Templates: templateList,
			NextToken: nextToken,
		},
	}, nil
}

func (o *Output) Subject() string {
	return "template"
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

func (o *Output) getActiveColumns() []manipulate.TemplateColumn {
	if len(o.columns) > 0 {
		return o.columns
	}
	return manipulate.GetDefaultColumns()
}

func (o *Output) getBriefTemplates() []manipulate.BriefTemplate {
	if len(o.data.Templates) == 0 {
		return []manipulate.BriefTemplate{}
	}

	return manipulate.BriefListFromResponse(o.data.Templates)
}

func (o *Output) Data() [][]string {
	// Get columns to display (respects --columns flag)
	columns := o.getActiveColumns()

	// Extract requested fields from templates list API
	templatesAligned := manipulate.CreateAlignedTemplates(
		manipulate.Data{Templates: o.data.Templates, NextToken: o.data.NextToken},
		columns,
	).Templates

	// Convert to string table and sort
	return manipulate.CreateTableData(templatesAligned, columns)
}

func (o *Output) GenerateTable() (string, error) {
	display, err := uitable.NewTableDisplay(o).Generate()
	if o.data.NextToken != nil {
		display += fmt.Sprintf("\r\n\n next token: %s", *o.data.NextToken)
	}
	return display, err
}

func (o *Output) getDataStructure() interface{} {
	if o.raw {
		return o.data
	}

	formatAlign := settings.GetInstance().GetListFormatAlign()
	if !formatAlign {
		// Use original brief format when format alignment is disabled
		return manipulate.BriefData{
			Templates: o.getBriefTemplates(),
			NextToken: o.data.NextToken,
		}
	}

	columns := o.getActiveColumns()
	return manipulate.CreateAlignedTemplates(manipulate.Data{Templates: o.data.Templates, NextToken: o.data.NextToken}, columns)
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

func (o *Output) GetList() []manipulate.BriefTemplate {
	return o.getBriefTemplates()
}

func HandleColumnsFlags(output listCommon.Output, cmd *cobra.Command) error {
	var err error
	userColumns := output.GetColumns()

	if len(userColumns) > 0 {
		_, err = manipulate.ParseTemplateColumns(userColumns)
		if err != nil {
			return err
		}
	}
	if output.IsSetDefaultColumns() {
		err = settings.GetInstance().SetTemplateColumns(userColumns)
		if err != nil {
			return err
		}
		if len(userColumns) == 0 {
			cmd.Printf(DefaultTemplateColumnsSet, manipulate.GetDefaultColumns())
		} else {
			cmd.Printf(DefaultTemplateColumnsSet, userColumns)
		}
	}
	return nil
}
