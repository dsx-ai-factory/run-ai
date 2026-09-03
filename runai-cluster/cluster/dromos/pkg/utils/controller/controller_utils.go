/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// AddDynamicWatchForGVK dynamically watches for objects of the given GVK
// and enqueues requests for the owner of the WorkloadRunner.
func AddDynamicWatchForGVK(
	manager *ctrl.Manager,
	controller *controller.Controller,
	gvk schema.GroupVersionKind,
) error {
	watchObj := &unstructured.Unstructured{}
	watchObj.SetGroupVersionKind(gvk)

	return (*controller).Watch(
		source.Kind(
			(*manager).GetCache(),
			watchObj,
			handler.TypedEnqueueRequestForOwner[*unstructured.Unstructured](
				(*manager).GetScheme(),
				(*manager).GetRESTMapper(),
				&dromosv1alpha1.WorkloadRunner{},
				handler.OnlyControllerOwner(),
			),
		),
	)
}
