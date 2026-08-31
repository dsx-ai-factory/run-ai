/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	diwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	infwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	iwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	runaiWorkloadRelevantUsage = "Submit"
)

func workloadsEventFilter() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return filterCreateDeleteEvents(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return filterUpdateEvents(e.ObjectOld, e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return filterCreateDeleteEvents(e.Object)
		},
	}
}

func filterUpdateEvents(objectOld, objectNew client.Object) bool {
	// TODO: for reviewer: we are using the pods of a workload to get the nodepool,
	// but here filtering in only workloads that changed phases.
	// I assume if the nodepool changed, phase changes as well so we will get the new nodepool either way.

	switch objectNew.(type) {
	case *v1alpha1.ExternalWorkload:
		return filterExternalWorkloadUpdate(objectOld.(*v1alpha1.ExternalWorkload), objectNew.(*v1alpha1.ExternalWorkload))
	case *twv2a1.TrainingWorkload:
		return filterTrainingWorkloadUpdate(objectOld.(*twv2a1.TrainingWorkload), objectNew.(*twv2a1.TrainingWorkload))
	case *iwv2a1.InteractiveWorkload:
		return filterInteractiveWorkloadUpdate(objectOld.(*iwv2a1.InteractiveWorkload), objectNew.(*iwv2a1.InteractiveWorkload))
	case *diwv2a1.DistributedWorkload:
		return filterDistributedWorkloadUpdate(objectOld.(*diwv2a1.DistributedWorkload), objectNew.(*diwv2a1.DistributedWorkload))
	case *infwv2a1.InferenceWorkload:
		return filterInferenceWorkloadUpdate(objectOld.(*infwv2a1.InferenceWorkload), objectNew.(*infwv2a1.InferenceWorkload))
	case *dinfw.DistributedInferenceWorkload:
		return filterDistributedInferenceWorkloadUpdate(objectOld.(*dinfw.DistributedInferenceWorkload), objectNew.(*dinfw.DistributedInferenceWorkload))
	}
	return false
}

func filterCreateDeleteEvents(object client.Object) bool {
	switch object.(type) {
	case *v1alpha1.ExternalWorkload:
		return true
	case *twv2a1.TrainingWorkload:
		return filterWorkload(object.(*twv2a1.TrainingWorkload).Spec.WorkloadFields)
	case *iwv2a1.InteractiveWorkload:
		return filterWorkload(object.(*iwv2a1.InteractiveWorkload).Spec.WorkloadFields)
	case *diwv2a1.DistributedWorkload:
		return filterWorkload(object.(*diwv2a1.DistributedWorkload).Spec.WorkloadFields)
	case *infwv2a1.InferenceWorkload:
		return filterWorkload(object.(*infwv2a1.InferenceWorkload).Spec.WorkloadFields)
	case *dinfw.DistributedInferenceWorkload:
		return object.(*dinfw.DistributedInferenceWorkload).Spec.Usage == runaiWorkloadRelevantUsage
	}
	return false
}

func filterWorkload(workloadFields field_groups.WorkloadFields) bool {
	return workloadFields.Usage == runaiWorkloadRelevantUsage
}

func filterExternalWorkloadUpdate(old, new *v1alpha1.ExternalWorkload) bool {
	return old.Status.ObservedStatus.Phase != new.Status.ObservedStatus.Phase
}

func filterWorkloadUpdate(oldStatus, newStatus common.WorkloadCommonStatus) bool {
	return oldStatus.Phase != newStatus.Phase
}

func filterTrainingWorkloadUpdate(old, new *twv2a1.TrainingWorkload) bool {
	return filterWorkload(new.Spec.WorkloadFields) && filterWorkloadUpdate(old.Status.WorkloadCommonStatus, new.Status.WorkloadCommonStatus)
}

func filterInteractiveWorkloadUpdate(old, new *iwv2a1.InteractiveWorkload) bool {
	return filterWorkload(new.Spec.WorkloadFields) && filterWorkloadUpdate(old.Status.WorkloadCommonStatus, new.Status.WorkloadCommonStatus)
}

func filterDistributedWorkloadUpdate(old, new *diwv2a1.DistributedWorkload) bool {
	return filterWorkload(new.Spec.WorkloadFields) && filterWorkloadUpdate(old.Status.WorkloadCommonStatus, new.Status.WorkloadCommonStatus)
}

func filterInferenceWorkloadUpdate(old, new *infwv2a1.InferenceWorkload) bool {
	return filterWorkload(new.Spec.WorkloadFields) && old.Status.Phase != new.Status.Phase
}

func filterDistributedInferenceWorkloadUpdate(old, new *dinfw.DistributedInferenceWorkload) bool {
	return (new.Spec.Usage == runaiWorkloadRelevantUsage) && old.Status.Phase != new.Status.Phase
}
