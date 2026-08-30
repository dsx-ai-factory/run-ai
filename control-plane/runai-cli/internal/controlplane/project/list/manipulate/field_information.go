package manipulate

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	openapi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	"k8s.io/utils/ptr"
)

var ColumnMap = map[ProjectColumn]FieldInformation{
	ColumnKeyName:            NameField,
	ColumnKeyStatus:          StatusField,
	ColumnKeyDepartment:      DepartmentField,
	ColumnKeyGPUQuota:        GPUQuotaField,
	ColumnKeyAllocatedGPUs:   AllocatedGPUsField,
	ColumnKeyAllocationRatio: AllocationRatioField,
}

var defaultColumns = []ProjectColumn{
	ColumnKeyName,
	ColumnKeyStatus,
	ColumnKeyDepartment,
	ColumnKeyGPUQuota,
	ColumnKeyAllocatedGPUs,
	ColumnKeyAllocationRatio,
}

func GetDefaultColumns() []ProjectColumn {
	return defaultColumns
}

type FieldInformation struct {
	columnTitle string
	extractFunc func(project *openapi.Project, projectRow *ProjectsAlignedRow)
	formatFunc  func(projectRow *ProjectsAlignedRow) string
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
	StatusField = FieldInformation{
		columnTitle: "Status",
		columnWidth: DefaultWidthPhase,
		extractFunc: extractStatusData,
		formatFunc:  formatStatusValue,
	}
	DepartmentField = FieldInformation{
		columnTitle: "Department",
		columnWidth: DefaultWidthDepartment,
		extractFunc: extractDepartmentData,
		formatFunc:  formatDepartmentValue,
	}
	GPUQuotaField = FieldInformation{
		columnTitle: "GPU Quota",
		columnWidth: DefaultWidthGPUQuota,
		extractFunc: extractGPUQuotaData,
		formatFunc:  formatGPUQuotaValue,
	}
	AllocatedGPUsField = FieldInformation{
		columnTitle: "Allocated GPUs",
		columnWidth: DefaultWidthAllocatedGPUs,
		extractFunc: extractAllocatedGPUsData,
		formatFunc:  formatAllocatedGPUsValue,
	}
	AllocationRatioField = FieldInformation{
		columnTitle: "GPU Allocation Ratio",
		columnWidth: DefaultWidthAllocationRatio,
		extractFunc: extractAllocationRatioData,
		formatFunc:  formatAllocationRatioValue,
	}
)

func GetColumns(columns []ProjectColumn) []ProjectColumn {
	if len(columns) != 0 {
		return columns
	}
	return defaultColumns
}

func ParseProjectColumns(columns []string) ([]ProjectColumn, error) {
	if len(columns) == 0 {
		return nil, nil
	}

	result := make([]ProjectColumn, 0, len(columns))
	var invalidColumns []string

	for _, columnName := range columns {
		trimmedKey := ProjectColumn(strings.ToLower(strings.TrimSpace(columnName)))

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

func extractNameData(project *openapi.Project, projectRow *ProjectsAlignedRow) {
	projectRow.Name = project.Name
}

func extractStatusData(project *openapi.Project, projectRow *ProjectsAlignedRow) {
	projectRow.Status = string(project.Status.GetPhase())
}

func extractDepartmentData(project *openapi.Project, projectRow *ProjectsAlignedRow) {
	if project.Parent.IsSet() {
		projectRow.Department = project.Parent.Get().Name
	}
}

func extractGPUQuotaData(project *openapi.Project, projectRow *ProjectsAlignedRow) {
	sumQuotaGpu := 0.0
	for _, resource := range project.Resources {
		sumQuotaGpu += resource.Gpu.Deserved
	}
	projectRow.GPUQuota = sumQuotaGpu
}

func extractAllocatedGPUsData(project *openapi.Project, projectRow *ProjectsAlignedRow) {
	sumAllocatedGpu := 0.0
	for _, quotaStatus := range project.Status.NodePoolQuotaStatuses {
		if quotaStatus.GetAllocated().Gpu != nil {
			sumAllocatedGpu += ptr.Deref[float64](quotaStatus.GetAllocated().Gpu, 0)
		}
	}
	projectRow.AllocatedGPUs = sumAllocatedGpu
}

func extractAllocationRatioData(project *openapi.Project, projectRow *ProjectsAlignedRow) {
	sumQuotaGpu := 0.0
	for _, resource := range project.Resources {
		sumQuotaGpu += resource.Gpu.Deserved
	}

	sumAllocatedGpu := 0.0
	for _, quotaStatus := range project.Status.NodePoolQuotaStatuses {
		if quotaStatus.GetAllocated().Gpu != nil {
			sumAllocatedGpu += ptr.Deref[float64](quotaStatus.GetAllocated().Gpu, 0)
		}
	}

	allocationRatioPercent := 0.0
	if sumQuotaGpu > 0 {
		allocationRatioPercent = (sumAllocatedGpu / sumQuotaGpu) * 100
	}
	projectRow.AllocationRatio = allocationRatioPercent
}

func formatNameValue(projectRow *ProjectsAlignedRow) string {
	return projectRow.Name
}

func formatStatusValue(projectRow *ProjectsAlignedRow) string {
	return projectRow.Status
}

func formatDepartmentValue(projectRow *ProjectsAlignedRow) string {
	if projectRow.Department != "" {
		return projectRow.Department
	}
	return noDataFormat
}

func formatGPUQuotaValue(projectRow *ProjectsAlignedRow) string {
	if projectRow.GPUQuota == -1 {
		return noDataFormat
	}
	return fmt.Sprintf("%.2f", projectRow.GPUQuota)
}

func formatAllocatedGPUsValue(projectRow *ProjectsAlignedRow) string {
	return fmt.Sprintf("%.2f", projectRow.AllocatedGPUs)
}

func formatAllocationRatioValue(projectRow *ProjectsAlignedRow) string {
	return fmt.Sprintf("%.2f%%", projectRow.AllocationRatio)
}
