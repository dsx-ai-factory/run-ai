/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/commonparsers"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var vmiPhaseToExternalWorkloadPhase = map[kubevirtv1.VirtualMachineInstancePhase]v1alpha1.ExternalWorkloadPhase{
	kubevirtv1.VmPhaseUnset: v1alpha1.ExWUnknownPhase,
	kubevirtv1.Pending:      v1alpha1.ExWCreatingPhase,
	kubevirtv1.Scheduling:   v1alpha1.ExWPendingPhase,
	kubevirtv1.Scheduled:    v1alpha1.ExWInitializingPhase,
	kubevirtv1.Running:      v1alpha1.ExWRunningPhase,
	kubevirtv1.Succeeded:    v1alpha1.ExWCompletedPhase,
	kubevirtv1.Failed:       v1alpha1.ExWFailedPhase,
	kubevirtv1.Unknown:      v1alpha1.ExWUnknownPhase,
}

type KubevirtVMITopOwnerHandler struct {
	cachedClient client.Client
}

func NewKubevirtVMITopOwnerHandler(cachedClient client.Client) *KubevirtVMITopOwnerHandler {
	return &KubevirtVMITopOwnerHandler{
		cachedClient: cachedClient,
	}
}

func (h *KubevirtVMITopOwnerHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	return utils.GetObject(ctx, h.cachedClient, name, namespace, func() client.Object { return &kubevirtv1.VirtualMachineInstance{} })
}

func (h *KubevirtVMITopOwnerHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	vmi, ok := object.(*kubevirtv1.VirtualMachineInstance)
	if !ok {
		return fmt.Errorf("failed converting object to kubevirtv1.VirtualMachineInstance")
	}

	externalWorkload.Spec.RequestedResources.Limits = vmi.Spec.Domain.Resources.Limits
	externalWorkload.Spec.RequestedResources.Requests = vmi.Spec.Domain.Resources.Requests
	externalWorkload.Spec.PriorityClassName, externalWorkload.Spec.Preemptibility = commonparsers.CalculatePriorityClassNameAndPreemptibility(ctx, h.cachedClient, vmi.Labels,
		vmi.GroupVersionKind().GroupKind(), vmi.Spec.PriorityClassName)
	commonparsers.FillRequestedNodePools(ctx, vmi.ObjectMeta, vmi.Kind, vmi.Spec.Affinity,
		h.cachedClient, externalWorkload)

	externalWorkload.Status.ObservedStatus.Phase = convertVMIPhasePhase(vmi.Status.Phase)
	externalWorkload.Status.ObservedStatus.K8sStatus = string(vmi.Status.Phase)
	externalWorkload.Status.ObservedStatus.Message = vmi.Status.Reason

	observedConditions := []v1alpha1.ObservedCondition{}
	for _, condition := range vmi.Status.Conditions {
		observedConditions = append(observedConditions, v1alpha1.ObservedCondition{
			Type:               string(condition.Type),
			Status:             condition.Status,
			Reason:             condition.Reason,
			Message:            condition.Message,
			LastTransitionTime: condition.LastTransitionTime,
		})
	}
	externalWorkload.Status.ObservedStatus.Conditions = observedConditions

	return nil
}

func convertVMIPhasePhase(vmiPhase kubevirtv1.VirtualMachineInstancePhase) v1alpha1.ExternalWorkloadPhase {
	phase, found := vmiPhaseToExternalWorkloadPhase[vmiPhase]
	if found {
		return phase
	}
	return v1alpha1.ExWUnknownPhase
}

func (h *KubevirtVMITopOwnerHandler) Filter(object client.Object) bool {
	vmi, ok := object.(*kubevirtv1.VirtualMachineInstance)
	if !ok {
		return false
	}
	return vmi.Spec.SchedulerName == constants.RunaiSchedulerName
}
