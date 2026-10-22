/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policies

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

const (
	versionLegacy     = common.PolicyVersionLegacy
	versionGVK        = common.PolicyVersionGVK
	actionCreate      = string(common.ActionCreate)
	actionUpdate      = string(common.ActionUpdate)
	actionDelete      = string(common.ActionDelete)
	fieldClusterID    = "clusterId"
	fieldDepartmentID = "departmentId"
	fieldProjectID    = "projectId"

	policyTypeDescription            = "Set policyVersion=legacy for native workload policies, or policyVersion=gvk for policies identified by Kubernetes Group, Version, and Kind (GVK). "
	policyScopeResolutionDescription = "Resolve cluster, department, and project names to IDs through list_clusters, list_departments, or list_projects. " +
		"Check all pages and ask the user if a name matches multiple IDs. "
)

func validatePolicyVersion(version string) error {
	if version != versionLegacy && version != versionGVK {
		return fmt.Errorf("policyVersion must be legacy (older named-workload policies) or gvk (GVK policies)")
	}
	return nil
}

func executePolicyOperation(ctx context.Context, client *runaiclient.Client, in manageWorkloadPolicyInput) (map[string]any, error) {
	if in.PolicyVersion == versionLegacy {
		return executeLegacyPolicyOperation(ctx, client, in)
	}
	return executeGVKPolicyOperation(ctx, client, in)
}
