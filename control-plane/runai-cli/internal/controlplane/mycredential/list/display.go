/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/mycredential/client"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/mycredential/list/manipulate"
	uitable "github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/table"
)

type Output struct {
	rawData []client.MyCredentialEntry
}

func NewOutput(entries []client.MyCredentialEntry) *Output {
	return &Output{
		rawData: entries,
	}
}

func (o *Output) Subject() string {
	return "my-credential"
}

func (o *Output) Columns() []table.Column {
	var tableColumns []table.Column
	for _, column := range manipulate.GetDefaultColumns() {
		fieldInfo := manipulate.ColumnMap[column]
		tableColumns = append(tableColumns, table.Column{
			Title: fieldInfo.GetColumnTitle(),
			Width: fieldInfo.GetColumnWidth(),
		})
	}
	return tableColumns
}

func (o *Output) Data() [][]string {
	columns := manipulate.GetDefaultColumns()
	aligned := manipulate.CreateAlignedMyCredentials(o.rawData, columns)
	return manipulate.CreateTableData(aligned.Credentials, columns)
}

func (o *Output) GenerateTable() (string, error) {
	return uitable.NewTableDisplay(o).Generate()
}
