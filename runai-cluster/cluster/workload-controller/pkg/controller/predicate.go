/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	logger "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	podutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/pod-utils"
	dwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	iw "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logger.Log.WithName("predicate")

// in order to reduce "noise" to the controller, we create a predicate which filters out
// changes which are not relevant for us to know.
func ignoreDeletionPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			switch e.ObjectOld.(type) {
			case *iw.InteractiveWorkload:
				return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
			case *twv1.TrainingWorkload:
				return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
			case *dwv2.DistributedWorkload:
				return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
			case *v1.PersistentVolumeClaim:
				// we want to ignore PVC status updates, since they are not relevant for us for now.
				return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
			default:
				if e.ObjectOld.GetResourceVersion() == e.ObjectNew.GetResourceVersion() {
					return false
				}
				return true
			}
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			return !e.DeleteStateUnknown
		},
	}
}

// DistributedWorkloadPodsFilter returns a function that filters pods and returns the distributed workload that the pod belongs to
func DistributedWorkloadPodsFilter() func(ctx context.Context, obj client.Object) []reconcile.Request {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		workloadKind := obj.GetLabels()[constants.LabelWorkloadKind]
		if workloadKind != constants.KindDistributedWorkload {
			return []reconcile.Request{}
		}

		workloadName := obj.GetLabels()[constants.LabelWorkloadName]
		if workloadName == "" {
			log.Error(nil, "pod label does not have a workload name")
			return []reconcile.Request{}
		}

		return []reconcile.Request{
			{NamespacedName: types.NamespacedName{
				Name:      workloadName,
				Namespace: obj.GetNamespace(),
			}},
		}
	}
}

// DistributedWorkloadPodPredicate returns a predicate that filters pod events for distributed workloads
// to only trigger reconciliation on meaningful status changes
func DistributedWorkloadPodPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			_, ok := e.ObjectOld.(*v1.Pod)
			if ok {
				return podutils.PodShouldUpdate(e)
			}
			return false
		},
	}
}
