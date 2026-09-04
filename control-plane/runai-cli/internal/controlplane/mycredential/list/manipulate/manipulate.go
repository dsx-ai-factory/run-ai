/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/mycredential/client"
)

func CreateAlignedMyCredentials(entries []client.MyCredentialEntry, columns []MyCredentialColumn) MyCredentialAlignedData {
	var aligned []MyCredentialAlignedRow

	for _, entry := range entries {
		row := MyCredentialAlignedRow{}
		for _, col := range columns {
			fieldInfo := ColumnMap[col]
			fieldInfo.extractFunc(&entry, &row)
		}
		aligned = append(aligned, row)
	}

	return MyCredentialAlignedData{Credentials: aligned}
}

func CreateTableData(credentials []MyCredentialAlignedRow, columns []MyCredentialColumn) [][]string {
	var data [][]string

	for _, credRow := range credentials {
		var row []string
		for _, col := range columns {
			fieldInfo := ColumnMap[col]
			row = append(row, fieldInfo.formatFunc(&credRow))
		}
		data = append(data, row)
	}

	return data
}
