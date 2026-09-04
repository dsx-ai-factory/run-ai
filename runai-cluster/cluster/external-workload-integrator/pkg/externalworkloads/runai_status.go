/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package externalworkloads

import (
	"context"
	"fmt"
	"math"

	"github.com/rs/zerolog/log"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/runaistatus"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	latestCreatedRevisionField = "latestCreatedRevisionName"
	latestReadyRevisionField   = "latestReadyRevisionName"
	statusReplicasField        = "statusReplicas"
	readyReplicasField         = "readyReplicas"
	networkingPluginName       = "networking-plugin"
	mnnvlPluginName            = "mnnvl"
)

func (h *ExternalWorkloadHandler) applyRunaiStatus(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload,
) {
	if externalWorkload.Spec.Type != inferenceWorkloadType &&
		externalWorkload.Spec.Type != distributedInferenceWorkloadType {
		return
	}

	delete(externalWorkload.Annotations, commonconstants.EwiStatusAnnotation)
	switch externalWorkload.Spec.Type {
	case inferenceWorkloadType:
		runaistatus.SetAdditionalFields(externalWorkload,
			[]string{latestCreatedRevisionField, latestReadyRevisionField}, nil)
	case distributedInferenceWorkloadType:
		runaistatus.SetAdditionalFields(externalWorkload,
			[]string{statusReplicasField, readyReplicasField}, nil)
	}
	wlrName := externalWorkload.Labels[commonconstants.WorkloadRunnerNameLabel]
	if wlrName == "" {
		return
	}

	wlr := &dromosv1alpha1.WorkloadRunner{}
	if err := h.cachedClient.Get(ctx, client.ObjectKey{Name: wlrName, Namespace: object.GetNamespace()}, wlr); err != nil {
		log.Warn().Msgf("runaistatus: failed getting WorkloadRunner <%s/%s>: %s",
			object.GetNamespace(), wlrName, err.Error())
		return
	}

	var err error
	switch externalWorkload.Spec.Type {
	case inferenceWorkloadType:
		err = h.applyInferenceStatus(ctx, object, wlr, externalWorkload)
	case distributedInferenceWorkloadType:
		err = h.applyDistributedInferenceStatus(ctx, object, wlr, externalWorkload)
	}
	if err != nil {
		log.Warn().Msgf("runaistatus: failed calculating %s status for <%s/%s>: %s",
			externalWorkload.Spec.Type, object.GetNamespace(), object.GetName(), err.Error())
		return
	}

	if externalWorkload.Annotations == nil {
		externalWorkload.Annotations = map[string]string{}
	}
	externalWorkload.Annotations[commonconstants.EwiStatusAnnotation] = "true"
}

func (h *ExternalWorkloadHandler) applyInferenceStatus(ctx context.Context, object client.Object,
	wlr *dromosv1alpha1.WorkloadRunner, externalWorkload *v1alpha1.ExternalWorkload,
) error {
	service, ok := object.(*servingv1.Service)
	if !ok {
		return fmt.Errorf("expected Knative Service, got %T", object)
	}

	input := runaistatus.InferenceInput{
		Service: &runaistatus.InferenceServiceState{
			LatestCreatedRevision: service.Status.LatestCreatedRevisionName,
			LatestReadyRevision:   service.Status.LatestReadyRevisionName,
		},
		BuildError: relevantCriticalIssueMessage(wlr),
	}
	for _, condition := range service.Status.Conditions {
		input.Service.Conditions = append(input.Service.Conditions, runaistatus.InferenceServiceCondition{
			Status: condition.Status, Message: condition.Message,
		})
	}
	if input.Service.LatestCreatedRevision != "" {
		pods := &corev1.PodList{}
		if err := h.cachedClient.List(ctx, pods,
			client.InNamespace(service.Namespace),
			client.MatchingLabels{knativeRevisionPodLabel: input.Service.LatestCreatedRevision}); err != nil {
			return fmt.Errorf("list Knative revision pods: %w", err)
		}
		input.Pods = pods.Items
	}

	phase, message := runaistatus.ComputeInferenceStatus(input)
	externalWorkload.Status.ObservedStatus.Phase = phase
	externalWorkload.Status.ObservedStatus.Message = message
	runaistatus.SetAdditionalFields(externalWorkload,
		[]string{latestCreatedRevisionField, latestReadyRevisionField},
		map[string]any{
			latestCreatedRevisionField: input.Service.LatestCreatedRevision,
			latestReadyRevisionField:   input.Service.LatestReadyRevision,
		})
	return nil
}

func (h *ExternalWorkloadHandler) applyDistributedInferenceStatus(ctx context.Context, object client.Object,
	wlr *dromosv1alpha1.WorkloadRunner, externalWorkload *v1alpha1.ExternalWorkload,
) error {
	lws, ok := object.(*unstructured.Unstructured)
	lwsGVK := workloadTypeGVKs[distributedInferenceWorkloadType]
	if !ok || lws.GroupVersionKind().GroupKind() != lwsGVK.GroupKind() {
		return fmt.Errorf("expected LeaderWorkerSet, got %T %s", object, object.GetObjectKind().GroupVersionKind())
	}

	input := runaistatus.DistributedInferenceInput{
		Exists: true, Replicas: 1, GroupSize: 1, BuildError: relevantCriticalIssueMessage(wlr),
	}
	if replicas, found, err := unstructured.NestedInt64(lws.Object, "spec", "replicas"); err != nil {
		return fmt.Errorf("read LWS replicas: %w", err)
	} else if found {
		if replicas < math.MinInt32 || replicas > math.MaxInt32 {
			return fmt.Errorf("LWS replicas %d are outside the int32 range", replicas)
		}
		input.Replicas = int32(replicas)
	}
	if size, found, err := unstructured.NestedInt64(lws.Object, "spec", "leaderWorkerTemplate", "size"); err != nil {
		return fmt.Errorf("read LWS group size: %w", err)
	} else if found {
		input.GroupSize = int(size)
	}

	pods := &corev1.PodList{}
	if err := h.cachedClient.List(ctx, pods,
		client.InNamespace(lws.GetNamespace()),
		client.MatchingLabels{leaderWorkerSetPodLabel: lws.GetName()}); err != nil {
		return fmt.Errorf("list LWS pods: %w", err)
	}
	input.Pods = pods.Items
	if runaistatus.MnnvlRequired(wlr) {
		input.MNNVLEnabled = true
		plugin := wlr.Status.Plugins[mnnvlPluginName]
		input.MNNVLReady = plugin != nil && plugin.Phase == dromosv1alpha1.PluginPhaseReady
	}

	networkingPhase := dromosv1alpha1.PluginPhase("")
	if plugin, ok := wlr.Status.Plugins[networkingPluginName]; ok && plugin != nil {
		networkingPhase = plugin.Phase
	}
	result := runaistatus.ComputeDistributedInferenceStatus(input, networkingPhase)
	externalWorkload.Status.ObservedStatus.Phase = result.Phase
	externalWorkload.Status.ObservedStatus.Message = result.Message
	externalWorkload.Status.ObservedStatus.K8sStatus = distributedInferenceK8sPhase(result.Phase)
	runaistatus.SetAdditionalFields(externalWorkload,
		[]string{statusReplicasField, readyReplicasField},
		map[string]any{
			statusReplicasField: result.Replicas,
			readyReplicasField:  result.ReadyReplicas,
		})
	return nil
}

func distributedInferenceK8sPhase(phase v1alpha1.ExternalWorkloadPhase) string {
	if phase == v1alpha1.ExWRunningPhase {
		return "Ready"
	}
	return string(phase)
}

func relevantCriticalIssueMessage(wlr *dromosv1alpha1.WorkloadRunner) *string {
	firstCritical := func(issues []dromosv1alpha1.Issue) *string {
		for _, currentIssue := range issues {
			if !currentIssue.IsCritical {
				continue
			}
			if currentIssue.Description != "" {
				return ptr.To(currentIssue.Description)
			}
			if currentIssue.Name != "" {
				return ptr.To(currentIssue.Name)
			}
		}
		return nil
	}
	if message := firstCritical(wlr.Status.Issues); message != nil {
		return message
	}
	for _, pluginName := range []string{networkingPluginName, mnnvlPluginName} {
		if plugin := wlr.Status.Plugins[pluginName]; plugin != nil {
			if message := firstCritical(plugin.Issues); message != nil {
				return message
			}
		}
	}
	return nil
}
