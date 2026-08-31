/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/mycredential/client"
)

var ColumnMap = map[MyCredentialColumn]FieldInformation{
	ColumnKeyType:        TypeField,
	ColumnKeyName:        NameField,
	ColumnKeyDescription: DescriptionField,
}

var defaultColumns = []MyCredentialColumn{
	ColumnKeyType,
	ColumnKeyName,
	ColumnKeyDescription,
}

func GetDefaultColumns() []MyCredentialColumn {
	return defaultColumns
}

type FieldInformation struct {
	columnTitle string
	extractFunc func(entry *client.MyCredentialEntry, row *MyCredentialAlignedRow)
	formatFunc  func(row *MyCredentialAlignedRow) string
	columnWidth int
}

func (f FieldInformation) GetColumnWidth() int {
	return f.columnWidth
}

func (f FieldInformation) GetColumnTitle() string {
	return f.columnTitle
}

var (
	TypeField = FieldInformation{
		columnTitle: "Type",
		columnWidth: 16,
		extractFunc: extractTypeData,
		formatFunc:  formatTypeValue,
	}
	NameField = FieldInformation{
		columnTitle: "Name",
		columnWidth: DefaultWidth,
		extractFunc: extractNameData,
		formatFunc:  formatNameValue,
	}
	DescriptionField = FieldInformation{
		columnTitle: "Description",
		columnWidth: DefaultWidth,
		extractFunc: extractDescriptionData,
		formatFunc:  formatDescriptionValue,
	}
)

func extractTypeData(entry *client.MyCredentialEntry, row *MyCredentialAlignedRow) {
	row.Type = entry.Type
}

func extractNameData(entry *client.MyCredentialEntry, row *MyCredentialAlignedRow) {
	row.Name = entry.Name
}

func extractDescriptionData(entry *client.MyCredentialEntry, row *MyCredentialAlignedRow) {
	row.Description = entry.Description
}

func formatTypeValue(row *MyCredentialAlignedRow) string {
	return row.Type
}

func formatNameValue(row *MyCredentialAlignedRow) string {
	return row.Name
}

func formatDescriptionValue(row *MyCredentialAlignedRow) string {
	return row.Description
}
