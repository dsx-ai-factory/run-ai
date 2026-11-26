/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/commonparsers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	topowner "github.com/dsx-ai-factory/run-ai/runai-cluster/common/top-owner"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	workflowTaskType            = "Task"
	workflowTaskAdditionalField = "workflowTask"
)

var argoNodeRetryPattern = regexp.MustCompile(`^(.+)\((\d+)\)$`)

type workflowTask struct {
	Workflow    string `json:"workflow"`
	Group       string `json:"group"`
	Task        string `json:"task"`
	Attempt     int    `json:"attempt"`
	TaskAttempt *int   `json:"taskAttempt,omitempty"`
}

type PodTopOwnerHandler struct {
	cachedClient client.Client
	draParser    *dra.DRAParser
}

func NewPodTopOwnerHandler(cachedClient client.Client, draParser *dra.DRAParser) *PodTopOwnerHandler {
	return &PodTopOwnerHandler{
		cachedClient: cachedClient,
		draParser:    draParser,
	}
}

func (h *PodTopOwnerHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	return utils.GetObject(ctx, h.cachedClient, name, namespace, func() client.Object { return &corev1.Pod{} })
}

func (h *PodTopOwnerHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	pod, ok := object.(*corev1.Pod)
	if !ok {
		return fmt.Errorf("failed converting object to corev1.Pod")
	}

	podTemplate := corev1.PodTemplateSpec{
		ObjectMeta: pod.ObjectMeta,
		Spec:       pod.Spec,
	}
	commonparsers.FillExternalWorkloadDataFromPodTemplateSpec(ctx, pod.ObjectMeta, pod.GroupVersionKind().GroupKind(), &podTemplate,
		nil, h.cachedClient, h.draParser, externalWorkload)

	externalWorkload.Spec.RequestedPods.Replicas = ptr.To(int32(1))

	setPodObservedStatus(&pod.Status, &externalWorkload.Status.ObservedStatus)

	if pod.Labels[constants.WorkflowLabelKey] != "" && len(pod.OwnerReferences) > 0 {
		externalWorkload.Spec.Type = workflowTaskType
		// Argo's node name holds the task and its retry, e.g. <workflow>.<task>(<retry>), while the pod name holds a hash
		if nodeName := pod.Annotations[constants.ArgoNodeNameAnnotationKey]; nodeName != "" {
			externalWorkload.Spec.OriginalWorkloadName = nodeName
			setWorkflowTaskAdditionalFields(pod.Labels[constants.ArgoWorkflowLabelKey], nodeName, externalWorkload)
		}
		// the workload id labels of a workflow pod belong to the workflow; they're set when the pod is handled as owned
		return nil
	}
	return h.addTopOwnerWorkloadIdLabelIfNeeded(ctx, pod, pod.UID, pod.Annotations)
}

// setWorkflowTaskAdditionalFields parses Argo's node name, <workflow>.<group>(<retry>).<task>(<retry>), where a single
// task is its own group and a node without a retry strategy has no (<retry>). Argo's task names can't contain dots or
// parentheses, so the path after the workflow's name splits unambiguously. Attempts count from 1, while Argo's retries
// count from 0.
func setWorkflowTaskAdditionalFields(workflowName, nodeName string, externalWorkload *v1alpha1.ExternalWorkload) {
	path, found := strings.CutPrefix(nodeName, workflowName+".")
	if workflowName == "" || !found || path == "" {
		return
	}
	elements := strings.Split(path, ".")
	group, groupRetry := parseArgoNodeElement(elements[0])
	task, taskRetry := parseArgoNodeElement(elements[len(elements)-1])

	details := workflowTask{
		Workflow: workflowName,
		Group:    group,
		Task:     task,
		Attempt:  ptr.Deref(groupRetry, 0) + 1,
	}
	if len(elements) > 1 && taskRetry != nil {
		details.TaskAttempt = ptr.To(*taskRetry + 1)
	}

	raw, err := json.Marshal(map[string]workflowTask{workflowTaskAdditionalField: details})
	if err != nil {
		log.Error().Msgf("Failed to marshal workflow task details of node <%s>, error: %s", nodeName, err.Error())
		return
	}
	externalWorkload.Spec.AdditionalFields = ptr.To(string(raw))
}

func parseArgoNodeElement(element string) (string, *int) {
	match := argoNodeRetryPattern.FindStringSubmatch(element)
	if match == nil {
		return element, nil
	}
	retry, err := strconv.Atoi(match[2])
	if err != nil {
		return element, nil
	}
	return match[1], &retry
}

func setPodObservedStatus(podStatus *corev1.PodStatus,
	observedStatus *v1alpha1.ExternalWorkloadObservedStatus) {
	if podStatus.Phase == corev1.PodSucceeded {
		observedStatus.Phase = v1alpha1.ExWCompletedPhase
	} else {
		observedStatus.Phase = v1alpha1.ExternalWorkloadPhase(podStatus.Phase)
	}

	observedStatus.K8sStatus = string(podStatus.Phase)
	observedStatus.Message = podStatus.Message

	observedConditions := []v1alpha1.ObservedCondition{}
	for _, condition := range podStatus.Conditions {
		observedConditions = append(observedConditions, v1alpha1.ObservedCondition{
			Type:               string(condition.Type),
			Status:             condition.Status,
			Reason:             condition.Reason,
			Message:            condition.Message,
			LastTransitionTime: condition.LastTransitionTime,
		})
	}

	observedStatus.Conditions = observedConditions
}

func (h *PodTopOwnerHandler) Filter(object client.Object) bool {
	pod, ok := object.(*corev1.Pod)
	if !ok {
		return false
	}
	return pod.Spec.SchedulerName == constants.RunaiSchedulerName
}

func (h *PodTopOwnerHandler) AdditionalHandler(ctx context.Context, object client.Object,
	topOwner *metav1.PartialObjectMetadata, externalWorkloadOfOwner *v1alpha1.ExternalWorkload) error {
	pod, ok := object.(*corev1.Pod)
	if !ok {
		return fmt.Errorf("failed converting object to corev1.Pod")
	}

	var err error = nil
	if topOwner != nil {
		innerErr := h.addTopOwnerWorkloadIdLabelIfNeeded(ctx, pod, topOwner.UID, topOwner.Annotations)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
		}
	}

	return err
}

func (h *PodTopOwnerHandler) addTopOwnerWorkloadIdLabelIfNeeded(ctx context.Context, pod *corev1.Pod,
	topOwnerUID types.UID, topOwnerAnnotations map[string]string) error {
	topOwnerWorkloadIdLabelValue, topOwnerWorkloadIdFound := pod.Labels[commonconstants.TopOwnerWorkloadIdLabel]
	clusterTopOwnerUid, clusterTopOwnerUidFound := pod.Labels[commonconstants.ClusterTopOwnerUidLabel]
	topOwnerUidLabelValue, topOwnerUidFound := pod.Labels[commonconstants.TopOwnerUidLabel]
	if topOwnerWorkloadIdFound && topOwnerWorkloadIdLabelValue != "" &&
		clusterTopOwnerUidFound && clusterTopOwnerUid != "" &&
		topOwnerUidFound && topOwnerUidLabelValue != "" {
		return nil
	}

	topOwnerWorkloadId := topowner.GetRunaiWorkloadId(topOwnerUID, topOwnerAnnotations)

	// NOTE: set deprecated as well, until we fully move to the new label
	patchBytes, err := utils.AddLabelsPatchBytes(map[string]string{
		commonconstants.TopOwnerWorkloadIdLabel: topOwnerWorkloadId,
		commonconstants.ClusterTopOwnerUidLabel: string(topOwnerUID),
		commonconstants.TopOwnerUidLabel:        topOwnerWorkloadId,
	})
	if err != nil {
		log.Error().Msgf("Failed to json.Marshal patch - add top owner label <%s> on pod <%s/%s>, error: %s",
			topOwnerWorkloadId, pod.Namespace, pod.Name, err.Error())
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	err = h.cachedClient.Patch(ctx, pod, patch)
	if err != nil {
		log.Error().Msgf("Failed to patch pod <%s/%s> with top owner label <%s>, error: %s",
			pod.Namespace, pod.Name, topOwnerWorkloadId, err.Error())
		return err
	}
	log.Info().Msgf("Patched pod <%s/%s> with top owner label <%s>",
		pod.Namespace, pod.Name, topOwnerWorkloadId)

	return nil
}
