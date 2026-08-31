/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mutators

import (
	"context"
	"encoding/json"
	"fmt"
	clusterConstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/common/topology"
	field_types "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	distw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	interactive "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	training "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	workloadCommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	commonErrors "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/errors"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/policy"
	"github.com/run-ai/runai/runai-cluster/common/constants"
	admissionV1 "k8s.io/api/admission/v1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type PhaseSetter interface {
	SetPhaseIfNeeded(jobWorkload client.Object) error
}

type WorkloadStatusUpdater struct {
}

func (as *WorkloadStatusUpdater) SetPhaseIfNeeded(jobWorkload client.Object) error {
	status := &common.WorkloadCommonStatus{}

	switch workloadCast := jobWorkload.(type) {
	case *interactive.InteractiveWorkload:
		status = &workloadCast.Status.WorkloadCommonStatus
	case *training.TrainingWorkload:
		status = &workloadCast.Status.WorkloadCommonStatus
	case *distw.DistributedWorkload:
		status = &workloadCast.Status.WorkloadCommonStatus
	default:
		return fmt.Errorf("couldn't extract status from object")
	}

	if status != nil {
		status.Phase = common.PhaseCreating.String()
	}

	return nil
}

type WorkloadMutatorHandler struct {
	workloadPolicy   policy.WorkloadPolicy
	getWorkloadObjFn func() client.Object
	clusterVersion   string
	StatusUpdater    PhaseSetter
	mnnvlEnabled     bool
	topologyEnabled  bool
	elasticCDEnabled bool
}

func NewWorkloadMutatorHandler(workloadPolicy policy.WorkloadPolicy, getWorkloadObjFn func() client.Object, clusterVersion string,
	statusUpdater PhaseSetter, mnnvlEnabled, topologyEnabled, elasticCDEnabled bool) *WorkloadMutatorHandler {
	return &WorkloadMutatorHandler{
		workloadPolicy:   workloadPolicy,
		getWorkloadObjFn: getWorkloadObjFn,
		clusterVersion:   clusterVersion,
		StatusUpdater:    statusUpdater,
		mnnvlEnabled:     mnnvlEnabled,
		topologyEnabled:  topologyEnabled,
		elasticCDEnabled: elasticCDEnabled,
	}
}

func (pm *WorkloadMutatorHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := logf.FromContext(ctx)
	switch req.Operation {
	case admissionV1.Create:
		workload := pm.getWorkloadObjFn()
		err := decodeRequest(req, workload)
		if err != nil {
			log.Error(err, "Failed to decode request")
			return admission.Errored(http.StatusServiceUnavailable, err)
		}
		return pm.HandleCreate(ctx, workload, req.Object.Raw)
	default:
		// in case of delete/update, we don't need to do anything
		return admission.Allowed("Ok")
	}
}

func (pm *WorkloadMutatorHandler) HandleCreate(ctx context.Context, jobWorkload client.Object, rawObject []byte) admission.Response {
	log := logf.FromContext(ctx)
	var err error
	if jobWorkload == nil {
		// this should never happen
		return admission.Denied("workload cannot be nil")
	}
	if err = annotateRequestedNodeAffinity(jobWorkload); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	if isControlPlaneManaged(jobWorkload) {
		log.Info("control plane managed workload is already patched and validated.")
	} else {
		err = pm.workloadPolicy.ApplyPolicy(ctx, jobWorkload)
		if err != nil {
			if commonErrors.IsWorkloadPolicyValidationError(err) {
				log.Error(err, "Failed to apply policy to workload")
				return admission.Denied(fmt.Sprintf("workload is invalid, error: %v", err))
			}
			log.Error(err, "Failed to apply policy to workload")
			return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to apply policy to workload %s, error: %v", jobWorkload.GetName(), err))
		}
	}
	// try extracting status from workload
	err = pm.StatusUpdater.SetPhaseIfNeeded(jobWorkload)
	if err != nil {
		// we don't want to fail the validation webhook for that
		log.Info("status extraction failed", "error", err.Error())
	}

	// for any future use case
	err = pm.addClusterVersionLabel(jobWorkload)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	err = setMnnvlFeatureLabel(jobWorkload, pm.mnnvlEnabled, pm.elasticCDEnabled)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if pm.topologyEnabled {
		err = copyTopologyAnnotations(ctx, jobWorkload)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to copy workload topology annotations %s, error: %v", jobWorkload.GetName(), err))
		}
	}

	policyPatchedWorkload, err := json.Marshal(jobWorkload)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to apply policy patch to workload %s, error: %v", jobWorkload.GetName(), err))
	}

	return admission.PatchResponseFromRaw(rawObject, policyPatchedWorkload)
}

func (pm *WorkloadMutatorHandler) addClusterVersionLabel(jobWorkload client.Object) error {
	labels := jobWorkload.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[workloadCommon.CreatedAtClusterVersionLabel] = pm.clusterVersion
	jobWorkload.SetLabels(labels)
	return nil
}

func decodeRequest(req admission.Request, object client.Object) error {
	err := json.Unmarshal(req.Object.Raw, object)
	if err != nil {
		return fmt.Errorf("failed to decode request: %v", err)
	}
	return nil
}

func isControlPlaneManaged(jobWorkload client.Object) bool {
	if value, exists := jobWorkload.GetLabels()[constants.ControlPlaneManagedLabel]; exists && value == "true" {
		return true
	}
	return false
}

func copyTopologyAnnotations(ctx context.Context, jobWorkload client.Object) error {
	dw, isDist := jobWorkload.(*distw.DistributedWorkload)
	if !isDist {
		return nil
	}

	return topology.CopySpecTopologyAnnotation(ctx, dw.GetAnnotations(), dw.Spec.Annotations)
}

// annotateRequestedNodeAffinity stores the user-submitted node affinity as an annotation on the
// workload before policy is applied, preserving the original request for downstream consumers.
// If the annotation already exists it is not overwritten.
func annotateRequestedNodeAffinity(jobWorkload client.Object) error {
	var nodeAffinityRequired *field_types.NodeSelector

	switch w := jobWorkload.(type) {
	case *training.TrainingWorkload:
		nodeAffinityRequired = w.Spec.NodeAffinityRequired
	case *interactive.InteractiveWorkload:
		nodeAffinityRequired = w.Spec.NodeAffinityRequired
	case *distw.DistributedWorkload:
		nodeAffinityRequired = w.Spec.NodeAffinityRequired
	default:
		return nil
	}

	if nodeAffinityRequired == nil {
		return nil
	}

	annotations := jobWorkload.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	if _, exists := annotations[clusterConstants.AnnotationRequestedNodeAffinity]; exists {
		return nil
	}

	data, err := json.Marshal(nodeAffinityRequired)
	if err != nil {
		return fmt.Errorf("failed to marshal requested node affinity: %w", err)
	}

	annotations[clusterConstants.AnnotationRequestedNodeAffinity] = string(data)
	jobWorkload.SetAnnotations(annotations)
	return nil
}
