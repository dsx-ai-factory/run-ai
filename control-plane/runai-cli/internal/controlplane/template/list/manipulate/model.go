/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	wcgenerated "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

// Data represents the raw template list response
type Data struct {
	Templates []wcgenerated.SupersetTemplate `json:"templates,omitempty" yaml:"templates"`
	NextToken *string                        `json:"nextToken,omitempty" yaml:"nextToken,omitempty"`
}

// BriefTemplate represents a simplified template for display
type BriefTemplate struct {
	Name                   string `json:"name,omitempty"`
	ScopeType              string `json:"scopeType,omitempty"`
	Image                  string `json:"image,omitempty"`
	GpuDevicesRequest      string `json:"gpuDevicesRequest,omitempty"`
	WorkloadType           string `json:"workloadType,omitempty"`
	MandatoryFieldsMissing *bool  `json:"mandatoryFieldsMissing,omitempty"`
	CreatedAt              string `json:"createdAt,omitempty"`
}

// BriefData represents simplified templates for JSON/YAML output
type BriefData struct {
	Templates []BriefTemplate `json:"templates,omitempty" yaml:"templates"`
	NextToken *string         `json:"nextToken,omitempty" yaml:"nextToken,omitempty"`
}

// TemplateAlignedRow represents a data row with proper fields for aligned output
type TemplateAlignedRow struct {
	Name                   string `json:"name,omitempty"`
	ScopeType              string `json:"scopeType,omitempty"`
	Image                  string `json:"image,omitempty"`
	GpuComputeRequest      string `json:"gpuComputeRequest,omitempty"`
	WorkloadType           string `json:"workloadType,omitempty"`
	MandatoryFieldsMissing *bool  `json:"mandatoryFieldsMissing,omitempty"`
	CreatedAt              string `json:"createdAt,omitempty"`
}

// TemplatesAlignedOutput represents templates output with column titles as keys
type TemplatesAlignedOutput struct {
	Templates []TemplateAlignedRow `json:"templates" yaml:"templates"`
	NextToken *string              `json:"nextToken,omitempty" yaml:"nextToken,omitempty"`
}

// newBriefFromTemplate converts a SupersetTemplate to BriefTemplate
func newBriefFromTemplate(tmpl wcgenerated.SupersetTemplate) BriefTemplate {
	brief := BriefTemplate{
		Name:         tmpl.Meta.GetName(),
		ScopeType:    string(tmpl.Meta.GetScopeType()),
		WorkloadType: string(tmpl.Meta.GetWorkloadType()),
		CreatedAt:    tmpl.Meta.CreatedAt.String(),
	}

	if tmpl.Spec != nil {
		if tmpl.Spec.HasImage() {
			brief.Image = tmpl.Spec.GetImage()
		}

		if tmpl.Spec.HasCompute() {
			compute := tmpl.Spec.GetCompute()
			if gpuRequest, ok := compute.GetGpuDevicesRequestOk(); ok && gpuRequest != nil {
				brief.GpuDevicesRequest = formatInt32(*gpuRequest)
			}
		}
	}

	// Extract is-runnable from template (store inverse as MandatoryFieldsMissing)
	if tmpl.HasRunnable() {
		runnable := tmpl.GetRunnable()
		if isRunnable, ok := runnable.GetIsRunnableOk(); ok && isRunnable != nil {
			// Store the inverse: isRunnable=true means mandatory fields are NOT missing (false)
			mandatoryFieldsMissing := !(*isRunnable)
			brief.MandatoryFieldsMissing = &mandatoryFieldsMissing
		} else if runnable.HasIssue() {
			// If issue is set (but isRunnable is not), treat as not runnable
			mandatoryFieldsMissing := true
			brief.MandatoryFieldsMissing = &mandatoryFieldsMissing
		}
	}

	return brief
}

// BriefListFromResponse converts a list of templates to brief format
func BriefListFromResponse(templates []wcgenerated.SupersetTemplate) []BriefTemplate {
	briefList := make([]BriefTemplate, 0, len(templates))
	for _, tmpl := range templates {
		briefList = append(briefList, newBriefFromTemplate(tmpl))
	}
	return briefList
}
