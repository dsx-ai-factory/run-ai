/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/field"
	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

// extractScope is a common helper to extract and format scope from datasource metadata
func extractScope(scope openapi.Scope) string {
	return field.CapitalizeFirst(string(scope))
}

// ParseDatasourceColumns validates and converts column names to DatasourceColumn types
func ParseDatasourceColumns(columnNames []string) ([]DatasourceColumn, error) {
	var columns []DatasourceColumn
	for _, name := range columnNames {
		normalizedName := strings.ToLower(strings.TrimSpace(name))
		column := DatasourceColumn(normalizedName)

		if _, exists := ColumnMap[column]; !exists {
			return nil, fmt.Errorf("invalid column name: %s", name)
		}
		columns = append(columns, column)
	}
	return columns, nil
}

// BriefDatasourceFromEntry converts a DatasourceListResponseEntry to BriefDatasource
func BriefDatasourceFromEntry(ds openapi.DatasourceListResponseEntry) BriefDatasource {
	return BriefDatasource{
		Name:  ds.Meta.Name,
		Type:  formatType(ds.Meta.Kind),
		Scope: extractScope(ds.Meta.Scope),
	}
}

// CreateAlignedDatasources transforms datasource entries into aligned format based on selected columns
func CreateAlignedDatasources(datasources []openapi.DatasourceListResponseEntry, columns []DatasourceColumn) DatasourcesAlignedData {
	var aligned []DatasourcesAlignedRow

	for _, ds := range datasources {
		row := DatasourcesAlignedRow{}
		// Use extract functions for each column
		for _, col := range columns {
			fieldInfo := ColumnMap[col]
			fieldInfo.extractFunc(&ds, &row)
		}
		aligned = append(aligned, row)
	}

	return DatasourcesAlignedData{Datasources: aligned}
}

// CreateTableData creates table rows from aligned datasources using format functions
func CreateTableData(datasources []DatasourcesAlignedRow, columns []DatasourceColumn) [][]string {
	var data [][]string

	for _, dsRow := range datasources {
		var row []string
		for _, col := range columns {
			fieldInfo := ColumnMap[col]
			row = append(row, fieldInfo.formatFunc(&dsRow))
		}
		data = append(data, row)
	}

	return data
}
