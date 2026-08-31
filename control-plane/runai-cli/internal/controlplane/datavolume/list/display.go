/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"encoding/json"
	"fmt"

	openapi "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"

	cliutils "github.com/run-ai/runai/control-plane/runai-cli/internal/utils"

	"github.com/charmbracelet/bubbles/table"
	listCommon "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/datavolume/list/manipulate"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/settings"
	uitable "github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/table"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type Output struct {
	rawData []openapi.DatavolumeNoSharedScopes
	raw     bool
	columns []manipulate.DatavolumeColumn
}

func ParseListResponse(response *openapi.GetDatavolumes200Response) (*Output, error) {
	if response == nil {
		return nil, fmt.Errorf("response is nil")
	}

	return &Output{
		rawData: response.Datavolumes,
	}, nil
}

func (o *Output) SetRaw(raw bool) {
	o.raw = raw
}

func (o *Output) SetColumns(columns []string) error {
	if len(columns) == 0 {
		columns = settings.GetInstance().GetDatavolumeColumns()
	}
	datavolumeColumns, err := manipulate.ParseDatavolumeColumns(columns)
	if err != nil {
		return err
	}
	o.columns = datavolumeColumns
	return nil
}

func (o *Output) Subject() string {
	return consts.DatavolumeStr
}

func (o *Output) getActiveColumns() []manipulate.DatavolumeColumn {
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
	aligned := manipulate.CreateAlignedDatavolumes(o.rawData, columns)
	data := manipulate.CreateTableData(aligned.Datavolumes, columns)

	cliutils.MultiColumnSort(data, []cliutils.SortColumn{
		{Index: 0, Order: cliutils.Ascending},
	})

	return data
}

func (o *Output) GenerateTable() (string, error) {
	return uitable.NewTableDisplay(o).Generate()
}

func (o *Output) getBriefDatavolumes() []manipulate.BriefDatavolume {
	if len(o.rawData) == 0 {
		return []manipulate.BriefDatavolume{}
	}

	var briefDV []manipulate.BriefDatavolume
	for _, dv := range o.rawData {
		briefDV = append(briefDV, manipulate.BriefDatavolumeFromEntry(dv))
	}
	return briefDV
}

func (o *Output) getDataStructure() interface{} {
	if o.raw {
		return manipulate.Data{Datavolumes: o.rawData}
	}

	formatAlign := settings.GetInstance().GetListFormatAlign()
	if !formatAlign {
		return manipulate.BriefData{Datavolumes: o.getBriefDatavolumes()}
	}

	columns := o.getActiveColumns()
	return manipulate.CreateAlignedDatavolumes(o.rawData, columns)
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

func (o *Output) GetList() []manipulate.BriefDatavolume {
	return o.getBriefDatavolumes()
}

func HandleColumnsFlags(output listCommon.Output, cmd *cobra.Command, defaultColumnsMsg string) error {
	var err error
	userColumns := output.GetColumns()

	if len(userColumns) > 0 {
		_, err = manipulate.ParseDatavolumeColumns(userColumns)
		if err != nil {
			return err
		}
	}
	if output.IsSetDefaultColumns() {
		err = settings.GetInstance().SetDatavolumeColumns(userColumns)
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
