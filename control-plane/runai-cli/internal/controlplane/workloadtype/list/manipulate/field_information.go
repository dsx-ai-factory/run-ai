package manipulate

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

var ColumnMap = map[WorkloadTypeColumn]FieldInformation{
	ColumnKeyName:           NameField,
	ColumnKeyCategory:       CategoryField,
	ColumnKeyPriority:       PriorityField,
	ColumnKeyPreemptibility: PreemptibilityField,
	ColumnKeyGroup:          GroupField,
	ColumnKeySource:         SourceField,
}

var defaultColumns = []WorkloadTypeColumn{
	ColumnKeyName,
	ColumnKeyCategory,
	ColumnKeyPriority,
	ColumnKeyPreemptibility,
	ColumnKeyGroup,
}

func GetDefaultColumns() []WorkloadTypeColumn {
	return defaultColumns
}

type FieldInformation struct {
	columnTitle string
	extractFunc func(wt *openapi.WorkloadTypeConfig, row *WorkloadTypeAlignedRow)
	formatFunc  func(row *WorkloadTypeAlignedRow) string
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
		columnWidth: DefaultWidthName,
		extractFunc: extractNameData,
		formatFunc:  formatNameValue,
	}
	CategoryField = FieldInformation{
		columnTitle: "Category",
		columnWidth: DefaultWidthCategory,
		extractFunc: extractCategoryData,
		formatFunc:  formatCategoryValue,
	}
	PriorityField = FieldInformation{
		columnTitle: "Priority",
		columnWidth: DefaultWidthPriority,
		extractFunc: extractPriorityData,
		formatFunc:  formatPriorityValue,
	}
	PreemptibilityField = FieldInformation{
		columnTitle: "Preemptibility",
		columnWidth: DefaultWidthPreemptibility,
		extractFunc: extractPreemptibilityData,
		formatFunc:  formatPreemptibilityValue,
	}
	GroupField = FieldInformation{
		columnTitle: "Group",
		columnWidth: DefaultWidthGroup,
		extractFunc: extractGroupData,
		formatFunc:  formatGroupValue,
	}
	SourceField = FieldInformation{
		columnTitle: "Source",
		columnWidth: DefaultWidthSource,
		extractFunc: extractSourceData,
		formatFunc:  formatSourceValue,
	}
)

func GetColumns(columns []WorkloadTypeColumn) []WorkloadTypeColumn {
	if len(columns) != 0 {
		return columns
	}
	return defaultColumns
}

func ParseWorkloadTypeColumns(columns []string) ([]WorkloadTypeColumn, error) {
	if len(columns) == 0 {
		return nil, nil
	}

	result := make([]WorkloadTypeColumn, 0, len(columns))
	var invalidColumns []string

	for _, columnName := range columns {
		trimmedKey := WorkloadTypeColumn(strings.ToLower(strings.TrimSpace(columnName)))

		if _, exists := ColumnMap[trimmedKey]; exists {
			result = append(result, trimmedKey)
		} else {
			invalidColumns = append(invalidColumns, fmt.Sprintf("\"%s\"", columnName))
		}
	}

	if len(invalidColumns) > 0 {
		return nil, fmt.Errorf(consts.ErrorInvalidColumns,
			strings.Join(invalidColumns, ", "))
	}

	return result, nil
}

func extractNameData(wt *openapi.WorkloadTypeConfig, row *WorkloadTypeAlignedRow) {
	row.Name = wt.Name
}

func extractCategoryData(wt *openapi.WorkloadTypeConfig, row *WorkloadTypeAlignedRow) {
	if wt.CategoryName != nil {
		row.CategoryName = *wt.CategoryName
	}
}

func extractPriorityData(wt *openapi.WorkloadTypeConfig, row *WorkloadTypeAlignedRow) {
	if wt.PriorityName != nil {
		row.PriorityName = *wt.PriorityName
	}
}

func extractPreemptibilityData(wt *openapi.WorkloadTypeConfig, row *WorkloadTypeAlignedRow) {
	row.Preemptibility = string(wt.Preemptibility)
}

func extractGroupData(wt *openapi.WorkloadTypeConfig, row *WorkloadTypeAlignedRow) {
	row.Group = wt.Group
}

func extractSourceData(wt *openapi.WorkloadTypeConfig, row *WorkloadTypeAlignedRow) {
	if wt.Source != nil {
		row.Source = *wt.Source
	}
}

func formatNameValue(row *WorkloadTypeAlignedRow) string {
	return row.Name
}

func formatCategoryValue(row *WorkloadTypeAlignedRow) string {
	if row.CategoryName != "" {
		return row.CategoryName
	}
	return noDataFormat
}

func formatPriorityValue(row *WorkloadTypeAlignedRow) string {
	if row.PriorityName != "" {
		return row.PriorityName
	}
	return noDataFormat
}

func formatPreemptibilityValue(row *WorkloadTypeAlignedRow) string {
	if row.Preemptibility != "" {
		return row.Preemptibility
	}
	return noDataFormat
}

func formatGroupValue(row *WorkloadTypeAlignedRow) string {
	if row.Group != "" {
		return row.Group
	}
	return noDataFormat
}

func formatSourceValue(row *WorkloadTypeAlignedRow) string {
	if row.Source != "" {
		return row.Source
	}
	return noDataFormat
}
