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

type listWorkloadPoliciesInput struct {
	PolicyVersion string   `json:"policyVersion" jsonschema:"API family: legacy for native workload policies, or gvk for Kubernetes Group, Version, Kind policies. Call both to list all visible policies."`
	WorkloadType  string   `json:"workloadType,omitempty" jsonschema:"Legacy: training, workspace, distributed, inference, distributedInference, nimService, or fallback."`
	Scope         string   `json:"scope,omitempty" jsonschema:"Legacy only. Exact scope: tenant, cluster, department, or project. Omit to include every scope."`
	ClusterID     string   `json:"clusterId,omitempty" jsonschema:"Legacy only. Filter by cluster UUID. Resolve a name through list_clusters."`
	DepartmentID  string   `json:"departmentId,omitempty" jsonschema:"Legacy only. Filter by department ID. Resolve a name through list_departments."`
	ProjectID     string   `json:"projectId,omitempty" jsonschema:"Legacy only. Filter by numeric project ID encoded as a string. Resolve a name through list_projects."`
	PolicyID      string   `json:"policyId,omitempty" jsonschema:"GVK only: policy UUID to retrieve one policy. Omit filterBy, sortBy, sortOrder, limit, and offset."`
	FilterBy      []string `json:"filterBy,omitempty" jsonschema:"GVK only: up to 10 ANDed expressions on id, type, or scope; || joins alternatives. Type uses group/kind/version. Omit with policyId."`
	SortBy        string   `json:"sortBy,omitempty" jsonschema:"GVK list only. Sort by id, type, scope, createdAt, createdBy, updatedAt, or updatedBy. Omit with policyId."`
	SortOrder     string   `json:"sortOrder,omitempty" jsonschema:"GVK list only. Sort direction when sortBy is set: asc (default) or desc. Omit with policyId."`
	Limit         int32    `json:"limit,omitempty" jsonschema:"GVK list only. Maximum policies in one page. Defaults to 50; maximum 500. Omit with policyId."`
	Offset        int32    `json:"offset,omitempty" jsonschema:"GVK list only. Default 0. After a full page, advance by the page size; stop after a shorter page. Omit with policyId."`
}

func (in listWorkloadPoliciesInput) Validate() error {
	if err := validatePolicyVersion(in.PolicyVersion); err != nil {
		return err
	}
	if in.PolicyVersion == versionGVK {
		return in.validateGVKList()
	}
	return in.validateLegacyList()
}

func (in listWorkloadPoliciesInput) validateGVKList() error {
	if in.WorkloadType != "" || in.Scope != "" || in.ClusterID != "" || in.DepartmentID != "" || in.ProjectID != "" {
		return fmt.Errorf("GVK filterBy supports only id, type, and scope level; omit workloadType, scope, and scope IDs. " +
			"Match clusterId, departmentId, or projectId in returned items across all pages")
	}
	if in.PolicyID != "" {
		if len(in.FilterBy) > 0 || in.SortBy != "" || in.SortOrder != "" || in.Limit != 0 || in.Offset != 0 {
			return fmt.Errorf("when policyId is set, omit filterBy, sortBy, sortOrder, limit, and offset")
		}
		return common.ValidateUUID("policyId", in.PolicyID)
	}

	return nil
}

func (in listWorkloadPoliciesInput) validateLegacyList() error {
	if in.PolicyID != "" || len(in.FilterBy) > 0 || in.SortBy != "" || in.SortOrder != "" || in.Limit != 0 || in.Offset != 0 {
		return fmt.Errorf("policyId, filterBy, sorting, limit, and offset are only supported for GVK policies (policyVersion=gvk)")
	}

	return nil
}

type listWorkloadPoliciesOutput struct {
	Items    []map[string]any `json:"items" jsonschema:"GVK policy objects with their declared rules. Contains one policy when policyId is set; empty for legacy calls."`
	Policies []map[string]any `json:"policies" jsonschema:"Legacy policy metadata, including names, IDs, and status. Empty for GVK calls."`
}

func listWorkloadPoliciesDescription() string {
	const (
		common = "List visible workload policy definitions or retrieve one GVK policy by ID. Call both policyVersion values for all policies. " +
			policyTypeDescription + "\n\n"
		legacy = "LEGACY: optionally filter by workloadType, scope, clusterId, departmentId, and projectId. Results in policies contain metadata such as names, IDs, and status. " +
			"IDs include descendants; combine scope with its matching ID to select where policies are defined. " +
			"Legacy results are metadata only; read settings with get_workload_effective_policy. Legacy does not use policyId or GVK filter/sort/pagination fields.\n\n"
		gvk = "GVK list: results in items include policy objects and declared rules. Filter with up to 10 filterBy expressions on id, type, or scope. " +
			"Expressions are ANDed; || joins alternative values, e.g. type==/Pod/v1||apps/Deployment/v1. Type is group/kind/version. Text operators are ==, !=, <=, >=, =@, !@, =^, and =$. " +
			"Sort with sortBy (id, type, scope, createdAt, createdBy, updatedAt, updatedBy) and sortOrder (asc or desc; default asc when sortBy is set). " +
			"Pagination uses limit (default 50, maximum 500) and offset (default 0). After a full page, advance offset by the page size; stop when fewer entries are returned. " +
			"To retrieve one GVK policy, provide policyId and omit filterBy, sortBy, sortOrder, limit, and offset. " +
			"Omit legacy workloadType and scope-ID inputs. Match scope IDs, descriptions, creators, and time ranges in returned metadata across all pages. " +
			"Older servers may reject UUID text matching or updatedAt sorting; compare returned metadata across all pages instead.\n\n"
		guidance = policyScopeResolutionDescription +
			"The unused result array is empty: items for legacy calls, policies for GVK calls. " +
			"Match legacy names in meta.name or GVK descriptions in description; neither uniquely identifies a policy. " +
			"Count results only after reading all pages; visibility is not authorship. " +
			"For creator questions, compare createdBy. For legacy inheritance, use get_workload_effective_policy even when no policy is listed at the target scope. " +
			"Use manage_workload_policy to create, update, or delete policies."
	)
	return common + legacy + gvk + guidance
}

func registerListWorkloadPolicies(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[listWorkloadPoliciesInput, listWorkloadPoliciesOutput]{
		Name:        "list_workload_policies",
		Description: listWorkloadPoliciesDescription(),
		ReadOnly:    true,
		Handler: func(ctx context.Context, in listWorkloadPoliciesInput) (listWorkloadPoliciesOutput, error) {
			return listWorkloadPoliciesRun(ctx, client, in)
		},
	})
}

func listWorkloadPoliciesRun(ctx context.Context, client *runaiclient.Client, in listWorkloadPoliciesInput) (listWorkloadPoliciesOutput, error) {
	out := listWorkloadPoliciesOutput{Policies: []map[string]any{}, Items: []map[string]any{}}
	var err error
	if in.PolicyVersion == versionLegacy {
		out.Policies, err = listLegacyPolicies(ctx, client, in)
		return out, err
	}
	if in.PolicyID == "" {
		out.Items, err = listGVKPolicies(ctx, client, in)
		return out, err
	}

	policy, err := getGVKPolicy(ctx, client, in.PolicyID)
	if err != nil {
		return out, err
	}
	out.Items = append(out.Items, policy)
	return out, nil
}

func listGVKPolicies(ctx context.Context, client *runaiclient.Client, in listWorkloadPoliciesInput) ([]map[string]any, error) {
	req := client.Policy.PolicyAPI.ListGvkPolicies(ctx).FilterBy(in.FilterBy).Limit(common.EffectiveLimit(in.Limit)).Offset(in.Offset)
	if in.SortBy != "" {
		req = req.SortBy(in.SortBy)
	}
	if in.SortOrder != "" {
		req = req.SortOrder(in.SortOrder)
	}
	_, resp, err := req.Execute()
	if err != nil {
		return nil, runaimcp.MapHTTPError(resp, err)
	}
	out := listWorkloadPoliciesOutput{Items: []map[string]any{}}
	if err := common.ReadJSONResponse(resp, &out); err != nil {
		return nil, fmt.Errorf("read policies: %w: %w", err, runaimcp.ErrUpstream)
	}
	if out.Items == nil {
		out.Items = []map[string]any{}
	}
	return out.Items, nil
}

func listLegacyPolicies(ctx context.Context, client *runaiclient.Client, in listWorkloadPoliciesInput) ([]map[string]any, error) {
	req := client.Policy.PolicyAPI.ListPolicies(ctx).IncludeFallbackPolicies(true).IncludeGvkPolicies(false)
	if in.Scope != "" {
		req = req.Scope(in.Scope)
	}
	if in.WorkloadType != "" {
		workloadType := in.WorkloadType
		if policy.PolicyType(workloadType).IsValid() {
			// Listing uses Training, DistributedInference, etc.; writes use lower camel case.
			workloadType = strings.ToUpper(workloadType[:1]) + workloadType[1:]
		}
		req = req.WorkloadType(workloadType)
	}
	if in.ClusterID != "" {
		req = req.ClusterId(in.ClusterID)
	}
	if in.DepartmentID != "" {
		req = req.DepartmentId(in.DepartmentID)
	}
	if in.ProjectID != "" {
		req = req.ProjectId(in.ProjectID)
	}
	_, resp, callErr := req.Execute()
	if callErr != nil {
		return nil, runaimcp.MapHTTPError(resp, callErr)
	}
	out := listWorkloadPoliciesOutput{Policies: []map[string]any{}}
	if err := common.ReadJSONResponse(resp, &out); err != nil {
		return nil, fmt.Errorf("read policies: %w: %w", err, runaimcp.ErrUpstream)
	}
	if out.Policies == nil {
		out.Policies = []map[string]any{}
	}
	return out.Policies, nil
}
