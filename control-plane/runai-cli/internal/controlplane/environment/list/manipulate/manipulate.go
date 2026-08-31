/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	"fmt"
	"strings"

	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

// ParseEnvironmentColumns validates and converts column names to EnvironmentColumn types
func ParseEnvironmentColumns(columnNames []string) ([]EnvironmentColumn, error) {
	var columns []EnvironmentColumn
	for _, name := range columnNames {
		normalizedName := strings.ToLower(strings.TrimSpace(name))
		column := EnvironmentColumn(normalizedName)

		if _, exists := ColumnMap[column]; !exists {
			return nil, fmt.Errorf("invalid column name: %s", name)
		}
		columns = append(columns, column)
	}
	return columns, nil
}

// BriefEnvironmentFromAsset converts an EnvironmentAsset to BriefEnvironment
func BriefEnvironmentFromAsset(env openapi.EnvironmentAsset) BriefEnvironment {
	// Extract scope
	scope := string(env.Meta.Scope)
	if scope == "" {
		scope = "Unknown"
	}

	// Extract image using the getter method
	image := ""
	if env.Spec.HasImage() {
		image = env.Spec.GetImage()
	}

	// Format tools
	tools := formatTools(env.Spec.GetConnections())

	// Format workload types from AssetMeta.WorkloadSupportedTypes
	workloadTypes := formatWorkloadTypes(env.Meta.WorkloadSupportedTypes)

	return BriefEnvironment{
		Name:          env.Meta.Name,
		Scope:         scope,
		Image:         image,
		Tools:         tools,
		WorkloadTypes: workloadTypes,
	}
}

// CreateAlignedEnvironments transforms environment assets into aligned format based on selected columns
func CreateAlignedEnvironments(environments []openapi.EnvironmentAsset, columns []EnvironmentColumn) EnvironmentsAlignedData {
	var aligned []EnvironmentsAlignedRow

	for _, env := range environments {
		row := EnvironmentsAlignedRow{}
		// Use extract functions for each column
		for _, col := range columns {
			fieldInfo := ColumnMap[col]
			fieldInfo.extractFunc(&env, &row)
		}
		aligned = append(aligned, row)
	}

	return EnvironmentsAlignedData{Environments: aligned}
}

// CreateTableData creates table rows from aligned environments using format functions
func CreateTableData(environments []EnvironmentsAlignedRow, columns []EnvironmentColumn) [][]string {
	var data [][]string

	for _, envRow := range environments {
		var row []string
		for _, col := range columns {
			fieldInfo := ColumnMap[col]
			row = append(row, fieldInfo.formatFunc(&envRow))
		}
		data = append(data, row)
	}

	return data
}
