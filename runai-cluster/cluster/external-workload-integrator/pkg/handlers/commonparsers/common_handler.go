/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package commonparsers

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/getters"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/resourceutils/resources"
	workloadTypeConfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/workload-type-config"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	monorepocommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/node-pool-utils/converter"
	topowner "github.com/dsx-ai-factory/run-ai/runai-cluster/common/top-owner"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
)

const (
	ReadyStatus     = "Ready"
	NotReadyStatus  = "NotReady"
	SuspendedStatus = "Suspended"
	ActiveStatus    = "Active"
	FailedStatus    = "Failed"
	SucceededStatus = "Succeeded"
)

func FillRequestedNodePools(ctx context.Context, objectMeta metav1.ObjectMeta, kind string,
	affinity *corev1.Affinity, cachedReader client.Reader, externalWorkload *v1alpha1.ExternalWorkload) {
	if len(externalWorkload.Spec.RequestedNodePools) > 0 {
		// requested nodepools cannot be updated if object already exists
		return
	}

	var err error
	var project *kaiorg.Project
	if externalWorkload.Spec.ProjectName != "" {
		project, err = utils.GetProject(ctx, externalWorkload.Spec.ProjectName, cachedReader)
	} else {
		project, err = utils.GetProjectFromNamespace(ctx, objectMeta.Namespace, cachedReader)
	}
	if err != nil {
		// even if error occurred, we still want to continue
		project = nil
		log.Warn().Msgf("For %s <%s/%s>, failed getting project, err: <%s>",
			kind, objectMeta.Namespace, objectMeta.Name, err.Error())
	}

	// only calculate if not already exists in spec.
	externalWorkload.Spec.RequestedNodePools = getRequestedNodePools(ctx, objectMeta, kind, affinity,
		project, cachedReader)
}

func getRequestedNodePools(ctx context.Context, objectMeta metav1.ObjectMeta, kind string,
	affinity *corev1.Affinity, project *kaiorg.Project, cachedReader client.Reader) []string {

	nodePoolSources := converter.NewNodePoolsSources(affinity, cachedReader, converter.IgnoreAffinityConversionErrors, nil, project)
	identifiers := converter.NodePoolIdentifiers{
		NodePoolAssignmentLabelKey: v1alpha1.NodePoolNameLabel,
		DefaultNodepoolName:        v1alpha1.DefaultNodePoolName,
		UnexistingNodepoolSentinel: v1alpha1.UnexistingRunaiNodePool,
		AnnotationNodepoolsKey:     monorepocommonconstants.AnnotationNodepools,
	}
	requestedNodePools, err := converter.GetRequestedNodePools(ctx, objectMeta, kind, identifiers, nodePoolSources)
	if err != nil {
		// even if error occurred, we still want to continue
		log.Warn().Msgf("For %s <%s/%s>, had an error when getting node pool names, err: <%s>",
			kind, objectMeta.Namespace, objectMeta.Name, err.Error())
	}

	return requestedNodePools
}

func FillDefaultNodePoolIfNeeded(externalWorkload *v1alpha1.ExternalWorkload) {
	if len(externalWorkload.Spec.RequestedNodePools) == 0 {
		externalWorkload.Spec.RequestedNodePools = []string{v1alpha1.DefaultNodePoolName}
	}
}

// MultiplyRequestedResourcesPerNumberOfRequestedPods multiplies the requested resources.
// It expects externalWorkload.Spec.RequestedResources to already be calculated for 1 pod.
func MultiplyRequestedResourcesPerNumberOfRequestedPods(numberOfRequestedPods int32,
	externalWorkload *v1alpha1.ExternalWorkload) {
	externalWorkload.Spec.RequestedResources = MultiplyResources(numberOfRequestedPods,
		externalWorkload.Spec.RequestedResources)
}

func MultiplyResources(multiplyFactor int32,
	requestedResources corev1.ResourceRequirements) corev1.ResourceRequirements {
	result := *((&requestedResources).DeepCopy())

	if multiplyFactor == 1 {
		return result
	}

	if multiplyFactor == 0 {
		return corev1.ResourceRequirements{}
	}

	resourcesToAdd := *((&requestedResources).DeepCopy())
	for i := 0; i < int(multiplyFactor)-1; i++ {
		result = resources.SumResourceRequirements(result, resourcesToAdd)
	}
	return result
}

// CalculateStatusByPods - if the workload is still in "progress" (Running but not ready, for example),
// we need to check the pods
func CalculateStatusByPods(ctx context.Context, cachedReader client.Reader,
	workload client.Object, observedStatus *v1alpha1.ExternalWorkloadObservedStatus) {
	topOwnerWorkloadId := topowner.GetRunaiWorkloadId(workload.GetUID(), workload.GetAnnotations())

	podsOfWorkload, err := getters.GetPodsOfWorkload(ctx,
		topOwnerWorkloadId, workload.GetNamespace(), cachedReader)
	if err != nil {
		log.Warn().Msgf("Failed getting pods of %s <%s/%s>, error: %s",
			workload.GetObjectKind().GroupVersionKind().String(), workload.GetNamespace(), workload.GetName(), err.Error())
		return
	}

	if len(podsOfWorkload) == 0 {
		return
	}

	// even if only 1 pod exists at the moment (the rest weren't created yet),
	// it shouldn't be scheduled and Initializing before the other pods
	// (pods of same workload are usually gang-scheduled).

	areAllPodsRunningOrInitializing := true
	for _, pod := range podsOfWorkload {
		if pod.Status.Phase != corev1.PodRunning {
			calculatedPhase := getPodCalculatedPhase(pod.Annotations)
			if calculatedPhase != "ContainerCreating" {
				areAllPodsRunningOrInitializing = false
				break
			}
		}
	}

	if areAllPodsRunningOrInitializing {
		// if got here, means the workload itself didn't reach the "Running" phase by its
		// status conditions (not "Ready" yet) - but all the pods are running/initializing;
		// so we mark the workload as "Initializing"
		observedStatus.Phase = v1alpha1.ExWInitializingPhase
	}
}

func SetObservedMetav1Conditions(statusConditions []metav1.Condition, observedStatus *v1alpha1.ExternalWorkloadObservedStatus) {
	observedConditions := []v1alpha1.ObservedCondition{}
	for _, condition := range statusConditions {
		observedConditions = append(observedConditions, v1alpha1.ObservedCondition{
			Type:    condition.Type,
			Status:  corev1.ConditionStatus(condition.Status),
			Reason:  condition.Reason,
			Message: condition.Message,
		})
	}

	observedStatus.Conditions = observedConditions
}

func CalculatePriorityClassNameAndPreemptibility(
	ctx context.Context, cachedReader client.Reader,
	topOwnerLabels map[string]string, topOwnerGroupKind schema.GroupKind, podSpecPriorityClassName string) (string, string) {
	priorityClassName, workloadTypeConfigForKind := calculatePriorityClassName(ctx, cachedReader, topOwnerLabels, topOwnerGroupKind, podSpecPriorityClassName)
	preemptibility := calculatePreemptibility(ctx, workloadTypeConfigForKind, cachedReader, topOwnerLabels, topOwnerGroupKind)

	return priorityClassName, preemptibility
}

// calculatePriorityClassName
// returns the calculated priority class name and the workload type config used (if any)
func calculatePriorityClassName(
	ctx context.Context, cachedReader client.Reader,
	topOwnerLabels map[string]string, topOwnerGroupKind schema.GroupKind, podSpecPriorityClassName string) (string, *workloadTypeConfig.WorkloadTypeConfig) {

	// this code is practically identical to the function "CalcPodGroupPriorityClass" in KAI-scheduler's podgrouper.
	// we don't want to use it directly because it would create an un-wanted dependency to the KAI-scheduler.

	if priorityClassName, found := topOwnerLabels[commonconstants.LabelPriorityClassName]; found {
		return priorityClassName, nil
	} else if len(podSpecPriorityClassName) != 0 {
		return podSpecPriorityClassName, nil
	}

	workloadTypeConfigForKind := getWorkloadTypeConfigForKind(ctx, cachedReader, topOwnerGroupKind)
	if workloadTypeConfigForKind == nil {
		return "", nil
	}

	return workloadTypeConfigForKind.PriorityName, workloadTypeConfigForKind
}

func calculatePreemptibility(
	ctx context.Context, workloadTypeConfigForKind *workloadTypeConfig.WorkloadTypeConfig,
	cachedReader client.Reader, topOwnerLabels map[string]string, topOwnerGroupKind schema.GroupKind) string {

	// this code is practically identical to the function "calcPodGroupPreemptibilityWithDefaults" in KAI-scheduler's podgrouper.
	// we don't want to use it directly because it would create an un-wanted dependency to the KAI-scheduler.

	if preemptibility, found := topOwnerLabels[commonconstants.LabelPreemptibilityName]; found {
		return preemptibility
	}

	if workloadTypeConfigForKind == nil {
		workloadTypeConfigForKind = getWorkloadTypeConfigForKind(ctx, cachedReader, topOwnerGroupKind)
	}
	if workloadTypeConfigForKind == nil {
		return ""
	}

	return workloadTypeConfigForKind.Preemptibility
}

func getWorkloadTypeConfigForKind(ctx context.Context, cachedReader client.Reader, groupKind schema.GroupKind) *workloadTypeConfig.WorkloadTypeConfig {
	if groupKind.Kind == "" {
		log.Warn().Msgf("Unable to get default priority class name: kind is empty")
		return nil
	}

	workloadConfig, err := workloadTypeConfig.GetWorkloadConfigByGroupKind(ctx, cachedReader,
		groupKind.Group, groupKind.Kind)
	if err != nil {
		log.Warn().Msgf("Failed to get workload config for GroupKind <%s>, error: %s",
			groupKind.String(), err.Error())
		return nil
	}

	return workloadConfig
}

func getPodCalculatedPhase(podAnnotations map[string]string) string {
	if podAnnotations == nil {
		return ""
	}
	calculatedPhase, found := podAnnotations[commonconstants.PodStatusAnnotation]
	if !found {
		return ""
	}
	return calculatedPhase
}
