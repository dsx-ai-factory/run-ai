/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
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
func setInitialScaleOnMinimumUpdate(previous map[string]any, service *servingv1.Service, canStartAtZero bool) error {
	ann := service.Spec.Template.Annotations
	if ann[autoscaling.MinScaleAnnotationKey] != "0" || ann[autoscaling.ClassAnnotationKey] != autoscaling.KPA || !routesToLatestRevision(service.Spec.Traffic) {
		return nil
	}
	if _, explicit := ann[autoscaling.InitialScaleAnnotationKey]; explicit || previous["apiVersion"] != service.APIVersion || previous["kind"] != service.Kind {
		return nil
	}

	// RawManifest is the submitted Service, before workloads-manager adds its
	// labels and defaults. Unknown fields must also disqualify a template change.
	raw, err := json.Marshal(previous["spec"])
	if err != nil {
		return fmt.Errorf("read previous inference Service spec: %w", err)
	}
	var prior servingv1.ServiceSpec
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&prior); err != nil {
		return nil
	}
	old := prior.Template.Annotations
	minimum, err := strconv.ParseUint(old[autoscaling.MinScaleAnnotationKey], 10, 32)
	if err != nil {
		return nil
	}
	initial, explicit := old[autoscaling.InitialScaleAnnotationKey]
	retain := minimum == 0 && initial == "0"
	if !retain && (minimum == 0 || explicit || !canStartAtZero) {
		return nil
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
		return nil
	}
	ann[autoscaling.InitialScaleAnnotationKey] = "0"
	return nil
}

func routesToLatestRevision(targets []servingv1.TrafficTarget) bool {
	if len(targets) == 0 {
		return true // Knative defaults to 100% traffic to the latest Ready Revision.
	}
	target := targets[0]
	return len(targets) == 1 && target.LatestRevision != nil && *target.LatestRevision && target.Percent != nil && *target.Percent == 100 && target.RevisionName == "" && target.Tag == "" && target.ConfigurationName == ""
}
