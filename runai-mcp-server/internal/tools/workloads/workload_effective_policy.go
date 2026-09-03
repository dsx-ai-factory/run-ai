/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

const (
	policyKindTraining    = "training"
	policyKindWorkspace   = "workspace"
	policyKindDistributed = "distributed"
	policyKindInference   = "inference"
)

type getWorkloadEffectivePolicyInput struct {
	ProjectID string `json:"projectId" jsonschema:"Run:ai project ID as a numeric string (e.g. \"4500002\")"`
	Kind      string `json:"kind"      jsonschema:"Singular workload kind: one of training | workspace | distributed | inference. Case-insensitive — same vocabulary as the type field returned by get_workload_status, lowercased"` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
}

func (in getWorkloadEffectivePolicyInput) Validate() error {
	if err := common.ValidateNumericString("projectId", in.ProjectID); err != nil {
		return err
	}
	switch strings.ToLower(strings.TrimSpace(in.Kind)) {
	case policyKindTraining, policyKindWorkspace, policyKindDistributed, policyKindInference:
		return nil
	default:
		return fmt.Errorf("kind must be one of training | workspace | distributed | inference (got %q)", in.Kind)
	}
}

type getWorkloadEffectivePolicyOutput struct {
	ProjectID string         `json:"projectId" jsonschema:"Project ID this policy applies to"`
	Kind      string         `json:"kind"      jsonschema:"Workload kind this policy applies to, lowercased (training | workspace | distributed | inference)"`
	Effective map[string]any `json:"effective" jsonschema:"Effective policy tree at the project level — the merged result of all policy layers — forwarded verbatim. Modern responses split it into {defaults, rules} (defaults.<key> for pre-merged values; rules.<key>.{required | forbidden | min | max | enum | regex, options, sourceOfRule, message} at leaves). Legacy responses come through as a flat tree with default + verb keys interleaved. The MCP server does not normalise either shape — walk the tree as-is. Empty object when there is no effective policy at this scope"` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
}

func registerGetWorkloadEffectivePolicy(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[getWorkloadEffectivePolicyInput, getWorkloadEffectivePolicyOutput]{
		Name: "get_workload_effective_policy",
		Description: "Call when the user asks what policy rules or defaults apply to a project's workloads of a given kind. " +
			"Do NOT call to validate a specific workload spec — that is a skill (audit_workload_against_policy). " +
			"Do NOT call to check if a workload is currently running — use get_workload_status.",
		ReadOnly:      true,
		MinAPIVersion: "2.18",
		Handler: func(ctx context.Context, in getWorkloadEffectivePolicyInput) (getWorkloadEffectivePolicyOutput, error) {
			return getWorkloadEffectivePolicyRun(ctx, client, in)
		},
	})
}

func getWorkloadEffectivePolicyRun(ctx context.Context, client *runaiclient.Client, in getWorkloadEffectivePolicyInput) (getWorkloadEffectivePolicyOutput, error) {
	kind := strings.ToLower(strings.TrimSpace(in.Kind))
	httpResp, err := dispatchPolicyByKind(ctx, client, kind, in.ProjectID)
	if err != nil {
		return getWorkloadEffectivePolicyOutput{}, fmt.Errorf("get %s policy for project %q: %w", kind, in.ProjectID, err)
	}
	effective, err := extractEffectivePolicy(httpResp)
	if err != nil {
		return getWorkloadEffectivePolicyOutput{}, fmt.Errorf("decode %s policy for project %q: %w", kind, in.ProjectID, err)
	}
	return getWorkloadEffectivePolicyOutput{
		ProjectID: in.ProjectID,
		Kind:      kind,
		Effective: effective,
	}, nil
}

// dispatchPolicyByKind returns the raw response of the kind-specific policy
func dispatchPolicyByKind(ctx context.Context, client *runaiclient.Client, kind, projectID string) (*http.Response, error) {
	const scope = "project"
	api := client.Policy.PolicyAPI
	var (
		resp *http.Response
		err  error
	)
	switch kind {
	case policyKindTraining:
		_, resp, err = api.GetTrainingPolicyV2(ctx).Scope(scope).ProjectId(projectID).Execute()
	case policyKindWorkspace:
		_, resp, err = api.GetWorkspacePolicyV2(ctx).Scope(scope).ProjectId(projectID).Execute()
	case policyKindDistributed:
		_, resp, err = api.GetDistributedPolicyV2(ctx).Scope(scope).ProjectId(projectID).Execute()
	case policyKindInference:
		_, resp, err = api.GetInferencePolicyV2(ctx).Scope(scope).ProjectId(projectID).Execute()
	default:
		return nil, fmt.Errorf("unsupported policy kind %q: %w", kind, runaimcp.ErrValidation)
	}
	if resp != nil && resp.StatusCode < 300 {
		return resp, nil
	}
	return resp, runaimcp.MapHTTPError(resp, err)
}

func extractEffectivePolicy(httpResp *http.Response) (map[string]any, error) {
	var raw struct {
		Effective map[string]any `json:"effective"`
	}
	if httpResp != nil && httpResp.Body != nil {
		if err := json.NewDecoder(httpResp.Body).Decode(&raw); err != nil && !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("decode policy: %w", err)
		}
	}
	if raw.Effective == nil {
		raw.Effective = map[string]any{}
	}
	return raw.Effective, nil
}
