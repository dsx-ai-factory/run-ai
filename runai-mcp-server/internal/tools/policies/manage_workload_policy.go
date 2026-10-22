/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policies

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

type manageWorkloadPolicyInput struct {
	Action                 string         `json:"action" jsonschema:"Required: create, update, or delete. Legacy create may overwrite an existing policy; GVK create rejects an existing target."`
	PolicyVersion          string         `json:"policyVersion" jsonschema:"Required API family: legacy for a native workload type, or gvk for a Kubernetes GVK policy."`
	PolicyID               string         `json:"policyId,omitempty" jsonschema:"GVK update/delete require this policy UUID from list_workload_policies.items[].id. Omit for legacy/create."`
	WorkloadType           string         `json:"workloadType,omitempty" jsonschema:"Legacy requires: training, workspace, distributed, inference, distributedInference, nimService, or fallback."`
	Scope                  string         `json:"scope,omitempty" jsonschema:"Legacy delete only. Exact scope: tenant, cluster, department, or project. Create/update take the scope from request.meta."`
	ClusterID              string         `json:"clusterId,omitempty" jsonschema:"Legacy delete only. Required cluster UUID when scope is cluster; otherwise omit."`
	DepartmentID           string         `json:"departmentId,omitempty" jsonschema:"Legacy delete: department ID for scope=department. Resolve names through list_departments."`
	ProjectID              string         `json:"projectId,omitempty" jsonschema:"Legacy delete: numeric project ID as a string for scope=project. Resolve names through list_projects."`
	Request                map[string]any `json:"request,omitempty" jsonschema:"API body required for create/update; omit for delete. GVK update needs policyId AND request. See tool description for body shapes."`
	ReplaceRules           bool           `json:"replaceRules,omitempty" jsonschema:"GVK update: false/default uses PATCH; true uses PUT with the full nonempty rule list. PATCH replaces matching rules and retains others."`
	ApproveClusterDeletion bool           `json:"approveClusterDeletion,omitempty" jsonschema:"Legacy except nimService: approve removal of copies on Run:ai 2.20 or earlier clusters; separate from confirm."`
	Confirm                bool           `json:"confirm,omitempty" jsonschema:"Update/delete: accept the prompt, or set true after approval when prompting is unavailable. Unused for create."`
}

func (in manageWorkloadPolicyInput) Validate() error {
	if !common.WriteAction(in.Action).Valid() {
		return fmt.Errorf("action must be create, update, or delete")
	}
	if err := validatePolicyVersion(in.PolicyVersion); err != nil {
		return err
	}
	if in.ReplaceRules && (in.PolicyVersion != versionGVK || in.Action != actionUpdate) {
		return fmt.Errorf("replaceRules is only supported when updating a GVK policy (policyVersion=gvk)")
	}
	if in.Action == actionDelete && in.Request != nil {
		return fmt.Errorf("request must not be set for delete")
	}
	if in.Action != actionDelete && in.Request == nil {
		return fmt.Errorf("request is required")
	}
	if in.PolicyVersion == versionLegacy {
		return in.validateLegacyPolicyInput()
	}
	return in.validateGVKPolicyInput()
}

func manageWorkloadPolicyDescription() string {
	const (
		common = "Create, update, or delete a workload policy at tenant, cluster, department, or project scope. " +
			policyTypeDescription + "\n\n"
		legacy = "LEGACY: select workloadType. " +
			"For create/update, request.meta contains name, scope, and its matching scope ID; request.policy holds defaults, rules, and supported asset assignments. " +
			"Create may overwrite an existing policy. Update uses PATCH; request.reset clears selected field paths. " +
			"Older servers require request.policy ({} if otherwise empty) for distributed reset-only calls. " +
			"Read settings with get_workload_effective_policy and preserve policy.imposedAssets when retaining assignments: older servers may clear omitted assets. " +
			"Delete uses top-level scope and scope IDs. " +
			"request.meta.projectId is a JSON number.\n\n"
		gvk = "GVK: create request contains workloadType (group, version, kind), scope, its matching scope ID, nonempty rules, and optional description. " +
			"For core resources such as v1/Pod, omit group or use an empty string. request.projectId is a JSON string. " +
			"Update requires policyId and request with description and/or rules. " +
			"PATCH replaces matching rules in full, identified by field, specSelector, itemSelector, and selectorParams; other rules remain. " +
			"replaceRules=true replaces the entire list. Remove rules by sending the remaining nonempty list, or delete the policy to remove all. " +
			"Update/delete use policyId and the stored scope; omit top-level workloadType and scope IDs. " +
			"Use get_workload_policy_mapping to discover rule fields, selectors, and enforcement types.\n\n"
		guidance = "Find policies with list_workload_policies; check both families when unknown. Names are not unique: resolve the family and exact target before update/delete, and ask if ambiguous. " +
			policyScopeResolutionDescription +
			"This tool has no dry-run. Use API request fields only; unknown fields are rejected. Update/delete require confirmation; create does not, including legacy overwrite. " +
			"Returns the API policy object for create/update, or {} for delete; failures return an error."
	)
	return common + legacy + gvk + guidance
}

func registerManageWorkloadPolicy(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[manageWorkloadPolicyInput, map[string]any]{
		Name:        "manage_workload_policy",
		Description: manageWorkloadPolicyDescription(),
		ReadOnly:    false,
		Destructive: true,
		Handler: func(ctx context.Context, in manageWorkloadPolicyInput) (map[string]any, error) {
			return manageWorkloadPolicyRun(ctx, client, in)
		},
	})
}

func manageWorkloadPolicyRun(ctx context.Context, client *runaiclient.Client, in manageWorkloadPolicyInput) (map[string]any, error) {
	if err := confirmPolicyOperation(ctx, in); err != nil {
		return nil, err
	}
	return executePolicyOperation(ctx, client, in)
}

func confirmPolicyOperation(ctx context.Context, in manageWorkloadPolicyInput) error {
	if in.Action == actionCreate {
		return nil
	}

	details := map[string]any{"policyVersion": in.PolicyVersion}
	if len(in.Request) > 0 {
		details["request"] = in.Request
	}
	if in.ReplaceRules {
		details["replaceRules"] = true
	}
	if in.ApproveClusterDeletion {
		details["approveClusterDeletion"] = true
	}
	target := in.WorkloadType
	if in.PolicyVersion == versionGVK {
		target = "GVK"
	}
	if in.PolicyVersion == versionLegacy {
		// Show the target separately because the request preview is truncated.
		meta, _ := in.Request["meta"].(map[string]any)
		if in.Action == actionDelete {
			meta = map[string]any{
				"scope": in.Scope, fieldClusterID: in.ClusterID,
				fieldDepartmentID: in.DepartmentID, fieldProjectID: in.ProjectID,
			}
		}
		for _, field := range []string{"scope", fieldClusterID, fieldDepartmentID, fieldProjectID} {
			if value, ok := meta[field]; ok {
				details[field] = value
			}
		}
	}
	return common.ConfirmWrite(ctx, in.Confirm, in.Action, "workload policy", target, in.PolicyID, details)
}
