/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package projectresources

import (
	"context"
	"fmt"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
)

// GPUEnabledPoolIDs returns all pools where the project and its parent permit GPU use.
func GPUEnabledPoolIDs(ctx context.Context, projects openapi.ProjectsAPI, projectID string) ([]string, error) {
	request := projects.GetProjectNodePoolResources(ctx).
		FilterBy([]string{"orgUnitId==" + projectID, "gpuAccessible==true"})
	var ids []string
	for offset := int32(0); ; {
		page, resp, err := request.Offset(offset).Limit(100).Execute()
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		if err != nil {
			return nil, err
		}
		if page == nil {
			return nil, fmt.Errorf("empty project node pool response")
		}
		for _, pool := range page.NodePoolResources {
			if pool.OrgUnitId != projectID {
				return nil, fmt.Errorf("unexpected project in node pool response")
			}
			ids = append(ids, pool.NodePoolId)
		}
		if page.GetNext() == 0 {
			return ids, nil
		}
		if page.GetNext() <= offset {
			return nil, fmt.Errorf("invalid next offset for project node pools: %d", page.GetNext())
		}
		offset = page.GetNext()
	}
}
