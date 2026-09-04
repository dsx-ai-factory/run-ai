/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func isRunAiPodPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return isRunAiPod(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return isRunAiPod(e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isRunAiPod(e.Object)
		},
	}
}

func isRunAiPod(object client.Object) bool {
	pod, isPod := object.(*v1.Pod)
	if isPod {
		return pod.Spec.SchedulerName == runaiSchedulerName
	}
	return false
}
