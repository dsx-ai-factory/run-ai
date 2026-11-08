/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
	inferenceRoutingNotSubmittableMessage       = "metadata.configuration.inferenceRouting is set by the server and cannot be submitted"
	loadAwareRoutingUnavailableMessage          = "load-aware routing is not available on this cluster"
	inferenceRoutingNetworkingOnlyUpdateMessage = "the networking of a workload with load-aware routing can change only through a workload update"
	inferenceRoutingFixedAtCreationMessage      = "load-aware routing is chosen when the workload is created and cannot be turned on or off by an update"
	inferenceRoutingPathPrefixMessage           = "loadAwareRoutingPathPrefix must start with '/' and can be set only when loadAwareRouting is true"
	inferenceRoutingUnsupportedWorkloadMessage  = "load-aware routing is supported only for distributed inference and NIM service workloads"
)

// inferenceRoutingAdapter selects the intent builder for a workload that opted in.
type inferenceRoutingAdapter int

const (
	inferenceRoutingUnsupported inferenceRoutingAdapter = iota
	inferenceRoutingDistributedInference
	inferenceRoutingNIMService
)

// inferenceRoutingAdapterFor selects the adapter from the server-established
// origin. The manifest labels and kind alone are caller-controlled.
// trustedNIMService is set only by the in-process NIM API, which composes the manifest.
func inferenceRoutingAdapterFor(trustedConvertedFromNative, trustedNIMService bool, resource *ResourceInfo) inferenceRoutingAdapter {
	switch {
	case resource == nil:
		return inferenceRoutingUnsupported
	case trustedConvertedFromNative && resource.Labels[constants.LabelWorkloadKind] == composer.DistributedInference:
		return inferenceRoutingDistributedInference
	case trustedNIMService && resource.Kind == nimServiceKind && resource.APIVersion == nimServiceAPIVersion:
		return inferenceRoutingNIMService
	}
	return inferenceRoutingUnsupported
}

// inferenceRoutingOptedIn reports whether the workload spec selects the inference
// endpoint path.
func inferenceRoutingOptedIn(routing *api.WorkloadV2ConfigurationInferenceRouting) bool {
	return routing != nil && convutils.Deref(routing.LoadAwareRouting, false)
}

// validateInferenceRoutingPathPrefix rejects a path prefix that does not start
// with a slash, or one set without the opt-in.
func validateInferenceRoutingPathPrefix(routing *api.WorkloadV2ConfigurationInferenceRouting) *api.SubmissionErrorV2 {
	if routing == nil || routing.LoadAwareRoutingPathPrefix == nil {
		return nil
	}
	if !inferenceRoutingOptedIn(routing) || !strings.HasPrefix(*routing.LoadAwareRoutingPathPrefix, "/") {
		return NewConfigurationError(inferenceRoutingPathPrefixMessage)
	}
	return nil
}

// validateInferenceRoutingOrigin rejects inferenceRouting outside the internal API
// and the in-process NIM API. Both fill it from the spec after project policy ran;
// a caller of the public API could otherwise bypass the project policy.
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
	adapter inferenceRoutingAdapter,
) (*inferenceendpoint.Defaults, *api.SubmissionErrorV2) {
	if err := validateInferenceRoutingPathPrefix(routing); err != nil {
		return nil, err
	}
	if !inferenceRoutingOptedIn(routing) {
		return nil, nil
	}
	if adapter == inferenceRoutingUnsupported {
		return nil, NewConfigurationError(inferenceRoutingUnsupportedWorkloadMessage)
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
	adapter inferenceRoutingAdapter,
	routing *api.WorkloadV2ConfigurationInferenceRouting,
	defaults *inferenceendpoint.Defaults,
	manifest map[string]interface{},
	configuration *models.WorkloadConfiguration,
	workloadName string,
) *api.SubmissionErrorV2 {
	if defaults == nil {
		return nil
	}
	// The distributed inference composer and the NIM API create one endpoint, for the serving port.
	if configuration == nil || configuration.Networking == nil || len(*configuration.Networking) != 1 {
		return NewConfigurationError("load-aware routing requires one serving port endpoint")
	}

	endpoint := &(*configuration.Networking)[0]
	pathPrefix := convutils.Deref(routing.LoadAwareRoutingPathPrefix, "")
	var inference *api.ExternalAccessInference
	var err error
	switch adapter {
	case inferenceRoutingDistributedInference:
		lws, decodeErr := decodeLeaderWorkerSet(manifest)
		if decodeErr != nil {
			return NewManifestError(fmt.Sprintf("load-aware routing: %v", decodeErr))
		}
		inference, err = inferenceendpoint.BuildDistributedInferenceIntent(lws, *endpoint, pathPrefix, defaults)
	case inferenceRoutingNIMService:
		inference, err = inferenceendpoint.BuildNIMServiceIntent(*endpoint, workloadName, pathPrefix, defaults)
	default:
		return NewConfigurationError(inferenceRoutingUnsupportedWorkloadMessage)
	}
	if err != nil {
		return NewConfigurationError(fmt.Sprintf("load-aware routing: %v", err))
	}
	endpoint.ExternalAccess.Inference = inference
	return nil
}

// validateInferenceRoutingUnchanged keeps a workload on the routing path it was
// created with. An update cannot add or remove the opt-in, because that would swap
// the composed HTTPRoute and the InferenceEndpoint route on a running workload.
func validateInferenceRoutingUnchanged(existing *models.WorkloadConfiguration, routing *api.WorkloadV2ConfigurationInferenceRouting) *api.SubmissionErrorV2 {
	if hasInferenceRouting(existing) != inferenceRoutingOptedIn(routing) {
		return NewConfigurationError(inferenceRoutingFixedAtCreationMessage)
	}
	return nil
}

// resolveNIMServiceInferenceRouting returns the NIM serving networking with the
// inference block resolved. The NIM update replaces the networking through a
// record modifier rather than the update request's configuration, so
// updateWorkloadV2 does not resolve the block for it.
func resolveNIMServiceInferenceRouting(
	ctx context.Context,
	routing *api.WorkloadV2ConfigurationInferenceRouting,
	clusterId string,
	workloadName string,
	networking *[]api.NetworkingEndpoint,
) (*models.NetworkingConfiguration, *api.SubmissionErrorV2) {
	configuration := &models.WorkloadConfiguration{Networking: cloneNetworkingConfig(networking)}
	defaults, err := resolveInferenceRoutingDefaults(ctx, routing, clusterId, inferenceRoutingNIMService)
	if err != nil {
		return nil, err
	}
	if err := applyInferenceRouting(inferenceRoutingNIMService, routing, defaults, nil, configuration, workloadName); err != nil {
		return nil, err
	}
	return configuration.Networking, nil
}

func hasInferenceRouting(configuration *models.WorkloadConfiguration) bool {
	if configuration == nil || configuration.Networking == nil {
		return false
	}
	for _, endpoint := range *configuration.Networking {
		if endpoint.ExternalAccess != nil && endpoint.ExternalAccess.Inference != nil {
			return true
		}
	}
	return false
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
