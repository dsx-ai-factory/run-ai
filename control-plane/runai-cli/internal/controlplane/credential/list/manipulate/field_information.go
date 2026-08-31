/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

var ColumnMap = map[CredentialColumn]FieldInformation{
	ColumnKeyName:    NameField,
	ColumnKeyType:    TypeField,
	ColumnKeyScope:   ScopeField,
	ColumnKeyK8sName: K8sNameField,
	ColumnKeyStatus:  StatusField,
}

var defaultColumns = []CredentialColumn{
	ColumnKeyName,
	ColumnKeyK8sName,
	ColumnKeyStatus,
	ColumnKeyType,
	ColumnKeyScope,
}

var titleCaser = cases.Title(language.English)

func GetDefaultColumns() []CredentialColumn {
	return defaultColumns
}

func GetColumns(columns []CredentialColumn) []CredentialColumn {
	if len(columns) > 0 {
		return columns
	}
	return GetDefaultColumns()
}

type FieldInformation struct {
	columnTitle string
	extractFunc func(cred *openapi.CredentialsListResponseEntry, credRow *CredentialsAlignedRow)
	formatFunc  func(credRow *CredentialsAlignedRow) string
	columnWidth int
}

func (f FieldInformation) GetColumnWidth() int {
	return f.columnWidth
}

func (f FieldInformation) GetColumnTitle() string {
	return f.columnTitle
}

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
	K8sNameField = FieldInformation{
		columnTitle: "K8s Name",
		columnWidth: DefaultWidth,
		extractFunc: extractK8sNameData,
		formatFunc:  formatK8sNameValue,
	}
	StatusField = FieldInformation{
		columnTitle: "Status",
		columnWidth: DefaultWidth,
		extractFunc: extractStatusData,
		formatFunc:  formatStatusValue,
	}
)

func extractNameData(cred *openapi.CredentialsListResponseEntry, credRow *CredentialsAlignedRow) {
	credRow.Name = cred.Meta.Name
}

func extractTypeData(cred *openapi.CredentialsListResponseEntry, credRow *CredentialsAlignedRow) {
	credRow.Type = formatType(cred.Meta.Kind)
}

func extractScopeData(cred *openapi.CredentialsListResponseEntry, credRow *CredentialsAlignedRow) {
	credRow.Scope = extractScope(cred.Meta.Scope)
}

func extractK8sNameData(cred *openapi.CredentialsListResponseEntry, credRow *CredentialsAlignedRow) {
	credRow.K8sName = extractK8sName(cred)
}

func formatNameValue(credRow *CredentialsAlignedRow) string {
	return credRow.Name
}

func formatTypeValue(credRow *CredentialsAlignedRow) string {
	return credRow.Type
}

func formatScopeValue(credRow *CredentialsAlignedRow) string {
	return credRow.Scope
}

func formatK8sNameValue(credRow *CredentialsAlignedRow) string {
	return credRow.K8sName
}

func extractStatusData(cred *openapi.CredentialsListResponseEntry, credRow *CredentialsAlignedRow) {
	credRow.Status = extractStatus(cred)
}

func formatStatusValue(credRow *CredentialsAlignedRow) string {
	return credRow.Status
}

func extractStatus(cred *openapi.CredentialsListResponseEntry) string {
	statusInfo, statusSet := cred.GetStatusOk()
	if !statusSet || statusInfo == nil {
		return ""
	}
	statusEnum, statusEnumSet := statusInfo.GetStatusOk()
	if !statusEnumSet || statusEnum == nil {
		return ""
	}
	return string(*statusEnum)
}

func extractK8sName(cred *openapi.CredentialsListResponseEntry) string {
	if !cred.HasClusterInfo() {
		return ""
	}
	clusterInfo := cred.GetClusterInfo()
	resources := clusterInfo.GetResources()
	if len(resources) == 0 {
		return ""
	}
	return resources[0].GetName()
}

func formatType(kind openapi.AssetKind) string {
	switch kind {
	case openapi.ASSETKIND_GENERIC_SECRET:
		return "Generic secret"
	case openapi.ASSETKIND_DOCKER_REGISTRY:
		return "Docker registry"
	case openapi.ASSETKIND_NGC_API_KEY:
		return "NGC API key"
	case openapi.ASSETKIND_PASSWORD:
		return "Password"
	case openapi.ASSETKIND_ACCESS_KEY:
		return "Access key"
	default:
		return titleCaser.String(string(kind))
	}
}
