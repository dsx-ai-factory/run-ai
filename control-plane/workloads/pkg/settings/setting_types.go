/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/manager"
)

type EnablePolicies struct {
	Enabled bool `json:"enabled"`
}

func (e *EnablePolicies) Category() string { return WorkloadsCategory }
func (e *EnablePolicies) Description() string {
	return "Enable to impose restrictions and default values on workloads' fields. For more information, see the <a data-doc-key='policies' target='_blank'>Policies guide</a>"
}
func (e *EnablePolicies) DisplayName() string         { return "Policies" }
func (e *EnablePolicies) Key() string                 { return string(api.WORKLOADSSETTINGSKEYENUM_ENABLE_POLICIES) }
func (e *EnablePolicies) SetDefault()                 { e.Enabled = true }
func (e *EnablePolicies) Stage() setting_pkg.Stage    { return setting_pkg.GA }
func (e *EnablePolicies) Validate(_ interface{}) bool { return true }
func (e *EnablePolicies) TransformForMigration(v interface{}) interface{} {
	if b, ok := v.(bool); ok {
		return EnablePolicies{Enabled: b}
	}
	return v
}

type WorkloadSubmissionV2Enabled struct {
	Enabled bool `json:"enabled"`
}

func (e *WorkloadSubmissionV2Enabled) Category() string { return WorkloadsCategory }
func (e *WorkloadSubmissionV2Enabled) Description() string {
	return "Enable submission of workload with the new API"
}
func (e *WorkloadSubmissionV2Enabled) DisplayName() string { return "Workload submission v2" }
func (e *WorkloadSubmissionV2Enabled) Key() string {
	return string(api.WORKLOADSSETTINGSKEYENUM_WORKLOAD_SUBMISSION_V2_ENABLED)
}
func (e *WorkloadSubmissionV2Enabled) SetDefault()                 { e.Enabled = true }
func (e *WorkloadSubmissionV2Enabled) Stage() setting_pkg.Stage    { return setting_pkg.Alpha }
func (e *WorkloadSubmissionV2Enabled) Validate(_ interface{}) bool { return true }
func (e *WorkloadSubmissionV2Enabled) TransformForMigration(v interface{}) interface{} {
	if b, ok := v.(bool); ok {
		return WorkloadSubmissionV2Enabled{Enabled: b}
	}
	return v
}

type EnableTemplatesV2 struct {
	Enabled bool `json:"enabled"`
}

func (e *EnableTemplatesV2) Category() string { return WorkloadsCategory }
func (e *EnableTemplatesV2) Description() string {
	return "Enable to create workload templates via the new flexible form. For more information, see the <a data-doc-key='workloadTemplates' target='_blank'>Workload Templates guide</a>"
}
func (e *EnableTemplatesV2) DisplayName() string { return "Flexible workload templates" }
func (e *EnableTemplatesV2) Key() string {
	return string(api.WORKLOADSSETTINGSKEYENUM_ENABLE_TEMPLATES_V2)
}
func (e *EnableTemplatesV2) SetDefault()                 { e.Enabled = true }
func (e *EnableTemplatesV2) Stage() setting_pkg.Stage    { return setting_pkg.GA }
func (e *EnableTemplatesV2) Validate(_ interface{}) bool { return true }
func (e *EnableTemplatesV2) TransformForMigration(v interface{}) interface{} {
	if b, ok := v.(bool); ok {
		return EnableTemplatesV2{Enabled: b}
	}
	return v
}

type EnableHybridSubmission struct {
	Enabled bool `json:"enabled"`
}

func (e *EnableHybridSubmission) Category() string { return WorkloadsCategory }
func (e *EnableHybridSubmission) Description() string {
	return "Enable to submit workloads via the new flexible form. Allow selecting an existing setup or starting from scratch, reviewing existing setups and understanding policy definitions."
}
func (e *EnableHybridSubmission) DisplayName() string { return "Flexible workload submission" }
func (e *EnableHybridSubmission) Key() string {
	return string(api.WORKLOADSSETTINGSKEYENUM_ENABLE_HYBRID_SUBMISSION)
}
func (e *EnableHybridSubmission) SetDefault()                 { e.Enabled = true }
func (e *EnableHybridSubmission) Stage() setting_pkg.Stage    { return setting_pkg.GA }
func (e *EnableHybridSubmission) Validate(_ interface{}) bool { return true }
func (e *EnableHybridSubmission) TransformForMigration(v interface{}) interface{} {
	if b, ok := v.(bool); ok {
		return EnableHybridSubmission{Enabled: b}
	}
	return v
}

type GpuOverProvisioning struct {
	Enabled bool `json:"enabled"`
}

func (e *GpuOverProvisioning) Category() string { return WorkloadsCategory }
func (e *GpuOverProvisioning) Description() string {
	return "Enable to set a GPU memory limit for workloads. To enable this on each Run:ai cluster, see the <a data-doc-key='dynamicFractions' target='_blank'>Dynamic GPU fractions guide</a>"
}
func (e *GpuOverProvisioning) DisplayName() string { return "GPU resource optimization" }
func (e *GpuOverProvisioning) Key() string {
	return string(api.WORKLOADSSETTINGSKEYENUM_GPU_OVER_PROVISIONING)
}
func (e *GpuOverProvisioning) SetDefault()                 { e.Enabled = false }
func (e *GpuOverProvisioning) Stage() setting_pkg.Stage    { return setting_pkg.Beta }
func (e *GpuOverProvisioning) Validate(_ interface{}) bool { return true }
func (e *GpuOverProvisioning) TransformForMigration(v interface{}) interface{} {
	if b, ok := v.(bool); ok {
		return GpuOverProvisioning{Enabled: b}
	}
	return v
}

type ApplicationSubmission struct {
	Enabled bool `json:"enabled"`
}

func (e *ApplicationSubmission) Category() string { return WorkloadsCategory }
func (e *ApplicationSubmission) Description() string {
	return "Enable to deploy agentic AI application stacks from the NGC catalog or Helm repositories, with resource overrides for multi-component pipelines."
}
func (e *ApplicationSubmission) DisplayName() string { return "AI application submission" }
func (e *ApplicationSubmission) Key() string {
	return string(api.WORKLOADSSETTINGSKEYENUM_APPLICATION_SUBMISSION)
}
func (e *ApplicationSubmission) SetDefault()                 { e.Enabled = true }
func (e *ApplicationSubmission) Stage() setting_pkg.Stage    { return setting_pkg.Alpha }
func (e *ApplicationSubmission) Validate(_ interface{}) bool { return true }
func (e *ApplicationSubmission) TransformForMigration(v interface{}) interface{} {
	if b, ok := v.(bool); ok {
		return ApplicationSubmission{Enabled: b}
	}
	return v
}

// AnyWorkloadSubmissionEnabled is a per-tenant flag that enables any-workload submission
// for all workload types. When enabled, all workload types are submitted via the
// any-workload path regardless of per-type config flags.
type AnyWorkloadSubmissionEnabled struct {
	Enabled bool `json:"enabled"`
}

func (e *AnyWorkloadSubmissionEnabled) Category() string { return WorkloadsCategory }
func (e *AnyWorkloadSubmissionEnabled) Description() string {
	return "Enable any workload submission for all workload types."
}
func (e *AnyWorkloadSubmissionEnabled) DisplayName() string { return "Any Workload Submission" }
func (e *AnyWorkloadSubmissionEnabled) Key() string {
	return "anyWorkloadSubmissionEnabled"
}
func (e *AnyWorkloadSubmissionEnabled) IsEnabled() bool             { return e.Enabled }
func (e *AnyWorkloadSubmissionEnabled) SetDefault()                 { e.Enabled = true }
func (e *AnyWorkloadSubmissionEnabled) Stage() setting_pkg.Stage    { return setting_pkg.Alpha }
func (e *AnyWorkloadSubmissionEnabled) Validate(_ interface{}) bool { return true }
func (e *AnyWorkloadSubmissionEnabled) TransformForMigration(v interface{}) interface{} {
	if b, ok := v.(bool); ok {
		return AnyWorkloadSubmissionEnabled{Enabled: b}
	}
	return v
}
