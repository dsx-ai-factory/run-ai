/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/common/dra"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/commonparsers"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	topowner "github.com/run-ai/runai/runai-cluster/common/top-owner"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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

	return h.addTopOwnerWorkloadIdLabelIfNeeded(ctx, pod, pod.UID, pod.Annotations)
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
