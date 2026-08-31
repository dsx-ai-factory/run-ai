/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"fmt"
	"net/http"

	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
)

type segmentedDistributedRoutingDecision struct {
	hasIntent bool
	validated bool
	method    SubmissionMethods
	err       *workloadsApi.Error
}

type anyWorkloadSubmissionGate func(
	ctx context.Context,
	tenantID int32,
	workloadType string,
	clusterCapabilities *checker.ClusterCapabilitiesChecker,
) SubmissionMethods

// segmentedDistributedSubmissionRouting validates segment support and records
// the ordinary distributed routing decision. The request does not override the
// Any Workload rollout gates: when those gates are disabled, a validated
// segmented request continues through the native distributed path.
func segmentedDistributedSubmissionRouting(
	ctx context.Context,
	tenantID int32,
	spec *workloadsApi.SupersetSpec,
	clusterCapabilities *checker.ClusterCapabilitiesChecker,
) segmentedDistributedRoutingDecision {
	return evaluateSegmentedDistributedSubmissionRouting(
		ctx,
		tenantID,
		spec,
		clusterCapabilities,
		determineSubmissionMethods,
	)
}

func evaluateSegmentedDistributedSubmissionRouting(
	ctx context.Context,
	tenantID int32,
	spec *workloadsApi.SupersetSpec,
	clusterCapabilities *checker.ClusterCapabilitiesChecker,
	anyWorkloadEnabled anyWorkloadSubmissionGate,
) segmentedDistributedRoutingDecision {
	decision := segmentedDistributedRoutingDecision{hasIntent: hasDistributedSegmentIntent(spec)}
	if !decision.hasIntent {
		return decision
	}

	if clusterCapabilities == nil {
		decision.err = workloadsApi.NewError(
			http.StatusServiceUnavailable,
			"failed to verify cluster support for segment topology-aware scheduling",
		)
		return decision
	}

	supported, capabilityErr := clusterCapabilities.ValidateClusterCapability(ctx, checker.MinVersionOfSegmentTopology)
	if capabilityErr != nil {
		decision.err = workloadsApi.NewError(
			http.StatusServiceUnavailable,
			fmt.Sprintf("failed to verify cluster support for segment topology-aware scheduling: %v", capabilityErr),
		)
		return decision
	}
	if !supported {
		decision.err = workloadsApi.NewError(
			http.StatusBadRequest,
			fmt.Sprintf(
				"segment topology-aware scheduling requires cluster version %s or later",
				checker.MinVersionForSupporting(checker.MinVersionOfSegmentTopology),
			),
		)
		return decision
	}

	decision.validated = true
	if anyWorkloadEnabled != nil {
		decision.method = anyWorkloadEnabled(ctx, tenantID, Distributed, clusterCapabilities)
	}
	return decision
}
