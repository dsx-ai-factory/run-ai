/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
)

var ColumnMap = map[DatasourceColumn]FieldInformation{
	ColumnKeyName:  NameField,
	ColumnKeyType:  TypeField,
	ColumnKeyScope: ScopeField,
}

var defaultColumns = []DatasourceColumn{
	ColumnKeyName,
	ColumnKeyType,
	ColumnKeyScope,
}

var titleCaser = cases.Title(language.English)

func GetDefaultColumns() []DatasourceColumn {
	return defaultColumns
}

func GetColumns(columns []DatasourceColumn) []DatasourceColumn {
	if len(columns) > 0 {
		return columns
	}
	return GetDefaultColumns()
}

type FieldInformation struct {
	columnTitle string
	extractFunc func(ds *openapi.DatasourceListResponseEntry, dsRow *DatasourcesAlignedRow)
	formatFunc  func(dsRow *DatasourcesAlignedRow) string
	columnWidth int
}

func (f FieldInformation) GetColumnWidth() int {
	return f.columnWidth
}

func (f FieldInformation) GetColumnTitle() string {
	return f.columnTitle
}

// Field definitions for datasource columns
var (
	NameField = FieldInformation{
		columnTitle: "Name",
		columnWidth: DefaultWidth,
		extractFunc: extractNameData,
		formatFunc:  formatNameValue,
	}
	TypeField = FieldInformation{
		columnTitle: "Type",
		columnWidth: DefaultWidth,
		extractFunc: extractTypeData,
		formatFunc:  formatTypeValue,
	}
	ScopeField = FieldInformation{
		columnTitle: "Scope",
		columnWidth: DefaultWidth,
		extractFunc: extractScopeData,
		formatFunc:  formatScopeValue,
	}
)

func extractNameData(ds *openapi.DatasourceListResponseEntry, dsRow *DatasourcesAlignedRow) {
	dsRow.Name = ds.Meta.Name
}

func extractTypeData(ds *openapi.DatasourceListResponseEntry, dsRow *DatasourcesAlignedRow) {
	dsRow.Type = formatType(ds.Meta.Kind)
}

func extractScopeData(ds *openapi.DatasourceListResponseEntry, dsRow *DatasourcesAlignedRow) {
	dsRow.Scope = extractScope(ds.Meta.Scope)
}

func formatNameValue(dsRow *DatasourcesAlignedRow) string {
	return dsRow.Name
}

func formatTypeValue(dsRow *DatasourcesAlignedRow) string {
	return dsRow.Type
}

func formatScopeValue(dsRow *DatasourcesAlignedRow) string {
	return dsRow.Scope
}

func formatType(kind openapi.AssetKind) string {
	switch kind {
	case openapi.ASSETKIND_NFS:
		return "Nfs"
	case openapi.ASSETKIND_PVC:
		return "Pvc"
	case openapi.ASSETKIND_S3:
		return "S3"
	case openapi.ASSETKIND_GIT:
		return "Git"
	case openapi.ASSETKIND_HOST_PATH:
		return "HostPath"
	case openapi.ASSETKIND_CONFIG_MAP:
		return "ConfigMap"
	case openapi.ASSETKIND_SECRET_VOLUME:
		return "Secret"
	case openapi.ASSETKIND_DATA_VOLUME:
		return "DataVolume"
	default:
		return titleCaser.String(string(kind))
	}
}
