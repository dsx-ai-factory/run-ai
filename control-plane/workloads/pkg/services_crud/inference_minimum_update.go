/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"k8s.io/apimachinery/pkg/api/equality"
	"knative.dev/serving/pkg/apis/autoscaling"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

// Some native settings become WorkloadRunner instructions outside the Service.
// Compare the full snapshot too, including environment values redacted from GET.
func onlyInferenceMinimumChanged(previous, desired *workloadsApi.SupersetSpec) bool {
	if previous == nil || desired == nil {
		return false
	}
	scale := desired.Autoscaling.Get()
	if scale == nil || scale.MinReplicas.Get() == nil || scale.GetMinReplicas() != 0 || scale.InitialReplicas.Get() != nil {
		return false
	}
	prior := gpucompat.LegacyWorkloadSpec(previous)
	priorScale := prior.GetAutoscaling()
	priorScale.MinReplicas = scale.MinReplicas
	prior.SetAutoscaling(priorScale)
	return reflect.DeepEqual(prior, gpucompat.LegacyWorkloadSpec(desired))
}

// A new Revision normally starts one pod even when its minimum is zero. If
// the previous revision holds the project's last GPU, that pod cannot start.
func setInitialScaleOnMinimumUpdate(previous map[string]any, service *servingv1.Service, canStartAtZero bool) (bool, *workloadsApi.SubmissionErrorV2) {
	ann := service.Spec.Template.Annotations
	if ann[autoscaling.MinScaleAnnotationKey] != "0" || ann[autoscaling.ClassAnnotationKey] != autoscaling.KPA || !routesToLatestRevision(service.Spec.Traffic) {
		return false, nil
	}
	if _, explicit := ann[autoscaling.InitialScaleAnnotationKey]; explicit || previous["apiVersion"] != service.APIVersion || previous["kind"] != service.Kind {
		return false, nil
	}

	// RawManifest is the submitted Service, before workloads-manager adds its
	// labels and defaults. Unknown fields must also disqualify a template change.
	raw, err := json.Marshal(previous["spec"])
	if err != nil {
		return false, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, fmt.Sprintf("read previous inference Service spec: %v", err))
	}
	var prior servingv1.ServiceSpec
	if err := json.Unmarshal(raw, &prior); err != nil {
		return false, nil
	}
	old := prior.Template.Annotations
	minimum, err := strconv.ParseUint(old[autoscaling.MinScaleAnnotationKey], 10, 32)
	if err != nil || old[autoscaling.ClassAnnotationKey] != autoscaling.KPA || !routesToLatestRevision(prior.Traffic) {
		return false, nil
	}
	// The native snapshot is written after workloads-manager accepts an update.
	// A minimum-only PATCH composed from a stale snapshot must not roll the image back.
	currentContainers, desiredContainers := prior.Template.Spec.Containers, service.Spec.Template.Spec.Containers
	if minimum > 0 && len(currentContainers) > 0 && len(desiredContainers) > 0 && currentContainers[0].Image != desiredContainers[0].Image {
		return false, workloadsApi.NewSubmissionErrorV2(http.StatusConflict, "Cannot lower minReplicas to 0 because the inference image differs from the latest accepted update. Retry after the image update synchronizes. If the conflict persists, submit the intended image separately.")
	}
	initial, explicit := old[autoscaling.InitialScaleAnnotationKey]
	retain := minimum == 0 && initial == "0"
	if !retain && (minimum == 0 || explicit || !canStartAtZero) {
		return false, nil
	}
	old[autoscaling.MinScaleAnnotationKey] = "0"
	delete(old, autoscaling.InitialScaleAnnotationKey)

	// Match the public GET's redaction on a copy. The complete typed inputs have
	// already established that the caller did not change these values.
	candidate := service.Spec.DeepCopy()
	if containers := candidate.Template.Spec.Containers; len(containers) > 0 {
		for i, variable := range containers[0].Env {
			if isSensitivePlaintextEnvironmentVariable(variable.Name) {
				containers[0].Env[i].Value = ""
			}
		}
	}
	if !equality.Semantic.DeepEqual(&prior, candidate) {
		return false, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&servingv1.ServiceSpec{}); err != nil {
		return false, nil
	}
	ann[autoscaling.InitialScaleAnnotationKey] = "0"
	return !retain, nil
}

func routesToLatestRevision(targets []servingv1.TrafficTarget) bool {
	if len(targets) == 0 {
		return true // Knative defaults to 100% traffic to the latest Ready Revision.
	}
	target := targets[0]
	return len(targets) == 1 && target.LatestRevision != nil && *target.LatestRevision && target.Percent != nil && *target.Percent == 100 && target.RevisionName == "" && target.Tag == "" && target.ConfigurationName == ""
}

// The reported version must acknowledge the same desired workload we are updating.
// An unversioned Running/Ready report may still describe the previous image.
func requireServingInferenceUpdate(ctx context.Context, tenantID uint, workloadID string, count *int64, isServing func(context.Context, uint, string, int64) (bool, error)) *workloadsApi.SubmissionErrorV2 {
	if count != nil {
		serving, err := isServing(ctx, tenantID, workloadID, *count)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Check serving revision of inference %s: %v", workloadID, err)
			return workloadsApi.NewSubmissionErrorV2(http.StatusServiceUnavailable, "Could not read inference rollout status. Retry the update.")
		}
		if serving {
			return nil
		}
	}
	return workloadsApi.NewSubmissionErrorV2(http.StatusConflict, "Cannot lower minReplicas to 0 until the current inference revision is serving. Wait for the image rollout to finish or roll it back, then retry.")
}
