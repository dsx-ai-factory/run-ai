/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"
	"fmt"
	"strings"

	orgunit "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

const (
	OrgTypeProject    = "project"
	OrgTypeDepartment = "department"
)

// AllocatedGPU returns the point-in-time allocated GPU count from an org-unit quota status,
// or nil when the platform reports no allocation.
func AllocatedGPU(qs *orgunit.QuotaStatus) *float64 {
	if qs == nil || qs.Allocated == nil {
		return nil
	}
	return qs.Allocated.Gpu
}

// AllocatedCPUCores returns allocated CPU normalized to cores (upstream ships millicores).
func AllocatedCPUCores(qs *orgunit.QuotaStatus) *float64 {
	if qs == nil || qs.Allocated == nil || qs.Allocated.Cpu == nil {
		return nil
	}
	v := MillicoresToCores(*qs.Allocated.Cpu)
	return &v
}

// AllocatedMemoryGiB returns allocated memory normalized to GiB. quotaStatus memory is always
// MB per the API (unlike resources.memory, which carries its own units).
func AllocatedMemoryGiB(qs *orgunit.QuotaStatus) *float64 {
	if qs == nil || qs.Allocated == nil || qs.Allocated.Memory == nil {
		return nil
	}
	v := MemoryToGiB(*qs.Allocated.Memory, orgunit.UNITS_MB)
	return &v
}

func ResolveOrgUnitID(ctx context.Context, client *runaiclient.Client, orgType, name, clusterID string) (string, error) {
	if orgType != OrgTypeProject && orgType != OrgTypeDepartment {
		return "", fmt.Errorf("orgType must be one of: %s, %s (got %q): %w",
			OrgTypeProject, OrgTypeDepartment, orgType, runaimcp.ErrValidation)
	}
	if err := ValidateNoCommas("orgName", name); err != nil {
		return "", fmt.Errorf("%v: %w", err, runaimcp.ErrValidation)
	}
	if clusterID != "" {
		if err := ValidateNoCommas("clusterId", clusterID); err != nil {
			return "", fmt.Errorf("%v: %w", err, runaimcp.ErrValidation)
		}
	}

	filters := []string{FilterEq("name", name)}
	if clusterID != "" {
		filters = append(filters, FilterEq("clusterId", clusterID))
	}

	type candidate struct{ id, clusterID string }
	var cands []candidate
	switch orgType {
	case OrgTypeProject:
		resp, httpResp, err := client.OrgUnit.ProjectsAPI.GetProjects(ctx).FilterBy(filters).Limit(2).Execute()
		if err != nil {
			return "", fmt.Errorf("resolve project name %q: %w", name, runaimcp.MapHTTPError(httpResp, err))
		}
		if resp == nil {
			return "", fmt.Errorf("resolve project name %q: nil response: %w", name, runaimcp.ErrUpstream)
		}
		for _, p := range resp.Projects {
			cands = append(cands, candidate{p.Id, p.ClusterId})
		}
	case OrgTypeDepartment:
		resp, httpResp, err := client.OrgUnit.DepartmentsAPI.GetDepartments(ctx).FilterBy(filters).Limit(2).Execute()
		if err != nil {
			return "", fmt.Errorf("resolve department name %q: %w", name, runaimcp.MapHTTPError(httpResp, err))
		}
		if resp == nil {
			return "", fmt.Errorf("resolve department name %q: nil response: %w", name, runaimcp.ErrUpstream)
		}
		for _, d := range resp.Departments {
			cands = append(cands, candidate{d.Id, d.ClusterId})
		}
	}

	switch len(cands) {
	case 0:
		return "", fmt.Errorf("%s %q: %w", orgType, name, runaimcp.ErrNotFound)
	case 1:
		return cands[0].id, nil
	default:
		matches := make([]string, 0, len(cands))
		for _, c := range cands {
			matches = append(matches, fmt.Sprintf("%s (cluster %s)", c.id, c.clusterID))
		}
		return "", fmt.Errorf("ambiguous_org_unit: %s name %q matches multiple org units (%s); narrow with clusterId: %w",
			orgType, name, strings.Join(matches, ", "), runaimcp.ErrValidation)
	}
}
