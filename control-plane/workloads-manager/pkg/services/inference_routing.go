/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"

	clustersapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	clustersClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/clusters"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/inferenceendpoint"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

const (
	inferenceRoutingNotSubmittableMessage = "metadata.configuration.inferenceRouting is set by the server and cannot be submitted"
	loadAwareRoutingUnavailableMessage    = "load-aware routing is not available on this cluster"
)

// inferenceRoutingOptedIn reports whether the workload spec selects the inference
// endpoint path. Any objectivePriority value opts in, including 0.
func inferenceRoutingOptedIn(routing *api.WorkloadV2ConfigurationInferenceRouting) bool {
	if routing == nil {
		return false
	}
	return convutils.Deref(routing.LoadAwareRouting, false) ||
		routing.ObjectivePriority != nil ||
		(routing.ModelRewrites != nil && len(*routing.ModelRewrites) > 0)
}

// validateInferenceRoutingOrigin rejects inferenceRouting outside the internal API.
// The workloads service fills it from the spec after project policy ran; a caller
// of the public API could otherwise set objective priorities and rewrites that
// the project policy does not allow.
func validateInferenceRoutingOrigin(configuration *api.WorkloadV2Configuration, trustedInternalCaller bool) *api.SubmissionErrorV2 {
	if trustedInternalCaller || configuration == nil || configuration.InferenceRouting == nil {
		return nil
	}
	return NewConfigurationError(inferenceRoutingNotSubmittableMessage)
}

// resolveInferenceRoutingDefaults returns nil when the workload did not opt in.
// An explicit opt-in is rejected, never ignored, when the cluster cannot serve it.
func resolveInferenceRoutingDefaults(
	ctx context.Context,
	routing *api.WorkloadV2ConfigurationInferenceRouting,
	clusterId string,
	trustedConvertedFromNative bool,
	manifestLabels map[string]string,
) (*inferenceendpoint.Defaults, *api.SubmissionErrorV2) {
	if !inferenceRoutingOptedIn(routing) {
		return nil, nil
	}
	if !trustedConvertedFromNative || manifestLabels[constants.LabelWorkloadKind] != composer.DistributedInference {
		return nil, NewConfigurationError("load-aware routing is supported only for distributed inference workloads")
	}

	cluster, clusterErr := clustersClient.Get().GetClusterByUuid(ctx, clusterId)
	if clusterErr != nil {
		return nil, NewSubmissionError(int(clusterErr.Code), fmt.Sprintf("failed to get cluster info: %s", clusterErr.Message), nil)
	}
	defaults, err := inferenceendpoint.ResolveDefaults(reportedClusterStatus(cluster.Status.Get()))
	if err != nil {
		return nil, NewConfigurationError(fmt.Sprintf("load-aware routing: %v", err))
	}
	if defaults == nil {
		return nil, NewConfigurationError(loadAwareRoutingUnavailableMessage)
	}
	return defaults, nil
}

// applyInferenceRouting sets externalAccess.inference on the serving endpoint. It
// runs after the caller's networking was validated, because that validation
// rejects a submitted inference block. defaults is nil when the workload did not opt in.
func applyInferenceRouting(
	routing *api.WorkloadV2ConfigurationInferenceRouting,
	defaults *inferenceendpoint.Defaults,
	manifest map[string]interface{},
	configuration *models.WorkloadConfiguration,
	workloadName string,
) *api.SubmissionErrorV2 {
	if defaults == nil {
		return nil
	}
	// The distributed inference composer creates one endpoint, for the serving port.
	if configuration == nil || configuration.Networking == nil || len(*configuration.Networking) != 1 {
		return NewConfigurationError("load-aware routing requires one serving port endpoint")
	}

	lws, decodeErr := decodeLeaderWorkerSet(manifest)
	if decodeErr != nil {
		return NewManifestError(fmt.Sprintf("load-aware routing: %v", decodeErr))
	}

	endpoint := &(*configuration.Networking)[0]
	protected := routing.ServingPortAuthorizationType != nil &&
		*routing.ServingPortAuthorizationType != api.SERVINGPORTACCESSAUTHORIZATIONTYPEENUM_PUBLIC
	inference, err := inferenceendpoint.BuildDistributedInferenceIntent(lws, *endpoint, protected, convutils.Deref(routing.ModelRewrites, nil), defaults)
	if err != nil {
		return NewConfigurationError(fmt.Sprintf("load-aware routing: %v", err))
	}
	// One objective lane per workload, named after the workload.
	if routing.ObjectivePriority != nil {
		objectives := []map[string]interface{}{{"model": workloadName, "priority": *routing.ObjectivePriority}}
		inference.Objectives = &objectives
	}
	endpoint.ExternalAccess.Inference = inference
	return nil
}

// reportedClusterStatus drops the control-plane connection state, which is the
// only field the displayed status adds to the status the cluster reports.
func reportedClusterStatus(status *clustersapi.ClusterDisplayedStatus) *clustersapi.ClusterReportedStatus {
	if status == nil {
		return nil
	}
	return &clustersapi.ClusterReportedStatus{
		Conditions:   status.Conditions,
		Operands:     status.Operands,
		Platform:     status.Platform,
		Config:       status.Config,
		Dependencies: status.Dependencies,
	}
}

func decodeLeaderWorkerSet(manifest map[string]interface{}) (*lwsv1.LeaderWorkerSet, error) {
	raw, err := json.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("encode LeaderWorkerSet: %w", err)
	}
	var lws lwsv1.LeaderWorkerSet
	if err := json.Unmarshal(raw, &lws); err != nil {
		return nil, fmt.Errorf("decode LeaderWorkerSet: %w", err)
	}
	return &lws, nil
}
