/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policies

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"

	policy "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

func (in manageWorkloadPolicyInput) validateGVKPolicyInput() error {
	// Reject tool fields that the GVK API call would otherwise ignore.
	if in.WorkloadType != "" || in.Scope != "" || in.ClusterID != "" || in.DepartmentID != "" || in.ProjectID != "" || in.ApproveClusterDeletion {
		return fmt.Errorf("top-level workloadType, scope, scope IDs, and approveClusterDeletion are legacy-only; GVK create takes workloadType and scope in request")
	}
	if in.Action == actionCreate {
		if in.PolicyID != "" {
			return fmt.Errorf("policyId must not be set for create")
		}
		return decodeGVKPolicyRequest(in.Request, &policy.CreateGvkPolicyReq{})
	}
	if err := common.ValidateUUID("policyId", in.PolicyID); err != nil {
		return err
	}
	if in.Action == actionDelete {
		return nil
	}
	return decodeGVKPolicyRequest(in.Request, &policy.UpdateGvkPolicyReq{})
}

func decodeGVKPolicyRequest(raw map[string]any, target any) error {
	data, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	request := maps.Clone(raw)
	rulesValue, hasRules := request["rules"]
	delete(request, "rules")
	if err := common.DecodeTypedRequest(request, target); err != nil {
		return err
	}
	if hasRules && rulesValue != nil {
		rules, ok := rulesValue.([]any)
		if !ok {
			return fmt.Errorf("request.rules must be an array")
		}
		for index, rule := range rules {
			if err := validateGVKPolicyRule(rule); err != nil {
				return fmt.Errorf("request.rules[%d]: %w", index, err)
			}
		}
	}
	return json.Unmarshal(data, target)
}

func validateGVKPolicyRule(value any) error {
	rule, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("rule must be an object")
	}
	rule = maps.Clone(rule)
	enforcements, ok := rule["enforcements"].([]any)
	if !ok {
		return fmt.Errorf("enforcements must be an array of restrictions or default values")
	}
	rule["enforcements"] = []any{}
	var fields policy.PolicyRule
	if err := common.DecodeTypedRequest(rule, &fields); err != nil {
		return err
	}
	for _, enforcement := range enforcements {
		if err := validateGVKPolicyEnforcement(enforcement); err != nil {
			return err
		}
	}
	return nil
}

func validateGVKPolicyEnforcement(value any) error {
	settings, ok := value.(map[string]any)
	if !ok || len(settings) != 1 {
		return fmt.Errorf("each enforcement must contain one setting, such as required or range")
	}
	var model any
	for kind := range settings {
		switch kind {
		case "blocked":
			model = &policy.PolicyEnforcementBlocked{}
		case "default":
			model = &policy.PolicyEnforcementDefault{}
		case "elements":
			model = &policy.PolicyEnforcementElements{}
		case "items":
			model = &policy.PolicyEnforcementItems{}
		case "length":
			model = &policy.PolicyEnforcementLength{}
		case "match":
			model = &policy.PolicyEnforcementMatch{}
		case "options":
			model = &policy.PolicyEnforcementOptions{}
		case "quantityRange":
			model = &policy.PolicyEnforcementQuantityRange{}
		case "range":
			model = &policy.PolicyEnforcementRange{}
		case "required":
			model = &policy.PolicyEnforcementRequired{}
		case "set":
			model = &policy.PolicyEnforcementSet{}
		default:
			return fmt.Errorf("unknown enforcement setting %q", kind)
		}
	}
	return common.DecodeTypedRequest(settings, model)
}

func getGVKPolicy(ctx context.Context, client *runaiclient.Client, policyID string) (map[string]any, error) {
	_, resp, err := client.Policy.PolicyAPI.GetGvkPolicy(ctx, policyID).Execute()
	return common.ReadAPIResponse(resp, err, false)
}

func executeGVKPolicyOperation(ctx context.Context, client *runaiclient.Client, in manageWorkloadPolicyInput) (map[string]any, error) {
	switch in.Action {
	case actionCreate:
		return createGVKPolicy(ctx, client, in.Request)
	case actionUpdate:
		return updateGVKPolicy(ctx, client, in)
	case actionDelete:
		resp, err := client.Policy.PolicyAPI.DeleteGvkPolicy(ctx, in.PolicyID).Execute()
		return common.ReadAPIResponse(resp, err, true)
	default:
		return nil, fmt.Errorf("unsupported operation %q: %w", in.Action, runaimcp.ErrValidation)
	}
}

func createGVKPolicy(ctx context.Context, client *runaiclient.Client, raw map[string]any) (map[string]any, error) {
	var request policy.CreateGvkPolicyReq
	if err := decodeGVKPolicyRequest(raw, &request); err != nil {
		return nil, fmt.Errorf("create GVK policy request: %w: %w", err, runaimcp.ErrValidation)
	}
	_, resp, err := client.Policy.PolicyAPI.CreateGvkPolicy(ctx).CreateGvkPolicyReq(request).Execute()
	return common.ReadAPIResponse(resp, err, false)
}

func updateGVKPolicy(ctx context.Context, client *runaiclient.Client, in manageWorkloadPolicyInput) (map[string]any, error) {
	var request policy.UpdateGvkPolicyReq
	if err := decodeGVKPolicyRequest(in.Request, &request); err != nil {
		return nil, fmt.Errorf("update GVK policy request: %w: %w", err, runaimcp.ErrValidation)
	}
	api := client.Policy.PolicyAPI
	if in.ReplaceRules {
		_, resp, err := api.ApplyGvkPolicy(ctx, in.PolicyID).UpdateGvkPolicyReq(request).Execute()
		return common.ReadAPIResponse(resp, err, false)
	}
	_, resp, err := api.PatchGvkPolicy(ctx, in.PolicyID).UpdateGvkPolicyReq(request).Execute()
	return common.ReadAPIResponse(resp, err, false)
}
