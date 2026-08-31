/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mutator

import (
	"context"
	"encoding/json"
	"fmt"
	version_utils "github.com/run-ai/runai/runai-cluster/cluster/common/utils/version"
	"net/http"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/common/topology"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/builders"
	mnnvltypes "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	workloadcommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	admissionv1 "k8s.io/api/admission/v1"
	"knative.dev/pkg/logging"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	validOk                 = "Ok"
	minVersionMnnvlMutation = "2.23.0" // older workloads will not be mutated for MNNVL
)

type DistributedInferenceMutatorHandler struct {
	k8sClient       client.Client
	mnnvlEnabled    bool
	topologyEnabled bool
	elasticEnabled  bool
}

func NewDistributedInferenceMutator(k8sClient client.Client, mnnvlEnabled, topologyEnabled, elasticEnabled bool) *DistributedInferenceMutatorHandler {
	return &DistributedInferenceMutatorHandler{
		k8sClient:       k8sClient,
		mnnvlEnabled:    mnnvlEnabled,
		topologyEnabled: topologyEnabled,
		elasticEnabled:  elasticEnabled,
	}
}

func (dm *DistributedInferenceMutatorHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Operation {
	case admissionv1.Create:
		return dm.HandleCreate(ctx, req)
	default:
		// in the case of update/delete, we don't need to do anything
		return admission.Allowed(validOk)
	}
}

func (dm *DistributedInferenceMutatorHandler) HandleCreate(ctx context.Context, req admission.Request) admission.Response {
	log := logging.FromContext(ctx)

	if req.Kind.Kind != constants.KindDistributedInferenceWorkload {
		return admission.Allowed(validOk)
	}

	// decode the request object
	workload := dinfw.DistributedInferenceWorkload{}
	err := json.Unmarshal(req.Object.Raw, &workload)
	if err != nil {
		log.Error(err, "failed to decode request")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// mutate the workload for MNNVL if required
	isPatched, err := dm.mutateForMnnvl(ctx, &workload)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if !isPatched {
		return admission.Allowed(validOk)
	}

	if dm.topologyEnabled {
		err = copyTopologyAnnotations(ctx, &workload)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to copy workload topology annotations %s, error: %v", workload.GetName(), err))
		}
	}

	// patch the workload and return it
	patchedWorkload, err := json.Marshal(workload)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to patch to workload %s, error: %v", workload.GetName(), err))
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, patchedWorkload)
}

func (dm *DistributedInferenceMutatorHandler) mutateForMnnvl(ctx context.Context, workload *dinfw.DistributedInferenceWorkload) (bool, error) {
	// if the workload already has MNNVL label with value None or Required, we don't need to mutate it
	mnnvlLabel, ok := workload.Labels[mnnvltypes.FeatureLabelKey]
	if ok && (mnnvlLabel == mnnvltypes.None || mnnvlLabel == mnnvltypes.Required) {
		return false, nil
	}

	// if the workload was created by a too-old cluster version, we don't need to mutate it
	if isOldWorkload(ctx, workload) {
		return false, nil
	}

	// otherwise, we need to mutate the workload:
	// if all the conditions for MNNVL are met, on both the cluster and workload levels,
	// we need to set the label to Required. Otherwise, we need to set it to None
	mnnvlMode := mnnvltypes.None
	if dm.mnnvlEnabled {
		isMnnvlSupported, err := dm.isMnnvlSupportedForWorkload(ctx, workload)
		if err != nil {
			return false, err
		}
		if isMnnvlSupported {
			mnnvlMode = mnnvltypes.Required
		}
	}

	if workload.Labels == nil {
		workload.Labels = make(map[string]string)
	}
	workload.Labels[mnnvltypes.FeatureLabelKey] = mnnvlMode

	return true, nil
}

// isOldWorkload checks if the workload was created by a too-old cluster version, that doesn't support MNNVL
func isOldWorkload(ctx context.Context, workload *dinfw.DistributedInferenceWorkload) bool {
	clusterVersion, ok := workload.Labels[workloadcommon.CreatedAtClusterVersionLabel]
	if !ok {
		return false
	}

	isOld, err := version_utils.VersionLessThan(clusterVersion, minVersionMnnvlMutation)
	if err != nil {
		log := logging.FromContext(ctx)
		log.Error(err, "failed to compare cluster version")
		return false
	}

	return isOld
}

func (dm *DistributedInferenceMutatorHandler) isMnnvlSupportedForWorkload(ctx context.Context, workload *dinfw.DistributedInferenceWorkload) (bool, error) {
	mnnvlBuilder := builders.NewMnnvlBuilder(dm.k8sClient, dm.elasticEnabled)

	// verify the workload has GPUs requested
	if mnnvlBuilder.GetNumGpuNodes(workload) == 0 {
		return false, nil
	}

	// verify the node pools support MNNVL
	isNodePoolsSupportMnnvl, err := mnnvlBuilder.IsNodePoolsSupportMnnvl(ctx, workload)
	if err != nil {
		return false, err
	}
	if !isNodePoolsSupportMnnvl {
		return false, nil
	}

	return true, nil
}

func copyTopologyAnnotations(ctx context.Context, workload *dinfw.DistributedInferenceWorkload) error {
	return topology.CopySpecTopologyAnnotation(ctx, workload.GetAnnotations(), workload.LeaderSpec.Annotations)
}
