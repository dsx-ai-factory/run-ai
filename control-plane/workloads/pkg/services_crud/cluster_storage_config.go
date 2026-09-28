/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"
)

// Uses the internal cluster endpoint, which carries no authorization, so the
// submission flow works for regular users who lack the cluster:read permission
// required by the external GetClusterByUuid endpoint.
func buildClusterStorageConfig(ctx context.Context, clusterId string) (composer.ClusterStorageConfig, *workloadsApi.SubmissionErrorV2) {
	cluster, getErr := clusterservice.Get().GetInternalCluster(ctx, clusterId)
	if getErr != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusServiceUnavailable, getErr.Error())
	}
	return composer.NewClusterStorageConfigFromCluster(cluster), nil
}
