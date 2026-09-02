/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	cliutils "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/utils"
	workloadservice "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"k8s.io/utils/ptr"
)

// ColumnMap maps column keys to their corresponding column definitions
var ColumnMap = map[WorkloadColumn]FieldInformation{
	ColumnKeyName:                NameField,
	ColumnKeyType:                TypeField,
	ColumnKeyFramework:           FrameworkField,
	ColumnKeyStatus:              StatusField,
	ColumnKeyProject:             ProjectField,
	ColumnKeyPreemptible:         PreemptibleField,
	ColumnKeyPods:                PodsField,
	ColumnKeyGPUAlloc:            GPUAllocationField,
	ColumnKeyCategory:            CategoryField,
	ColumnKeyPriority:            PriorityField,
	ColumnKeyDepartment:          DepartmentField,
	ColumnKeySubmittedBy:         SubmittedByField,
	ColumnKeyCreatedAt:           CreatedAtField,
	ColumnKeyTotalPendingTime:    TotalPendingTimeField,
	ColumnKeyAllocatedGPUMemory:  AllocatedGPUMemoryField,
	ColumnKeyAllocatedCPUCompute: AllocatedCPUComputeField,
	ColumnKeyAllocatedCPUMemory:  AllocatedCPUMemoryField,
	ColumnKeyCompletedAt:         CompletedAtField,
	ColumnKeyUrls:                UrlsField,
	ColumnKeyImage:               ImageField,
	ColumnKeyCurrentNodepool:     CurrentNodepoolField,
	ColumnKeyRequestedNodepool:   RequestedNodepoolField,
}
var defaultColumns = []WorkloadColumn{
	ColumnKeyName,
	ColumnKeyType,
	ColumnKeyFramework,
	ColumnKeyStatus,
	ColumnKeyProject,
	ColumnKeyPreemptible,
	ColumnKeyPods,
	ColumnKeyGPUAlloc,
}

func GetDefaultColumns() []WorkloadColumn {
	return defaultColumns
}

// FieldInformation represents a column in the workload list with its field name and display title
type FieldInformation struct {
	columnTitle string
	extractFunc func(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow)
	formatFunc  func(workloadRow *WorkloadsAlignedRow) string
	columnWidth int
}

func (f FieldInformation) GetColumnWidth() int {
	return f.columnWidth
}
func (f FieldInformation) GetColumnTitle() string {
	return f.columnTitle
}

func GetColumns(columns []WorkloadColumn, deleted bool) []WorkloadColumn {
	if len(columns) != 0 {
		return columns
	}
	columns = defaultColumns
	if !deleted {
		return columns
	}
	cols := make([]WorkloadColumn, 0)
	for _, col := range columns {
		if col != ColumnKeyStatus {
			cols = append(cols, col)
		}
	}

	return cols // Return the modified slice
}
func CreateAlignedWorkloads(workloadsRaw Data, columns []WorkloadColumn) WorkloadsAlignedOutput {
	workloadsAligned := make([]WorkloadsAlignedRow, 0)
	for _, workload := range workloadsRaw.Workloads {
		workloadRow := &WorkloadsAlignedRow{}
		for _, column := range columns {
			ColumnMap[column].extractFunc(ptr.To(workload), workloadRow)
		}
		workloadsAligned = append(workloadsAligned, *workloadRow)
	}
	return WorkloadsAlignedOutput{
		Workloads: workloadsAligned,
		NextToken: workloadsRaw.NextToken,
	}
}
func CreateTableData(workloadsAligned []WorkloadsAlignedRow, columns []WorkloadColumn) [][]string {
	var (
		data [][]string
	)
	for _, workload := range workloadsAligned {
		var workloadRow []string
		for _, column := range columns {
			workloadRow = append(workloadRow, ColumnMap[column].formatFunc(ptr.To(workload)))
		}
		data = append(data, workloadRow)
	}
	cliutils.MultiColumnSort(data, []cliutils.SortColumn{
		{Index: nameColumnIndex, Order: cliutils.Ascending},
	})
	return data
}

// Non-trivial extract functions
func getPodsValue(workload *workloadservice.Workload) *Pods {
	requestedPods := workload.RequestedPods
	runningPods := &workload.RunningPods
	if requestedPods == nil {
		return &Pods{RunningPods: runningPods}
	}
	if requestedPods.Number.IsSet() {
		return &Pods{
			RunningPods:   runningPods,
			RequestedPods: requestedPods.Number.Get(),
		}
	}

	if requestedPods.Min.IsSet() && requestedPods.Max.IsSet() {
		return &Pods{
			RunningPods:      runningPods,
			RequestedPodsMax: requestedPods.Max.Get(),
			RequestedPodsMin: requestedPods.Min.Get(),
		}
	}

	if requestedPods.Parallelism.IsSet() {
		return &Pods{
			RunningPods: runningPods,
			Parallelism: requestedPods.Parallelism.Get(),
		}
	}
	if requestedPods.Completions.IsSet() {
		return &Pods{
			RunningPods: runningPods,
			Completions: requestedPods.Completions.Get(),
		}
	}
	return &Pods{RunningPods: runningPods}
}

func extractFrameworkData(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow) {
	framework := RunaiStandardFramework
	if workload.DistributedFramework != nil {
		framework = *workload.DistributedFramework
	}
	workloadRow.DistributedFramework = &framework
}

func extractStatusData(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow) {
	if workload.Phase != "" && workload.Phase.IsValid() {
		workloadRow.Status = &workload.Phase
	} else {
		workloadRow.Status, _ = workloadservice.NewPhaseFromValue(string(workloadservice.PHASE_UNKNOWN))
	}
}

func extractPreemptibleData(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow) {
	if workload.Preemptible.IsSet() {
		workloadRow.Preemptible = workload.Preemptible.Get()
	}
}

// Non-trivial format functions
func formatTypeValue(workloadRow *WorkloadsAlignedRow) string {
	workloadType := workloadRow.Type
	if workloadType == api.Distributed.StringCapitalize() {
		workloadType = api.Training.StringCapitalize()
	}
	return workloadType
}

func formatStatusValue(workloadRow *WorkloadsAlignedRow) string {
	if workloadRow.Status != nil {
		return string(*workloadRow.Status)
	}
	return ""
}

func formatPreemptibleValue(workloadRow *WorkloadsAlignedRow) string {
	if workloadRow.Preemptible == nil {
		return noDataFormat
	}
	if *workloadRow.Preemptible {
		return "Yes"
	} else {
		return "No"
	}

}

func formatPodsValue(workloadRow *WorkloadsAlignedRow) string {
	pods := workloadRow.Pods

	// Get running pods value or default to 0
	runningPods := int32(0)
	if pods.RunningPods != nil {
		runningPods = *pods.RunningPods
	}

	// Priority order based on what fields are set
	if pods.RequestedPods != nil {
		// Fixed number: "3/5"
		return fmt.Sprintf("%d/%d", runningPods, *pods.RequestedPods)
	}

	if pods.RequestedPodsMin != nil && pods.RequestedPodsMax != nil {
		// Range: "3/2-8"
		return fmt.Sprintf("%d/%d-%d", runningPods, *pods.RequestedPodsMin, *pods.RequestedPodsMax)
	}

	if pods.Parallelism != nil {
		return fmt.Sprintf("%d/%d", runningPods, *pods.Parallelism)
	}

	if pods.Completions != nil {
		return fmt.Sprintf("%d/-", runningPods)
	}

	if pods.RequestedPodsMin == nil && pods.RequestedPods == nil && pods.RequestedPodsMax == nil &&
		pods.Parallelism == nil && pods.Completions == nil {
		return fmt.Sprintf("%d/-", runningPods)
	}

	return noDataFormat
}

// Extract functions for all field types
func extractNameData(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow) {
	workloadRow.Name = workload.Name
}

func extractTypeData(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow) {
	workloadRow.Type = workload.Type
}

func extractProjectData(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow) {
	workloadRow.Project = workload.ProjectName
}

func extractCategoryData(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow) {
	if workload.Category != "" {
		workloadRow.Category = workload.Category
	}
}

func extractPriorityData(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow) {
	if workload.Priority.IsSet() {
		workloadRow.Priority = *workload.Priority.Get()
	}
}

func extractDepartmentData(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow) {
	if workload.DepartmentName != "" {
		workloadRow.Department = workload.DepartmentName
	}
}

func extractSubmittedByData(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow) {
	if workload.SubmittedBy != nil {
		workloadRow.SubmittedBy = *workload.SubmittedBy
	}
}

func extractCreatedAtData(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow) {
	if !workload.CreatedAt.IsZero() {
		workloadRow.CreatedAt = ptr.To(workload.CreatedAt)
	}
}

func extractTotalPendingTimeData(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow) {
	if workload.TotalPendingTimeSeconds.IsSet() {
		workloadRow.TotalPendingTimeSeconds = *workload.TotalPendingTimeSeconds.Get()
	}
}

func extractAllocatedGPUMemoryData(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow) {
	if workload.AllocatedResources.IsSet() && workload.AllocatedResources.Get().GpuMemory.IsSet() {
		workloadRow.AllocatedGPUMemory = *workload.AllocatedResources.Get().GpuMemory.Get()
	}
}

func extractAllocatedCPUComputeData(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow) {
	if workload.AllocatedResources.IsSet() && workload.AllocatedResources.Get().Cpu.IsSet() {
		workloadRow.AllocatedCPUCompute = *workload.AllocatedResources.Get().Cpu.Get()
	}
}

func extractAllocatedCPUMemoryData(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow) {
	if workload.AllocatedResources.IsSet() && workload.AllocatedResources.Get().CpuMemory.IsSet() {
		workloadRow.AllocatedCPUMemory = *workload.AllocatedResources.Get().CpuMemory.Get()
	}
}

func extractCompletedAtData(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow) {
	if workload.CompletedAt.IsSet() {
		workloadRow.CompletedAt = ptr.To(*workload.CompletedAt.Get())
	}
}

func extractPodsData(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow) {
	workloadRow.Pods = getPodsValue(workload)
}

func extractGPUAllocationData(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow) {
	workloadRow.GpuAllocation = workload.AllocatedResources.Get().GetGpu()
}

func extractUrlsData(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow) {
	if len(workload.Urls) > 0 {
		// Join URLs with comma separator
		workloadRow.Urls = strings.Join(workload.Urls, ", ")
	}
}

func extractImageData(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow) {
	if len(workload.Images) > 0 {
		workloadRow.Image = strings.Join(workload.Images, ", ")
	}
}

func extractCurrentNodepoolData(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow) {
	if len(workload.CurrentNodePools) > 0 {
		workloadRow.CurrentNodepool = strings.Join(workload.CurrentNodePools, ", ")
	}
}

func extractRequestedNodepoolData(workload *workloadservice.Workload, workloadRow *WorkloadsAlignedRow) {
	if len(workload.RequestedNodePools) > 0 {
		workloadRow.RequestedNodepool = strings.Join(workload.RequestedNodePools, ", ")
	}
}

// Format functions for all field types
func formatNameValue(workloadRow *WorkloadsAlignedRow) string {
	return workloadRow.Name
}

func formatFrameworkValue(workloadRow *WorkloadsAlignedRow) string {
	return *workloadRow.DistributedFramework
}

func formatProjectValue(workloadRow *WorkloadsAlignedRow) string {
	return workloadRow.Project
}

func formatCategoryValue(workloadRow *WorkloadsAlignedRow) string {
	if workloadRow.Category != "" {
		return workloadRow.Category
	}
	return noDataFormat
}

func formatPriorityValue(workloadRow *WorkloadsAlignedRow) string {
	if workloadRow.Priority != 0 {
		return fmt.Sprintf("%d", workloadRow.Priority)
	}
	return noDataFormat
}

func formatDepartmentValue(workloadRow *WorkloadsAlignedRow) string {
	if workloadRow.Department != "" {
		return workloadRow.Department
	}
	return noDataFormat
}

func formatSubmittedByValue(workloadRow *WorkloadsAlignedRow) string {
	if workloadRow.SubmittedBy != "" {
		return workloadRow.SubmittedBy
	}
	return noDataFormat
}

func formatCreatedAtValue(workloadRow *WorkloadsAlignedRow) string {
	if workloadRow.CreatedAt != nil && !workloadRow.CreatedAt.IsZero() {
		return workloadRow.CreatedAt.Format(timeFieldsLayout)
	}
	return noDataFormat
}

func formatTotalPendingTimeValue(workloadRow *WorkloadsAlignedRow) string {
	if workloadRow.TotalPendingTimeSeconds > 0 {
		return fmt.Sprintf("%d", workloadRow.TotalPendingTimeSeconds)
	}
	return noDataFormat
}

func formatAllocatedGPUMemoryValue(workloadRow *WorkloadsAlignedRow) string {
	if workloadRow.AllocatedGPUMemory != "" {
		return workloadRow.AllocatedGPUMemory
	}
	return noDataFormat
}

func formatAllocatedCPUComputeValue(workloadRow *WorkloadsAlignedRow) string {
	if workloadRow.AllocatedCPUCompute != 0 {
		return fmt.Sprintf("%.2f", workloadRow.AllocatedCPUCompute)
	}
	return noDataFormat
}

func formatAllocatedCPUMemoryValue(workloadRow *WorkloadsAlignedRow) string {
	if workloadRow.AllocatedCPUMemory != "" {
		return workloadRow.AllocatedCPUMemory
	}
	return noDataFormat
}

func formatCompletedAtValue(workloadRow *WorkloadsAlignedRow) string {
	if workloadRow.CompletedAt != nil && !workloadRow.CompletedAt.IsZero() {
		return workloadRow.CompletedAt.Format(timeFieldsLayout)
	}
	return noDataFormat
}

func formatGPUAllocationValue(workloadRow *WorkloadsAlignedRow) string {
	return fmt.Sprintf("%.2f", workloadRow.GpuAllocation)
}

func formatUrlsValue(workloadRow *WorkloadsAlignedRow) string {
	if workloadRow.Urls != "" {
		return workloadRow.Urls
	}
	return noDataFormat
}

func formatImageValue(workloadRow *WorkloadsAlignedRow) string {
	if workloadRow.Image != "" {
		return workloadRow.Image
	}
	return noDataFormat
}

func formatCurrentNodepoolValue(workloadRow *WorkloadsAlignedRow) string {
	if workloadRow.CurrentNodepool != "" {
		return workloadRow.CurrentNodepool
	}
	return noDataFormat
}

func formatRequestedNodepoolValue(workloadRow *WorkloadsAlignedRow) string {
	if workloadRow.RequestedNodepool != "" {
		return workloadRow.RequestedNodepool
	}
	return noDataFormat
}
