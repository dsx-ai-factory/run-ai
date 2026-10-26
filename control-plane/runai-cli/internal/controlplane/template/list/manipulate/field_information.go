/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	"fmt"

	wcgenerated "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

const (
	DefaultWidthName                   = 25
	DefaultWidthScope                  = 15
	DefaultWidthImage                  = 35
	DefaultWidthGpuComputeRequest      = 25
	DefaultWidthWorkloadType           = 18
	DefaultWidthMandatoryFieldsMissing = 28
	DefaultWidthCreatedAt              = 30
)

// Field definitions for template columns
var (
	NameField = FieldInformation{
		columnTitle: "Template",
		columnWidth: DefaultWidthName,
		extractFunc: extractNameData,
		formatFunc:  formatNameValue,
	}
	ScopeField = FieldInformation{
		columnTitle: "Scope",
		columnWidth: DefaultWidthScope,
		extractFunc: extractScopeData,
		formatFunc:  formatScopeValue,
	}
	ImageField = FieldInformation{
		columnTitle: "Image",
		columnWidth: DefaultWidthImage,
		extractFunc: extractImageData,
		formatFunc:  formatImageValue,
	}
	GpuComputeRequestField = FieldInformation{
		columnTitle: "GPU Compute Request",
		columnWidth: DefaultWidthGpuComputeRequest,
		extractFunc: extractGpuComputeRequestData,
		formatFunc:  formatGpuComputeRequestValue,
	}
	WorkloadTypeField = FieldInformation{
		columnTitle: "Workload Type",
		columnWidth: DefaultWidthWorkloadType,
		extractFunc: extractWorkloadTypeData,
		formatFunc:  formatWorkloadTypeValue,
	}
	MandatoryFieldsMissingField = FieldInformation{
		columnTitle: "Mandatory fields missing",
		columnWidth: DefaultWidthMandatoryFieldsMissing,
		extractFunc: extractMandatoryFieldsMissingData,
		formatFunc:  formatMandatoryFieldsMissingValue,
	}
	CreatedAtField = FieldInformation{
		columnTitle: "Created At",
		columnWidth: DefaultWidthCreatedAt,
		extractFunc: extractCreatedAtData,
		formatFunc:  formatCreatedAtValue,
	}
)

// Extract functions for template fields
func extractNameData(template *wcgenerated.SupersetTemplate, templateRow *TemplateAlignedRow) {
	templateRow.Name = template.Meta.GetName()
}

func extractScopeData(template *wcgenerated.SupersetTemplate, templateRow *TemplateAlignedRow) {
	templateRow.ScopeType = string(template.Meta.GetScopeType())
}

func extractImageData(template *wcgenerated.SupersetTemplate, templateRow *TemplateAlignedRow) {
	if template.Spec != nil && template.Spec.HasImage() {
		templateRow.Image = template.Spec.GetImage()
	} else {
		templateRow.Image = ""
	}
}

func extractGpuComputeRequestData(template *wcgenerated.SupersetTemplate, templateRow *TemplateAlignedRow) {
	if template.Spec != nil && template.Spec.HasCompute() {
		compute := template.Spec.GetCompute()
		if gpuRequest, ok := compute.GetGpuDevicesRequestOk(); ok && gpuRequest != nil {
			templateRow.GpuComputeRequest = fmt.Sprintf("%d", *gpuRequest)
		} else {
			templateRow.GpuComputeRequest = ""
		}
	} else {
		templateRow.GpuComputeRequest = ""
	}
}

func extractWorkloadTypeData(template *wcgenerated.SupersetTemplate, templateRow *TemplateAlignedRow) {
	templateRow.WorkloadType = string(template.Meta.GetWorkloadType())
}

func extractMandatoryFieldsMissingData(template *wcgenerated.SupersetTemplate, templateRow *TemplateAlignedRow) {
	if template.HasRunnable() {
		runnable := template.GetRunnable()
		if isRunnable, ok := runnable.GetIsRunnableOk(); ok && isRunnable != nil {
			// Store the inverse: isRunnable=true means mandatory fields are NOT missing (false)
			mandatoryFieldsMissing := !(*isRunnable)
			templateRow.MandatoryFieldsMissing = &mandatoryFieldsMissing
		} else if runnable.HasIssue() {
			// If issue is set (but isRunnable is not), treat as not runnable
			mandatoryFieldsMissing := true
			templateRow.MandatoryFieldsMissing = &mandatoryFieldsMissing
		}
	}
}

func extractCreatedAtData(template *wcgenerated.SupersetTemplate, templateRow *TemplateAlignedRow) {
	templateRow.CreatedAt = template.Meta.CreatedAt.String()
}

// Format functions for template fields
func formatNameValue(templateRow *TemplateAlignedRow) string {
	return templateRow.Name
}

func formatScopeValue(templateRow *TemplateAlignedRow) string {
	return templateRow.ScopeType
}

func formatImageValue(templateRow *TemplateAlignedRow) string {
	if templateRow.Image == "" {
		return "-"
	}
	return templateRow.Image
}

func formatGpuComputeRequestValue(templateRow *TemplateAlignedRow) string {
	if templateRow.GpuComputeRequest == "" {
		return "0"
	}
	return templateRow.GpuComputeRequest
}

func formatWorkloadTypeValue(templateRow *TemplateAlignedRow) string {
	return templateRow.WorkloadType
}

func formatMandatoryFieldsMissingValue(templateRow *TemplateAlignedRow) string {
	if templateRow.MandatoryFieldsMissing == nil {
		return "-"
	}
	if *templateRow.MandatoryFieldsMissing {
		return "Yes"
	}
	return "No"
}

func formatCreatedAtValue(templateRow *TemplateAlignedRow) string {
	return templateRow.CreatedAt
}
