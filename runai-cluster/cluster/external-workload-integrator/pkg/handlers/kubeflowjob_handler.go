/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"

	kubeflowv2beta1 "github.com/kubeflow/mpi-operator/pkg/apis/kubeflow/v2beta1"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/resourceutils/resources"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/commonparsers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubeflowJobKindName string

const (
	KubeflowJobKindNameTFJob      KubeflowJobKindName = commonconstants.KindTFJob
	KubeflowJobKindNamePyTorchJob KubeflowJobKindName = commonconstants.KindPyTorchJob
	KubeflowJobKindNameXGBoostJob KubeflowJobKindName = commonconstants.KindXGBoostJob
	KubeflowJobKindNameJAXJob     KubeflowJobKindName = commonconstants.KindJAXJob
	KubeflowJobKindNameMPIJobV1   KubeflowJobKindName = "MPIJobV1"
	KubeflowJobKindNameMPIJobV2   KubeflowJobKindName = "MPIJobV2"
)

type KubeflowJobTopOwnerHandler struct {
	cachedClient   client.Client
	objectKindName KubeflowJobKindName
}

type KubeflowJobFieldsForParsing struct {
	objectMeta       metav1.ObjectMeta
	groupKind        schema.GroupKind
	replicaSpecs     map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaSpec
	runPolicy        *kubeflowv1.RunPolicy
	runPolicyV2beta1 *kubeflowv2beta1.RunPolicy
	jobStatus        *kubeflowv1.JobStatus
}

func NewKubeflowJobTopOwnerHandler(cachedClient client.Client, objKind KubeflowJobKindName) *KubeflowJobTopOwnerHandler {
	return &KubeflowJobTopOwnerHandler{
		cachedClient:   cachedClient,
		objectKindName: objKind,
	}
}

func kindToNewEmptyObj(kindName KubeflowJobKindName) client.Object {
	switch kindName {
	case KubeflowJobKindNameTFJob:
		{
			return &kubeflowv1.TFJob{}
		}
	case KubeflowJobKindNamePyTorchJob:
		{
			return &kubeflowv1.PyTorchJob{}
		}
	case KubeflowJobKindNameXGBoostJob:
		{
			return &kubeflowv1.XGBoostJob{}
		}
	case KubeflowJobKindNameJAXJob:
		{
			return &kubeflowv1.JAXJob{}
		}
	case KubeflowJobKindNameMPIJobV1:
		{
			return &kubeflowv1.MPIJob{}
		}
	case KubeflowJobKindNameMPIJobV2:
		{
			return &kubeflowv2beta1.MPIJob{}
		}
	}
	return nil
}

func (h *KubeflowJobTopOwnerHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	obj := kindToNewEmptyObj(h.objectKindName)
	return utils.GetObject(ctx, h.cachedClient, name, namespace, func() client.Object { return obj })
}

func (h *KubeflowJobTopOwnerHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	kubeflowJobFieldsForParsing := &KubeflowJobFieldsForParsing{}

	switch object.(type) {
	case *kubeflowv1.TFJob:
		kubeflowJobFieldsForParsing.objectMeta = object.(*kubeflowv1.TFJob).ObjectMeta
		kubeflowJobFieldsForParsing.replicaSpecs = object.(*kubeflowv1.TFJob).Spec.TFReplicaSpecs
		kubeflowJobFieldsForParsing.runPolicy = &object.(*kubeflowv1.TFJob).Spec.RunPolicy
		kubeflowJobFieldsForParsing.jobStatus = &object.(*kubeflowv1.TFJob).Status
	case *kubeflowv1.PyTorchJob:
		kubeflowJobFieldsForParsing.objectMeta = object.(*kubeflowv1.PyTorchJob).ObjectMeta
		kubeflowJobFieldsForParsing.replicaSpecs = object.(*kubeflowv1.PyTorchJob).Spec.PyTorchReplicaSpecs
		kubeflowJobFieldsForParsing.runPolicy = &object.(*kubeflowv1.PyTorchJob).Spec.RunPolicy
		kubeflowJobFieldsForParsing.jobStatus = &object.(*kubeflowv1.PyTorchJob).Status
	case *kubeflowv1.XGBoostJob:
		kubeflowJobFieldsForParsing.objectMeta = object.(*kubeflowv1.XGBoostJob).ObjectMeta
		kubeflowJobFieldsForParsing.replicaSpecs = object.(*kubeflowv1.XGBoostJob).Spec.XGBReplicaSpecs
		kubeflowJobFieldsForParsing.runPolicy = &object.(*kubeflowv1.XGBoostJob).Spec.RunPolicy
		kubeflowJobFieldsForParsing.jobStatus = &object.(*kubeflowv1.XGBoostJob).Status
	case *kubeflowv1.JAXJob:
		kubeflowJobFieldsForParsing.objectMeta = object.(*kubeflowv1.JAXJob).ObjectMeta
		kubeflowJobFieldsForParsing.replicaSpecs = object.(*kubeflowv1.JAXJob).Spec.JAXReplicaSpecs
		kubeflowJobFieldsForParsing.runPolicy = &object.(*kubeflowv1.JAXJob).Spec.RunPolicy
		kubeflowJobFieldsForParsing.jobStatus = &object.(*kubeflowv1.JAXJob).Status
	case *kubeflowv1.MPIJob:
		kubeflowJobFieldsForParsing.objectMeta = object.(*kubeflowv1.MPIJob).ObjectMeta
		kubeflowJobFieldsForParsing.replicaSpecs = object.(*kubeflowv1.MPIJob).Spec.MPIReplicaSpecs
		kubeflowJobFieldsForParsing.runPolicy = &object.(*kubeflowv1.MPIJob).Spec.RunPolicy
		kubeflowJobFieldsForParsing.jobStatus = &object.(*kubeflowv1.MPIJob).Status
	case *kubeflowv2beta1.MPIJob:
		kubeflowJobFieldsForParsing = getKubeflowJobFieldsForParsingForMpiJobV2beta1(object.(*kubeflowv2beta1.MPIJob))
	default:
		return fmt.Errorf("failed converting object to kubeflowv1 jobs")
	}

	kubeflowJobFieldsForParsing.groupKind = object.GetObjectKind().GroupVersionKind().GroupKind()
	h.parseKubeflowJob(ctx, object, kubeflowJobFieldsForParsing, externalWorkload)

	return nil
}

func (h *KubeflowJobTopOwnerHandler) parseKubeflowJob(ctx context.Context,
	object client.Object, kubeflowJobFieldsForParsing *KubeflowJobFieldsForParsing, externalWorkload *v1alpha1.ExternalWorkload) {
	h.parseKubeflowJobSpec(ctx, kubeflowJobFieldsForParsing, externalWorkload)
	commonparsers.FillDefaultNodePoolIfNeeded(externalWorkload)

	setKubeflowJobObservedStatus(kubeflowJobFieldsForParsing.jobStatus, &externalWorkload.Status.ObservedStatus)
	h.applyRunaiDistributedStatus(ctx, object, kubeflowJobFieldsForParsing, externalWorkload)
}

func (h *KubeflowJobTopOwnerHandler) parseKubeflowJobSpec(ctx context.Context, kubeflowJobFieldsForParsing *KubeflowJobFieldsForParsing,
	externalWorkload *v1alpha1.ExternalWorkload) {
	var replicas int32 = 0
	externalWorkload.Spec.RequestedResources = corev1.ResourceRequirements{}

	for replicaType, replicaSpec := range kubeflowJobFieldsForParsing.replicaSpecs {
		if replicaType != "Master" && externalWorkload.Spec.ImageName == "" {
			commonparsers.CalculateImageInfo(&replicaSpec.Template, externalWorkload)
		}
		currentReplicas := utils.DerefInt32(replicaSpec.Replicas, 1)
		replicas += currentReplicas

		podRequestedResources := commonparsers.GetPodRequestedResources(&replicaSpec.Template)
		workerRequestedResources := commonparsers.MultiplyResources(currentReplicas, podRequestedResources)

		externalWorkload.Spec.RequestedResources = resources.SumResourceRequirements(
			externalWorkload.Spec.RequestedResources, workerRequestedResources)

		commonparsers.FillRequestedNodePools(ctx, kubeflowJobFieldsForParsing.objectMeta, kubeflowJobFieldsForParsing.groupKind.Kind, replicaSpec.Template.Spec.Affinity,
			h.cachedClient, externalWorkload)
	}

	externalWorkload.Spec.RequestedPods.Replicas = utils.Int32PtrOrNil(replicas)

	podSpecPriorityClassName := ""
	if kubeflowJobFieldsForParsing.runPolicy != nil && kubeflowJobFieldsForParsing.runPolicy.SchedulingPolicy != nil &&
		kubeflowJobFieldsForParsing.runPolicy.SchedulingPolicy.PriorityClass != "" {
		podSpecPriorityClassName = kubeflowJobFieldsForParsing.runPolicy.SchedulingPolicy.PriorityClass
	} else if kubeflowJobFieldsForParsing.runPolicyV2beta1 != nil && kubeflowJobFieldsForParsing.runPolicyV2beta1.SchedulingPolicy != nil &&
		kubeflowJobFieldsForParsing.runPolicyV2beta1.SchedulingPolicy.PriorityClass != "" {
		podSpecPriorityClassName = kubeflowJobFieldsForParsing.runPolicyV2beta1.SchedulingPolicy.PriorityClass
	}
	externalWorkload.Spec.PriorityClassName, externalWorkload.Spec.Preemptibility = commonparsers.CalculatePriorityClassNameAndPreemptibility(ctx, h.cachedClient, kubeflowJobFieldsForParsing.objectMeta.Labels,
		kubeflowJobFieldsForParsing.groupKind, podSpecPriorityClassName)
}

func (h *KubeflowJobTopOwnerHandler) Filter(object client.Object) bool {
	switch object.(type) {
	case *kubeflowv1.TFJob:
		return filterKubeflowReplicasSpecs(object.(*kubeflowv1.TFJob).Spec.TFReplicaSpecs)
	case *kubeflowv1.PyTorchJob:
		return filterKubeflowReplicasSpecs(object.(*kubeflowv1.PyTorchJob).Spec.PyTorchReplicaSpecs)
	case *kubeflowv1.XGBoostJob:
		return filterKubeflowReplicasSpecs(object.(*kubeflowv1.XGBoostJob).Spec.XGBReplicaSpecs)
	case *kubeflowv1.JAXJob:
		return filterKubeflowReplicasSpecs(object.(*kubeflowv1.JAXJob).Spec.JAXReplicaSpecs)
	case *kubeflowv1.MPIJob:
		return filterKubeflowReplicasSpecs(object.(*kubeflowv1.MPIJob).Spec.MPIReplicaSpecs)
	case *kubeflowv2beta1.MPIJob:
		return filterKubeflowReplicasSpecs(convertMPIReplicaSpecs(object.(*kubeflowv2beta1.MPIJob).Spec.MPIReplicaSpecs))
	default:
		return false
	}
}

func setKubeflowJobObservedStatus(jobStatus *kubeflowv1.JobStatus,
	observedStatus *v1alpha1.ExternalWorkloadObservedStatus) {
	observedConditions := []v1alpha1.ObservedCondition{}
	foundNonCreatingStatus := false
	for _, condition := range jobStatus.Conditions {
		observedConditions = append(observedConditions, v1alpha1.ObservedCondition{
			Type:               string(condition.Type),
			Status:             condition.Status,
			Reason:             condition.Reason,
			Message:            condition.Message,
			LastTransitionTime: condition.LastTransitionTime,
		})

		if !foundNonCreatingStatus && condition.Type == kubeflowv1.JobCreated && condition.Status == corev1.ConditionTrue {
			observedStatus.Phase = v1alpha1.ExWPendingPhase
			observedStatus.K8sStatus = string(kubeflowv1.JobCreated)
			observedStatus.Message = condition.Message
		}

		if condition.Type == kubeflowv1.JobRunning && condition.Status == corev1.ConditionTrue {
			observedStatus.Phase = v1alpha1.ExWRunningPhase
			observedStatus.K8sStatus = string(kubeflowv1.JobRunning)
			observedStatus.Message = condition.Message
			foundNonCreatingStatus = true
		}

		if condition.Type == kubeflowv1.JobRestarting && condition.Status == corev1.ConditionTrue {
			observedStatus.Phase = v1alpha1.ExWPendingPhase
			observedStatus.K8sStatus = string(kubeflowv1.JobRestarting)
			observedStatus.Message = condition.Message
			foundNonCreatingStatus = true
		}

		if condition.Type == kubeflowv1.JobSucceeded && condition.Status == corev1.ConditionTrue {
			observedStatus.Phase = v1alpha1.ExWCompletedPhase
			observedStatus.K8sStatus = string(kubeflowv1.JobSucceeded)
			observedStatus.Message = condition.Message
			foundNonCreatingStatus = true
		}

		if condition.Type == kubeflowv1.JobConditionType(kubeflowv2beta1.JobSuspended) && condition.Status == corev1.ConditionTrue {
			observedStatus.Phase = v1alpha1.ExWStoppedPhase
			observedStatus.K8sStatus = string(kubeflowv2beta1.JobSuspended)
			observedStatus.Message = condition.Message
			foundNonCreatingStatus = true
		}

		if condition.Type == kubeflowv1.JobFailed && condition.Status == corev1.ConditionTrue {
			observedStatus.Phase = v1alpha1.ExWFailedPhase
			observedStatus.K8sStatus = string(kubeflowv1.JobFailed)
			observedStatus.Message = condition.Message
			foundNonCreatingStatus = true
		}
	}

	observedStatus.Conditions = observedConditions
}

func filterKubeflowReplicasSpecs(replicaSpecs map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaSpec) bool {
	for _, replicaSpec := range replicaSpecs {
		if replicaSpec.Template.Spec.SchedulerName != constants.RunaiSchedulerName {
			return false
		}
	}
	return true
}

func getKubeflowJobFieldsForParsingForMpiJobV2beta1(mpiJob *kubeflowv2beta1.MPIJob) *KubeflowJobFieldsForParsing {
	jobStatusConditions := []kubeflowv1.JobCondition{}
	for _, condition := range mpiJob.Status.Conditions {
		jobStatusConditions = append(jobStatusConditions, kubeflowv1.JobCondition{
			Type:               kubeflowv1.JobConditionType(condition.Type),
			Status:             condition.Status,
			Reason:             condition.Reason,
			Message:            condition.Message,
			LastTransitionTime: condition.LastTransitionTime,
		})
	}

	kubeflowJobFieldsForParsing := &KubeflowJobFieldsForParsing{
		objectMeta:       mpiJob.ObjectMeta,
		replicaSpecs:     convertMPIReplicaSpecs(mpiJob.Spec.MPIReplicaSpecs),
		runPolicyV2beta1: &mpiJob.Spec.RunPolicy,
		jobStatus:        &kubeflowv1.JobStatus{Conditions: jobStatusConditions},
	}
	return kubeflowJobFieldsForParsing
}

func convertMPIReplicaSpecs(mpiReplicaSpecs map[kubeflowv2beta1.MPIReplicaType]*kubeflowv2beta1.ReplicaSpec) map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaSpec {
	replicaSpecs := map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaSpec{}
	for replicaType, replicaSpec := range mpiReplicaSpecs {
		replicaSpecs[kubeflowv1.ReplicaType(replicaType)] = &kubeflowv1.ReplicaSpec{
			Replicas:      replicaSpec.Replicas,
			Template:      replicaSpec.Template,
			RestartPolicy: kubeflowv1.RestartPolicy(replicaSpec.RestartPolicy),
		}
	}
	return replicaSpecs
}
