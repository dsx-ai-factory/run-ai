/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"encoding/json"
	"math"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	constants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	specReplicasAdditionalField  = "specReplicas"
	workersAdditionalField       = "workers"
	restartPolicyAdditionalField = "restartPolicy"
	startupPolicyAdditionalField = "startupPolicy"
)

var distributedInferenceSpecAdditionalFields = []string{
	specReplicasAdditionalField,
	workersAdditionalField,
	restartPolicyAdditionalField,
	startupPolicyAdditionalField,
}

func projectProfileCompatibilityFields(wlr *dromosv1alpha1.WorkloadRunner, fields *openapi.WorkloadSyncFields) {
	template := parseWorkloadRunnerTemplate(wlr)
	if template == nil || fields == nil {
		return
	}
	projectInferenceEngineCompatibilityFields(template.GetLabels(), fields)

	var containerPaths [][]string
	switch {
	case template.GroupVersionKind().Group == constants.KnativeServiceGroup &&
		template.GetKind() == constants.KnativeServiceKind:
		containerPaths = [][]string{{"spec", "template", "spec", "containers"}}
	case template.GroupVersionKind().Group == constants.LwsGroup && template.GetKind() == constants.LwsKind:
		containerPaths = [][]string{
			{"spec", "leaderWorkerTemplate", "leaderTemplate", "spec", "containers"},
			{"spec", "leaderWorkerTemplate", "workerTemplate", "spec", "containers"},
		}
		projectLWSRequestedPods(template, fields)
	default:
		return
	}

	seen := make(map[string]struct{})
	images := make([]string, 0)
	for _, path := range containerPaths {
		containers, found, err := unstructured.NestedSlice(template.Object, path...)
		if err != nil || !found {
			continue
		}
		for _, rawContainer := range containers {
			container, ok := rawContainer.(map[string]any)
			if !ok {
				continue
			}
			image, _ := container["image"].(string)
			if image == "" {
				continue
			}
			if _, exists := seen[image]; exists {
				continue
			}
			seen[image] = struct{}{}
			images = append(images, image)
		}
	}
	if len(images) > 0 {
		fields.Images = images
	}
}

func projectDistributedInferenceCompatibilityFields(
	wlr *dromosv1alpha1.WorkloadRunner,
	fields *openapi.WorkloadSyncFields,
) {
	template := parseWorkloadRunnerTemplate(wlr)
	if template == nil || fields == nil ||
		template.GroupVersionKind().Group != constants.LwsGroup ||
		template.GetKind() != constants.LwsKind {
		return
	}

	additionalFields := map[string]any{}
	if fields.AdditionalFields != nil {
		if err := json.Unmarshal([]byte(*fields.AdditionalFields), &additionalFields); err != nil {
			return
		}
	}
	for _, field := range distributedInferenceSpecAdditionalFields {
		delete(additionalFields, field)
	}

	if replicas, found := nestedTemplateInt64(template.Object, "spec", "replicas"); found &&
		replicas >= constants.MinDistributedInferenceReplicas && replicas <= math.MaxInt32 {
		additionalFields[specReplicasAdditionalField] = replicas
	}
	if size, found := nestedTemplateInt64(template.Object, "spec", "leaderWorkerTemplate", "size"); found {
		workers := size - 1
		if workers >= constants.MinDistributedInferenceWorkers && workers <= math.MaxInt32 {
			additionalFields[workersAdditionalField] = workers
		}
	}
	if restartPolicy, found, err := unstructured.NestedString(
		template.Object, "spec", "leaderWorkerTemplate", "restartPolicy",
	); err == nil && found && validDistributedInferenceRestartPolicy(restartPolicy) {
		additionalFields[restartPolicyAdditionalField] = restartPolicy
	}
	if startupPolicy, found, err := unstructured.NestedString(
		template.Object, "spec", "startupPolicy",
	); err == nil && found && validDistributedInferenceStartupPolicy(startupPolicy) {
		additionalFields[startupPolicyAdditionalField] = startupPolicy
	}

	fields.AdditionalFields = getWorkloadAdditionalFields(additionalFields)
}

func parseWorkloadRunnerTemplate(wlr *dromosv1alpha1.WorkloadRunner) *unstructured.Unstructured {
	if wlr == nil || len(wlr.Spec.Template.Raw) == 0 {
		return nil
	}
	object := make(map[string]any)
	if err := json.Unmarshal(wlr.Spec.Template.Raw, &object); err != nil {
		return nil
	}
	return &unstructured.Unstructured{Object: object}
}

func validDistributedInferenceRestartPolicy(policy string) bool {
	return policy == constants.DistributedInferenceRestartPolicyRecreateGroupOnPodRestart ||
		policy == constants.DistributedInferenceRestartPolicyNone
}

func validDistributedInferenceStartupPolicy(policy string) bool {
	return policy == constants.DistributedInferenceStartupPolicyLeaderCreated ||
		policy == constants.DistributedInferenceStartupPolicyLeaderReady
}

func projectInferenceEngineCompatibilityFields(labels map[string]string, fields *openapi.WorkloadSyncFields) {
	additionalFields := map[string]any{}
	if fields.AdditionalFields != nil {
		if err := json.Unmarshal([]byte(*fields.AdditionalFields), &additionalFields); err != nil {
			return
		}
	}

	if labels[constants.LabelNIMWorkload] == constants.BoolTrueAsString {
		additionalFields[constants.AdditionalNIMWorkload] = constants.BoolTrueAsString
	}
	if labels[constants.LabelVLLMWorkload] == constants.BoolTrueAsString {
		additionalFields[constants.AdditionalVLLMWorkload] = constants.BoolTrueAsString
	}

	fields.AdditionalFields = getWorkloadAdditionalFields(additionalFields)
}

func projectLWSRequestedPods(template *unstructured.Unstructured, fields *openapi.WorkloadSyncFields) {
	replicas := int64(1)
	if value, found := nestedTemplateInt64(template.Object, "spec", "replicas"); found {
		replicas = value
	}
	groupSize := int64(1)
	if value, found := nestedTemplateInt64(template.Object, "spec", "leaderWorkerTemplate", "size"); found {
		groupSize = value
	}
	if replicas < 0 || groupSize <= 0 || replicas > math.MaxInt32/groupSize {
		return
	}
	total := replicas * groupSize
	if fields.RequestedPods == nil {
		fields.RequestedPods = openapi.NewRequestedPods()
	}
	fields.RequestedPods.SetNumber(int32(total)) //nolint:gosec // total is bounded by math.MaxInt32 above
}

func nestedTemplateInt64(object map[string]any, fields ...string) (int64, bool) {
	value, found, err := unstructured.NestedFieldNoCopy(object, fields...)
	if err != nil || !found {
		return 0, false
	}
	switch number := value.(type) {
	case int64:
		return number, true
	case float64:
		converted := int64(number)
		return converted, float64(converted) == number
	default:
		return 0, false
	}
}
