/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/field"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
)

func extractScope(scope openapi.Scope) string {
	return field.CapitalizeFirst(string(scope))
}

func ParseCredentialColumns(columnNames []string) ([]CredentialColumn, error) {
	var columns []CredentialColumn
	for _, name := range columnNames {
		normalizedName := strings.ToLower(strings.TrimSpace(name))
		column := CredentialColumn(normalizedName)

		if _, exists := ColumnMap[column]; !exists {
			return nil, fmt.Errorf("invalid column name: %s", name)
		}
		columns = append(columns, column)
	}
	return columns, nil
}

func BriefCredentialFromEntry(cred openapi.CredentialsListResponseEntry) BriefCredential {
	return BriefCredential{
		Name:    cred.Meta.Name,
		Type:    formatType(cred.Meta.Kind),
		Scope:   extractScope(cred.Meta.Scope),
		K8sName: extractK8sName(&cred),
		Status:  extractStatus(&cred),
	}
}

func CreateAlignedCredentials(credentials []openapi.CredentialsListResponseEntry, columns []CredentialColumn) CredentialsAlignedData {
	var aligned []CredentialsAlignedRow

	for _, cred := range credentials {
		row := CredentialsAlignedRow{}
		for _, col := range columns {
			fieldInfo := ColumnMap[col]
			fieldInfo.extractFunc(&cred, &row)
		}
		aligned = append(aligned, row)
	}

	return CredentialsAlignedData{Credentials: aligned}
}

func CreateTableData(credentials []CredentialsAlignedRow, columns []CredentialColumn) [][]string {
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
