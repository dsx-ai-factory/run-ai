/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/client"
)

const workloadsPageLimit = int32(500)

func ProjectWorkloadIDs(ctx context.Context, clients *client.Clients, projectID, workloadType string) ([]string, error) {
	var ids []string
	for offset := int32(0); ; offset += workloadsPageLimit {
		page, _, err := clients.Workloads.WorkloadsAPI.GetWorkloads(ctx).
			FilterBy([]string{"projectId==" + projectID, "type==" + workloadType}).
			Limit(workloadsPageLimit).Offset(offset).Execute()
		if err != nil {
			return nil, fmt.Errorf("listing %s workloads of project %s: %w", workloadType, projectID, err)
		}
		for _, workload := range page.Workloads {
			ids = append(ids, workload.Id)
		}
		if len(page.Workloads) < int(workloadsPageLimit) {
			return ids, nil
		}
	}
}
