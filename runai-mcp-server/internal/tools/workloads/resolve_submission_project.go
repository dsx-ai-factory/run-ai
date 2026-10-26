/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"fmt"
	"strings"
	"time"

	orgunit "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

const namespaceResolutionTimeout = 30 * time.Second

// resolveSubmissionProject fetches the project metadata needed to submit and
// verifies an explicitly supplied cluster matches that project.
func resolveSubmissionProject(ctx context.Context, client *runaiclient.Client, projectID, clusterID string) (*orgunit.Project, error) {
	project, err := getSubmissionProject(ctx, client, projectID)
	if err != nil {
		return nil, err
	}
	if clusterID != "" && !strings.EqualFold(project.ClusterId, clusterID) {
		return nil, fmt.Errorf("project %q belongs to cluster %q, not %q: %w", projectID, project.ClusterId, clusterID, runaimcp.ErrValidation)
	}
	return project, nil
}

func resolveYAMLSubmissionProject(ctx context.Context, client *runaiclient.Client, projectID, clusterID, namespace string) (*orgunit.Project, error) {
	if projectID != "" {
		return resolveSubmissionProject(ctx, client, projectID, "")
	}
	return resolveProjectForNamespace(ctx, client, clusterID, namespace)
}

func resolveProjectForNamespace(ctx context.Context, client *runaiclient.Client, clusterID, namespace string) (*orgunit.Project, error) {
	resolveCtx, cancel := context.WithTimeout(ctx, namespaceResolutionTimeout)
	defer cancel()

	offset := int32(0)
	for {
		resp, httpResp, err := client.OrgUnit.ProjectsAPI.GetProjectsMetadata(resolveCtx).
			FilterBy([]string{common.FilterEq("clusterId", clusterID)}).
			Limit(common.UpstreamPageSize).
			Offset(offset).
			Execute()
		if err != nil {
			return nil, fmt.Errorf("resolve project for namespace %q: %w", namespace, runaimcp.MapHTTPError(httpResp, err))
		}
		if resp == nil {
			return nil, fmt.Errorf("resolve project for namespace %q: nil response: %w", namespace, runaimcp.ErrUpstream)
		}
		for i := range resp.Projects {
			project := &resp.Projects[i]
			projectNamespace := project.GetNamespace()
			if projectNamespace == "" {
				projectNamespace = "runai-" + project.Name
			}
			if projectNamespace == namespace {
				resolved, err := getSubmissionProject(resolveCtx, client, project.Id)
				if err != nil {
					return nil, err
				}
				return resolved, nil
			}
		}
		if resp.Next == nil || *resp.Next <= offset {
			break
		}
		if len(resp.Projects) == 0 {
			return nil, fmt.Errorf("resolve project for namespace %q: empty page reported a next offset: %w", namespace, runaimcp.ErrUpstream)
		}
		offset = *resp.Next
	}
	return nil, fmt.Errorf("project for cluster %q and namespace %q: %w", clusterID, namespace, runaimcp.ErrNotFound)
}

func getSubmissionProject(ctx context.Context, client *runaiclient.Client, projectID string) (*orgunit.Project, error) {
	project, resp, err := client.OrgUnit.ProjectsAPI.GetProject(ctx, projectID).Execute()
	if err != nil {
		return nil, fmt.Errorf("get project %q: %w", projectID, runaimcp.MapHTTPError(resp, err))
	}
	if project == nil {
		return nil, fmt.Errorf("get project %q: nil response: %w", projectID, runaimcp.ErrUpstream)
	}
	return project, nil
}
