/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	policy "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

const (
	policyKindTraining             = string(policy.POLICYTYPE_TRAINING)
	policyKindWorkspace            = string(policy.POLICYTYPE_WORKSPACE)
	policyKindDistributed          = string(policy.POLICYTYPE_DISTRIBUTED)
	policyKindInference            = string(policy.POLICYTYPE_INFERENCE)
	policyKindDistributedInference = string(policy.POLICYTYPE_DISTRIBUTED_INFERENCE)
	policyKindNimService           = string(policy.POLICYTYPE_NIM_SERVICE)
	policyKindFallback             = string(policy.POLICYTYPE_FALLBACK)
)

type getWorkloadEffectivePolicyInput struct {
	Kind         string `json:"kind" jsonschema:"Legacy workload type: training, workspace, distributed, inference, distributedInference, nimService, or fallback. Case-insensitive."`
	Scope        string `json:"scope,omitempty" jsonschema:"Defaults to project. Exact scope: tenant, cluster, department, or project."`
	ClusterID    string `json:"clusterId,omitempty" jsonschema:"Cluster UUID required for scope=cluster. Resolve names through list_clusters."`
	DepartmentID string `json:"departmentId,omitempty" jsonschema:"Department ID required for scope=department. Resolve names through list_departments."`
	ProjectID    string `json:"projectId,omitempty" jsonschema:"Numeric project ID as a string, required for scope=project. Resolve names through list_projects."`
}

func (in getWorkloadEffectivePolicyInput) targetScope() string {
	if in.Scope == "" {
		return string(policy.SCOPETYPE_PROJECT)
	}
	return in.Scope
}

func (in getWorkloadEffectivePolicyInput) Validate() error {
	if err := common.ValidatePolicyScope(in.targetScope(), in.ClusterID, in.DepartmentID, in.ProjectID); err != nil {
		return err
	}
	_, err := policy.NewPolicyTypeFromValue(normalizePolicyKind(in.Kind))
	return err
}

type getWorkloadEffectivePolicyOutput struct {
	ProjectID string         `json:"projectId,omitempty" jsonschema:"Target project ID, when scope is project."`
	Scope     string         `json:"scope" jsonschema:"Scope at which the policy was read."`
	Kind      string         `json:"kind" jsonschema:"Legacy workload type in the spelling used by manage_workload_policy.workloadType."`
	Policy    map[string]any `json:"policy" jsonschema:"Settings declared at the target scope, including defaults, rules, and asset assignments."`
	Effective map[string]any `json:"effective" jsonschema:"Merged policy at the target scope, including inherited settings. This is not an update request."`
}

func registerGetWorkloadEffectivePolicy(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[getWorkloadEffectivePolicyInput, getWorkloadEffectivePolicyOutput]{
		Name: "get_workload_effective_policy",
		Description: "Read policy rules and defaults that apply at a scope, including inheritance. " +
			"Select a legacy workload kind; scope defaults to project and also supports tenant, cluster, and department with its matching ID. " +
			"The policy result contains settings declared at that scope; effective includes inherited settings. Read policy.imposedAssets before an edit when assignments must be preserved. " +
			"Use list_workload_policies for policy discovery, IDs, names, and GVK policies. Resolve scope names through list_clusters, list_departments, or list_projects; ask if ambiguous. " +
			"This tool does not validate workload specs or change policy settings.",
		ReadOnly:      true,
		MinAPIVersion: "2.18",
		Handler: func(ctx context.Context, in getWorkloadEffectivePolicyInput) (getWorkloadEffectivePolicyOutput, error) {
			return getWorkloadEffectivePolicyRun(ctx, client, in)
		},
	})
}

func getWorkloadEffectivePolicyRun(ctx context.Context, client *runaiclient.Client, in getWorkloadEffectivePolicyInput) (getWorkloadEffectivePolicyOutput, error) {
	out := getWorkloadEffectivePolicyOutput{
		ProjectID: in.ProjectID, Scope: in.targetScope(), Kind: normalizePolicyKind(in.Kind),
		Policy: map[string]any{}, Effective: map[string]any{},
	}
	resp, err := getLegacyPolicyResponse(ctx, client, out.Kind, in)
	if err != nil {
		return out, fmt.Errorf("get %s policy: %w", out.Kind, err)
	}
	if err := common.ReadJSONResponse(resp, &out); err != nil && !errors.Is(err, io.EOF) {
		return out, fmt.Errorf("read %s policy: %w: %w", out.Kind, err, runaimcp.ErrUpstream)
	}
	if out.Policy == nil {
		out.Policy = map[string]any{}
	}
	if out.Effective == nil {
		out.Effective = map[string]any{}
	}
	return out, nil
}

func normalizePolicyKind(kind string) string {
	kind = strings.TrimSpace(kind)
	for _, candidate := range policy.AllowedPolicyTypeEnumValues {
		if strings.EqualFold(kind, string(candidate)) {
			return string(candidate)
		}
	}
	return kind
}

func getLegacyPolicyResponse(ctx context.Context, client *runaiclient.Client, kind string, in getWorkloadEffectivePolicyInput) (*http.Response, error) {
	api := client.Policy.PolicyAPI
	var resp *http.Response
	var err error
	switch kind {
	case policyKindTraining:
		_, resp, err = api.GetTrainingPolicyV2(ctx).Scope(in.targetScope()).ClusterId(in.ClusterID).DepartmentId(in.DepartmentID).ProjectId(in.ProjectID).Execute()
	case policyKindWorkspace:
		_, resp, err = api.GetWorkspacePolicyV2(ctx).Scope(in.targetScope()).ClusterId(in.ClusterID).DepartmentId(in.DepartmentID).ProjectId(in.ProjectID).Execute()
	case policyKindDistributed:
		_, resp, err = api.GetDistributedPolicyV2(ctx).Scope(in.targetScope()).ClusterId(in.ClusterID).DepartmentId(in.DepartmentID).ProjectId(in.ProjectID).Execute()
	case policyKindInference:
		_, resp, err = api.GetInferencePolicyV2(ctx).Scope(in.targetScope()).ClusterId(in.ClusterID).DepartmentId(in.DepartmentID).ProjectId(in.ProjectID).Execute()
	case policyKindDistributedInference:
		_, resp, err = api.GetDistributedInferencePolicyV2(ctx).Scope(in.targetScope()).ClusterId(in.ClusterID).DepartmentId(in.DepartmentID).ProjectId(in.ProjectID).Execute()
	case policyKindNimService:
		_, resp, err = api.GetNimServicePolicy(ctx).Scope(in.targetScope()).ClusterId(in.ClusterID).DepartmentId(in.DepartmentID).ProjectId(in.ProjectID).Execute()
	case policyKindFallback:
		_, resp, err = api.GetFallbackPolicy(ctx).Scope(in.targetScope()).ClusterId(in.ClusterID).DepartmentId(in.DepartmentID).ProjectId(in.ProjectID).Execute()
	default:
		return nil, fmt.Errorf("unsupported policy kind %q: %w", kind, runaimcp.ErrValidation)
	}
	// Older response shapes may not decode into the generated model; read the raw body on success.
	if resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp, nil
	}
	return resp, runaimcp.MapHTTPError(resp, err)
}
