package manipulate

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
)

// Column definitions for workload list display
var (
	NameField = FieldInformation{
		columnTitle: "Workload",
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
	FrameworkField = FieldInformation{
		columnTitle: "Framework",
		columnWidth: DefaultWidth,
		extractFunc: extractFrameworkData,
		formatFunc:  formatFrameworkValue,
	}
	StatusField = FieldInformation{
		columnTitle: "Status",
		columnWidth: DefaultWidth,
		extractFunc: extractStatusData,
		formatFunc:  formatStatusValue,
	}
	ProjectField = FieldInformation{
		columnTitle: "Project",
		columnWidth: DefaultWidth,
		extractFunc: extractProjectData,
		formatFunc:  formatProjectValue,
	}
	CategoryField = FieldInformation{
		columnTitle: "Category",
		columnWidth: DefaultWidth,
		extractFunc: extractCategoryData,
		formatFunc:  formatCategoryValue,
	}
	PriorityField = FieldInformation{
		columnTitle: "Priority",
		columnWidth: DefaultWidth,
		extractFunc: extractPriorityData,
		formatFunc:  formatPriorityValue,
	}
	DepartmentField = FieldInformation{
		columnTitle: "Department",
		columnWidth: DefaultWidth,
		extractFunc: extractDepartmentData,
		formatFunc:  formatDepartmentValue,
	}
	SubmittedByField = FieldInformation{
		columnTitle: "Created By",
		columnWidth: DefaultWidth,
		extractFunc: extractSubmittedByData,
		formatFunc:  formatSubmittedByValue,
	}
	CreatedAtField = FieldInformation{
		columnTitle: "Creation Time",
		columnWidth: DefaultWidthPreemptible,
		extractFunc: extractCreatedAtData,
		formatFunc:  formatCreatedAtValue,
	}
	TotalPendingTimeField = FieldInformation{
		columnTitle: "Total Pending Time (s)",
		columnWidth: DefaultWidthTotalPendingTime,
		extractFunc: extractTotalPendingTimeData,
		formatFunc:  formatTotalPendingTimeValue,
	}
	AllocatedGPUMemoryField = FieldInformation{
		columnTitle: "GPU Memory Alloc.",
		columnWidth: DefaultWidthMemoryAllocation,
		extractFunc: extractAllocatedGPUMemoryData,
		formatFunc:  formatAllocatedGPUMemoryValue,
	}
	AllocatedCPUComputeField = FieldInformation{
		columnTitle: "CPU Compute Alloc.",
		columnWidth: DefaultWidthMemoryAllocation,
		extractFunc: extractAllocatedCPUComputeData,
		formatFunc:  formatAllocatedCPUComputeValue,
	}
	AllocatedCPUMemoryField = FieldInformation{
		columnTitle: "CPU Memory Alloc.",
		columnWidth: DefaultWidthMemoryAllocation,
		extractFunc: extractAllocatedCPUMemoryData,
		formatFunc:  formatAllocatedCPUMemoryValue,
	}
	CompletedAtField = FieldInformation{
		columnTitle: "Completion Time",
		columnWidth: DefaultWidth + 5, // Wider for timestamps
		extractFunc: extractCompletedAtData,
		formatFunc:  formatCompletedAtValue,
	}
	PreemptibleField = FieldInformation{
		columnTitle: "Preemptible",
		columnWidth: DefaultWidthPreemptible,
		extractFunc: extractPreemptibleData,
		formatFunc:  formatPreemptibleValue,
	}
	PodsField = FieldInformation{
		columnTitle: "Running/Req. Pods",
		columnWidth: DefaultWidthMemoryAllocation,
		extractFunc: extractPodsData,
		formatFunc:  formatPodsValue,
	}
	GPUAllocationField = FieldInformation{
		columnTitle: "GPU Alloc.",
		columnWidth: DefaultWidthGpuAllocation,
		extractFunc: extractGPUAllocationData,
		formatFunc:  formatGPUAllocationValue,
	}
	UrlsField = FieldInformation{
		columnTitle: "URLs",
		columnWidth: DefaultWidth,
		extractFunc: extractUrlsData,
		formatFunc:  formatUrlsValue,
	}
	ImageField = FieldInformation{
		columnTitle: "Image",
		columnWidth: DefaultWidth,
		extractFunc: extractImageData,
		formatFunc:  formatImageValue,
	}
	CurrentNodepoolField = FieldInformation{
		columnTitle: "Current Nodepool",
		columnWidth: DefaultWidthPreemptible,
		extractFunc: extractCurrentNodepoolData,
		formatFunc:  formatCurrentNodepoolValue,
	}
	RequestedNodepoolField = FieldInformation{
		columnTitle: "Requested Nodepool",
		columnWidth: DefaultWidthPreemptible,
		extractFunc: extractRequestedNodepoolData,
		formatFunc:  formatRequestedNodepoolValue,
	}
)

func ParseWorkloadColumns(columns []string) ([]WorkloadColumn, error) {
	// If no columns specified, return nil slice
	if len(columns) == 0 {
		return nil, nil
	}

	result := make([]WorkloadColumn, 0, len(columns))
	var invalidColumns []string

	for _, columnName := range columns {
		trimmedKey := WorkloadColumn(strings.ToLower(strings.TrimSpace(columnName)))

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
