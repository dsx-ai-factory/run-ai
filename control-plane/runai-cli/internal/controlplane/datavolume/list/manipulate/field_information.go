package manipulate

import (
	openapi "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"
	"k8s.io/utils/ptr"
)

var ColumnMap = map[DatavolumeColumn]FieldInformation{
	ColumnKeyName:      NameField,
	ColumnKeyStatus:    StatusField,
	ColumnKeyScope:     ScopeField,
	ColumnKeyOriginPvc: OriginPvcField,
}

var defaultColumns = []DatavolumeColumn{
	ColumnKeyName,
	ColumnKeyStatus,
	ColumnKeyScope,
	ColumnKeyOriginPvc,
}

func GetDefaultColumns() []DatavolumeColumn {
	return defaultColumns
}

func GetColumns(columns []DatavolumeColumn) []DatavolumeColumn {
	if len(columns) > 0 {
		return columns
	}
	return GetDefaultColumns()
}

type FieldInformation struct {
	columnTitle string
	extractFunc func(dv *openapi.DatavolumeNoSharedScopes, dvRow *DatavolumesAlignedRow)
	formatFunc  func(dvRow *DatavolumesAlignedRow) string
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
	ScopeField = FieldInformation{
		columnTitle: "Scope",
		columnWidth: DefaultWidth,
		extractFunc: extractScopeData,
		formatFunc:  formatScopeValue,
	}
	OriginPvcField = FieldInformation{
		columnTitle: "Origin Pvc",
		columnWidth: DefaultWidth,
		extractFunc: extractOriginPvcData,
		formatFunc:  formatOriginPvcValue,
	}
	StatusField = FieldInformation{
		columnTitle: "Status",
		columnWidth: DefaultWidth,
		extractFunc: extractStatusData,
		formatFunc:  formatStatusValue,
	}
)

func extractNameData(dv *openapi.DatavolumeNoSharedScopes, dvRow *DatavolumesAlignedRow) {
	dvRow.Name = dv.Name
}

func extractScopeData(dv *openapi.DatavolumeNoSharedScopes, dvRow *DatavolumesAlignedRow) {
	dvRow.Scope = extractScope(dv)
}

func extractOriginPvcData(dv *openapi.DatavolumeNoSharedScopes, dvRow *DatavolumesAlignedRow) {
	dvRow.OriginPvc = dv.OriginPvcName
}

func formatNameValue(dvRow *DatavolumesAlignedRow) string {
	return dvRow.Name
}

func formatScopeValue(dvRow *DatavolumesAlignedRow) string {
	return dvRow.Scope
}

func formatOriginPvcValue(dvRow *DatavolumesAlignedRow) string {
	return dvRow.OriginPvc
}

func extractStatusData(dv *openapi.DatavolumeNoSharedScopes, dvRow *DatavolumesAlignedRow) {
	dvRow.Status = datavolumePhase(dv)
}

func formatStatusValue(dvRow *DatavolumesAlignedRow) string {
	return datavolumeStatusString(dvRow.Status)
}

func extractScope(dv *openapi.DatavolumeNoSharedScopes) string {
	if ptr.Deref(dv.ClusterId, "") != "" {
		return "Cluster"
	}
	if ptr.Deref(dv.DepartmentId, "") != "" {
		return "Department"
	}
	return "Project"
}

func datavolumePhase(dv *openapi.DatavolumeNoSharedScopes) string {
	if dv == nil {
		return ""
	}
	return string(dv.Status.GetPhase())
}

func datavolumeStatusString(phase string) string {
	switch openapi.DatavolumePhase(phase) {
	case openapi.DATAVOLUMEPHASE_READY:
		return "No issues found"
	case openapi.DATAVOLUMEPHASE_NOT_READY:
		return "Issues found"
	default:
		return ""
	}
}
