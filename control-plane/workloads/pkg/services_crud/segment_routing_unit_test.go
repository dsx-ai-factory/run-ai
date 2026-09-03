/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"net/http"
	"strings"
	"testing"

	workloadsTesting "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
)

func TestSegmentedDistributedSubmissionRouting(t *testing.T) {
	ctx := context.Background()
	const tenantID int32 = 7

	t.Run("registers the feature at cluster version 2.27", func(t *testing.T) {
		if actual := checker.MinVersionForSupporting(checker.MinVersionOfSegmentTopology); actual != "2.27.0" {
			t.Fatalf("expected segment topology minimum version 2.27.0, got %q", actual)
		}
	})

	t.Run("does not change routing for a non-segmented workload", func(t *testing.T) {
		decision := evaluateSegmentedDistributedSubmissionRouting(
			ctx,
			tenantID,
			&workloadsApi.SupersetSpec{},
			nil,
			nil,
		)

		if decision.hasIntent || decision.validated || decision.method != (SubmissionMethods{}) || decision.err != nil {
			t.Fatalf("expected no segment-specific routing decision, got %+v", decision)
		}
	})

	t.Run("rejects a cluster older than 2.27", func(t *testing.T) {
		decision := evaluateSegmentedDistributedSubmissionRouting(
			ctx,
			tenantID,
			newSegmentedPyTorchSpec(),
			newSegmentRoutingCapabilityChecker(t, "2.26.9"),
			nil,
		)

		assertSegmentRoutingError(t, decision, http.StatusBadRequest, "requires cluster version 2.27.0 or later")
	})

	t.Run("keeps a validated segment request on the native path when distributed any-workload submission is disabled", func(t *testing.T) {
		decision := evaluateSegmentedDistributedSubmissionRouting(
			ctx,
			tenantID,
			newSegmentedPyTorchSpec(),
			newSegmentRoutingCapabilityChecker(t, "2.27.0"),
			func(context.Context, int32, string, *checker.ClusterCapabilitiesChecker) SubmissionMethods {
				return SubmissionMethodNative
			},
		)

		if !decision.hasIntent || !decision.validated || decision.method != SubmissionMethodNative || decision.err != nil {
			t.Fatalf("expected validated native segment routing, got %+v", decision)
		}
	})

	t.Run("enables an eligible segment request through the existing distributed gate", func(t *testing.T) {
		gateCalled := false
		capabilityChecker := newSegmentRoutingCapabilityChecker(t, "2.27.0")

		decision := evaluateSegmentedDistributedSubmissionRouting(
			ctx,
			tenantID,
			newSegmentedPyTorchSpec(),
			capabilityChecker,
			func(_ context.Context, actualTenantID int32, workloadType string, actualChecker *checker.ClusterCapabilitiesChecker) SubmissionMethods {
				gateCalled = true
				if actualTenantID == tenantID && workloadType == Distributed && actualChecker == capabilityChecker {
					return SubmissionMethodAnyWorkload
				}
				return (SubmissionMethods{})
			},
		)

		if !decision.hasIntent || !decision.validated || decision.method != SubmissionMethodAnyWorkload || decision.err != nil {
			t.Fatalf("expected segmented WLM routing to be enabled, got %+v", decision)
		}
		if !gateCalled {
			t.Fatal("expected segmented routing to consult the existing any-workload submission gate")
		}
	})

	t.Run("fails closed when cluster support cannot be checked", func(t *testing.T) {
		decision := evaluateSegmentedDistributedSubmissionRouting(
			ctx,
			tenantID,
			newSegmentedPyTorchSpec(),
			nil,
			nil,
		)

		assertSegmentRoutingError(t, decision, http.StatusServiceUnavailable, "failed to verify cluster support")
	})

	t.Run("preserves a capability lookup failure as service unavailable", func(t *testing.T) {
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient())
		checker.InitClusterCapabilitiesChecker("", workloadsTesting.GetDefaultClustersCacheConfig(), "")

		decision := evaluateSegmentedDistributedSubmissionRouting(
			ctx,
			tenantID,
			newSegmentedPyTorchSpec(),
			checker.NewClusterCapabilitiesChecker("missing-cluster"),
			nil,
		)

		assertSegmentRoutingError(t, decision, http.StatusServiceUnavailable, "failed to verify cluster support")
	})
}

func TestSegmentRoutingErrorIsUsedByTopologyEnrichment(t *testing.T) {
	expected := workloadsApi.NewError(http.StatusBadRequest, "targeted segment routing error")
	routing := segmentedDistributedRoutingDecision{
		hasIntent: true,
		err:       expected,
	}

	actual := updateWorkloadTopologyAnnotationsForRouting(
		context.Background(),
		nil,
		newSegmentedPyTorchSpec(),
		nil,
		routing,
	)
	if actual != expected {
		t.Fatalf("expected topology enrichment to return the routing error, got %+v", actual)
	}
}

func TestDistributedSubmitterUsesCachedSegmentRoutingDecision(t *testing.T) {
	submitter := &AnyWorkloadSubmitterDistributed{}
	wc := &workloadContext{
		workloadType: Distributed,
		spec:         newSegmentedPyTorchSpec(),
		segmentRouting: &segmentedDistributedRoutingDecision{
			hasIntent: true,
			validated: true,
			method:    SubmissionMethodAnyWorkload,
		},
		// A nil checker proves DetermineSubmissionMethods did not repeat either the capability
		// validation or the ordinary any-workload submission gate.
		clusterCapabilities: nil,
	}

	if submitter.DetermineSubmissionMethods(context.Background(), wc) != SubmissionMethodAnyWorkload {
		t.Fatal("expected the cached eligible segment-routing decision to be reused")
	}

	wc.segmentRouting.method = SubmissionMethodNative
	if submitter.DetermineSubmissionMethods(context.Background(), wc) != SubmissionMethodNative {
		t.Fatal("expected a native gate decision to keep segmented submission on the native path")
	}

	wc.segmentRouting = nil
	if submitter.DetermineSubmissionMethods(context.Background(), wc) != (SubmissionMethods{}) {
		t.Fatal("expected segmented submission without a precomputed routing decision to fail closed")
	}
}

func newSegmentRoutingCapabilityChecker(t *testing.T, version string) *checker.ClusterCapabilitiesChecker {
	t.Helper()
	clusterID := strings.ReplaceAll(t.Name(), "/", "-")
	checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
		WithCluster(clusterID, workloadsTesting.BuildDisplayedCluster(clusterID, version, nil)))
	checker.InitClusterCapabilitiesChecker("", workloadsTesting.GetDefaultClustersCacheConfig(), "")
	return checker.NewClusterCapabilitiesChecker(clusterID)
}

func assertSegmentRoutingError(t *testing.T, decision segmentedDistributedRoutingDecision, code int, message string) {
	t.Helper()
	if !decision.hasIntent || decision.validated || decision.method != (SubmissionMethods{}) || decision.err == nil {
		t.Fatalf("expected disabled segment routing with an error, got %+v", decision)
	}
	if int(decision.err.Code) != code {
		t.Fatalf("expected routing error code %d, got %d", code, decision.err.Code)
	}
	if !strings.Contains(decision.err.Message, message) {
		t.Fatalf("expected routing error containing %q, got %q", message, decision.err.Message)
	}
}
