/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
)

const (
	submitSourceNative          = "native"
	submitSourceYAML            = "yaml"
	submitWorkloadMinAPIVersion = "2.23"
)

type submitWorkloadInput struct {
	Source  string                    `json:"source"          jsonschema:"Submission API to use. One of: native, yaml."`
	Native  *nativeWorkloadSubmission `json:"native,omitempty" jsonschema:"Native Run:ai workload submission. Required only when source=native."`
	YAML    *yamlWorkloadSubmission   `json:"yaml,omitempty"   jsonschema:"YAML Workloads V2 submission. Required only when source=yaml."`
	DryRun  bool                      `json:"dryRun,omitempty" jsonschema:"Ask the Workloads V2 API to validate a YAML submission without creating it. Supported only when source=yaml."`
	Confirm bool                      `json:"confirm,omitempty" jsonschema:"Required for a real submission when the client cannot show an interactive confirmation prompt."`
}

func (in submitWorkloadInput) Validate() error {
	switch in.Source {
	case submitSourceNative:
		return validateNativeSubmissionInput(in)
	case submitSourceYAML:
		return validateYAMLSubmissionInput(in)
	default:
		return fmt.Errorf("source must be one of: native, yaml")
	}
}

type submitWorkloadOutput struct {
	DryRun       bool   `json:"dryRun"                 jsonschema:"True when a YAML submission was validated by the Workloads V2 API without being created."`
	ID           string `json:"id,omitempty"            jsonschema:"Created workload ID. Empty for any dry-run."`
	Name         string `json:"name,omitempty"          jsonschema:"Created or requested workload name."`
	WorkloadType string `json:"workloadType,omitempty"  jsonschema:"Native workload type or Kubernetes manifest kind."`
	ProjectID    string `json:"projectId,omitempty"      jsonschema:"Project that owns the workload, when known."`
	ClusterID    string `json:"clusterId,omitempty"      jsonschema:"Cluster where the workload is submitted, when known."`
	NameIsPrefix bool   `json:"-"`
}

func registerSubmitWorkload(r *runaimcp.Registry, client *runaiclient.Client) {
	yamlSupported := r.SupportsAPIVersion(submitWorkloadMinAPIVersion, "")
	runaimcp.Register(r, runaimcp.Tool[submitWorkloadInput, submitWorkloadOutput]{
		Name:        "submit_workload",
		Description: submitWorkloadDescription(yamlSupported),
		ReadOnly:    false,
		Destructive: false,
		Handler: func(ctx context.Context, in submitWorkloadInput) (submitWorkloadOutput, error) {
			return submitWorkloadRun(ctx, client, in, yamlSupported)
		},
	})
}

func submitWorkloadDescription(yamlSupported bool) string {
	if !yamlSupported {
		return "Call when the user asks to submit a new native Run:ai workload. This control-plane version supports source=native only; source=yaml requires control-plane version " +
			submitWorkloadMinAPIVersion + " or newer. Use native.workloadType and the complete corresponding creation request. A submission requires confirmation."
	}
	return "Call when the user asks to submit a new workload through one of two submission paths: a native Run:ai workload API or a Kubernetes YAML manifest. " +
		"Use source=native to call a native Run:ai workload API. Provide native.workloadType and the complete corresponding creation request for a Training, Workspace, Distributed, Inference, or DistributedInference workload. " + //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
		"Use source=yaml to submit a Kubernetes YAML manifest through Workloads V2. Provide Workloads V2 metadata and manifestYAML for any supported Kubernetes workload kind. Both paths preserve all fields accepted by their upstream APIs. " + //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
		"dryRun=true is supported only for YAML submissions and is validated by the Workloads V2 API. " +
		"A real submission requires confirmation. " +
		"Do NOT use this tool to list or inspect existing workloads; use list_workloads, get_workload_spec, or get_workload_status instead."
}

func submitWorkloadRun(ctx context.Context, client *runaiclient.Client, in submitWorkloadInput, yamlSupported bool) (submitWorkloadOutput, error) {
	if in.Source == submitSourceNative {
		return submitNativeWorkload(ctx, client, in)
	}
	if !yamlSupported {
		return submitWorkloadOutput{}, fmt.Errorf("source=yaml requires control-plane version %s or newer: %w",
			submitWorkloadMinAPIVersion, runaimcp.ErrValidation)
	}
	return submitYAMLWorkload(ctx, client, in)
}
