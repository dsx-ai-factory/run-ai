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
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/credential/list/manipulate"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/settings"
	uitable "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/table"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type Output struct {
	rawData []openapi.CredentialsListResponseEntry
	raw     bool
	columns []manipulate.CredentialColumn
}

func ParseListResponse(response *openapi.CredentialsListResponse) (*Output, error) {
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
		columns = settings.GetInstance().GetCredentialColumns()
	}
	credentialColumns, err := manipulate.ParseCredentialColumns(columns)
	if err != nil {
		return err
	}
	o.columns = credentialColumns
	return nil
}

func (o *Output) Subject() string {
	return consts.CredentialStr
}

func (o *Output) getActiveColumns() []manipulate.CredentialColumn {
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
	aligned := manipulate.CreateAlignedCredentials(o.rawData, columns)
	data := manipulate.CreateTableData(aligned.Credentials, columns)

	cliutils.MultiColumnSort(data, []cliutils.SortColumn{
		{Index: 0, Order: cliutils.Ascending},
	})

	return data
}

func (o *Output) GenerateTable() (string, error) {
	return uitable.NewTableDisplay(o).Generate()
}

func (o *Output) getBriefCredentials() []manipulate.BriefCredential {
	if len(o.rawData) == 0 {
		return []manipulate.BriefCredential{}
	}

	var briefCreds []manipulate.BriefCredential
	for _, cred := range o.rawData {
		briefCreds = append(briefCreds, manipulate.BriefCredentialFromEntry(cred))
	}
	return briefCreds
}

func (o *Output) getDataStructure() interface{} {
	if o.raw {
		return manipulate.Data{Credentials: o.rawData}
	}

	formatAlign := settings.GetInstance().GetListFormatAlign()
	if !formatAlign {
		return manipulate.BriefData{Credentials: o.getBriefCredentials()}
	}

	columns := o.getActiveColumns()
	return manipulate.CreateAlignedCredentials(o.rawData, columns)
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

func (o *Output) GetList() []manipulate.BriefCredential {
	return o.getBriefCredentials()
}

func HandleColumnsFlags(output listCommon.Output, cmd *cobra.Command, defaultColumnsMsg string) error {
	var err error
	userColumns := output.GetColumns()

	if len(userColumns) > 0 {
		_, err = manipulate.ParseCredentialColumns(userColumns)
		if err != nil {
			return err
		}
	}
	if output.IsSetDefaultColumns() {
		err = settings.GetInstance().SetCredentialColumns(userColumns)
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
