/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"fmt"
	"strconv"
	"strings"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/util"
	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	config2 "knative.dev/serving/pkg/apis/config"
	knv1 "knative.dev/serving/pkg/apis/serving/v1"
)

type RevisionParser struct {
}

func NewRevisionParser() *RevisionParser {
	return &RevisionParser{}
}

func (p *RevisionParser) ParseRevisionToServerScheme(revision *knv1.Revision, workloadID string, project *util.ProjectData) (*openapi.RevisionSyncFields, error) {
	if workloadID == "" {
		return nil, fmt.Errorf("workload ID is empty")
	}

	projectId, err := util.GetProjectId(project)
	if err != nil {
		return nil, err
	}
	departmentId, err := util.GetDepartmentId(project)
	if err != nil {
		return nil, err
	}

	userContainer, err := getUserContainer(revision.Spec.Containers)
	if err != nil {
		return nil, err
	}

	var phase openapi.Phase
	var phaseMessage *string
	if phaseStr, ok := revision.Status.Annotations[commonconstants.RunaiPhaseAnnotationKey]; ok && phaseStr != "" {
		phase = openapi.Phase(phaseStr)
	}
	if !phase.IsValid() {
		phase = openapi.PHASE_UNKNOWN
	}
	if phaseMsgStr, ok := revision.Status.Annotations[commonconstants.RunaiPhaseMessageAnnotationKey]; ok {
		phaseMessage = ptr.To(phaseMsgStr)
	}

	return &openapi.RevisionSyncFields{
		Type:                 "Revision",
		Name:                 revision.Name,
		Id:                   string(revision.UID),
		WorkloadId:           workloadID,
		ClusterId:            config.Config().ClusterUuid,
		DepartmentId:         departmentId,
		ProjectId:            projectId,
		CreatedAt:            revision.CreationTimestamp.Time,
		RequestedResources:   p.parseRevisionRequestedResources(userContainer.Resources),
		RequestedPods:        parseRevisionRequestedPods(revision),
		Images:               []string{userContainer.Image},
		EnvironmentVariables: getContainersEnvVars(userContainer),
		Command:              ptr.To(strings.Join(userContainer.Command, " ")),
		Arguments:            ptr.To(strings.Join(userContainer.Args, " ")),
		Conditions:           parseRevisionStatusToServerScheme(revision.Status.Conditions),
		Phase:                phase,
		PhaseMessage:         phaseMessage,
		ChildrenIds:          nil,
		AdditionalFields:     nil,
	}, nil
}

func parseRevisionStatusToServerScheme(conditionReferences duckv1.Conditions) []openapi.Condition {
	conditionScheme := []openapi.Condition{}
	for _, conditionReference := range conditionReferences {
		conditionScheme = append(conditionScheme, openapi.Condition{
			Type:    string(conditionReference.Type),
			Status:  string(conditionReference.Status),
			Reason:  ptr.To(conditionReference.Reason),
			Message: ptr.To(conditionReference.Message),
		})
	}
	return conditionScheme
}

func getUserContainer(containers []corev1.Container) (corev1.Container, error) {
	if len(containers) == 0 {
		return corev1.Container{}, fmt.Errorf("no containers")
	}

	for _, container := range containers {
		if strings.HasPrefix(container.Name, config2.DefaultUserContainerName) {
			return container, nil
		}
	}

	// If the "user-container" is not found, use the first container instead
	return containers[0], nil
}

func getContainersEnvVars(container corev1.Container) *map[string]string {
	envVars := &map[string]string{}
	for _, envVar := range container.Env {
		(*envVars)[envVar.Name] = envVar.Value
	}

	return envVars
}

func parseRevisionRequestedPods(revision *knv1.Revision) *openapi.RequestedPods {
	result := &openapi.RequestedPods{}

	if min, ok := revision.Annotations["autoscaling.knative.dev/min-scale"]; ok && min != "" {
		if parsed, err := strconv.Atoi(min); err == nil {
			result.Min = *openapi.NewNullableInt32(ptr.To(int32(parsed)))
		} else {
			zap.S().Errorf("Revision %s doesn't have a valid min scale annotation. Error %v", revision.UID, err)
		}
	}

	if max, ok := revision.Annotations["autoscaling.knative.dev/max-scale"]; ok && max != "" {
		if parsed, err := strconv.Atoi(max); err == nil {
			result.Max = *openapi.NewNullableInt32(ptr.To(int32(parsed)))
		} else {
			zap.S().Errorf("Revision %s doesn't have a valid max scale annotation. Error %v", revision.UID, err)
		}
	}

	return result
}
func (p *RevisionParser) parseRevisionRequestedResources(requestedResources corev1.ResourceRequirements) openapi.NullableWorkloadRequestResources {
	podNullableRequestedResources := parseRequestedResources(requestedResources.Requests,
		requestedResources.Limits)
	podRequestedResources := podNullableRequestedResources.Get()
	if podRequestedResources == nil {
		return openapi.NullableWorkloadRequestResources{}
	}

	result := openapi.WorkloadRequestResources{
		Cpu:               podRequestedResources.Cpu,
		CpuMemory:         podRequestedResources.CpuMemory,
		Gpu:               podRequestedResources.Gpu,
		GpuMemory:         podRequestedResources.GpuMemory,
		GpuRequestType:    podRequestedResources.GpuRequestType,
		ExtendedResources: podRequestedResources.ExtendedResources,
	}
	nullableResult := openapi.NullableWorkloadRequestResources{}
	nullableResult.Set(&result)
	return nullableResult
}
