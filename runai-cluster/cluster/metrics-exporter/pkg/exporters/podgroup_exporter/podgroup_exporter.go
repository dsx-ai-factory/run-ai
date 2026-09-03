/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package podgroup_exporter

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/metrics_exporter/pkg/exporters/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/metrics_exporter/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/scheduling/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	topOwnerAnnotation            = "kai.scheduler/top-owner-metadata"
	topOwnerAnnotationBC          = "run.ai/top-owner-metadata"
	calculatedStatusAnnotation    = "runai-calculated-status"
	allocatedGPUsAnnotation       = "runai-current-allocated-gpus"
	requestedGPUsAnnotation       = "runai-current-requested-gpus"
	requestedGPUsMemoryAnnotation = "runai-current-requested-gpus-memory"
	requestedMigDevicesAnnotation = "runai-podgroup-requested-mig-devices"
	usedNodesAnnotation           = "runai-used-nodes"
	runningPods                   = "runai-running-pods"
	userAnnotationOrLabel         = "user"

	priorityInferenceNumber              = 125
	priorityBuildNumber                  = 100
	priorityInteractivePreemptibleNumber = 75
	priorityTrainNumber                  = 50
	defaultPriority                      = 50

	jobTypeTrain                  = "Train"
	jobTypeInteractive            = "Interactive"
	jobTypeInference              = "Inference"
	jobTypeInteractivePreemptible = "Interactive-Preemptible"
)

type TopOwnerMetadata struct {
	Name    string `yaml:"name"`
	UID     string `yaml:"uid"`
	Group   string `yaml:"group"`
	Version string `yaml:"version"`
	Kind    string `yaml:"kind"`
}

type PodGroupExporter struct {
	client   client.Client
	podGroup *kaiv2alpha2.PodGroup
	ctx      context.Context
}

func NewPodGroupExporter(ctx context.Context, client client.Client,
	podgroup *kaiv2alpha2.PodGroup) *PodGroupExporter {
	return &PodGroupExporter{
		client:   client,
		podGroup: podgroup,
		ctx:      ctx,
	}
}

func (pe *PodGroupExporter) GetName() string {
	return pe.podGroup.Name
}

func (pe *PodGroupExporter) GetNamespace() string {
	return pe.podGroup.Namespace
}

func (pe *PodGroupExporter) GetQueueName() string {
	return pe.podGroup.Spec.Queue
}

func (pe *PodGroupExporter) GetProjectName() (string, error) {
	queue, err := pe.getQueueOfPodGroup()
	if err != nil {
		return "", err
	}

	projectName := common.GetProjectNameOfQueueFromLabels(queue.Labels)
	if projectName == "" {
		return "", fmt.Errorf("could not find %s label on queue %s", common.ProjectLabel, queue.Name)
	}
	return projectName, nil
}

func (pe *PodGroupExporter) GetNodePoolName() string {
	return utils.GetNodePoolNameFromLabels(pe.podGroup.Labels)
}

func (pe *PodGroupExporter) GetDepartmentName() (string, error) {
	queue, err := pe.getQueueOfPodGroup()
	if err != nil {
		return "", err
	}

	if queue.Spec.ParentQueue == "" {
		return "", fmt.Errorf("could not find department for podgroup %s/%s: parent queue of queue <%s> was empty",
			pe.podGroup.Namespace, pe.podGroup.Name, queue.Name)
	}

	parentQueue, err := pe.getQueue(queue.Spec.ParentQueue)
	if err != nil {
		return "", err
	}

	departmentName := common.GetDepartmentNameOfQueueFromLabels(parentQueue.Labels)
	if departmentName == "" {
		return "",
			fmt.Errorf("could not find department for podgroup %s/%s: "+
				"parent queue <%s> of queue <%s> didn't have department name label",
				pe.podGroup.Namespace, pe.podGroup.Name, parentQueue.Name, queue.Name)
	}

	return departmentName, nil
}

func (pe *PodGroupExporter) getQueueOfPodGroup() (*kaiv2.Queue, error) {
	return pe.getQueue(pe.podGroup.Spec.Queue)
}

func (pe *PodGroupExporter) getQueue(queueName string) (*kaiv2.Queue, error) {
	var queue kaiv2.Queue
	err := pe.client.Get(pe.ctx, client.ObjectKey{Name: queueName}, &queue)
	if err != nil {
		return nil, fmt.Errorf("could not get queue %s", queueName)
	}

	return &queue, nil
}

func (pe *PodGroupExporter) GetUser() string {
	user := ""
	if userAnnotation, found := pe.podGroup.Annotations[userAnnotationOrLabel]; found && userAnnotation != "" {
		user = userAnnotation
	} else if userLabel, found := pe.podGroup.Labels[userAnnotationOrLabel]; found {
		user = userLabel
	}
	return user
}

func (pe *PodGroupExporter) GetNodes() (string, error) {
	nodes, found := pe.podGroup.Annotations[usedNodesAnnotation]
	if !found {
		return "", fmt.Errorf("could not find %s annotation", usedNodesAnnotation)
	}
	if strings.Contains(nodes, ",") {
		nodes = "-" // multiple nodes give no information on metric, and overload mimir
	}
	return nodes, nil
}

func (pe *PodGroupExporter) GetStatus() (string, error) {
	status, found := pe.podGroup.Annotations[calculatedStatusAnnotation]
	if !found {
		return "", fmt.Errorf("could not find %s annotation", calculatedStatusAnnotation)
	}
	return status, nil
}

func (pe *PodGroupExporter) GetNumRunningPods() (int64, error) {
	valString, found := pe.podGroup.Annotations[runningPods]
	if !found {
		return 0, fmt.Errorf("could not find %s annotation", runningPods)
	}
	val, err := strconv.ParseInt(valString, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse %s annotation value, error: %w", valString, err)
	}

	return val, nil
}

func (pe *PodGroupExporter) GetAllocatedGPUs() (float64, error) {
	return utils.ParseFloatFromAnnotation(pe.podGroup.Annotations, allocatedGPUsAnnotation)
}

func (pe *PodGroupExporter) GetRequestedGPUs() (float64, error) {
	return utils.ParseFloatFromAnnotation(pe.podGroup.Annotations, requestedGPUsAnnotation)
}

func (pe *PodGroupExporter) GetRequestedGPUsMemory() (float64, error) {
	return utils.ParseFloatFromAnnotation(pe.podGroup.Annotations, requestedGPUsMemoryAnnotation)
}

func (pe *PodGroupExporter) GetWorkloadType() string {
	priority := pe.getPodGroupPriority(pe.podGroup)

	switch priority {
	case priorityInferenceNumber:
		return jobTypeInference
	case priorityBuildNumber:
		if len(pe.podGroup.OwnerReferences) > 0 && pe.podGroup.OwnerReferences[0].Kind == "Deployment" {
			// For backward compatibility
			return jobTypeInference
		}
		return jobTypeInteractive
	case priorityInteractivePreemptibleNumber:
		return jobTypeInteractivePreemptible
	case priorityTrainNumber:
		return jobTypeTrain
	default:
		log.Printf("unexpected priority level: %s, pod group: %s/%s",
			pe.podGroup.Spec.PriorityClassName, pe.podGroup.Namespace, pe.podGroup.Name)
		return jobTypeTrain
	}
}

func (pe *PodGroupExporter) GetTopOwnerMetadata() (*TopOwnerMetadata, error) {
	topOwnerMetadataString, found := pe.podGroup.Annotations[topOwnerAnnotation]
	if found {
		return unmarshallMetadata(topOwnerMetadataString)
	}
	topOwnerMetadataString, found = pe.podGroup.Annotations[topOwnerAnnotationBC]
	if found {
		return unmarshallMetadata(topOwnerMetadataString)
	}

	return nil, fmt.Errorf("top owner annotation not found for podgroup %s/%s",
		pe.podGroup.Namespace, pe.podGroup.Name)
}

func (pe *PodGroupExporter) GetRequestedMigDevices() (map[string]int64, error) {
	valString, found := pe.podGroup.Annotations[requestedMigDevicesAnnotation]
	if !found {
		return nil, fmt.Errorf("could not find annotation '%s'", requestedMigDevicesAnnotation)
	}
	var migDevices map[string]int64
	err := json.Unmarshal([]byte(valString), &migDevices)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal value '%s' of annotation %s",
			valString, requestedMigDevicesAnnotation)
	}
	return migDevices, nil
}

func (pe *PodGroupExporter) getPodGroupPriority(podGroup *kaiv2alpha2.PodGroup) int32 {
	podGroupDefaultPriority := pe.getDefaultPriority()
	var priorityClass v1.PriorityClass
	err := pe.client.Get(pe.ctx, client.ObjectKey{Name: podGroup.Spec.PriorityClassName}, &priorityClass)
	if err != nil {
		log.Printf("Couldn't find priorityClass %s for podGroup %s/%s, error: %v. Using default priority %d",
			podGroup.Spec.PriorityClassName, podGroup.Namespace, podGroup.Name, err, podGroupDefaultPriority)
		return podGroupDefaultPriority
	}
	return priorityClass.Value
}

func (pe *PodGroupExporter) getDefaultPriority() int32 {
	var priorityClassList v1.PriorityClassList
	err := pe.client.List(pe.ctx, &priorityClassList)
	if err != nil {
		log.Printf("error listing priorityclasses: %v", err)
		return defaultPriority
	}

	for _, pc := range priorityClassList.Items {
		if pc.GlobalDefault {
			return pc.Value
		}
	}
	return defaultPriority
}

func (pe *PodGroupExporter) GetNumBoundNonReleasingPods() (int, error) {
	podGroupPods := corev1.PodList{}
	err := pe.client.List(pe.ctx, &podGroupPods, client.MatchingFields{utils.PodGroupAnnotationForPodPath: pe.podGroup.Name})
	if err != nil {
		return 0, fmt.Errorf("failed to get pods related to pod group %s/%s. Error: %w",
			pe.podGroup.Namespace, pe.podGroup.Name, err)
	}

	boundPodsCounter := 0
	for _, pod := range podGroupPods.Items {
		if pod.DeletionTimestamp == nil && pod.Spec.NodeName != "" {
			boundPodsCounter += 1
		}
	}
	return boundPodsCounter, err
}

func unmarshallMetadata(metadataString string) (*TopOwnerMetadata, error) {
	var metadata TopOwnerMetadata
	err := yaml.Unmarshal([]byte(metadataString), &metadata)
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}
