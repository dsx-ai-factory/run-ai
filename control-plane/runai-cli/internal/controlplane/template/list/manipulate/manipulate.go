/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	"fmt"
	"strings"

	cliutils "github.com/run-ai/runai/control-plane/runai-cli/internal/utils"
	wcgenerated "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

// TemplateColumn represents a column in the template list
type TemplateColumn string

const (
	ColumnKeyName                   TemplateColumn = "name"
	ColumnKeyScopeType              TemplateColumn = "scopetype"
	ColumnKeyImage                  TemplateColumn = "image"
	ColumnKeyGpuComputeRequest      TemplateColumn = "gpucomputerequest"
	ColumnKeyWorkloadType           TemplateColumn = "workloadtype"
	ColumnKeyMandatoryFieldsMissing TemplateColumn = "mandatoryfieldsmissing"
	ColumnKeyCreatedAt              TemplateColumn = "createdat"
)

// FieldInformation represents a column in the template list with its field name and display title
type FieldInformation struct {
	columnTitle string
	extractFunc func(template *wcgenerated.SupersetTemplate, templateRow *TemplateAlignedRow)
	formatFunc  func(templateRow *TemplateAlignedRow) string
	columnWidth int
}

func (f FieldInformation) GetColumnWidth() int {
	return f.columnWidth
}

func (f FieldInformation) GetColumnTitle() string {
	return f.columnTitle
}

// ColumnMap maps column keys to their corresponding column definitions
var ColumnMap = map[TemplateColumn]FieldInformation{
	ColumnKeyName:                   NameField,
	ColumnKeyScopeType:              ScopeField,
	ColumnKeyImage:                  ImageField,
	ColumnKeyGpuComputeRequest:      GpuComputeRequestField,
	ColumnKeyWorkloadType:           WorkloadTypeField,
	ColumnKeyMandatoryFieldsMissing: MandatoryFieldsMissingField,
	ColumnKeyCreatedAt:              CreatedAtField,
}

var defaultColumns = []TemplateColumn{
	ColumnKeyName,
	ColumnKeyScopeType,
	ColumnKeyImage,
	ColumnKeyGpuComputeRequest,
	ColumnKeyWorkloadType,
	ColumnKeyMandatoryFieldsMissing,
}

func GetDefaultColumns() []TemplateColumn {
	return defaultColumns
}

// ParseTemplateColumns parses and validates the provided column names
func ParseTemplateColumns(columns []string) ([]TemplateColumn, error) {
	var parsedColumns []TemplateColumn

	for _, column := range columns {
		columnKey := TemplateColumn(strings.ToLower(strings.TrimSpace(column)))
		if _, exists := ColumnMap[columnKey]; !exists {
			return nil, fmt.Errorf("unknown column: %s", column)
		}
		parsedColumns = append(parsedColumns, columnKey)
	}

	return parsedColumns, nil
}

// CreateAlignedTemplates transforms raw API template data into display-ready rows
// by extracting only the requested columns.
func CreateAlignedTemplates(templatesRaw Data, columns []TemplateColumn) TemplatesAlignedOutput {
	templatesAligned := make([]TemplateAlignedRow, 0)

	for _, tmpl := range templatesRaw.Templates {
		templateRow := TemplateAlignedRow{}

		for _, column := range columns {
			ColumnMap[column].extractFunc(&tmpl, &templateRow)
		}

		templatesAligned = append(templatesAligned, templateRow)
	}

	return TemplatesAlignedOutput{
		Templates: templatesAligned,
		NextToken: templatesRaw.NextToken,
	}
}

// CreateTableData converts aligned rows into a 2D string array for table rendering.
// Process:
//   - Converts each TemplateAlignedRow into a string array (one row per template)
//   - Uses ColumnMap formatFunc to convert field values to display strings
//   - Sorts by template name (if name column is present)
func CreateTableData(templatesAligned []TemplateAlignedRow, columns []TemplateColumn) [][]string {
	var data [][]string

	// Convert each aligned row to string array using format functions
	for _, tmpl := range templatesAligned {
		var templateRow []string
		for _, column := range columns {
			// Each column's formatFunc knows how to convert its field to a string
			templateRow = append(templateRow, ColumnMap[column].formatFunc(&tmpl))
		}
		data = append(data, templateRow)
	}

	// Find the name column index for sorting
	nameIndex := -1
	for i, column := range columns {
		if column == ColumnKeyName {
			nameIndex = i
			break
		}
	}

	// Sort by name alphabetically if name column exists
	if nameIndex >= 0 {
		cliutils.MultiColumnSort(data, []cliutils.SortColumn{
			{Index: nameIndex, Order: cliutils.Ascending},
		})
	}

	return data
}

// formatInt32 formats an int32 value
func formatInt32(val int32) string {
	return fmt.Sprintf("%d", val)
}
