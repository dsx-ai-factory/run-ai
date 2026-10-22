/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policies

import (
	"context"
	"fmt"
	"strings"

	policy "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

type policyMappingFieldRef struct {
	FieldName    string `json:"fieldName"              jsonschema:"Required field name from the preceding minimal mapping response."`
	SpecSelector string `json:"specSelector,omitempty" jsonschema:"Spec selector name that exposes the field. Omit to return the field under every supported spec selector."`
	ItemSelector string `json:"itemSelector,omitempty" jsonschema:"Item selector name for an itemized field. Omit for fields without item selectors."`
}

type getWorkloadPolicyMappingInput struct {
	WorkloadType common.GVK              `json:"workloadType"       jsonschema:"Kubernetes Group, Version, and Kind whose policy mapping to inspect."`
	FieldsRef    []policyMappingFieldRef `json:"fieldsRef,omitempty" jsonschema:"Field references to inspect in detail. Omit for the initial lookup of available fields and selectors without JQ expressions."`
}

func (in getWorkloadPolicyMappingInput) Validate() error {
	if err := in.WorkloadType.Validate(); err != nil {
		return err
	}
	for i, ref := range in.FieldsRef {
		if strings.TrimSpace(ref.FieldName) == "" {
			return fmt.Errorf("fieldsRef[%d].fieldName is required", i)
		}
	}
	return nil
}

type getWorkloadPolicyMappingOutput struct {
	Mapping map[string]any `json:"mapping" jsonschema:"Policy API response. The initial lookup returns field and selector names with descriptions; fieldsRef returns expressions, types, selector details, and possible enforcements."`
}

func registerGetWorkloadPolicyMapping(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[getWorkloadPolicyMappingInput, getWorkloadPolicyMappingOutput]{
		Name: "get_workload_policy_mapping",
		Description: "Use when creating or updating GVK policies and the supported fields, selectors, or enforcement settings are unknown. " +
			"Call without fieldsRef to discover field and selector names, then provide fieldsRef for expressions, types, and supported enforcement settings. " +
			"Use the results to build manage_workload_policy.request.rules. This lookup is optional, applies only to GVK policies, and does not validate a policy.",
		ReadOnly: true,
		Handler: func(ctx context.Context, in getWorkloadPolicyMappingInput) (getWorkloadPolicyMappingOutput, error) {
			return getWorkloadPolicyMappingRun(ctx, client, in)
		},
	})
}

func getWorkloadPolicyMappingRun(ctx context.Context, client *runaiclient.Client, in getWorkloadPolicyMappingInput) (getWorkloadPolicyMappingOutput, error) {
	var mapping map[string]any
	var err error
	// Without field references, discover available fields and selectors.
	// With field references, get details for those fields, including supported enforcement settings.
	if len(in.FieldsRef) == 0 {
		mapping, err = getMinimalPolicyMapping(ctx, client, in)
	} else {
		mapping, err = getPolicyFieldMapping(ctx, client, in)
	}
	return getWorkloadPolicyMappingOutput{Mapping: mapping}, err
}

func getMinimalPolicyMapping(ctx context.Context, client *runaiclient.Client, in getWorkloadPolicyMappingInput) (map[string]any, error) {
	var request policy.GetMappingMinimalInfoReq
	if err := common.DecodeTypedRequest(in, &request); err != nil {
		return nil, fmt.Errorf("mapping request: %w: %w", err, runaimcp.ErrValidation)
	}
	_, resp, err := client.Policy.PolicyAPI.GetMappingMinimal(ctx).GetMappingMinimalInfoReq(request).Execute()
	return common.ReadAPIResponse(resp, err, false)
}

func getPolicyFieldMapping(ctx context.Context, client *runaiclient.Client, in getWorkloadPolicyMappingInput) (map[string]any, error) {
	var request policy.GetMappingFieldInfoReq
	if err := common.DecodeTypedRequest(in, &request); err != nil {
		return nil, fmt.Errorf("mapping request: %w: %w", err, runaimcp.ErrValidation)
	}
	_, resp, err := client.Policy.PolicyAPI.GetMappingFieldInfo(ctx).GetMappingFieldInfoReq(request).Execute()
	return common.ReadAPIResponse(resp, err, false)
}
